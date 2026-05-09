package tui

import (
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"github.com/jesseduffield/gocui"
)

func TestViewportPlacement_GivenUserViewSelection_WhenPressingZT_ThenItPlacesTheSelectionAtTheTopOfTheViewport(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{Users: given_manyItems("user", 40)}))
	gui, userView, targetIndex := given_userViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewUserName, userView, 'z', 't')

	actualView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheTopOfTheViewport(t, actualView, targetIndex, len(subject.model.VisibleUsers()))
}

func TestViewportPlacement_GivenUserViewSelection_WhenPressingZZ_ThenItPlacesTheSelectionAtTheCenterOfTheViewport(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{Users: given_manyItems("user", 40)}))
	gui, userView, targetIndex := given_userViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewUserName, userView, 'z', 'z')

	actualView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheCenterOfTheViewport(t, actualView, targetIndex, len(subject.model.VisibleUsers()))
}

func TestViewportPlacement_GivenUserViewSelection_WhenPressingZB_ThenItPlacesTheSelectionAtTheBottomOfTheViewport(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{Users: given_manyItems("user", 40)}))
	gui, userView, targetIndex := given_userViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewUserName, userView, 'z', 'b')

	actualView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheBottomOfTheViewport(t, actualView, targetIndex, len(subject.model.VisibleUsers()))
}

func TestViewportPlacement_GivenDetailCursor_WhenPressingZT_ThenItPlacesTheCursorAtTheTopOfTheViewport(t *testing.T) {
	subject := NewProgramWithModel(given_detailPlacementModel())
	gui, detailView, targetLine := given_detailViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewDetailName, detailView, 'z', 't')

	actualView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewPlacesCursorAtTheTopOfTheViewport(t, actualView, targetLine, 40)
}

func TestViewportPlacement_GivenDetailCursor_WhenPressingZZ_ThenItPlacesTheCursorAtTheCenterOfTheViewport(t *testing.T) {
	subject := NewProgramWithModel(given_detailPlacementModel())
	gui, detailView, targetLine := given_detailViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewDetailName, detailView, 'z', 'z')

	actualView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewPlacesCursorAtTheCenterOfTheViewport(t, actualView, targetLine, 40)
}

func TestViewportPlacement_GivenDetailCursor_WhenPressingZB_ThenItPlacesTheCursorAtTheBottomOfTheViewport(t *testing.T) {
	subject := NewProgramWithModel(given_detailPlacementModel())
	gui, detailView, targetLine := given_detailViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewDetailName, detailView, 'z', 'b')

	actualView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewPlacesCursorAtTheBottomOfTheViewport(t, actualView, targetLine, 40)
}

func TestViewportPlacement_GivenActionsPopupSelection_WhenPressingZT_ThenItPlacesTheSelectionAtTheTopOfThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui, popupView, targetIndex := given_actionsPopupPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewActionsPopupName, popupView, 'z', 't')

	actualView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheTopOfTheViewport(t, actualView, targetIndex, subject.currentActionsPopupRenderedLineCount())
}

func TestViewportPlacement_GivenActionsPopupSelection_WhenPressingZZ_ThenItPlacesTheSelectionAtTheCenterOfThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui, popupView, targetIndex := given_actionsPopupPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewActionsPopupName, popupView, 'z', 'z')

	actualView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheCenterOfTheViewport(t, actualView, targetIndex, subject.currentActionsPopupRenderedLineCount())
}

func TestViewportPlacement_GivenActionsPopupSelection_WhenPressingZB_ThenItPlacesTheSelectionAtTheBottomOfThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui, popupView, targetIndex := given_actionsPopupPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewActionsPopupName, popupView, 'z', 'b')

	actualView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheBottomOfTheViewport(t, actualView, targetIndex, subject.currentActionsPopupRenderedLineCount())
}

