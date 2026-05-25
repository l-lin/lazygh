package clipboard

import (
	"errors"
	"testing"
)

func TestNewSystemReader_GivenANewReader_WhenCreating_ThenItConfiguresTheSystemClipboardReadFunction(t *testing.T) {
	actual := NewSystemReader()

	if actual == nil {
		t.Fatal("expected a reader instance")
	}
	if actual.readAll == nil {
		t.Fatal("expected the system clipboard read function to be configured")
	}
}

func TestReadText_GivenSystemReader_WhenReading_ThenItDelegatesToTheUnderlyingClipboardFunction(t *testing.T) {
	readCalled := false
	subject := SystemReader{readAll: func() (string, error) {
		readCalled = true
		return "https://example.test/pr/42", nil
	}}

	actual, actualErr := subject.ReadText()

	then_noError(t, actualErr)
	if actual != "https://example.test/pr/42" {
		t.Fatalf("expected text %q, actual %q", "https://example.test/pr/42", actual)
	}
	if !readCalled {
		t.Fatal("expected clipboard reader to be called")
	}
}

func TestReadText_GivenClipboardFailure_WhenReading_ThenItReturnsTheError(t *testing.T) {
	expected := errors.New("boom")
	subject := SystemReader{readAll: func() (string, error) {
		return "", expected
	}}

	_, actualErr := subject.ReadText()

	if !errors.Is(actualErr, expected) {
		t.Fatalf("expected error %v, actual %v", expected, actualErr)
	}
}
