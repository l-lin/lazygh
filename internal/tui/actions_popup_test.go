package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestKeybindingSpecs_GivenProgram_WhenListingActionsPopupBindings_ThenAOpensItFromEveryMainViewAndThePopupSupportsSearchNavigationAndClose(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingKeyExists(t, actual, viewUserName, 'a')
	then_bindingKeyExists(t, actual, viewPullRequestsName, 'a')
	then_bindingKeyExists(t, actual, viewDetailName, 'a')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: '/', handler: subject.focusActionsPopupSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'j', handler: subject.moveActionsPopupSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyArrowDown, handler: subject.moveActionsPopupSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'k', handler: subject.moveActionsPopupSelectionUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyArrowUp, handler: subject.moveActionsPopupSelectionUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'n', handler: subject.nextActionsPopupSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'N', handler: subject.previousActionsPopupSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyEnter, handler: subject.executeSelectedActionsPopupAction})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyAltEnter, handler: subject.submitSelectedActionsPopupAction})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyEsc, handler: subject.closeActionsPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'q', handler: subject.closeActionsPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyCtrlN, handler: subject.moveActionsPopupSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyArrowDown, handler: subject.moveActionsPopupSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyCtrlP, handler: subject.moveActionsPopupSelectionUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyArrowUp, handler: subject.moveActionsPopupSelectionUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyEnter, handler: subject.executeSelectedActionsPopupAction})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyAltEnter, handler: subject.submitSelectedActionsPopupAction})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupSearchName, key: gocui.KeyEsc, handler: subject.closeActionsPopup})
	then_bindingDoesNotExist(t, actual, viewActionsPopupSearchName, 'q')
	then_bindingDoesNotExist(t, actual, viewActionsPopupSearchName, gocui.KeyTab)
}

func TestActionsPopup_GivenPullRequestsView_WhenOpening_ThenItShowsGroupedReviewPullRequestAndThemeActionsAndTakesFocus(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "Actions") {
		t.Fatalf("expected popup title to contain %q, actual %q", "Actions", popupView.Title)
	}
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"Review",
		actionsPopupLabel(actionsPopupStartReviewIcon, "Start review"),
		actionsPopupLabel(actionsPopupReviewStoryIcon, "Start review as story"),
		actionsPopupLabel(actionsPopupReviewApproveIcon, "Review: Approve PR"),
		actionsPopupLabel(actionsPopupReviewCommentIcon, "Review: Comment on PR"),
		actionsPopupLabel(actionsPopupReviewRequestChangesIcon, "Review: Request changes"),
		"Pull request",
		actionsPopupLabel(actionsPopupYankPullRequestURLIcon, "Yank URL to clipboard"),
		actionsPopupLabel(actionsPopupOpenPullRequestBrowserIcon, "Open PR in browser"),
		actionsPopupLabel(actionsPopupOpenPullRequestByURLIcon, "Open PR from URL"),
		actionsPopupLabel(actionsPopupCustomSearchIcon, "Custom search"),
		actionsPopupLabel(actionsPopupCommentOnPullRequestIcon, "Comment on PR"),
		actionsPopupLabel(actionsPopupEditPullRequestIcon, "Edit PR title"),
		actionsPopupLabel(actionsPopupEditPullRequestIcon, "Edit PR description"),
		"Theme",
		actionsPopupLabel(actionsPopupChangeThemeIcon, themePickerActionTitle),
		"Cache",
		actionsPopupLabel(actionsPopupRefreshPullRequestIcon, pullRequestRefreshActionTitle),
		actionsPopupLabel(actionsPopupRefreshPullRequestIcon, pullRequestListRefreshActionTitle),
	})
	if strings.Contains(popupView.Buffer(), "Review PR from URL") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "Review PR from URL", popupView.Buffer())
	}
	if strings.Contains(popupView.Buffer(), "12 of 12 actions") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "12 of 12 actions", popupView.Buffer())
	}
	if popupView.Footer != "" {
		t.Fatalf("expected popup footer to stay empty without a search query, actual %q", popupView.Footer)
	}
	_, actualCursorY := popupView.Cursor()
	if actualCursorY != 1 {
		t.Fatalf("expected the first selectable action to start below the header, actual cursor row %d", actualCursorY)
	}

	then_viewExists(t, gui, viewActionsPopupSearchName)
}

