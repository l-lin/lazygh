package clipboard

import (
	"errors"
	"testing"
)

func TestWriteText_GivenSystemWriter_WhenWriting_ThenItDelegatesToTheUnderlyingClipboardFunction(t *testing.T) {
	writeCalled := false
	subject := SystemWriter{writeAll: func(actual string) error {
		writeCalled = true
		if actual != "https://example.test/pr/42" {
			t.Fatalf("expected text %q, actual %q", "https://example.test/pr/42", actual)
		}
		return nil
	}}

	actualErr := subject.WriteText("https://example.test/pr/42")

	then_noError(t, actualErr)
	if !writeCalled {
		t.Fatal("expected clipboard writer to be called")
	}
}

func TestWriteText_GivenClipboardFailure_WhenWriting_ThenItReturnsTheError(t *testing.T) {
	expected := errors.New("boom")
	subject := SystemWriter{writeAll: func(string) error {
		return expected
	}}

	actualErr := subject.WriteText("https://example.test/pr/42")

	if !errors.Is(actualErr, expected) {
		t.Fatalf("expected error %v, actual %v", expected, actualErr)
	}
}

func then_noError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Fatalf("expected no error, actual %v", actual)
	}
}
