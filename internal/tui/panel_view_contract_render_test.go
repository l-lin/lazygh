package tui

import (
	"strings"
	"testing"
)

func TestPanelViewContracts_GivenBrowserModeRender_WhenRendering_ThenViewsZeroThroughThreeKeepTheirVisibleNumbers(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	for _, viewNumber := range []panelViewNumber{mainPanelViewZero, sidePanelViewOne, sidePanelViewTwo, sidePanelViewThree} {
		then_panelViewShowsVisibleNumber(t, gui, viewNumber)
	}
}

func TestPanelViewContracts_GivenBrowserModeProgram_WhenSwitchingBetweenSidePanelAndViewZero_ThenOnlyViewZeroShowsTheMainCursor(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if gui.Cursor {
		t.Fatal("expected side-panel focus to hide the main cursor")
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), '0')(gui, nil))
	then_currentPanelViewIs(t, gui, mainPanelViewZero)
	if !gui.Cursor {
		t.Fatal("expected view 0 focus to show the main cursor")
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(mainPanelViewZero), '1')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewOne)
	if gui.Cursor {
		t.Fatal("expected side-panel focus to hide the main cursor again")
	}
}

func TestPanelViewContracts_GivenOverlayTransitions_WhenRendering_ThenBottomRightKeyHintsFollowTheActivePanelAndOverlay(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_statusLineKeyHintsAreRightAligned(t, gui, "?: help, /: search, a: action")

	then_noError(t, subject.openActionsPopup(gui, nil))
	then_currentViewNameIs(t, gui, viewActionsPopupName)
	then_statusLineKeyHintsAreRightAligned(t, gui, "/: search, Enter: execute, Escape: cancel")

	subject.model.UpdateActionsPopupSearch("comment on pr", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "comment on pr"))
	then_noError(t, subject.refreshViews(gui))
	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))
	then_currentViewNameIs(t, gui, viewModalEditorName)
	then_statusLineKeyHintsAreRightAligned(t, gui, "Alt+Enter: submit, Ctrl+G: editor, Escape: cancel")
}

func TestPanelViewContracts_GivenActionsAndTextEditing_WhenOpeningOverlays_ThenActionsUsePopupChromeAndEditingUsesInputBoxPopupChrome(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !popupView.Frame {
		t.Fatal("expected actions to use framed popup chrome")
	}
	if popupView.Editable {
		t.Fatal("expected actions popup to stay read-only")
	}
	if !strings.Contains(popupView.Title, "Actions") {
		t.Fatalf("expected actions popup title to contain %q, actual %q", "Actions", popupView.Title)
	}

	subject.model.UpdateActionsPopupSearch("comment on pr", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "comment on pr"))
	then_noError(t, subject.refreshViews(gui))
	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))
	then_viewDoesNotExist(t, gui, viewActionsPopupName)

	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !modalView.Frame {
		t.Fatal("expected text editing to use framed input-box chrome")
	}
	if !modalView.Editable {
		t.Fatal("expected the input-box popup to stay editable")
	}
	if modalView.Footer != "" {
		t.Fatalf("expected the input-box popup footer to stay empty, actual %q", modalView.Footer)
	}
	if !strings.Contains(modalView.Title, pullRequestCommentComposerTitle) {
		t.Fatalf("expected the input-box popup title to contain %q, actual %q", pullRequestCommentComposerTitle, modalView.Title)
	}
}