func TestActionsPopup_GivenPullRequestLevelReactionAction_WhenOpening_ThenItKeepsAddReactionInsideThePullRequestGroupAfterReview(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {ID: "PR_kwDOA", Title: "First PR", Number: 42, Body: "Body 42", State: "OPEN"},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	buffer := popupView.Buffer()
	reviewHeaderCount := 0
	pullRequestHeaderCount := 0
	for _, line := range strings.Split(strings.TrimSpace(buffer), "\n") {
		switch strings.TrimSpace(line) {
		case actionsPopupGroupReview:
			reviewHeaderCount++
		case actionsPopupGroupPullRequest:
			pullRequestHeaderCount++
		}
	}
	if reviewHeaderCount != 1 {
		t.Fatalf("expected exactly one %q header, actual %d in %q", actionsPopupGroupReview, reviewHeaderCount, buffer)
	}
	if pullRequestHeaderCount != 1 {
		t.Fatalf("expected exactly one %q header, actual %d in %q", actionsPopupGroupPullRequest, pullRequestHeaderCount, buffer)
	}
	reviewHeaderIndex := strings.Index(buffer, actionsPopupGroupReview)
	pullRequestHeaderIndex := strings.Index(buffer, actionsPopupGroupPullRequest)
	addReactionIndex := strings.Index(buffer, reactionPickerTitle)
	themeHeaderIndex := strings.Index(buffer, actionsPopupGroupTheme)
	if reviewHeaderIndex < 0 || pullRequestHeaderIndex < 0 || addReactionIndex < 0 || themeHeaderIndex < 0 {
		t.Fatalf("expected popup buffer to contain %q, %q, %q, and %q, actual %q", actionsPopupGroupReview, actionsPopupGroupPullRequest, reactionPickerTitle, actionsPopupGroupTheme, buffer)
	}
	if !(reviewHeaderIndex < pullRequestHeaderIndex && pullRequestHeaderIndex < addReactionIndex && addReactionIndex < themeHeaderIndex) {
		t.Fatalf("expected %q to stay inside the %q group after %q and before %q, actual %q", reactionPickerTitle, actionsPopupGroupPullRequest, actionsPopupGroupReview, actionsPopupGroupTheme, buffer)
	}
}

func TestActionsPopup_GivenConnectedUserDetail_WhenOpening_ThenItShowsTheGlobalActionsAndTakesFocus(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.pullRequestCache = &fakePersistentPullRequestCache{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"Theme",
		actionsPopupLabel(actionsPopupChangeThemeIcon, themePickerActionTitle),
		"Cache",
		actionsPopupLabel(actionsPopupRefreshPullRequestIcon, pullRequestListRefreshActionTitle),
		actionsPopupLabel(iconDelete, "Clear cache"),
	})
	then_viewExists(t, gui, viewActionsPopupSearchName)
}

func TestActionsPopup_GivenOpenPopup_WhenStartingSearchAndTyping_ThenItShowsABorderlessInlinePromptAndFiltersTheMatchingActionsLive(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
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
	if actual := strings.TrimSpace(searchView.Buffer()); actual != ">" {
		t.Fatalf("expected popup search buffer %q, actual %q", ">", actual)
	}
	if searchView.InnerHeight() != 1 {
		t.Fatalf("expected the popup search prompt to expose one visible content row, actual %d", searchView.InnerHeight())
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	searchX0, searchY0, searchX1, _, actualErr := gui.ViewPosition(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	popupX0, popupY0, popupX1, _, actualErr := gui.ViewPosition(viewActionsPopupName)
	then_noError(t, actualErr)
	if searchX0 != popupX0 || searchX1 != popupX1 {
		t.Fatalf("expected the popup search prompt to share the action list width, search=(%d,%d) popup=(%d,%d)", searchX0, searchX1, popupX0, popupX1)
	}
	if searchY0 >= popupY0 {
		t.Fatalf("expected the popup search prompt to stay above the action list, search y=%d popup y=%d", searchY0, popupY0)
	}

	for _, ch := range "clipboard" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "1 of 15 actions") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "1 of 15 actions", popupView.Buffer())
	}
	popupChromeView, actualErr := gui.View(viewActionsPopupChromeName)
	then_noError(t, actualErr)
	if popupChromeView.Footer != "1 of 15 actions" {
		t.Fatalf("expected popup footer %q, actual %q", "1 of 15 actions", popupChromeView.Footer)
	}
	if !strings.Contains(popupView.Buffer(), "Yank URL to clipboard") {
		t.Fatalf("expected popup buffer to keep %q visible, actual %q", "Yank URL to clipboard", popupView.Buffer())
	}
	for _, unexpected := range []string{"Comment on PR", "Edit PR title"} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected popup buffer to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}