func TestViewportPlacement_GivenRemappedSideViewportPlacementBindings_WhenPressingXTXXAndXB_ThenItUsesTheConfiguredTopCenterAndBottomPlacementKeys(t *testing.T) {
	subject := given_programWithKeymapOverrides(NewModel(SeedData{Users: given_manyItems("user", 40)}), appconfig.KeymapOverrides{
		"side": {
			"place_selection_at_viewport_top":    {"xt"},
			"recenter_selection":                 {"xx"},
			"place_selection_at_viewport_bottom": {"xb"},
		},
	})
	gui, userView, targetIndex := given_userViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewUserName, userView, 'x', 't')
	actualUserView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheTopOfTheViewport(t, actualUserView, targetIndex, len(subject.model.VisibleUsers()))

	when_pressingKeySequence(t, subject, gui, viewUserName, actualUserView, 'x', 'x')
	actualUserView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheCenterOfTheViewport(t, actualUserView, targetIndex, len(subject.model.VisibleUsers()))

	when_pressingKeySequence(t, subject, gui, viewUserName, actualUserView, 'x', 'b')
	actualUserView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewPlacesSelectionAtTheBottomOfTheViewport(t, actualUserView, targetIndex, len(subject.model.VisibleUsers()))
}

func TestViewportPlacement_GivenRemappedDetailViewportPlacementBindings_WhenPressingMTMMAndMB_ThenItUsesTheConfiguredTopCenterAndBottomPlacementKeys(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_detailPlacementModel(), appconfig.KeymapOverrides{
		"detail": {
			"place_cursor_at_viewport_top":    {"mt"},
			"recenter_cursor":                 {"mm"},
			"place_cursor_at_viewport_bottom": {"mb"},
		},
	})
	gui, detailView, targetLine := given_detailViewPlacementScenario(t, subject)
	defer gui.Close()

	when_pressingKeySequence(t, subject, gui, viewDetailName, detailView, 'm', 't')
	actualDetailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewPlacesCursorAtTheTopOfTheViewport(t, actualDetailView, targetLine, 40)

	when_pressingKeySequence(t, subject, gui, viewDetailName, actualDetailView, 'm', 'm')
	actualDetailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewPlacesCursorAtTheCenterOfTheViewport(t, actualDetailView, targetLine, 40)

	when_pressingKeySequence(t, subject, gui, viewDetailName, actualDetailView, 'm', 'b')
	actualDetailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewPlacesCursorAtTheBottomOfTheViewport(t, actualDetailView, targetLine, 40)
}

func given_userViewPlacementScenario(t *testing.T, subject *Program) (*gocui.Gui, *gocui.View, int) {
	t.Helper()

	gui := given_headlessGuiWithSize(t, 120, 12)
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

	return gui, userView, targetIndex
}

func given_detailPlacementModel() *Model {
	model := NewModel(SeedData{Users: []Item{{Title: "user-1", Detail: given_multilineDetail(40)}}})
	model.OpenDetail()
	return model
}

func given_detailViewPlacementScenario(t *testing.T, subject *Program) (*gocui.Gui, *gocui.View, int) {
	t.Helper()

	gui := given_headlessGuiWithSize(t, 120, 12)
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

	return gui, detailView, targetLine
}

func given_actionsPopupPlacementScenario(t *testing.T, subject *Program) (*gocui.Gui, *gocui.View, int) {
	t.Helper()

	gui := given_headlessGuiWithSize(t, 120, 12)
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	targetActionIndex := popupView.InnerHeight() + 1
	if targetActionIndex >= len(subject.currentActionsPopupActions()) {
		targetActionIndex = len(subject.currentActionsPopupActions()) - 1
	}
	for range targetActionIndex {
		actualErr = subject.moveActionsPopupSelectionDown(gui, popupView)
		then_noError(t, actualErr)
	}
	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)

	return gui, popupView, subject.currentActionsPopupSelectedRenderedLine()
}

func when_pressingKeySequence(t *testing.T, subject *Program, gui *gocui.Gui, viewName string, view *gocui.View, keys ...any) {
	t.Helper()

	actualView := view
	for _, key := range keys {
		handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewName, key)
		actualErr := handler(gui, actualView)
		then_noError(t, actualErr)
		actualView, actualErr = gui.View(viewName)
		then_noError(t, actualErr)
	}
}

