package tui

import "testing"

func TestUpdate_GivenMsgOpenPullRequestCustomSearchEditorRequested_WhenPullRequestsContextIsActive_ThenItClearsPendingPrefixAndOpensThePrefilledModal(t *testing.T) {
	subject := given_pullRequestCustomSearchProgram(&fakePullRequestDetailLoader{})
	subject.navigationState.pendingSelectionKeySequence = keySequenceState{pendingTarget: keySequenceTarget{viewName: viewPullRequestsName}}

	actual := Update(subject, MsgOpenPullRequestCustomSearchEditorRequested{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.modalEditorVisible() || !subject.overlayState.modalEditor.isLineEditor() {
		t.Fatal("expected the reducer to open the custom search editor")
	}
	if actualText := subject.overlayState.modalEditor.Text(); actualText != "--author @me --state open --sort updated --order desc" {
		t.Fatalf("expected modal editor text %q, actual %q", "--author @me --state open --sort updated --order desc", actualText)
	}
	if actualPending := subject.navigationState.pendingSelectionKeySequence.pendingTarget; actualPending != (keySequenceTarget{}) {
		t.Fatalf("expected pending selection prefix to be cleared, actual %+v", actualPending)
	}
}

func TestUpdate_GivenMsgExecuteSelectedActionsPopupActionRequested_WhenAPopupActionIsSelected_ThenItClearsPendingPrefixAndRoutesTheRequestedMessage(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	actions := subject.currentActionsPopupActions()
	subject.model.OpenActionsPopup(len(actions))
	subject.model.UpdateActionsPopupSearch("change theme", matchingActionsPopupIndexes(actions, "change theme"))
	subject.navigationState.pendingSelectionKeySequence = keySequenceState{pendingTarget: keySequenceTarget{viewName: viewActionsPopupName}}

	actual := Update(subject, MsgExecuteSelectedActionsPopupActionRequested{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.themePickerVisible() {
		t.Fatal("expected the selected popup action to route through update and open the theme picker")
	}
	if actualPending := subject.navigationState.pendingSelectionKeySequence.pendingTarget; actualPending != (keySequenceTarget{}) {
		t.Fatalf("expected pending selection prefix to be cleared, actual %+v", actualPending)
	}
}