func TestActionsPopup_GivenTitleSearch_WhenHighlighting_ThenItCanStillJumpToReviewAndEditActionsWithoutHidingTheOthers(t *testing.T) {
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
	expectedReviewIndexes := matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestReviewApprovalTitle)
	for _, ch := range pullRequestReviewApprovalTitle {
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
	if len(expectedReviewIndexes) != 1 || subject.model.ActionsPopupSelectedActionIndex() != expectedReviewIndexes[0] {
		t.Fatalf("expected selected action index %v, actual %d", expectedReviewIndexes, subject.model.ActionsPopupSelectedActionIndex())
	}

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr = gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	expectedEditIndexes := matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestTitleEditorTitle)
	for _, ch := range pullRequestTitleEditorTitle {
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
	if strings.Contains(popupView.Buffer(), "Yank URL to clipboard") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "Yank URL to clipboard", popupView.Buffer())
	}
	if len(expectedEditIndexes) == 0 || subject.model.ActionsPopupSelectedActionIndex() != expectedEditIndexes[0] {
		t.Fatalf("expected selected action index to jump to %v, actual %d", expectedEditIndexes, subject.model.ActionsPopupSelectedActionIndex())
	}
}

func TestActionsPopup_GivenSearchQuery_WhenRendering_ThenTheMatchUsesAReadableThemeForeground(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.ResolvePaletteWithPreset("catppuccin-frappe", theme.Palette{}))

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
	for _, ch := range "theme" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	matchLineIndex := given_viewLineIndexContaining(t, popupView, themePickerActionTitle)
	then_viewLineSegmentHasForegroundColor(t, gui, viewActionsPopupName, matchLineIndex, "theme", given_themeColorHex(t, theme.BackgroundHex), "search match readable foreground")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewActionsPopupName, matchLineIndex, "theme", given_themeColorHex(t, theme.SearchHighlightHex), "search match background")
}

func TestActionsPopup_GivenStartReviewActionSelected_WhenGitHubRefusesToOpenThePendingReview_ThenItKeepsThePopupOpenAndShowsATransientErrorPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewErr: errors.New("review refused")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = &capturingAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("review mode", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "review mode"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !reflect.DeepEqual(loader.startReviewCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected start review calls %v, actual %v", []string{"acme/widgets#42"}, loader.startReviewCalls)
	}
	if subject.actionsPopupErrorMessage != "" {
		t.Fatalf("expected popup error message to stay empty, actual %q", subject.actionsPopupErrorMessage)
	}
	if !subject.transientErrorPopupVisible() {
		t.Fatal("expected the transient error popup state to stay visible")
	}
	then_transientErrorPopupContains(t, gui, "review refused")
	if subject.reviewSession.active {
		t.Fatal("expected review mode to stay inactive after the error")
	}
}

func TestActionsPopup_GivenSelectedPullRequestWithoutPendingReview_WhenOpening_ThenItHidesCancelPendingReview(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "Cancel pending review") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "Cancel pending review", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), "Start review") {
		t.Fatalf("expected popup buffer to keep %q visible, actual %q", "Start review", popupView.Buffer())
	}
	if !reflect.DeepEqual(loader.getPendingReviewCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected pending review lookup calls %v, actual %v", []string{"acme/widgets#42"}, loader.getPendingReviewCalls)
	}
}

func TestActionsPopup_GivenSelectedPullRequestWithPendingReviewLoadingInBackground_WhenLookupSucceeds_ThenItRefreshesToShowCancelPendingReview(t *testing.T) {
	loader := &fakePullRequestDetailLoader{reviewKeyByPendingID: map[string]string{"PRR_pending": "acme/widgets#42"}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialQueuedRuns := len(asyncRunner.runs)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != initialQueuedRuns {
		t.Fatalf("expected the popup to reuse the already queued background work, actual queued runs %d from %d", len(asyncRunner.runs), initialQueuedRuns)
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "Cancel pending review") {
		t.Fatalf("expected popup buffer to hide %q before the lookup completes, actual %q", "Cancel pending review", popupView.Buffer())
	}

	queuedRuns := append([]func(){}, asyncRunner.runs...)
	for _, run := range queuedRuns {
		run()
	}

	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	startLineIndex := given_viewLineIndexContaining(t, popupView, "Start review")
	storyLineIndex := given_viewLineIndexContaining(t, popupView, "Start review as story")
	cancelLineIndex := given_viewLineIndexContaining(t, popupView, "Cancel pending review")
	if storyLineIndex != startLineIndex+1 || cancelLineIndex != storyLineIndex+1 {
		t.Fatalf("expected start review and start review as story to stay on top before cancel pending review, actual %q", popupView.Buffer())
	}
	if !reflect.DeepEqual(loader.getPendingReviewCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected pending review lookup calls %v, actual %v", []string{"acme/widgets#42"}, loader.getPendingReviewCalls)
	}
}

func TestActionsPopup_GivenCancelPendingReviewActionSelected_WhenExecuting_ThenItDeletesThePendingReviewInvalidatesTheCachesAndClosesThePopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pendingPullRequestReviewCache["acme/widgets#42"] = pendingPullRequestReviewState{id: "PRR_pending"}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"})}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: reviewDiffData{}}
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("cancel pending review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "cancel pending review"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after canceling the pending review")
	}
	if !reflect.DeepEqual(loader.deletePullRequestReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected deleted pending review ids %v, actual %v", []string{"PRR_pending"}, loader.deletePullRequestReviewIDs)
	}
	if _, ok := subject.pullRequestDetailCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull request detail to be invalidated after canceling the pending review")
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull request diff to be invalidated after canceling the pending review")
	}
	summary, ok := subject.currentPullRequestSummary()
	if !ok {
		t.Fatal("expected the selected pull request summary to stay available")
	}
	pendingState, known := subject.pendingPullRequestReviewStateForSummary(summary)
	if !known {
		t.Fatal("expected the pending review state to stay known after canceling the review")
	}
	if strings.TrimSpace(pendingState.id) != "" {
		t.Fatalf("expected the pending review state to be cleared, actual %+v", pendingState)
	}
	then_statusLineContains(t, gui, "Pending review canceled")
}

