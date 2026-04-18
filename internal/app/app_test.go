package app

import (
	"errors"
	"testing"
)

func TestRun_GivenRunner_WhenRunning_ThenDelegatesToRunner(t *testing.T) {
	runner := &fakeRunner{}
	subject := New(runner)

	actualErr := when_running(subject)

	then_noError(t, actualErr)

	if !runner.runCalled {
		t.Fatal("expected runner to be called")
	}
}

func TestRun_GivenRunnerError_WhenRunning_ThenReturnsTheError(t *testing.T) {
	expected := errors.New("boom")
	runner := &fakeRunner{runErr: expected}
	subject := New(runner)

	actual := when_running(subject)
	if !errors.Is(actual, expected) {
		t.Fatalf("expected error %v, actual %v", expected, actual)
	}
}

type fakeRunner struct {
	runCalled bool
	runErr    error
}

func (runner *fakeRunner) Run() error {
	runner.runCalled = true
	return runner.runErr
}

func when_running(subject *App) error {
	return subject.Run()
}

func then_noError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Fatalf("expected no error, actual %v", actual)
	}
}
