package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingSearchBindings_ThenSlashOpensSearchFromAnyMainPaneAndSearchViewSupportsEnterAndEscape(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: '/', handler: subject.openSearch})
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyEnter, handler: subject.submitSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyEsc, handler: subject.cancelSearch})
}

func TestSearchPrompt_GivenRenderedProgram_WhenOpeningSearch_ThenThePromptTakesFocus(t *testing.T) {
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
	if !strings.Contains(searchView.Title, "Search connected user") {
		t.Fatalf("expected search title to mention the connected user, actual %q", searchView.Title)
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
