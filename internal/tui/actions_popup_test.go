package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestKeybindingSpecs_GivenProgram_WhenListingActionsPopupBindings_ThenAOpensItOnlyInPullRequestContextsAndThePopupSupportsSearchNavigationAndClose(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'a', handler: subject.openActionsPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'a', handler: subject.openActionsPopup})
	then_bindingDoesNotExist(t, actual, viewUserName, 'a')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: '/', handler: subject.focusActionsPopupSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'j', handler: subject.moveActionsPopupSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyArrowDown, handler: subject.moveActionsPopupSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'k', handler: subject.moveActionsPopupSelectionUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyArrowUp, handler: subject.moveActionsPopupSelectionUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyEnter, handler: subject.executeSelectedActionsPopupAction})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyEsc, handler: subject.closeActionsPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyEnter, handler: subject.focusActionsPopupList})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyEsc, handler: subject.closeActionsPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyTab, handler: subject.focusActionsPopupList})
}

func TestActionsPopup_GivenPullRequestsView_WhenOpening_ThenItShowsAllRequestedPullRequestActionsAndTakesFocus(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewActionsPopupName)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "Actions") {
		t.Fatalf("expected popup title to contain %q, actual %q", "Actions", popupView.Title)
	}
	for _, expected := range []string{
		"Comment on PR",
		"Yank URL to clipboard",
		"Open PR in browser",
		"Start review",
		"Review: Approve PR",
		"Review: Comment on PR",
		"Review: Request changes",
		"Edit PR title",
		"Edit PR description",
	} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	if !strings.Contains(popupView.Buffer(), "9 of 9 actions") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "9 of 9 actions", popupView.Buffer())
	}

	then_viewDoesNotExist(t, gui, viewActionsPopupSearchName)
}

func TestActionsPopup_GivenConnectedUserDetail_WhenOpening_ThenItDoesNothing(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_viewDoesNotExist(t, gui, viewActionsPopupSearchName)
}

func TestActionsPopup_GivenOpenPopup_WhenStartingSearchAndTyping_ThenItShowsABorderlessBottomPromptAndFiltersTheActionsLive(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !gui.Cursor {
		t.Fatal("expected the cursor to be visible while the popup search input is focused")
	}
	then_viewDoesNotExist(t, gui, viewSearchName)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	if searchView.Frame {
		t.Fatal("expected the popup search prompt to be borderless")
	}
	if searchView.Title != "" {
		t.Fatalf("expected an empty popup search title, actual %q", searchView.Title)
	}
	if actual := strings.TrimSpace(searchView.Buffer()); actual != "/" {
		t.Fatalf("expected popup search buffer %q, actual %q", "/", actual)
	}
	if searchView.InnerHeight() != 1 {
		t.Fatalf("expected the popup search prompt to expose one visible content row, actual %d", searchView.InnerHeight())
	}
	if searchView.InnerWidth() != 120 {
		t.Fatalf("expected the popup search prompt to span the full width, actual %d", searchView.InnerWidth())
	}

	_, _, _, detailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	if detailY1 != 28 {
		t.Fatalf("expected detail view to stop above the popup prompt at y=%d, actual y=%d", 28, detailY1)
	}

	for _, ch := range "clipboard" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "1 of 9 actions") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "1 of 9 actions", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), "Yank URL to clipboard") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "Yank URL to clipboard", popupView.Buffer())
	}
	for _, unexpected := range []string{"Comment on PR", "Edit PR title"} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected popup buffer to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}
func TestActionsPopup_GivenKeywordSearch_WhenFiltering_ThenItCanFindReviewAndEditActions(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "lgtm" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Review: Approve PR") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "Review: Approve PR", popupView.Buffer())
	}
	if strings.Contains(popupView.Buffer(), "Yank URL to clipboard") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "Yank URL to clipboard", popupView.Buffer())
	}

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr = gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "rename" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Edit PR title") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "Edit PR title", popupView.Buffer())
	}
}

func TestActionsPopup_GivenStartReviewActionSelected_WhenGitHubRefusesToOpenThePendingReview_ThenItKeepsThePopupOpenAndShowsTheFailure(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewErr: errors.New("review refused")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("start review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "start review"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "review refused") {
		t.Fatalf("expected popup title to contain %q, actual %q", "review refused", popupView.Title)
	}
	if subject.reviewSession.active {
		t.Fatal("expected review mode to stay inactive after the error")
	}
}

func TestActionsPopup_GivenTitleSearchOnTheSelectedRow_WhenFiltering_ThenItKeepsSearchBackgroundOnTheMatchAndSelectionBackgroundElsewhere(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "approve" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewActionsPopupName, 2, "Approve")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewActionsPopupName, 2, "Review: ")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewActionsPopupName, 2, "Approve")
	then_viewLineSegmentIsBold(t, gui, viewActionsPopupName, 2, "Approve")
	then_viewLineSegmentIsBold(t, gui, viewActionsPopupName, 2, "Review: ")
}