func TestActionsPopup_GivenCancelPendingReviewActionSelected_WhenGitHubRejectsCancel_ThenItShowsATransientErrorPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{deletePullRequestReviewErr: errors.New("cancel refused")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pendingPullRequestReviewCache["acme/widgets#42"] = pendingPullRequestReviewState{id: "PRR_pending"}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"})}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: reviewDiffData{}}
	subject.asyncRunner = &capturingAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("cancel pending review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "cancel pending review"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !reflect.DeepEqual(loader.deletePullRequestReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected deleted pending review ids %v, actual %v", []string{"PRR_pending"}, loader.deletePullRequestReviewIDs)
	}
	if subject.actionsPopupErrorMessage != "" {
		t.Fatalf("expected popup error message to stay empty, actual %q", subject.actionsPopupErrorMessage)
	}
	then_transientErrorPopupContains(t, gui, "cancel refused")
	if _, ok := subject.pullRequestDetailCache["acme/widgets#42"]; !ok {
		t.Fatal("expected the cached pull request detail to stay available after the failure")
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; !ok {
		t.Fatal("expected the cached pull request diff to stay available after the failure")
	}
}

func TestActionsPopup_GivenCancelPendingReviewActionSelectedFromRequestedPullRequestList_WhenExecuting_ThenItReloadsTheActivePullRequestList(t *testing.T) {
	loader := &fakePullRequestDetailLoader{reviewKeyByPendingID: map[string]string{"PRR_pending": "acme/widgets#42"}}
	model := given_model()
	model.FocusPullRequestsView()
	subject := given_pullRequestCommentProgram(model, loader)
	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{
		{Label: "My PRs", Command: []string{"search", "prs", "--author", "@me", "--state", "open"}},
		{Label: "Requested", Command: []string{"search", "prs", "--review-requested", "@me", "--state", "open"}},
	})
	requestedSummary := githubcli.PullRequest{Title: "Requested PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42"}
	subject.model.SetPullRequestRows(RequestedPullRequestsTab, []PullRequestRow{requestedPullRequestRow(requestedSummary)})
	subject.model.activePullRequestTab = RequestedPullRequestsTab
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.pendingPullRequestReviewCache["acme/widgets#42"] = pendingPullRequestReviewState{id: "PRR_pending"}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("cancel pending review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "cancel pending review"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.deletePullRequestReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected deleted pending review ids %v, actual %v", []string{"PRR_pending"}, loader.deletePullRequestReviewIDs)
	}
	if !reflect.DeepEqual(loader.listPullRequestCommands, [][]string{{"search", "prs", "--review-requested", "@me", "--state", "open"}}) {
		t.Fatalf("expected pull request reload commands %v, actual %v", [][]string{{"search", "prs", "--review-requested", "@me", "--state", "open"}}, loader.listPullRequestCommands)
	}
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(pullRequestsView.Buffer(), "Requested PR") {
		t.Fatalf("expected the requested pull request list to stop showing %q after canceling the review, actual %q", "Requested PR", pullRequestsView.Buffer())
	}
	if !strings.Contains(pullRequestsView.Buffer(), reviewRequestedPullRequestsEmptyTitle) {
		t.Fatalf("expected the requested pull request list to refresh to %q, actual %q", reviewRequestedPullRequestsEmptyTitle, pullRequestsView.Buffer())
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

	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	matchLineIndex := given_viewLineIndexContaining(t, popupView, "Review: Approve PR")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewActionsPopupName, matchLineIndex, "Approve")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewActionsPopupName, matchLineIndex, "Review: ")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewActionsPopupName, matchLineIndex, "Approve")
	then_viewLineSegmentIsBold(t, gui, viewActionsPopupName, matchLineIndex, "Approve")
	then_viewLineSegmentIsBold(t, gui, viewActionsPopupName, matchLineIndex, "Review: ")
}

