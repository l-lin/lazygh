package tui

import (
	"testing"
	"time"
)

func TestActionsPopupWidgetState_GivenStaleConfirmationAndError_WhenUpdatingTheChrome_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := actionsPopupWidgetState{
		errorMessage:                      "stale",
		pendingConfirmationActionID:       "old-action",
		assigneePickerSearchDebounceDelay: 123 * time.Millisecond,
	}

	confirmed := subject.withPendingConfirmation(" " + clearCacheActionTitle + " ")
	reported := confirmed.withErrorMessage(" boom ")
	cleared := reported.withPendingConfirmationCleared().withoutErrorMessage()

	if actual := confirmed.pendingConfirmationActionID; actual != clearCacheActionTitle {
		t.Fatalf("expected pending confirmation %q, actual %q", clearCacheActionTitle, actual)
	}
	if actual := confirmed.errorMessage; actual != "" {
		t.Fatalf("expected confirmation to clear the popup error, actual %q", actual)
	}
	if actual := reported.errorMessage; actual != "boom" {
		t.Fatalf("expected popup error message %q, actual %q", "boom", actual)
	}
	if actual := reported.pendingConfirmationActionID; actual != clearCacheActionTitle {
		t.Fatalf("expected reported confirmation %q, actual %q", clearCacheActionTitle, actual)
	}
	if actual := cleared.pendingConfirmationActionID; actual != "" {
		t.Fatalf("expected pending confirmation %q, actual %q", "", actual)
	}
	if actual := cleared.errorMessage; actual != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actual)
	}
	if actual := cleared.assigneePickerSearchDebounceDelay; actual != 123*time.Millisecond {
		t.Fatalf("expected debounce delay %s, actual %s", 123*time.Millisecond, actual)
	}
	if actual := subject.pendingConfirmationActionID; actual != "old-action" {
		t.Fatalf("expected the original confirmation %q, actual %q", "old-action", actual)
	}
	if actual := subject.errorMessage; actual != "stale" {
		t.Fatalf("expected the original popup error %q, actual %q", "stale", actual)
	}
}

func TestActionsPopupWidgetState_GivenSearchAndPickerChrome_WhenOpeningAnAssigneePickerAndClosing_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	target := pullRequestAssigneePickerTarget{repository: "acme/widgets", number: 42}
	subject := actionsPopupWidgetState{
		searchEditor:                      newLineEditor("stale"),
		searchEditorVisible:               true,
		errorMessage:                      "stale",
		pendingConfirmationActionID:       clearCacheActionTitle,
		reactionPicker:                    &reactionPickerState{target: pullRequestReactionActionTarget{repository: "acme/widgets", number: 42, subjectID: "subject-1"}},
		themePicker:                       &themePickerState{},
		assigneePickerLoad:                &assigneePickerLoadState{target: target, command: "gh api graphql"},
		assigneePickerSearchDebounceDelay: 123 * time.Millisecond,
	}

	opened := subject.withAssigneePickerOpened(target, "viewer", "Viewer")
	closed := opened.withPopupClosed()

	if opened.assigneePicker == nil {
		t.Fatal("expected the assignee picker to be opened")
	}
	if actual := opened.assigneePicker.target.repository; actual != "acme/widgets" {
		t.Fatalf("expected opened repository %q, actual %q", "acme/widgets", actual)
	}
	if actual := opened.assigneePicker.viewerLogin; actual != "viewer" {
		t.Fatalf("expected opened viewer login %q, actual %q", "viewer", actual)
	}
	if opened.searchEditorVisible {
		t.Fatal("expected opening the assignee picker to clear the popup search editor")
	}
	if actual := opened.errorMessage; actual != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actual)
	}
	if actual := opened.pendingConfirmationActionID; actual != "" {
		t.Fatalf("expected pending confirmation %q, actual %q", "", actual)
	}
	if opened.reactionPicker != nil {
		t.Fatal("expected opening the assignee picker to clear the reaction picker")
	}
	if opened.themePicker != nil {
		t.Fatal("expected opening the assignee picker to clear the theme picker")
	}
	if opened.assigneePickerLoad != nil {
		t.Fatal("expected opening the assignee picker to clear any stale assignee-picker load state")
	}
	if closed.assigneePicker != nil || closed.reactionPicker != nil || closed.themePicker != nil || closed.assigneePickerLoad != nil {
		t.Fatalf("expected closing the popup to clear every picker state, actual %+v", closed)
	}
	if closed.searchEditorVisible {
		t.Fatal("expected closing the popup to clear the popup search editor")
	}
	if actual := closed.errorMessage; actual != "" {
		t.Fatalf("expected closed popup error message %q, actual %q", "", actual)
	}
	if actual := closed.pendingConfirmationActionID; actual != "" {
		t.Fatalf("expected closed pending confirmation %q, actual %q", "", actual)
	}
	if actual := closed.assigneePickerSearchDebounceDelay; actual != 123*time.Millisecond {
		t.Fatalf("expected debounce delay %s, actual %s", 123*time.Millisecond, actual)
	}
	if !subject.searchEditorVisible {
		t.Fatal("expected the original popup search editor to stay visible")
	}
	if actual := subject.errorMessage; actual != "stale" {
		t.Fatalf("expected the original popup error %q, actual %q", "stale", actual)
	}
	if actual := subject.pendingConfirmationActionID; actual != clearCacheActionTitle {
		t.Fatalf("expected the original pending confirmation %q, actual %q", clearCacheActionTitle, actual)
	}
	if subject.reactionPicker == nil {
		t.Fatal("expected the original reaction picker to stay intact")
	}
	if subject.themePicker == nil {
		t.Fatal("expected the original theme picker to stay intact")
	}
	if subject.assigneePickerLoad == nil {
		t.Fatal("expected the original assignee-picker load state to stay intact")
	}
}
