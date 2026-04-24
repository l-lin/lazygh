package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestKeybindingSpecs_GivenProgram_WhenListingRecenterBindings_ThenUserPullRequestsDetailAndActionsPopupSupportZZ(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'z', handler: subject.recenterSideSelection})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'z', handler: subject.recenterSideSelection})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'z', handler: subject.armInlineConversationTogglePrefix})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'z', handler: subject.recenterActionsPopupSelection})
}

func TestRecenter_GivenUserViewSelection_WhenPressingZZ_ThenTheSelectionMovesToTheMiddleOfTheViewport(t *testing.T) {
	model := NewModel(SeedData{Users: given_manyItems("user", 40)})
	subject := NewProgramWithModel(model)
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)

	targetIndex := userView.InnerHeight() + 3
	for range targetIndex {
		subject.model.MoveSelectionDown()
	}
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	userView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)

	recenterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewUserName, 'z')
	actualErr = recenterHandler(gui, userView)
	then_noError(t, actualErr)
	actualErr = recenterHandler(gui, userView)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	userView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)

	_, actualOriginY := userView.Origin()
	_, actualCursorY := userView.Cursor()
	expectedOriginY := centeredViewportOrigin(targetIndex, userView.InnerHeight(), len(subject.model.VisibleUsers()))
	expectedCursorY := targetIndex - expectedOriginY
	if actualOriginY != expectedOriginY {
		t.Fatalf("expected user origin y %d, actual %d", expectedOriginY, actualOriginY)
	}
	if actualCursorY != expectedCursorY {
		t.Fatalf("expected user cursor y %d, actual %d", expectedCursorY, actualCursorY)
	}
}

func TestRecenter_GivenDetailCursor_WhenPressingZZ_ThenTheCursorMovesToTheMiddleOfTheViewport(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "user-1", Detail: given_multilineDetail(40)}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	targetLine := detailView.InnerHeight() + 3
	for range targetLine {
		actualErr = subject.moveSelectionDown(gui, detailView)
		then_noError(t, actualErr)
	}
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)

	recenterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	actualErr = recenterHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = recenterHandler(gui, detailView)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)

	_, actualOriginY := detailView.Origin()
	_, actualCursorY := detailView.Cursor()
	expectedOriginY := centeredViewportOrigin(targetLine, detailView.InnerHeight(), 40)
	expectedCursorY := targetLine - expectedOriginY
	if actualOriginY != expectedOriginY {
		t.Fatalf("expected detail origin y %d, actual %d", expectedOriginY, actualOriginY)
	}
	if actualCursorY != expectedCursorY {
		t.Fatalf("expected detail cursor y %d, actual %d", expectedCursorY, actualCursorY)
	}
}

func TestRecenter_GivenActionsPopupSelection_WhenPressingZZ_ThenTheSelectionMovesToTheMiddleOfThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	targetIndex := popupView.InnerHeight() + 1
	if targetIndex >= len(subject.currentActionsPopupActions()) {
		targetIndex = len(subject.currentActionsPopupActions()) - 1
	}
	for range targetIndex {
		actualErr = subject.moveActionsPopupSelectionDown(gui, popupView)
		then_noError(t, actualErr)
	}
	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)

	recenterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'z')
	actualErr = recenterHandler(gui, popupView)
	then_noError(t, actualErr)
	actualErr = recenterHandler(gui, popupView)
	then_noError(t, actualErr)
	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)

	_, actualOriginY := popupView.Origin()
	_, actualCursorY := popupView.Cursor()
	expectedOriginY := centeredViewportOrigin(targetIndex, popupView.InnerHeight(), len(subject.currentActionsPopupActions()))
	expectedCursorY := targetIndex - expectedOriginY
	if actualOriginY != expectedOriginY {
		t.Fatalf("expected actions popup origin y %d, actual %d", expectedOriginY, actualOriginY)
	}
	if actualCursorY != expectedCursorY {
		t.Fatalf("expected actions popup cursor y %d, actual %d", expectedCursorY, actualCursorY)
	}
}

func given_manyItems(prefix string, count int) []Item {
	items := make([]Item, 0, count)
	for index := range count {
		items = append(items, Item{Title: fmt.Sprintf("%s-%d", prefix, index+1), Detail: fmt.Sprintf("%s detail %d", prefix, index+1)})
	}
	return items
}

func given_multilineDetail(lineCount int) string {
	lines := make([]string, 0, lineCount)
	for index := range lineCount {
		lines = append(lines, fmt.Sprintf("line %d", index+1))
	}
	return strings.Join(lines, "\n")
}