func TestActionsPopupSearch_GivenFilteredResults_WhenPressingEnter_ThenItExecutesTheSelectedAction(t *testing.T) {
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

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	expectedIndexes := matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "clipboard")
	for _, ch := range "clipboard" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	actualErr = actualHandler(gui, searchView)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_viewDoesNotExist(t, gui, viewActionsPopupSearchName)
	if len(expectedIndexes) != 1 {
		t.Fatalf("expected one matching action index, actual %v", expectedIndexes)
	}
	if !reflect.DeepEqual(clipboardWriter.writes, []string{"https://github.com/acme/widgets/pull/42"}) {
		t.Fatalf("expected one clipboard write, actual %v", clipboardWriter.writes)
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
	if actual := strings.TrimSpace(searchView.Buffer()); actual != ">" {
		t.Fatalf("expected popup search buffer %q, actual %q", ">", actual)
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "12 of 12 actions") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "12 of 12 actions", popupView.Buffer())
	}
	if popupView.Footer != "" {
		t.Fatalf("expected popup footer to stay empty when the query is cleared, actual %q", popupView.Footer)
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
	actualErr = subject.afterStateChange(gui)
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

func TestActionsPopup_GivenSubmittedSearch_WhenRepeatingWithNAndN_ThenItMovesBetweenMatchingActionsWithoutHidingTheOtherRows(t *testing.T) {
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
	for _, ch := range "edit" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	actualErr = subject.focusActionsPopupList(gui, nil)
	then_noError(t, actualErr)

	expectedIndexes := matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit")
	if len(expectedIndexes) != 2 {
		t.Fatalf("expected 2 matching action indexes, actual %v", expectedIndexes)
	}
	if subject.model.ActionsPopupSelectedActionIndex() != expectedIndexes[0] {
		t.Fatalf("expected selected action index %d, actual %d", expectedIndexes[0], subject.model.ActionsPopupSelectedActionIndex())
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "Start review") {
		t.Fatalf("expected the non-matching actions to be hidden, actual %q", popupView.Buffer())
	}

	nextHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'n')
	previousHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'N')

	actualErr = nextHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.ActionsPopupSelectedActionIndex() != expectedIndexes[1] {
		t.Fatalf("expected selected action index %d, actual %d", expectedIndexes[1], subject.model.ActionsPopupSelectedActionIndex())
	}

	actualErr = nextHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.ActionsPopupSelectedActionIndex() != expectedIndexes[0] {
		t.Fatalf("expected selected action index %d, actual %d", expectedIndexes[0], subject.model.ActionsPopupSelectedActionIndex())
	}

	actualErr = previousHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.ActionsPopupSelectedActionIndex() != expectedIndexes[1] {
		t.Fatalf("expected selected action index %d, actual %d", expectedIndexes[1], subject.model.ActionsPopupSelectedActionIndex())
	}
}

