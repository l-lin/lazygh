package tui

import "testing"

func TestUpdate_GivenMsgAppStarted_WhenApplying_ThenItMarksTheProgramAsStarted(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgAppStarted{})

	if !subject.appStarted {
		t.Fatal("expected the program to be marked as started")
	}
	if len(actual) != 0 {
		t.Fatalf("expected no explicit commands from startup update, actual %d", len(actual))
	}
}

func TestUpdate_GivenMsgNextSideView_WhenApplying_ThenItMovesFocusToTheNextSidePane(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	Update(subject, MsgNextSideView{})

	if actual := subject.model.Focus(); actual != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actual)
	}
}

func TestUpdate_GivenSearchLifecycleMessages_WhenApplying_ThenItFlowsThroughTheCentralReducer(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	Update(subject, MsgOpenSearch{Query: ""})
	Update(subject, MsgSearchDraftChanged{Query: "1"})
	Update(subject, MsgSubmitSearch{})

	if subject.model.SearchActive() {
		t.Fatal("expected the search to be inactive after submit")
	}
	if actual := subject.model.UserSearchQuery(); actual != "1" {
		t.Fatalf("expected the applied user search query %q, actual %q", "1", actual)
	}
	if subject.searchEditor != nil {
		t.Fatal("expected the search editor to be cleared after submit")
	}
}

func TestUpdate_GivenActionsPopupLifecycleMessages_WhenApplying_ThenItResetsEphemeralPopupState(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.actionsPopupSearchEditor = newLineEditor("stale")
	subject.actionsPopupErrorMessage = "boom"
	subject.reactionPicker = &reactionPickerState{}
	subject.themePicker = &themePickerState{}
	subject.assigneePicker = &assigneePickerState{}
	subject.assigneePickerLoad = &assigneePickerLoadState{}

	Update(subject, MsgOpenActionsPopup{ActionCount: 3})
	Update(subject, MsgMoveActionsPopupSelection{Delta: 1})
	Update(subject, MsgCloseActionsPopup{})

	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to be hidden after close")
	}
	if actual := subject.model.ActionsPopupSelectedActionIndex(); actual != 0 {
		t.Fatalf("expected the popup selection to reset after close, actual %d", actual)
	}
	if subject.actionsPopupSearchEditor != nil {
		t.Fatal("expected the popup search editor to be cleared")
	}
	if subject.actionsPopupErrorMessage != "" {
		t.Fatalf("expected the popup error message to be cleared, actual %q", subject.actionsPopupErrorMessage)
	}
	if subject.reactionPicker != nil || subject.themePicker != nil || subject.assigneePicker != nil || subject.assigneePickerLoad != nil {
		t.Fatal("expected popup-only picker state to be cleared")
	}
}
