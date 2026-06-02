package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingDetailSearchFollowBindings_ThenTheDetailAndPullRequestsViewsExposeTheirOwnHandlers(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'n', handler: subject.nextDetailSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'N', handler: subject.previousDetailSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'n', handler: subject.nextPullRequestsSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'N', handler: subject.previousPullRequestsSearchMatch})
	then_bindingDoesNotExist(t, actual, viewSearchName, 'n')
	then_bindingDoesNotExist(t, actual, viewSearchName, 'N')
}

func TestSubmitSearch_GivenDetailSearchMatchAfterTheCurrentCursor_WhenSubmitting_ThenItMovesTheCursorAndViewportToTheMatch(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: given_detailSearchBody(2, 12)}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGuiWithSize(t, 80, 10)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	for range 10 {
		actualErr = subject.moveSelectionDown(gui, detailView)
		then_noError(t, actualErr)
	}
	expectedViewportHeight := detailView.InnerHeight()
	expectedCurrentOriginRow := subject.detailState.viewState.originRow

	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	when_typingSearchQuery(t, subject, searchView, "Alpha")

	actualErr = subject.submitSearch(gui, searchView)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	originX, _ := detailView.Origin()
	cursorX, cursorY := detailView.Cursor()
	actualOriginRow := subject.detailState.viewState.originRow
	expectedTargetRow := 15
	expectedOriginRow := visibleViewportOrigin(expectedTargetRow, expectedCurrentOriginRow, expectedViewportHeight, 20)
	if originX != 0 || actualOriginRow != expectedOriginRow {
		t.Fatalf("expected detail origin 0,%d after following the submitted search, actual %d,%d", expectedOriginRow, originX, actualOriginRow)
	}
	if cursorX != 0 || cursorY != expectedTargetRow-expectedOriginRow {
		t.Fatalf("expected detail cursor 0,%d after following the submitted search, actual %d,%d", expectedTargetRow-expectedOriginRow, cursorX, cursorY)
	}
}

func TestDetailSearchFollow_GivenNoActiveDetailSearch_WhenRepeatingForwardOrBackward_ThenItLeavesTheCursorAndViewportUnchanged(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: strings.Join([]string{"one", "two", "three", "four"}, "\n")}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, detailView)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, detailView)
	then_noError(t, actualErr)

	expectedOriginX, _ := detailView.Origin()
	expectedOriginRow := subject.detailState.viewState.originRow
	expectedCursorX, expectedCursorY := detailView.Cursor()

	actualErr = subject.nextDetailSearchMatch(gui, detailView)
	then_noError(t, actualErr)
	actualErr = subject.previousDetailSearchMatch(gui, detailView)
	then_noError(t, actualErr)

	actualOriginX, _ := detailView.Origin()
	actualOriginRow := subject.detailState.viewState.originRow
	actualCursorX, actualCursorY := detailView.Cursor()
	if actualOriginX != expectedOriginX || actualOriginRow != expectedOriginRow {
		t.Fatalf("expected detail origin %d,%d without an active search, actual %d,%d", expectedOriginX, expectedOriginRow, actualOriginX, actualOriginRow)
	}
	if actualCursorX != expectedCursorX || actualCursorY != expectedCursorY {
		t.Fatalf("expected detail cursor %d,%d without an active search, actual %d,%d", expectedCursorX, expectedCursorY, actualCursorX, actualCursorY)
	}
}

func given_detailSearchBody(matchRows ...int) string {
	rows := make([]string, 0, 20)
	for rowIndex := range 20 {
		rowText := fmt.Sprintf("detail line %02d", rowIndex)
		if indexOfInt(matchRows, rowIndex) >= 0 {
			rowText = "Alpha " + rowText
		}
		rows = append(rows, rowText)
	}

	return strings.Join(rows, "\n")
}

func when_typingSearchQuery(t *testing.T, subject *Program, view *gocui.View, query string) {
	t.Helper()

	for _, character := range query {
		actualHandled := subject.editSearch(view, 0, character, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(character))
		}
	}
}