func TestActionsPopupSearch_GivenFilteredResults_WhenPressingEnter_ThenItStopsSearchingWithoutExecutingTheAction(t *testing.T) {
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "clipboard" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	actualErr = actualHandler(gui, searchView)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	then_viewDoesNotExist(t, gui, viewActionsPopupSearchName)
	if subject.model.ActionsPopupSearchActive() {
		t.Fatal("expected the popup search to stop")
	}
	if subject.model.ActionsPopupSelectedActionIndex() != 1 {
		t.Fatalf("expected selected action index 1, actual %d", subject.model.ActionsPopupSelectedActionIndex())
	}
	if len(clipboardWriter.writes) != 0 {
		t.Fatalf("expected no clipboard writes, actual %v", clipboardWriter.writes)
	}
}

func TestActionsPopupSearch_GivenAppliedViewSearches_WhenStartingThePopupSearch_ThenItClearsTheOtherViewHighlights(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.userSearchQuery = "dummy"
	model.pullRequestSearchQueries[MyPullRequestsTab] = "widgets"
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	then_viewExists(t, gui, viewUserFooterName)
	then_viewExists(t, gui, viewPullRequestsFooterName)

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	if subject.model.UserSearchQuery() != "" {
		t.Fatalf("expected the user search query to be cleared, actual %q", subject.model.UserSearchQuery())
	}
	if subject.model.PullRequestSearchQuery(MyPullRequestsTab) != "" {
		t.Fatalf("expected the pull request search query to be cleared, actual %q", subject.model.PullRequestSearchQuery(MyPullRequestsTab))
	}
	then_viewDoesNotExist(t, gui, viewUserFooterName)
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}

func TestActionsPopup_GivenExistingFilter_WhenStartingANewSearch_ThenItClearsThePreviousPromptText(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "clip" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	actualErr = subject.focusActionsPopupList(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr = gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(searchView.Buffer()); actual != "/" {
		t.Fatalf("expected popup search buffer %q, actual %q", "/", actual)
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "9 of 9 actions") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "9 of 9 actions", popupView.Buffer())
	}
}

func TestActionsPopup_GivenFocusedSearchRow_WhenPressingTab_ThenItReturnsToTheActionList(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)

	actualErr = subject.focusActionsPopupList(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewActionsPopupName)
	if subject.model.ActionsPopupSearchActive() {
		t.Fatal("expected the popup search to be unfocused")
	}
}

func TestActionsPopup_GivenFilteredActions_WhenHandlingArrowBindings_ThenTheyFollowTheVisibleResults(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	reviewIndexes := matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "review")
	subject.model.UpdateActionsPopupSearch("review", reviewIndexes)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyArrowDown)
	actualErr = downHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.ActionsPopupSelectedActionIndex() != reviewIndexes[1] {
		t.Fatalf("expected selected action index %d, actual %d", reviewIndexes[1], subject.model.ActionsPopupSelectedActionIndex())
	}

	upHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyArrowUp)
	actualErr = upHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.ActionsPopupSelectedActionIndex() != reviewIndexes[0] {
		t.Fatalf("expected selected action index %d, actual %d", reviewIndexes[0], subject.model.ActionsPopupSelectedActionIndex())
	}
}

func TestActionsPopup_GivenCommentActionSelected_WhenExecuting_ThenItReusesTheCommentComposer(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewModalEditorName)
	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Title, pullRequestCommentComposerTitle) {
		t.Fatalf("expected composer title to contain %q, actual %q", pullRequestCommentComposerTitle, composerView.Title)
	}
}

func TestActionsPopup_GivenYankActionSelected_WhenExecuting_ThenItReusesTheCopyPathAndClosesThePopup(t *testing.T) {
	model := given_pullRequestCommentModel()
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.MoveActionsPopupSelectionDown()
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	if len(clipboardWriter.writes) != 1 || clipboardWriter.writes[0] != "https://github.com/acme/widgets/pull/42" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"https://github.com/acme/widgets/pull/42"}, clipboardWriter.writes)
	}
	pullRequestsFooterView, actualErr := gui.View("pull-requests-footer")
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsFooterView.Buffer(), yankSuccessMessage) {
		t.Fatalf("expected pull request footer to contain %q, actual %q", yankSuccessMessage, pullRequestsFooterView.Buffer())
	}
}

func TestHelpPopup_GivenPullRequestContext_WhenTogglingHelp_ThenItListsTheActionsShortcut(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	if !strings.Contains(helpView.Buffer(), "PR actions") {
		t.Fatalf("expected help buffer to contain %q, actual %q", "PR actions", helpView.Buffer())
	}
}

func TestActionsPopup_GivenDetailFocus_WhenClosing_ThenItReturnsToTheDetailPaneCleanly(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.closeActionsPopup(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_viewDoesNotExist(t, gui, viewActionsPopupSearchName)
}

func given_actionsPopupPullRequest() githubcli.PullRequest {
	return githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Original body"}
}
