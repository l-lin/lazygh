package tui

import "testing"

func TestUpdate_GivenMsgActionsPopupPageRequested_WhenApplying_ThenItReturnsATypedPageSizeCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	actual := Update(subject, MsgActionsPopupPageRequested{Kind: pageNavigationKindHalfDown})

	if len(actual) != 1 {
		t.Fatalf("expected one actions-popup page command, actual %d", len(actual))
	}
	if _, ok := actual[0].(resolveActionsPopupPageSizeCmd); !ok {
		t.Fatalf("expected a resolveActionsPopupPageSizeCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgActionsPopupPageResolved_WhenApplying_ThenItMovesSelectionAndReturnsATypedViewportCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	initialIndex := subject.model.actionsPopup.selectedActionIndex
	expected := adjustVisibleSelection(initialIndex, subject.model.actionsPopup.filteredActionIndexes, pageDelta(6))

	actual := Update(subject, MsgActionsPopupPageResolved{Kind: pageNavigationKindHalfDown, PageSize: 6})

	if actualIndex := subject.model.actionsPopup.selectedActionIndex; actualIndex != expected {
		t.Fatalf("expected actions-popup selection index %d, actual %d", expected, actualIndex)
	}
	if len(actual) != 1 {
		t.Fatalf("expected one actions-popup viewport command, actual %d", len(actual))
	}
	command, ok := actual[0].(actionsPopupViewportCmd)
	if !ok {
		t.Fatalf("expected an actionsPopupViewportCmd, actual %T", actual[0])
	}
	if command.Placement != viewportPlacementCenter {
		t.Fatalf("expected centered popup viewport placement %v, actual %v", viewportPlacementCenter, command.Placement)
	}
}

func TestUpdate_GivenMsgActionsPopupViewportRequested_WhenApplying_ThenItReturnsATypedViewportCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	actual := Update(subject, MsgActionsPopupViewportRequested{Placement: viewportPlacementTop})

	if len(actual) != 1 {
		t.Fatalf("expected one actions-popup viewport command, actual %d", len(actual))
	}
	command, ok := actual[0].(actionsPopupViewportCmd)
	if !ok {
		t.Fatalf("expected an actionsPopupViewportCmd, actual %T", actual[0])
	}
	if command.Placement != viewportPlacementTop {
		t.Fatalf("expected viewport placement %v, actual %v", viewportPlacementTop, command.Placement)
	}
}
