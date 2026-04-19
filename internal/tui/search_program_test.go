package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingSearchBindings_ThenSlashOpensSearchFromAnyMainPaneAndSearchViewSupportsSubmitAndEscapeShortcuts(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: '/', handler: subject.openSearch})
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyEnter, handler: subject.submitSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyCtrlJ, handler: subject.submitSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyEsc, handler: subject.cancelSearch})
}

func TestSearchPrompt_GivenRenderedProgram_WhenOpeningSearch_ThenThePromptRendersAsABorderlessSlashAtTheBottom(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewSearchName)
	if !gui.Cursor {
		t.Fatal("expected the cursor to be visible while search input is focused")
	}

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	if searchView.Frame {
		t.Fatal("expected the search prompt to be borderless")
	}
	if searchView.Title != "" {
		t.Fatalf("expected an empty search title, actual %q", searchView.Title)
	}
	if actual := strings.TrimSpace(searchView.Buffer()); actual != "/" {
		t.Fatalf("expected search buffer %q, actual %q", "/", actual)
	}

	if searchView.InnerHeight() != 1 {
		t.Fatalf("expected the search prompt to expose one visible content row, actual %d", searchView.InnerHeight())
	}
	if searchView.InnerWidth() != 120 {
		t.Fatalf("expected the search prompt to span the full width, actual %d", searchView.InnerWidth())
	}

	_, _, _, detailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	if detailY1 != 28 {
		t.Fatalf("expected detail view to stop above the bottom prompt at y=%d, actual y=%d", 28, detailY1)
	}
}

func TestSearchPrompt_GivenOpenSearch_WhenHandlingNavigationActions_ThenUnderlyingStateDoesNotChange(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)

	actualErr = subject.nextSideView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusPullRequestsView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, searchView)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewSearchName)
	if subject.model.Focus() != FocusUserView {
		t.Fatalf("expected focus %v, actual %v", FocusUserView, subject.model.Focus())
	}
	if subject.model.SelectedUserIndex() != 0 {
		t.Fatalf("expected selected user index 0, actual %d", subject.model.SelectedUserIndex())
	}
}

func TestSearchPrompt_GivenPreviouslyAppliedQuery_WhenOpeningSearchAgain_ThenItStartsEmpty(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	actualHandled := subject.editSearch(searchView, 0, '1', gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected typing to be handled")
	}

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewSearchName, gocui.KeyCtrlJ)
	actualErr = actualHandler(gui, searchView)
	then_noError(t, actualErr)

	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr = gui.View(viewSearchName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(searchView.Buffer()); actual != "/" {
		t.Fatalf("expected search buffer %q, actual %q", "/", actual)
	}
	if subject.model.SearchDraft() != "" {
		t.Fatalf("expected an empty search draft, actual %q", subject.model.SearchDraft())
	}

	actualErr = subject.cancelSearch(gui, nil)
	then_noError(t, actualErr)
	if subject.model.UserSearchQuery() != "1" {
		t.Fatalf("expected applied query %q after canceling, actual %q", "1", subject.model.UserSearchQuery())
	}
}

func TestSearchPrompt_GivenOpenSearch_WhenSubmittingWithControlJ_ThenItClosesThePromptAndAppliesTheQuery(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	actualHandled := subject.editSearch(searchView, 0, '1', gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected typing to be handled")
	}

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewSearchName, gocui.KeyCtrlJ)
	actualErr = actualHandler(gui, searchView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
	then_viewDoesNotExist(t, gui, viewSearchName)
	if subject.model.UserSearchQuery() != "1" {
		t.Fatalf("expected applied query %q, actual %q", "1", subject.model.UserSearchQuery())
	}
}

func given_handlerForBinding(t *testing.T, specs []keybindingSpec, expectedView string, expectedKey any) func(*gocui.Gui, *gocui.View) error {
	t.Helper()

	for _, spec := range specs {
		if spec.viewName == expectedView && spec.key == expectedKey {
			return spec.handler
		}
	}

	t.Fatalf("expected binding for view %q and key %v", expectedView, expectedKey)
	return nil
}