func TestActionsPopup_GivenFilteredActions_WhenPressingGGOrG_ThenItMovesToTheFirstOrLastVisibleResult(t *testing.T) {
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
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	bottomHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'G')
	actualErr = bottomHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.ActionsPopupSelectedActionIndex() != reviewIndexes[len(reviewIndexes)-1] {
		t.Fatalf("expected selected action index %d, actual %d", reviewIndexes[len(reviewIndexes)-1], subject.model.ActionsPopupSelectedActionIndex())
	}

	topHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'g')
	actualErr = topHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.ActionsPopupSelectedActionIndex() != reviewIndexes[len(reviewIndexes)-1] {
		t.Fatalf("expected the first g to arm the motion without moving selection, actual %d", subject.model.ActionsPopupSelectedActionIndex())
	}

	actualErr = topHandler(gui, nil)
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
	subject.model.UpdateActionsPopupSearch(pullRequestCommentComposerTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestCommentComposerTitle))
	actualErr = subject.afterStateChange(gui)
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
	subject.model.UpdateActionsPopupSearch("clipboard", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "clipboard"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	if len(clipboardWriter.writes) != 1 || clipboardWriter.writes[0] != "https://github.com/acme/widgets/pull/42" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"https://github.com/acme/widgets/pull/42"}, clipboardWriter.writes)
	}
	then_statusLineContains(t, gui, yankSuccessMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
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
	if !strings.Contains(helpView.Buffer(), "Actions") {
		t.Fatalf("expected help buffer to contain %q, actual %q", "Actions", helpView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTabCursorOnAnInlineThread_WhenOpening_ThenItShowsTheResolveInlineCommentAction(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "General feedback",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
						Body:      "Inline thread body",
						CreatedAt: "2026-04-18T10:30:00Z",
						DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42":            "Rendered body 42",
		"General feedback":   "Rendered general feedback",
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Mark inline comment as resolved") {
		t.Fatalf("expected the popup to contain the inline-thread resolve action, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTabCursorOnAnOwnedInlineComment_WhenOpening_ThenItShowsUpdateAndDeleteInlineCommentActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						ID:              "PRRC_1",
						ViewerDidAuthor: true,
						Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
						Body:            "Inline thread body",
						CreatedAt:       "2026-04-18T10:30:00Z",
						DiffHunk:        "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42":            "Rendered body 42",
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), inlineCommentUpdateEditorTitle) {
		t.Fatalf("expected the popup to contain the inline comment update action, actual %q", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), inlineCommentDeleteActionTitle) {
		t.Fatalf("expected the popup to contain the inline comment delete action, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTabCursorOutsideInlineComments_WhenOpening_ThenItHidesTheResolveInlineCommentAction(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "General feedback",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
						Body:      "Inline thread body",
						CreatedAt: "2026-04-18T10:30:00Z",
						DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42":            "Rendered body 42",
		"General feedback":   "Rendered general feedback",
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered general feedback")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "Mark inline comment as") {
		t.Fatalf("expected the popup to hide inline-thread resolution actions away from inline comments, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTabResolveInlineCommentAction_WhenExecuting_ThenItRefreshesTheThreadStateAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
						Body:      "Inline thread body",
						CreatedAt: "2026-04-18T10:30:00Z",
						DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42":            "Rendered body 42",
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.resolveReviewThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected resolved thread ids %v, actual %v", []string{"thread-1"}, loader.resolveReviewThreadIDs)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, inlineCommentResolvedSuccessMessage)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Resolved") || strings.Contains(detailView.Buffer(), "@reviewer-inline  2026-04-18 10:30 UTC  Resolved") {
		t.Fatalf("expected the detail buffer to refresh with the resolved state on the header, actual %q", detailView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTabResolveInlineCommentAction_WhenGitHubRejectsTheResolution_ThenItShowsATransientErrorPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		resolveReviewThreadErr: errors.New("resolve refused"),
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"General feedback":   "Rendered general feedback",
		"Inline thread body": "Rendered inline thread body",
	}}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.asyncRunner = &capturingAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !reflect.DeepEqual(loader.resolveReviewThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected resolved thread ids %v, actual %v", []string{"thread-1"}, loader.resolveReviewThreadIDs)
	}
	if subject.actionsPopupErrorMessage != "" {
		t.Fatalf("expected popup error message to stay empty, actual %q", subject.actionsPopupErrorMessage)
	}
	then_transientErrorPopupContains(t, gui, "resolve refused")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Resolved") {
		t.Fatalf("expected the detail buffer to keep the thread unresolved after the failure, actual %q", detailView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTabResolveInlineCommentAction_WhenSwitchingToChangesTab_ThenItRefreshesTheThreadStateThereToo(t *testing.T) {
	diff := githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/internal/tui/render.go b/internal/tui/render.go",
			"index 1111111..2222222 100644",
			"--- a/internal/tui/render.go",
			"+++ b/internal/tui/render.go",
			"@@ -42,2 +42,2 @@",
			" context line",
			"-old line",
			"+new line",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 1}},
		Threads: []githubcli.PullRequestReviewThread{{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     43,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
				Body:      "Inline thread body",
				CreatedAt: "2026-04-18T10:30:00Z",
			}},
		}},
	}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
						Body:      "Inline thread body",
						CreatedAt: "2026-04-18T10:30:00Z",
						DiffHunk:  "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
					}},
				}},
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Resolved") || !strings.Contains(detailView.Buffer(), "internal/tui/render.go:43") {
		t.Fatalf("expected the changes tab to show the resolved inline thread header, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered inline thread body") {
		t.Fatalf("expected resolved threads to stay folded in the changes tab, actual %q", detailView.Buffer())
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42"}, loader.diffCalls)
	}
}

