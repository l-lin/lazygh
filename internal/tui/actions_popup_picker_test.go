package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestActionsPopup_GivenOpenPopup_WhenRenderingThePicker_ThenItFocusesTheInlineSearchAboveTheActionList(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !gui.Cursor {
		t.Fatal("expected the cursor to stay visible while the picker input is focused")
	}

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	_, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if searchView.Frame {
		t.Fatal("expected the picker input to stay borderless inside the popup")
	}

	searchX0, searchY0, searchX1, _, actualErr := gui.ViewPosition(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	popupX0, popupY0, popupX1, _, actualErr := gui.ViewPosition(viewActionsPopupName)
	then_noError(t, actualErr)
	if searchX0 != popupX0 || searchX1 != popupX1 {
		t.Fatalf("expected the picker input to share the popup width, search=(%d,%d) popup=(%d,%d)", searchX0, searchX1, popupX0, popupX1)
	}
	if searchY0 >= popupY0 {
		t.Fatalf("expected the picker input to render above the action list, search y=%d popup y=%d", searchY0, popupY0)
	}
}

func TestActionsPopup_GivenPickerSearchQuery_WhenTyping_ThenItFiltersTheVisibleActionsLive(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "clipboard" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Yank URL to clipboard") {
		t.Fatalf("expected the filtered popup to keep %q visible, actual %q", "Yank URL to clipboard", popupView.Buffer())
	}
	for _, unexpected := range []string{"Comment on PR", "Edit PR title", "Change theme"} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the filtered popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenPickerSearchFocus_WhenUsingCtrlNAndCtrlP_ThenItMovesThroughTheFilteredActionsWithoutChangingTheQuery(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "edit" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	expectedIndexes := matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit")
	if len(expectedIndexes) < 2 {
		t.Fatalf("expected at least two filtered edit actions, actual %v", expectedIndexes)
	}
	if subject.model.ActionsPopupSelectedActionIndex() != expectedIndexes[0] {
		t.Fatalf("expected the picker to start on filtered action index %d, actual %d", expectedIndexes[0], subject.model.ActionsPopupSelectedActionIndex())
	}

	moveDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyCtrlN)
	then_noError(t, moveDownHandler(gui, searchView))
	if subject.model.ActionsPopupSelectedActionIndex() != expectedIndexes[1] {
		t.Fatalf("expected ctrl+n to move to filtered action index %d, actual %d", expectedIndexes[1], subject.model.ActionsPopupSelectedActionIndex())
	}
	if actual := subject.currentActionsPopupSearchText(); actual != "edit" {
		t.Fatalf("expected ctrl+n to keep the picker query %q, actual %q", "edit", actual)
	}

	moveUpHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyCtrlP)
	then_noError(t, moveUpHandler(gui, searchView))
	if subject.model.ActionsPopupSelectedActionIndex() != expectedIndexes[0] {
		t.Fatalf("expected ctrl+p to move back to filtered action index %d, actual %d", expectedIndexes[0], subject.model.ActionsPopupSelectedActionIndex())
	}
}

func TestActionsPopup_GivenPickerSearchFocus_WhenUsingCtrlDAndCtrlU_ThenTheyKeepEditingTheQuery(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "copy" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	if !subject.editActionsPopupSearch(searchView, gocui.KeyCtrlB, 0, gocui.ModNone) {
		t.Fatal("expected ctrl+b to be handled by the picker input")
	}
	if !subject.editActionsPopupSearch(searchView, gocui.KeyCtrlB, 0, gocui.ModNone) {
		t.Fatal("expected ctrl+b to be handled by the picker input")
	}
	if !subject.editActionsPopupSearch(searchView, gocui.KeyCtrlD, 0, gocui.ModNone) {
		t.Fatal("expected ctrl+d to be handled by the picker input")
	}
	if actual := subject.currentActionsPopupSearchText(); actual != "coy" {
		t.Fatalf("expected ctrl+d to delete the next character in the picker query, actual %q", actual)
	}
	if !subject.editActionsPopupSearch(searchView, gocui.KeyCtrlU, 0, gocui.ModNone) {
		t.Fatal("expected ctrl+u to be handled by the picker input")
	}
	if actual := subject.currentActionsPopupSearchText(); actual != "y" {
		t.Fatalf("expected ctrl+u to delete back to the start of the picker query, actual %q", actual)
	}
}

func TestActionsPopup_GivenPickerSearchFocus_WhenPressingEnter_ThenItExecutesTheSelectedAction(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range pullRequestCommentComposerTitle {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	executeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	then_noError(t, executeHandler(gui, searchView))
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected enter to close the picker after executing the selected action")
	}
	if !subject.modalEditorVisible() {
		t.Fatal("expected enter to open the pull request comment composer")
	}
}
