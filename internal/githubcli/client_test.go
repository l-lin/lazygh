package githubcli

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"
)

func TestRunGH_GivenRunnerFailure_WhenExecutingTheCommand_ThenItReturnsTheClassifiedError(t *testing.T) {
	runner := &fakeRunner{
		stderr: []byte("To get started with GitHub CLI, please run: gh auth login"),
		err:    errors.New("exit status 4"),
	}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.runGH("gh api user", "api", "user")

	then_commandIs(t, runner, "gh", []string{"api", "user"})
	if !errors.Is(actualErr, ErrUnauthenticated) {
		t.Fatalf("expected error %v, actual %v", ErrUnauthenticated, actualErr)
	}
}

func TestRunGHWithInput_GivenStandardInput_WhenExecutingTheCommand_ThenItDelegatesToTheInputRunner(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("ok")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.runGHWithInput("gh pr comment", []byte("Ship it"), "pr", "comment", "42")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "comment", "42"})
	then_stdinIs(t, runner, "Ship it")
	if string(actual.Stdout) != "ok" {
		t.Fatalf("expected stdout %q, actual %q", "ok", string(actual.Stdout))
	}
}

func TestGetConnectedUser_GivenValidGhResponse_WhenFetching_ThenReturnsTheConnectedUser(t *testing.T) {
	runner := &fakeRunner{
		stdout: []byte(`{"login":"octocat","name":"Mona Lisa Octocat","bio":"Mascot on call","company":"GitHub","location":"The Internet","public_repos":8,"followers":42,"html_url":"https://github.com/octocat"}`),
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetConnectedUser()

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "user"})

	expected := ConnectedUser{
		Login:       "octocat",
		Name:        "Mona Lisa Octocat",
		Bio:         "Mascot on call",
		Company:     "GitHub",
		Location:    "The Internet",
		PublicRepos: 8,
		Followers:   42,
		URL:         "https://github.com/octocat",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected user %+v, actual %+v", expected, actual)
	}
}

func TestGetConnectedUser_GivenInvalidJSON_WhenFetching_ThenReturnsAnInvalidResponseError(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"login":`)}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetConnectedUser()

	if !errors.Is(actualErr, ErrInvalidConnectedUserResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidConnectedUserResponse, actualErr)
	}
}

func TestGetConnectedUser_GivenAuthenticationPrompt_WhenFetching_ThenReturnsAnUnauthenticatedError(t *testing.T) {
	runner := &fakeRunner{
		stderr: []byte("To get started with GitHub CLI, please run: gh auth login"),
		err:    errors.New("exit status 4"),
	}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetConnectedUser()

	if !errors.Is(actualErr, ErrUnauthenticated) {
		t.Fatalf("expected error %v, actual %v", ErrUnauthenticated, actualErr)
	}
}

func TestGetConnectedUser_GivenGhMissingFromPath_WhenFetching_ThenReturnsAnUnavailableError(t *testing.T) {
	runner := &fakeRunner{err: exec.ErrNotFound}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetConnectedUser()

	if !errors.Is(actualErr, ErrUnavailable) {
		t.Fatalf("expected error %v, actual %v", ErrUnavailable, actualErr)
	}
}

func TestGetConnectedUser_GivenEmptyUserPayload_WhenFetching_ThenReturnsAnEmptyResponseError(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"login":""}`)}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetConnectedUser()

	if !errors.Is(actualErr, ErrEmptyConnectedUser) {
		t.Fatalf("expected error %v, actual %v", ErrEmptyConnectedUser, actualErr)
	}
}

type fakeRunner struct {
	stdout    []byte
	stderr    []byte
	err       error
	name      string
	args      []string
	stdin     []byte
	responses []fakeCommandResponse
	calls     []fakeCommandCall
}

type fakeCommandResponse struct {
	stdout []byte
	stderr []byte
	err    error
}

type fakeCommandCall struct {
	name  string
	args  []string
	stdin []byte
}

func (runner *fakeRunner) Run(name string, args ...string) (CommandResult, error) {
	runner.recordCall(name, args, nil)
	response := runner.nextResponse()
	return CommandResult{Stdout: response.stdout, Stderr: response.stderr}, response.err
}

func (runner *fakeRunner) RunWithInput(name string, input []byte, args ...string) (CommandResult, error) {
	runner.recordCall(name, args, input)
	response := runner.nextResponse()
	return CommandResult{Stdout: response.stdout, Stderr: response.stderr}, response.err
}

func (runner *fakeRunner) recordCall(name string, args []string, input []byte) {
	runner.name = name
	runner.args = append([]string(nil), args...)
	if input == nil {
		runner.stdin = nil
	} else {
		runner.stdin = append([]byte(nil), input...)
	}
	runner.calls = append(runner.calls, fakeCommandCall{
		name:  runner.name,
		args:  append([]string(nil), runner.args...),
		stdin: append([]byte(nil), runner.stdin...),
	})
}

func (runner *fakeRunner) nextResponse() fakeCommandResponse {
	if len(runner.responses) == 0 {
		return fakeCommandResponse{stdout: runner.stdout, stderr: runner.stderr, err: runner.err}
	}

	response := runner.responses[0]
	runner.responses = runner.responses[1:]
	return response
}

func then_commandIs(t *testing.T, runner *fakeRunner, expectedName string, expectedArgs []string) {
	t.Helper()

	if runner.name != expectedName || !reflect.DeepEqual(runner.args, expectedArgs) {
		t.Fatalf("expected command %s %v, actual %s %v", expectedName, expectedArgs, runner.name, runner.args)
	}
}

func then_stdinIs(t *testing.T, runner *fakeRunner, expected string) {
	t.Helper()

	if string(runner.stdin) != expected {
		t.Fatalf("expected stdin %q, actual %q", expected, string(runner.stdin))
	}
}

func then_commandsAre(t *testing.T, runner *fakeRunner, expected []fakeCommandCall) {
	t.Helper()

	if !reflect.DeepEqual(runner.calls, expected) {
		t.Fatalf("expected command calls %+v, actual %+v", expected, runner.calls)
	}
}

func then_noError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Fatalf("expected no error, actual %v", actual)
	}
}