func TestResolveInlineComment_GivenBrowserCommentsTabAction_WhenSubmittingOptimistically_ThenItKeepsTheThreadHeaderVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
						Body:      "Inline thread body",
						CreatedAt: "2026-04-18T10:30:00Z",
						DiffHunk:  "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh before the queued run, actual %v", loader.detailCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Resolved") {
		t.Fatalf("expected detail buffer to contain the optimistic resolved header, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered inline thread body") {
		t.Fatalf("expected the resolved thread body to stay hidden after the optimistic refresh, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTabCursorOnTheSecondInlineThread_WhenResolving_ThenItTargetsTheMatchingThreadAfterCompactRendering(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{
					{
						ID:         "thread-1",
						IsResolved: true,
						Path:       "internal/tui/render.go",
						Line:       43,
						DiffSide:   "RIGHT",
						Comments: []githubcli.PullRequestComment{{
							Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline-one"},
							Body:      "First inline thread body",
							CreatedAt: "2026-04-18T10:30:00Z",
							DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
						}},
					},
					{
						ID:       "thread-2",
						Path:     "internal/tui/model.go",
						Line:     57,
						DiffSide: "RIGHT",
						Comments: []githubcli.PullRequestComment{{
							Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline-two"},
							Body:      "Second inline thread body",
							CreatedAt: "2026-04-18T11:00:00Z",
							DiffHunk:  "@@ -56,2 +56,2 @@\n old line\n-old value\n+new value",
						}},
					},
				},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42":                   "Rendered body 42",
		"First inline thread body":  "Rendered first inline thread body",
		"Second inline thread body": "Rendered second inline thread body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered second inline thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.resolveReviewThreadIDs, []string{"thread-2"}) {
		t.Fatalf("expected resolved thread ids %v, actual %v", []string{"thread-2"}, loader.resolveReviewThreadIDs)
	}
}

func TestActionsPopup_GivenReviewModeCursorOnAResolvedInlineThread_WhenOpening_ThenItShowsTheUnresolveInlineCommentAction(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:         "thread-1",
		IsResolved: true,
		Path:       "internal/tui/render.go",
		Line:       3,
		DiffSide:   "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
			DiffHunk:  "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go:3")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Mark inline comment as unresolved") {
		t.Fatalf("expected the popup to contain the inline-thread unresolve action, actual %q", popupView.Buffer())
	}
}

func TestResolveInlineComment_GivenReviewModeAction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
			DiffHunk:  "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh before the queued run, actual %v", loader.diffCalls)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(filesView.Buffer(), "Loading file tree...") {
		t.Fatalf("expected files buffer to avoid the loading state, actual %q", filesView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Resolved") {
		t.Fatalf("expected detail buffer to contain the optimistic resolved header, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the resolved thread body to stay hidden after the optimistic refresh, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
}

func TestUnresolveInlineComment_GivenReviewModeAction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:         "thread-1",
		IsResolved: true,
		Path:       "internal/tui/render.go",
		Line:       3,
		DiffSide:   "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
			DiffHunk:  "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go:3")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark inline comment as unresolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as unresolved"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh before the queued run, actual %v", loader.diffCalls)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(filesView.Buffer(), "Loading file tree...") {
		t.Fatalf("expected files buffer to avoid the loading state, actual %q", filesView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Unresolved") {
		t.Fatalf("expected detail buffer to contain the optimistic unresolved header, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the unresolved thread body to reappear after the optimistic refresh, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
}

func TestActionsPopup_GivenReviewModeOwnedInlineCommentUpdateAction_WhenExecuting_ThenItOpensTheEditorSeededWithTheCurrentMarkdown(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:              "PRRC_1",
			ViewerDidAuthor: true,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
			Body:            "**Thread body**",
			CreatedAt:       "2026-04-20T10:00:00Z",
			DiffHunk:        "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"**Thread body**": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("edit inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	editorView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(editorView.Title, inlineCommentUpdateEditorTitle) {
		t.Fatalf("expected the editor title to contain %q, actual %q", inlineCommentUpdateEditorTitle, editorView.Title)
	}
	if !strings.Contains(editorView.Buffer(), "**Thread body**") {
		t.Fatalf("expected the editor buffer to contain the raw markdown %q, actual %q", "**Thread body**", editorView.Buffer())
	}
}

func TestEditInlineComment_GivenBrowserCommentsTabSubmit_WhenSubmittingOptimistically_ThenItKeepsTheRenderedThreadVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						ID:              "PRRC_1",
						ViewerDidAuthor: true,
						Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
						Body:            "Original body",
						CreatedAt:       "2026-04-18T10:30:00Z",
						DiffHunk:        "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Original body": "Rendered original body",
		"Updated body":  "Rendered updated body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("edit inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Updated body")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh before the queued run, actual %v", loader.detailCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered updated body") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered updated body", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
}

func TestEditInlineComment_GivenSuccessfulSubmit_WhenSubmitting_ThenItRefreshesTheRenderedReviewThreadAndShowsFeedback(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:              "PRRC_1",
			ViewerDidAuthor: true,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
			Body:            "Original body",
			CreatedAt:       "2026-04-20T10:00:00Z",
			DiffHunk:        "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Original body": "Rendered original body",
		"Updated body":  "Rendered updated body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("edit inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Updated body")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.updateReviewCommentIDs, []string{"PRRC_1"}) {
		t.Fatalf("expected updated inline comment ids %v, actual %v", []string{"PRRC_1"}, loader.updateReviewCommentIDs)
	}
	if !reflect.DeepEqual(loader.updateReviewCommentBodies, []string{"Updated body"}) {
		t.Fatalf("expected updated inline comment bodies %v, actual %v", []string{"Updated body"}, loader.updateReviewCommentBodies)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered updated body") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered updated body", detailView.Buffer())
	}
	then_statusLineContains(t, gui, inlineCommentUpdatedSuccessMessage)
}

