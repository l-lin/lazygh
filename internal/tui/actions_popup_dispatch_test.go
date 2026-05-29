package tui

import "testing"

func TestUpdate_GivenMsgActionsPopupActionRequested_WhenApplying_ThenItRoutesTheTypedRequestedMessage(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)
	subject.actionsPopupWidget.searchEditor = newLineEditor("theme")
	subject.actionsPopupWidget.searchEditorVisible = true
	subject.actionsPopupWidget.errorMessage = "stale"
	subject.actionsPopupWidget.pendingConfirmationActionID = clearCacheActionTitle

	actual := Update(subject, MsgActionsPopupActionRequested{Action: actionsPopupAction{
		id:        "change-theme",
		requested: MsgOpenThemePickerRequested{},
	}})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.themePickerVisible() {
		t.Fatal("expected the typed action request to open the theme picker")
	}
	if !subject.model.ActionsPopupVisible() {
		t.Fatal("expected the typed action request to reopen the actions popup for the picker")
	}
	if subject.actionsPopupWidget.hasSearchEditor() {
		t.Fatal("expected the typed action request to clear the popup search editor")
	}
	if actualMessage := subject.actionsPopupWidget.errorMessage; actualMessage != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actualMessage)
	}
	if actualConfirmation := subject.actionsPopupWidget.pendingConfirmationActionID; actualConfirmation != "" {
		t.Fatalf("expected pending confirmation %q, actual %q", "", actualConfirmation)
	}
}

func TestUpdate_GivenMsgOpenReactionPickerRequested_WhenApplying_ThenItClearsPopupChromeAndReopensForReactions(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)
	subject.actionsPopupWidget.searchEditor = newLineEditor("react")
	subject.actionsPopupWidget.searchEditorVisible = true
	subject.actionsPopupWidget.errorMessage = "stale"
	subject.actionsPopupWidget.pendingConfirmationActionID = clearCacheActionTitle

	actual := Update(subject, MsgOpenReactionPickerRequested{Target: pullRequestReactionActionTarget{repository: "acme/widgets", number: 42, subjectID: "subject-1"}})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.reactionPickerVisible() {
		t.Fatal("expected the popup to reopen with the reaction picker visible")
	}
	if !subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to stay visible for the reaction picker")
	}
	if subject.actionsPopupWidget.hasSearchEditor() {
		t.Fatal("expected the reaction picker to clear the popup search editor")
	}
	if actualMessage := subject.actionsPopupWidget.errorMessage; actualMessage != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actualMessage)
	}
	if actualConfirmation := subject.actionsPopupWidget.pendingConfirmationActionID; actualConfirmation != "" {
		t.Fatalf("expected pending confirmation %q, actual %q", "", actualConfirmation)
	}
	if actualTarget := subject.actionsPopupWidget.reactionPicker.target.subjectID; actualTarget != "subject-1" {
		t.Fatalf("expected reaction target %q, actual %q", "subject-1", actualTarget)
	}
}
