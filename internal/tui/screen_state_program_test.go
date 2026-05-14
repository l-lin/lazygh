package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestScreenStateShell_GivenBrowserModeProgram_WhenHandlingNumberAndTabBindings_ThenTheProjectedScreenStateMatchesTheVisibleBehavior(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	then_screenFocusIs(t, subject.screenState(), ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)

	then_noError(t, given_handlerForScreenStateBinding(t, subject, panelViewName(sidePanelViewTwo), '1')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewOne)
	then_screenFocusIs(t, subject.screenState(), ActivePanelSide, sidePanelViewOne, FocusUserView)

	then_noError(t, given_handlerForScreenStateBinding(t, subject, panelViewName(sidePanelViewOne), '2')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	then_screenFocusIs(t, subject.screenState(), ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)

	then_noError(t, given_handlerForScreenStateBinding(t, subject, panelViewName(sidePanelViewTwo), ']')(gui, nil))
	then_viewTabLabelIs(t, subject.screenState(), sidePanelViewTwo, "Requested")

	then_noError(t, given_handlerForScreenStateBinding(t, subject, panelViewName(sidePanelViewTwo), '0')(gui, nil))
	then_currentPanelViewIs(t, gui, mainPanelViewZero)
	then_screenFocusIs(t, subject.screenState(), ActivePanelMain, mainPanelViewZero, FocusDetailView)
	if !subject.screenState().AllowsMainCursor() {
		t.Fatal("expected view 0 to keep the main cursor after the numeric jump")
	}
}

func TestScreenStateShell_GivenNilModelProgram_WhenRendering_ThenItStartsOnViewTwo(t *testing.T) {
	subject := NewProgramWithModelAndDeps(nil, AppDeps{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	then_screenFocusIs(t, subject.screenState(), ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)
}

func TestScreenStateShell_GivenOverlayTransitions_WhenHandlingActionsAndModalEditing_ThenTheProjectedKeyHintContextMatchesTheVisibleOverlay(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if actual := subject.screenState().KeyHintContext(); actual != KeyHintContextSidePanel {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextSidePanel, actual)
	}

	then_noError(t, given_handlerForScreenStateBinding(t, subject, panelViewName(sidePanelViewTwo), 'a')(gui, nil))
	if actual := subject.screenState().KeyHintContext(); actual != KeyHintContextActionsPopup {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextActionsPopup, actual)
	}
	then_currentViewNameIs(t, gui, viewActionsPopupName)

	subject.model.UpdateActionsPopupSearch("comment on pr", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "comment on pr"))
	then_noError(t, subject.refreshViews(gui))
	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))
	then_currentViewNameIs(t, gui, viewModalEditorName)
	if actual := subject.screenState().KeyHintContext(); actual != KeyHintContextModalEditor {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextModalEditor, actual)
	}

	then_noError(t, given_handlerForScreenStateBinding(t, subject, viewModalEditorName, gocui.KeyEsc)(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if actual := subject.screenState().KeyHintContext(); actual != KeyHintContextSidePanel {
		t.Fatalf("expected key hint context %v after closing the modal editor, actual %v", KeyHintContextSidePanel, actual)
	}
}
