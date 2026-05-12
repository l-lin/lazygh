package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestPanelViewContracts_GivenBrowserModeModel_WhenChangingTheActiveSideViewAndSelection_ThenViewZeroTracksThatSideSelection(t *testing.T) {
	subject := given_model()

	when_focusingPanelViewNumber(subject, sidePanelViewOne)
	when_selectingNextSidePanelRow(subject)
	then_viewZeroContentIs(t, subject, "User detail 2")

	when_focusingPanelViewNumber(subject, sidePanelViewTwo)
	when_selectingNextSidePanelRow(subject)
	then_viewZeroContentIs(t, subject, "My PR detail 2")

	when_focusingPanelViewNumber(subject, sidePanelViewThree)
	when_selectingNextSidePanelRow(subject)
	then_viewZeroContentIs(t, subject, "Notification detail 2")
}

func TestPanelViewContracts_GivenBrowserModeModel_WhenCyclingTheSidePanel_ThenViewsOneTwoAndThreeWrapInOrder(t *testing.T) {
	subject := given_model()

	when_focusingPanelViewNumber(subject, sidePanelViewOne)
	subject.NextSideView()
	then_modelFocusesPanelViewNumber(t, subject, sidePanelViewTwo)

	subject.NextSideView()
	then_modelFocusesPanelViewNumber(t, subject, sidePanelViewThree)

	subject.NextSideView()
	then_modelFocusesPanelViewNumber(t, subject, sidePanelViewOne)

	subject.PreviousSideView()
	then_modelFocusesPanelViewNumber(t, subject, sidePanelViewThree)
}

func TestPanelViewContracts_GivenBrowserModeProgram_WhenPressingSideViewKeys_ThenHAndShiftTabMoveLeftWhileLAndTabMoveRight(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_currentPanelViewIs(t, gui, sidePanelViewOne)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewOne), 'l')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), "", gocui.KeyTab)(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewThree)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewThree), 'h')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), "", gocui.KeyBacktab)(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewOne)
}

func TestPanelViewContracts_GivenBrowserModeProgram_WhenPressingNumberShortcuts_ThenZeroOneTwoAndThreeJumpToTheirPanelViews(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), '0')(gui, nil))
	then_currentPanelViewIs(t, gui, mainPanelViewZero)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(mainPanelViewZero), '1')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewOne)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewOne), '2')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), '3')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewThree)
}

func TestPanelViewContracts_GivenBrowserModeProgram_WhenPressingBrackets_ThenOnlyTabbedViewsUseThemForTheirLocalTabs(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_bindingDoesNotExist(t, subject.keybindingSpecs(), panelViewName(sidePanelViewOne), '[')
	then_bindingDoesNotExist(t, subject.keybindingSpecs(), panelViewName(sidePanelViewThree), ']')

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), ']')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if actual := subject.model.ActivePullRequestTab(); actual != RequestedPullRequestsTab {
		t.Fatalf("expected view 2 to switch to tab %v, actual %v", RequestedPullRequestsTab, actual)
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), '[')(gui, nil))
	if actual := subject.model.ActivePullRequestTab(); actual != MyPullRequestsTab {
		t.Fatalf("expected view 2 to switch back to tab %v, actual %v", MyPullRequestsTab, actual)
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), '0')(gui, nil))
	then_currentPanelViewIs(t, gui, mainPanelViewZero)
	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(mainPanelViewZero), ']')(gui, nil))
	then_currentPanelViewIs(t, gui, mainPanelViewZero)
	if actual := subject.activeDetailTab; actual != CommentsDetailTab {
		t.Fatalf("expected view 0 to switch to tab %v, actual %v", CommentsDetailTab, actual)
	}
}

func when_selectingNextSidePanelRow(subject *Model) {
	subject.MoveSelectionDown()
}

func then_viewZeroContentIs(t *testing.T, subject *Model, expected string) {
	t.Helper()

	if actual := subject.DetailContent(); actual != expected {
		t.Fatalf("expected view 0 content %q, actual %q", expected, actual)
	}
}
