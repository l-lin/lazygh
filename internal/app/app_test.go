package app

import (
	"bytes"
	"testing"
)

func TestRun_GivenAppWithBuffer_WhenRunning_ThenWritesBootstrapMessage(t *testing.T) {
	stdout := given_buffer()
	subject := New(stdout)

	actualErr := when_running(subject)

	then_noError(t, actualErr)

	expected := "lazygh is bootstrapped. TUI work starts in TODO 02.\n"
	actual := stdout.String()
	if actual != expected {
		t.Fatalf("expected %q, actual %q", expected, actual)
	}
}

func given_buffer() *bytes.Buffer {
	return &bytes.Buffer{}
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