func then_listViewPlacesSelectionAtTheTopOfTheViewport(t *testing.T, view *gocui.View, expectedSelectedIndex int, lineCount int) {
	t.Helper()
	then_listViewPlacesSelectionAtViewportOrigin(t, view, expectedSelectedIndex, expectedViewportTopOrigin(expectedSelectedIndex, view.InnerHeight(), lineCount))
}

func then_listViewPlacesSelectionAtTheCenterOfTheViewport(t *testing.T, view *gocui.View, expectedSelectedIndex int, lineCount int) {
	t.Helper()
	then_listViewPlacesSelectionAtViewportOrigin(t, view, expectedSelectedIndex, centeredViewportOrigin(expectedSelectedIndex, view.InnerHeight(), lineCount))
}

func then_listViewPlacesSelectionAtTheBottomOfTheViewport(t *testing.T, view *gocui.View, expectedSelectedIndex int, lineCount int) {
	t.Helper()
	then_listViewPlacesSelectionAtViewportOrigin(t, view, expectedSelectedIndex, expectedViewportBottomOrigin(expectedSelectedIndex, view.InnerHeight(), lineCount))
}

func then_detailViewPlacesCursorAtTheTopOfTheViewport(t *testing.T, view *gocui.View, expectedLine int, lineCount int) {
	t.Helper()
	then_detailViewPlacesCursorAtViewportOrigin(t, view, expectedLine, expectedViewportTopOrigin(expectedLine, view.InnerHeight(), lineCount))
}

func then_detailViewPlacesCursorAtTheCenterOfTheViewport(t *testing.T, view *gocui.View, expectedLine int, lineCount int) {
	t.Helper()
	then_detailViewPlacesCursorAtViewportOrigin(t, view, expectedLine, centeredViewportOrigin(expectedLine, view.InnerHeight(), lineCount))
}

func then_detailViewPlacesCursorAtTheBottomOfTheViewport(t *testing.T, view *gocui.View, expectedLine int, lineCount int) {
	t.Helper()
	then_detailViewPlacesCursorAtViewportOrigin(t, view, expectedLine, expectedViewportBottomOrigin(expectedLine, view.InnerHeight(), lineCount))
}

func then_listViewPlacesSelectionAtViewportOrigin(t *testing.T, view *gocui.View, expectedSelectedIndex int, expectedOriginY int) {
	t.Helper()

	_, actualOriginY := view.Origin()
	_, actualCursorY := view.Cursor()
	expectedCursorY := expectedSelectedIndex - expectedOriginY
	if actualOriginY != expectedOriginY {
		t.Fatalf("expected origin y %d, actual %d", expectedOriginY, actualOriginY)
	}
	if actualCursorY != expectedCursorY {
		t.Fatalf("expected cursor y %d, actual %d", expectedCursorY, actualCursorY)
	}
}

func then_detailViewPlacesCursorAtViewportOrigin(t *testing.T, view *gocui.View, expectedLine int, expectedOriginY int) {
	t.Helper()

	_, actualOriginY := view.Origin()
	_, actualCursorY := view.Cursor()
	expectedCursorY := expectedLine - expectedOriginY
	if actualOriginY != expectedOriginY {
		t.Fatalf("expected detail origin y %d, actual %d", expectedOriginY, actualOriginY)
	}
	if actualCursorY != expectedCursorY {
		t.Fatalf("expected detail cursor y %d, actual %d", expectedCursorY, actualCursorY)
	}
}

func expectedViewportTopOrigin(selectedIndex int, visibleHeight int, lineCount int) int {
	return clampInt(selectedIndex, 0, maxInt(0, lineCount-maxInt(1, visibleHeight)))
}

func expectedViewportBottomOrigin(selectedIndex int, visibleHeight int, lineCount int) int {
	visibleHeight = maxInt(1, visibleHeight)
	return clampInt(selectedIndex-(visibleHeight-1), 0, maxInt(0, lineCount-visibleHeight))
}
