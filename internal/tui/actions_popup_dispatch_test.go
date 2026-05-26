package tui

import "testing"

func TestUpdate_GivenMsgActionsPopupActionRequested_WhenApplying_ThenItRoutesTheTypedRequestedMessage(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)

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
}