func TestDeleteInlineComment_GivenBrowserCommentsTabAction_WhenSubmittingOptimistically_ThenItKeepsTheCommentsTabVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     43,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						ID:              "PRRC_1",
						ViewerDidAuthor: true,
						Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
						Body:            "Thread body",
						CreatedAt:       "2026-04-18T10:30:00Z",
						DiffHunk:        "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("delete inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "delete inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh before the queued run, actual %v", loader.detailCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected detail buffer to remove %q, actual %q", "Rendered thread body", detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 1)
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
}

func TestActionsPopup_GivenReviewModeOwnedInlineCommentDeleteAction_WhenExecuting_ThenItDeletesTheCommentAndRefreshesTheDiff(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:              "PRRC_1",
			ViewerDidAuthor: true,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
			Body:            "Thread body",
			CreatedAt:       "2026-04-20T10:00:00Z",
			DiffHunk:        "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("delete inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "delete inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.deleteReviewCommentIDs, []string{"PRRC_1"}) {
		t.Fatalf("expected deleted inline comment ids %v, actual %v", []string{"PRRC_1"}, loader.deleteReviewCommentIDs)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected detail buffer to remove %q, actual %q", "Rendered thread body", detailView.Buffer())
	}
	then_statusLineContains(t, gui, inlineCommentDeletedSuccessMessage)
}

func TestEditInlineComment_GivenReviewModeSubmit_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:              "PRRC_1",
			ViewerDidAuthor: true,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
			Body:            "Original body",
			CreatedAt:       "2026-04-20T10:00:00Z",
			DiffHunk:        "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Original body": "Rendered original body",
		"Updated body":  "Rendered updated body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("edit inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Updated body")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh before the queued run, actual %v", loader.diffCalls)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(filesView.Buffer(), "Loading file tree...") {
		t.Fatalf("expected files buffer to avoid the loading state, actual %q", filesView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered updated body") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered updated body", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
}

func TestDeleteInlineComment_GivenReviewModeAction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:              "PRRC_1",
			ViewerDidAuthor: true,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
			Body:            "Thread body",
			CreatedAt:       "2026-04-20T10:00:00Z",
			DiffHunk:        "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("delete inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "delete inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh before the queued run, actual %v", loader.diffCalls)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(filesView.Buffer(), "Loading file tree...") {
		t.Fatalf("expected files buffer to avoid the loading state, actual %q", filesView.Buffer())
	}
	if strings.Contains(filesView.Buffer(), "render.go  1") {
		t.Fatalf("expected files buffer to drop the optimistic comment count, actual %q", filesView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected detail buffer to remove %q, actual %q", "Rendered thread body", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
}

func TestActionsPopup_GivenDetailFocus_WhenPressingQ_ThenItReturnsToTheDetailPaneCleanly(t *testing.T) {
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
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'q')
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_viewDoesNotExist(t, gui, viewActionsPopupSearchName)
}

func actionsPopupLabel(icon string, title string) string {
	return actionsPopupAction{icon: icon, title: title}.label()
}

func then_popupBufferContainsOrderedActionLines(t *testing.T, buffer string, expected []string) {
	t.Helper()

	actual := strings.Split(strings.TrimSpace(buffer), "\n")
	if len(actual) != len(expected) {
		t.Fatalf("expected %d popup action lines, actual %d: %q", len(expected), len(actual), buffer)
	}
	for index, expectedLine := range expected {
		actualLine := strings.TrimSpace(actual[index])
		if actualLine != strings.TrimSpace(expectedLine) {
			t.Fatalf("expected popup action line %d to be %q, actual %q", index, expectedLine, actual[index])
		}
	}
}

func given_actionsPopupPullRequest() githubcli.PullRequest {
	return githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Original body"}
}
