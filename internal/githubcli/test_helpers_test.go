package githubcli

import (
	"reflect"
	"sync"
	"testing"
)

type fakeRunner struct {
	mu        sync.Mutex
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
	runner.mu.Lock()
	defer runner.mu.Unlock()

	runner.recordCall(name, args, nil)
	response := runner.nextResponse()
	return CommandResult{Stdout: response.stdout, Stderr: response.stderr}, response.err
}

func (runner *fakeRunner) RunWithInput(name string, input []byte, args ...string) (CommandResult, error) {
	runner.mu.Lock()
	defer runner.mu.Unlock()

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
