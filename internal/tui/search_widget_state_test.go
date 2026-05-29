package tui

import "testing"

func TestSearchWidgetState_GivenEditorAndDirectionTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := searchWidgetState{editor: newLineEditor("draft"), editorVisible: true, detailReversed: true}

	opened := searchWidgetState{}.withEditorOpened(" query ")
	updated, ok := subject.withEditorIntentApplied(newLineEditorInsertRuneIntent('!'))
	if !ok {
		t.Fatal("expected the search editor intent to apply")
	}
	cleared := updated.withEditorCleared()
	directionReset := cleared.withDetailSearchDirection(false)

	if !opened.hasEditor() {
		t.Fatal("expected opening the editor to make it visible")
	}
	if actual := opened.editor.Text(); actual != " query " {
		t.Fatalf("expected opened editor text %q, actual %q", " query ", actual)
	}
	if actual := updated.editor.Text(); actual != "draft!" {
		t.Fatalf("expected updated editor text %q, actual %q", "draft!", actual)
	}
	if !updated.detailReversed {
		t.Fatal("expected editor input to preserve the existing detail-search direction")
	}
	if cleared.hasEditor() {
		t.Fatal("expected clearing the editor to hide it")
	}
	if actual := cleared.editor.Text(); actual != "" {
		t.Fatalf("expected cleared editor text %q, actual %q", "", actual)
	}
	if directionReset.detailReversed {
		t.Fatal("expected the detail-search direction to reset to forward")
	}
	if actual := subject.editor.Text(); actual != "draft" {
		t.Fatalf("expected the original editor text %q, actual %q", "draft", actual)
	}
	if !subject.hasEditor() {
		t.Fatal("expected the original editor visibility to stay true")
	}
	if !subject.detailReversed {
		t.Fatal("expected the original detail-search direction to stay reversed")
	}
}
