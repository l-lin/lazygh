package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestRefreshOverlayView_GivenAHiddenOverlay_WhenRefreshing_ThenItDeletesTheStaleView(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_, actualErr := gui.SetView("stale-overlay", 0, 0, 10, 3, 0)
	then_unknownViewOrNoError(t, actualErr)

	actualErr = subject.refreshOverlayView(gui, false, "stale-overlay", func(_ *gocui.View) {
		t.Fatal("expected hidden overlay refresh to skip configuration")
	}, func(_ *gocui.View) {
		t.Fatal("expected hidden overlay refresh to skip rendering")
	})
	then_noError(t, actualErr)
	then_viewDoesNotExist(t, gui, "stale-overlay")
}

func TestRefreshActionsPopupViews_GivenAClosedPopup_WhenRefreshing_ThenItDeletesTheStaleViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_, actualErr := gui.SetView(viewActionsPopupName, 0, 0, 20, 10, 0)
	then_unknownViewOrNoError(t, actualErr)
	_, actualErr = gui.SetView(viewActionsPopupSearchName, 0, 11, 20, 12, 0)
	then_unknownViewOrNoError(t, actualErr)

	actualErr = subject.refreshActionsPopupViews(gui)
	then_noError(t, actualErr)
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_viewDoesNotExist(t, gui, viewActionsPopupSearchName)
}

func then_unknownViewOrNoError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil && !isUnknownViewError(actual) {
		t.Fatalf("expected no error or unknown view, actual %v", actual)
	}
}
