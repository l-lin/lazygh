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
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyCtrlS, handler: subject.submitSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewSearchName, key: gocui.KeyEsc, handler: subject.cancelSearch})
}

func TestSearchPrompt_GivenPullRequestsFocus_WhenOpeningSearch_ThenThePromptUsesTheGlobalStatusLineWithoutShrinkingTheLayout(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	_, _, _, pullRequestsY1, actualErr := gui.ViewPosition(viewPullRequestsName)
	then_noError(t, actualErr)
	statusX0, statusY0, statusX1, statusY1, actualErr := gui.ViewPosition(viewStatusLineName)
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

	searchX0, searchY0, searchX1, searchY1, actualErr := gui.ViewPosition(viewSearchName)
	then_noError(t, actualErr)
	if searchX0 != statusX0 || searchX1 != statusX1 || searchY0 != statusY0 || searchY1 != statusY1 {
		t.Fatalf("expected the search prompt to reuse the status line, actual=(%d,%d)-(%d,%d) expected=(%d,%d)-(%d,%d)", searchX0, searchY0, searchX1, searchY1, statusX0, statusY0, statusX1, statusY1)
	}

	_, _, _, actualPullRequestsY1, actualErr := gui.ViewPosition(viewPullRequestsName)
	then_noError(t, actualErr)
	if actualPullRequestsY1 != pullRequestsY1 {
		t.Fatalf("expected the pull requests pane height to stay unchanged, actual y1=%d expected y1=%d", actualPullRequestsY1, pullRequestsY1)
	}
}

func TestSearchPrompt_GivenDetailFocus_WhenOpeningSearch_ThenThePromptUsesTheGlobalStatusLineWithoutShrinkingTheLayout(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	_, _, _, detailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	statusX0, statusY0, statusX1, statusY1, actualErr := gui.ViewPosition(viewStatusLineName)
	then_noError(t, actualErr)

	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewSearchName)

	searchX0, searchY0, searchX1, searchY1, actualErr := gui.ViewPosition(viewSearchName)
	then_noError(t, actualErr)
	if searchX0 != statusX0 || searchX1 != statusX1 || searchY0 != statusY0 || searchY1 != statusY1 {
		t.Fatalf("expected the search prompt to reuse the status line, actual=(%d,%d)-(%d,%d) expected=(%d,%d)-(%d,%d)", searchX0, searchY0, searchX1, searchY1, statusX0, statusY0, statusX1, statusY1)
	}

	_, _, _, actualDetailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	if actualDetailY1 != detailY1 {
		t.Fatalf("expected the detail pane height to stay unchanged, actual y1=%d expected y1=%d", actualDetailY1, detailY1)
	}
}

func TestSearchFooter_GivenSubmittedPullRequestsSearch_WhenRendering_ThenTheAppliedQueryMovesToThePaneFooter(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.StartSearch()
	model.UpdateSearchDraft("2")
	model.SubmitSearch()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	footerView, actualErr := gui.View("pull-requests-footer")
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(footerView.Buffer()); actual != "/2 (1 match)  •  ?: Help, /: Search, a: Action" {
		t.Fatalf("expected pull requests footer %q, actual %q", "/2 (1 match)  •  ?: Help, /: Search, a: Action", actual)
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(pullRequestsView.Title, "/2") || strings.Contains(pullRequestsView.TitlePrefix, "/2") {
		t.Fatalf("expected the pull requests title to stay stable, actual title=%q prefix=%q", pullRequestsView.Title, pullRequestsView.TitlePrefix)
	}
}

func TestSearchFooter_GivenSubmittedDetailSearch_WhenRendering_ThenTheAppliedQueryMovesToThePaneFooter(t *testing.T) {
	model := NewModel(SeedData{
		Users: []Item{{Title: "dummy-user-1", Detail: "Alpha detail line"}},
	})
	model.OpenDetail()
	model.StartSearch()
	model.UpdateSearchDraft("Alpha")
	model.SubmitSearch()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	footerView, actualErr := gui.View("detail-footer")
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(footerView.Buffer()); actual != "/Alpha (1 match)  •  ?: Help, /: Search" {
		t.Fatalf("expected detail footer %q, actual %q", "/Alpha (1 match)  •  ?: Help, /: Search", actual)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Title, "/Alpha") || strings.Contains(detailView.TitlePrefix, "/Alpha") {
		t.Fatalf("expected the detail title to stay stable, actual title=%q prefix=%q", detailView.Title, detailView.TitlePrefix)
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

func TestSearchPrompt_GivenOpenSearch_WhenSubmittingWithControlS_ThenItClosesThePromptAndAppliesTheQuery(t *testing.T) {
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

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewSearchName, gocui.KeyCtrlS)
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
