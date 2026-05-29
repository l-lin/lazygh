package tui

import "testing"

func TestModalEditorState_GivenStaleErrorsAndEditorIntents_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	lineEditor := newLineModalEditorState("Prompt", "draft")
	lineEditor.errorMessage = "stale"
	lineReported := lineEditor.withErrorMessage(" boom ")
	lineCleared := lineReported.withoutErrorMessage()
	lineUpdated, ok := lineReported.withLineEditorIntentApplied(newLineEditorInsertRuneIntent('!'))
	if !ok {
		t.Fatal("expected the line-editor intent to apply")
	}

	multilineEditor := newModalEditorState("Comment", "hello")
	multilineEditor.errorMessage = "stale"
	multilineUpdated, ok := multilineEditor.withMultilineEditorIntentApplied(newMultilineEditorInsertRuneIntent('!'))
	if !ok {
		t.Fatal("expected the multiline-editor intent to apply")
	}

	if actual := lineReported.errorMessage; actual != "boom" {
		t.Fatalf("expected line-editor error message %q, actual %q", "boom", actual)
	}
	if actual := lineCleared.errorMessage; actual != "" {
		t.Fatalf("expected cleared line-editor error message %q, actual %q", "", actual)
	}
	if actual := lineUpdated.Text(); actual != "draft!" {
		t.Fatalf("expected updated line-editor text %q, actual %q", "draft!", actual)
	}
	if actual := lineUpdated.errorMessage; actual != "" {
		t.Fatalf("expected updated line-editor error message %q, actual %q", "", actual)
	}
	if actual := multilineUpdated.Text(); actual != "hello!" {
		t.Fatalf("expected updated multiline text %q, actual %q", "hello!", actual)
	}
	if actual := multilineUpdated.errorMessage; actual != "" {
		t.Fatalf("expected updated multiline error message %q, actual %q", "", actual)
	}
	if actual := lineEditor.Text(); actual != "draft" {
		t.Fatalf("expected the original line-editor text %q, actual %q", "draft", actual)
	}
	if actual := lineEditor.errorMessage; actual != "stale" {
		t.Fatalf("expected the original line-editor error message %q, actual %q", "stale", actual)
	}
	if actual := multilineEditor.Text(); actual != "hello" {
		t.Fatalf("expected the original multiline text %q, actual %q", "hello", actual)
	}
	if actual := multilineEditor.errorMessage; actual != "stale" {
		t.Fatalf("expected the original multiline error message %q, actual %q", "stale", actual)
	}
}

func TestModalEditorState_GivenExternalEditorText_WhenApplying_ThenItNormalizesSingleLineAndClearsTheErrorWithoutMutatingTheOriginal(t *testing.T) {
	subject := newLineModalEditorState("Prompt", "draft")
	subject.errorMessage = "boom"

	actual := subject.withTextFromExternalEditor("Alpha\nBeta\n")

	if actual := actual.Text(); actual != "Alpha Beta" {
		t.Fatalf("expected external-editor text %q, actual %q", "Alpha Beta", actual)
	}
	if actual := actual.errorMessage; actual != "" {
		t.Fatalf("expected cleared external-editor error message %q, actual %q", "", actual)
	}
	if actual := subject.Text(); actual != "draft" {
		t.Fatalf("expected the original text %q, actual %q", "draft", actual)
	}
	if actual := subject.errorMessage; actual != "boom" {
		t.Fatalf("expected the original error message %q, actual %q", "boom", actual)
	}
}
