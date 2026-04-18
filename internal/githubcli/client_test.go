package githubcli

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"
)

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
	stdout []byte
	stderr []byte
	err    error
	name   string
	args   []string
}

func (runner *fakeRunner) Run(name string, args ...string) (CommandResult, error) {
	runner.name = name
	runner.args = append([]string(nil), args...)
	return CommandResult{Stdout: runner.stdout, Stderr: runner.stderr}, runner.err
}

func then_commandIs(t *testing.T, runner *fakeRunner, expectedName string, expectedArgs []string) {
	t.Helper()

	if runner.name != expectedName || !reflect.DeepEqual(runner.args, expectedArgs) {
		t.Fatalf("expected command %s %v, actual %s %v", expectedName, expectedArgs, runner.name, runner.args)
	}
}

func then_noError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Fatalf("expected no error, actual %v", actual)
	}
}
