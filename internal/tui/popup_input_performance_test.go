package tui

import "testing"

func TestDispatchEditorMessage_GivenLineModalEditorOpenOnReviewDescription_WhenHandlingInputRepeatedly_ThenItStaysBelowTheRegressionCeiling(t *testing.T) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(t)
	defer gui.Close()
	then_noError(t, subject.openLineModalEditor(gui, "Prompt", "draft"))

	actual := testing.AllocsPerRun(10, func() {
		subject.setModalEditorState(newLineModalEditorState("Prompt", "draft"))
		if handled := subject.dispatchEditorMessage(MsgModalEditorLineInputRequested{Intent: newLineEditorInsertRuneIntent('!')}); !handled {
			t.Fatal("expected the modal editor input to be handled")
		}
	})

	const expectedMaximumAllocsPerRun = 6000.0
	if actual > expectedMaximumAllocsPerRun {
		t.Fatalf("expected modal-editor input allocations to stay below %.0f allocs/run, actual %.2f", expectedMaximumAllocsPerRun, actual)
	}
}

func TestDispatchEditorMessage_GivenActionsPopupSearchOpenOnReviewDescription_WhenHandlingInputRepeatedly_ThenItStaysBelowTheRegressionCeiling(t *testing.T) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(t)
	defer gui.Close()
	then_noError(t, subject.openActionsPopup(gui, nil))
	then_noError(t, subject.focusActionsPopupSearch(gui, nil))

	actual := testing.AllocsPerRun(10, func() {
		subject.model.UpdateActionsPopupSearch("", nil)
		subject.actionsPopupWidget = subject.actionsPopupWidget.withSearchEditorOpened("")
		if handled := subject.dispatchEditorMessage(MsgActionsPopupSearchInputRequested{Intent: newLineEditorInsertRuneIntent('r')}); !handled {
			t.Fatal("expected the actions popup search input to be handled")
		}
	})

	const expectedMaximumAllocsPerRun = 12000.0
	if actual > expectedMaximumAllocsPerRun {
		t.Fatalf("expected actions-popup search input allocations to stay below %.0f allocs/run, actual %.2f", expectedMaximumAllocsPerRun, actual)
	}
}
