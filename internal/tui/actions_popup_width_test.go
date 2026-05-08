package tui

import "testing"

func TestActionsPopup_GivenDefaultLayout_WhenOpening_ThenItUsesAHalfSizedPopupWidth(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	x0, _, x1, _, actualErr := gui.ViewPosition(viewActionsPopupName)
	then_noError(t, actualErr)
	if actual := x1 - x0 + 1; actual != 30 {
		t.Fatalf("expected actions popup width %d, actual %d", 30, actual)
	}
}
