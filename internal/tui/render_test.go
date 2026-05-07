package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestLayout_GivenFreshProgram_WhenRendering_ThenCreatesThreeViewsAndPlacesDetailOnTheRight(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	then_viewExists(t, gui, viewDetailName)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_currentViewNameIs(t, gui, viewUserName)

	detailX0, _, _, _, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	userX0, userY0, userX1, userY1, actualErr := gui.ViewPosition(viewUserName)
	then_noError(t, actualErr)
	pullRequestsX0, pullRequestsY0, pullRequestsX1, _, actualErr := gui.ViewPosition(viewPullRequestsName)
	then_noError(t, actualErr)

	if detailX0 <= userX1 || detailX0 <= pullRequestsX1 {
		t.Fatalf("expected detail view to be on the right, actual detailX0=%d userX1=%d pullRequestsX1=%d", detailX0, userX1, pullRequestsX1)
	}
	if userX0 != 0 || pullRequestsX0 != 0 || userY0 != 0 {
		t.Fatalf("expected side views to start on the left edge, actual userX0=%d pullRequestsX0=%d userY0=%d", userX0, pullRequestsX0, userY0)
	}
	if userY1 >= pullRequestsY0 {
		t.Fatalf("expected user view above pull requests view, actual userY1=%d pullRequestsY0=%d", userY1, pullRequestsY0)
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if pullRequestsView.TitlePrefix != "[2]" {
		t.Fatalf("expected pull requests title prefix %q, actual %q", "[2]", pullRequestsView.TitlePrefix)
	}
	then_tabsAre(t, pullRequestsView, []string{"My PRs", "Requested"}, 0)
}

func TestLayout_GivenWideTerminal_WhenRendering_ThenTheDetailPaneGetsAboutSixtyFivePercentOfTheWidth(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGuiWithSize(t, 100, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	_, _, userX1, _, actualErr := gui.ViewPosition(viewUserName)
	then_noError(t, actualErr)
	detailX0, _, detailX1, _, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)

	sidebarWidth := userX1 + 1
	detailWidth := detailX1 - detailX0 + 1
	if sidebarWidth != 35 {
		t.Fatalf("expected sidebar width %d, actual %d", 35, sidebarWidth)
	}
	if detailWidth != 65 {
		t.Fatalf("expected detail width %d, actual %d", 65, detailWidth)
	}
}

func TestLayout_GivenNarrowTerminal_WhenRendering_ThenThePaneWidthsStillRespectTheMinimums(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGuiWithSize(t, 80, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	_, _, userX1, _, actualErr := gui.ViewPosition(viewUserName)
	then_noError(t, actualErr)
	detailX0, _, detailX1, _, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)

	sidebarWidth := userX1 + 1
	detailWidth := detailX1 - detailX0 + 1
	if sidebarWidth != 32 {
		t.Fatalf("expected sidebar width %d, actual %d", 32, sidebarWidth)
	}
	if detailWidth < 40 {
		t.Fatalf("expected detail width to stay at least %d, actual %d", 40, detailWidth)
	}
}

func TestPullRequestsTitle_GivenKnownCounts_WhenRendering_ThenItShowsCountsForBothTabs(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.myPullRequestsCount = 3
	subject.myPullRequestsCountKnown = true
	subject.requestedPullRequestsCount = 12
	subject.requestedPullRequestsCountKnown = true

	actual := subject.pullRequestsTabLabels()
	if len(actual) != 2 || actual[0] != "My PRs (3)" || actual[1] != "Requested (12)" {
		t.Fatalf("expected tab labels %v, actual %v", []string{"My PRs (3)", "Requested (12)"}, actual)
	}
}

func TestLayout_GivenFreshProgram_WhenRendering_ThenUsesActiveAndInactiveViewColorsWithoutSelectionBackground(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)

	activeTextColor := gocui.GetColor(theme.ActiveTextHex)
	activeBorderColor := gocui.GetColor(theme.ActiveBorderHex)
	inactiveBorderColor := gocui.GetColor(theme.InactiveBorderHex)

	if gui.SelFrameColor != activeBorderColor {
		t.Fatalf("expected active border color %v, actual %v", activeBorderColor, gui.SelFrameColor)
	}
	if gui.FrameColor != inactiveBorderColor {
		t.Fatalf("expected inactive border color %v, actual %v", inactiveBorderColor, gui.FrameColor)
	}
	if gui.SelBgColor != gocui.ColorDefault {
		t.Fatalf("expected no active frame background, actual %v", gui.SelBgColor)
	}

	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)

	if userView.FgColor != activeTextColor {
		t.Fatalf("expected active user text color %v, actual %v", activeTextColor, userView.FgColor)
	}
	if detailView.FgColor != activeTextColor {
		t.Fatalf("expected inactive detail text color %v, actual %v", activeTextColor, detailView.FgColor)
	}
	if pullRequestsView.FgColor != activeTextColor {
		t.Fatalf("expected inactive pull request text color %v, actual %v", activeTextColor, pullRequestsView.FgColor)
	}
	if !userView.Highlight {
		t.Fatal("expected the active side view to be highlighted")
	}
	if detailView.Highlight {
		t.Fatal("expected the inactive detail view to avoid highlight background")
	}
	if pullRequestsView.Highlight {
		t.Fatal("expected the inactive pull requests view to avoid highlight background")
	}
	selectedLineBackground := gocui.GetColor(theme.SelectedLineBackgroundHex)
	if userView.SelBgColor != selectedLineBackground {
		t.Fatalf("expected active line background color %v, actual %v", selectedLineBackground, userView.SelBgColor)
	}
}

func TestLayout_GivenDetailFocus_WhenRendering_ThenTheSourceViewKeepsTheSelectedLineBackground(t *testing.T) {
	model := given_model()
	model.NextSideView()
	model.MoveSelectionDown()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)

	if userView.Highlight {
		t.Fatal("expected non-source view highlight to stay off")
	}
	if !pullRequestsView.Highlight {
		t.Fatal("expected source view highlight to stay on while detail is focused")
	}
	if !pullRequestsView.HighlightInactive {
		t.Fatal("expected source view to keep inactive highlight while detail is focused")
	}
}

func TestLayout_GivenPullRequestListStatusIcon_WhenRendering_ThenTheRepositoryAndTitleKeepTheirCurrentColors(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(githubcli.PullRequest{
		Title:      "First PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
	})})
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, 0, "acme/widgets#42", given_themeColorHex(t, theme.PullRequestReferenceHex), "pull request reference")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, 0, "First PR", given_themeColorHex(t, theme.PullRequestTitleHex), "pull request title")
}

func TestLayout_GivenPullRequestListStatusIcon_WhenRenderingTheSelectedRow_ThenItStaysVisibleWithItsStatusColor(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(githubcli.PullRequest{
		Title:      "First PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
	})})
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, 0, "", given_themeColorHex(t, theme.PullRequestStatusOpenHex), "pull request status icon")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, 0, "")
}

func TestLayout_GivenSuccessfulMergeChecks_WhenRendering_ThenTheListRowUsesTheSuccessBackground(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(githubcli.PullRequest{
		Title:                  "Approved PR",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "APPROVED",
		Mergeable:              "MERGEABLE",
		MergeStateStatus:       "CLEAN",
		StatusCheckRollupState: "SUCCESS",
	})})
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewLineSegmentHasBackgroundColor(t, gui, viewPullRequestsName, 0, " acme/widgets#42 Approved PR", given_themeColorHex(t, theme.SuccessBackgroundHex), "approved pull request background")
}

func TestLayout_GivenFailingMergeChecks_WhenRendering_ThenTheListRowUsesTheFailureBackground(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(githubcli.PullRequest{
		Title:                  "Blocked PR",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "CHANGES_REQUESTED",
		MergeStateStatus:       "BLOCKED",
		StatusCheckRollupState: "FAILURE",
	})})
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewLineSegmentHasBackgroundColor(t, gui, viewPullRequestsName, 0, " acme/widgets#42 Blocked PR", given_themeColorHex(t, theme.FailureBackgroundHex), "failed merge checks pull request background")
}

func TestLayout_GivenPullRequestReviewTeams_WhenRendering_ThenTheListRowDoesNotShowThem(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(githubcli.PullRequest{
		Title:      "Need teams",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
		ReviewRequests: []githubcli.PullRequestReviewRequest{
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-one"}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "VIBE", Slug: "vibe", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "P3C", Slug: "p3c", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "FYP", Slug: "fyp", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
		},
	})})
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(pullRequestsView.Buffer(), detailReviewRequestsIcon+" VIBE, P3C, FYP") || strings.Contains(pullRequestsView.Buffer(), "VIBE") {
		t.Fatalf("expected the pull request list to omit requested review teams, actual %q", pullRequestsView.Buffer())
	}
}

func TestLayout_GivenFreshProgram_WhenRendering_ThenUsesRoundBordersForAllViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	expected := []rune{'─', '│', '╭', '╮', '╰', '╯'}
	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		view, actualErr := gui.View(viewName)
		then_noError(t, actualErr)
		if string(view.FrameRunes) != string(expected) {
			t.Fatalf("expected round frame runes %q for %s, actual %q", string(expected), viewName, string(view.FrameRunes))
		}
	}
}

func TestLayout_GivenFreshProgram_WhenRendering_ThenConnectedUserViewHasOneContentLine(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if userView.InnerHeight() != 1 {
		t.Fatalf("expected connected user inner height 1, actual %d", userView.InnerHeight())
	}
}

func TestLayout_GivenFreshProgram_WhenRendering_ThenItReservesTheLastTerminalLineForTheGlobalStatusLine(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	statusView, actualErr := gui.View("status-line")
	then_noError(t, actualErr)
	if statusView.InnerHeight() != 1 {
		t.Fatalf("expected status line inner height 1, actual %d", statusView.InnerHeight())
	}

	_, statusY0, _, _, actualErr := gui.ViewPosition("status-line")
	then_noError(t, actualErr)
	if statusY0 != 28 {
		t.Fatalf("expected status line y0 %d, actual %d", 28, statusY0)
	}
}

func TestLayout_GivenPullRequestsLoadingState_WhenRendering_ThenThePanesShowOnlyASpinnerAndTheStatusLineShowsTheGhCommand(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	subject.myPullRequestsLoading = true
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected pull requests buffer to contain spinner %q, actual %q", string(loadingSpinnerFrames[0]), pullRequestsView.Buffer())
	}
	if strings.Contains(pullRequestsView.Buffer(), myPullRequestsLoadingTitle) {
		t.Fatalf("expected pull requests buffer to hide %q, actual %q", myPullRequestsLoadingTitle, pullRequestsView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "My PRs tab") {
		t.Fatalf("expected detail buffer to hide the tab loading label, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), myPullRequestsLoadingDetail) {
		t.Fatalf("expected detail buffer to hide the loading command, actual %q", detailView.Buffer())
	}
	if strings.TrimSpace(detailView.Buffer()) != string(loadingSpinnerFrames[0]) {
		t.Fatalf("expected detail buffer %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}

	statusView, actualErr := gui.View("status-line")
	then_noError(t, actualErr)
	expectedStatus := string(loadingSpinnerFrames[0]) + " " + myPullRequestsLoadingDetail
	if actual := strings.TrimSpace(statusView.Buffer()); actual != expectedStatus {
		t.Fatalf("expected status line %q, actual %q", expectedStatus, actual)
	}
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, a: Action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}

func TestRefreshViews_GivenALoadingPullRequestsSpinner_WhenAdvancingTheFrame_ThenTheRenderedSpinnerChanges(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	subject.advanceLoadingSpinnerFrame()
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), string(loadingSpinnerFrames[1])) {
		t.Fatalf("expected pull requests buffer to contain spinner %q after advancing, actual %q", string(loadingSpinnerFrames[1]), pullRequestsView.Buffer())
	}
	if strings.Contains(pullRequestsView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected pull requests buffer to drop spinner %q after advancing, actual %q", string(loadingSpinnerFrames[0]), pullRequestsView.Buffer())
	}
}

func TestLayout_GivenDetailFocusOnPullRequests_WhenRendering_ThenShowsTheSelectedPullRequestInTheDetailPane(t *testing.T) {
	model := given_model()
	model.NextSideView()
	model.MoveSelectionDown()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	actualBuffer := detailView.Buffer()
	if !strings.Contains(actualBuffer, "My PR detail 2") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "My PR detail 2", actualBuffer)
	}
	if strings.Contains(actualBuffer, "ctrl+c quit") {
		t.Fatalf("expected detail buffer to omit the inline help text, actual %q", actualBuffer)
	}
}

func TestHelpPopup_GivenDetailFocus_WhenTogglingHelp_ThenItShowsCurrentViewAndGlobalKeybindings(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewHelpName)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	for _, expected := range []string{"--- Local ---", "--- Global ---", "h/j/k/l/<up>/<down>", "Move cursor", "w/e/b", "Next/end/previous word", "n/N", "Next/previous match", "v/V", "Start char/line visual selection", "<esc>/q", "Exit visual / return", "?", "Toggle help", "tab", "Switch side view"} {
		if !strings.Contains(actualBuffer, expected) {
			t.Fatalf("expected help buffer to contain %q, actual %q", expected, actualBuffer)
		}
	}
}

func TestHelpPopup_GivenVisibleHelp_WhenTogglingAgain_ThenThePopupClosesAndFocusReturnsToTheUnderlyingView(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewHelpName)

	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
	then_viewDoesNotExist(t, gui, viewHelpName)
}

func TestHelpPopup_GivenVisibleHelp_WhenPressingQ_ThenThePopupClosesAndFocusReturnsToTheUnderlyingView(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewHelpName)

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewHelpName, 'q')
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
	then_viewDoesNotExist(t, gui, viewHelpName)
}

func TestHelpPopup_GivenVisibleHelp_WhenHandlingSideViewShortcuts_ThenTheUnderlyingFocusDoesNotChange(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewHelpName)

	actualErr = subject.nextSideView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewHelpName)
	if subject.model.Focus() != FocusUserView {
		t.Fatalf("expected underlying focus %v, actual %v", FocusUserView, subject.model.Focus())
	}
}

func TestSwitchToSpecificView_GivenRenderedProgram_WhenHandlingViewShortcuts_ThenCurrentViewMatchesShortcut(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.focusPullRequestsView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
}

func TestOpenDetailAndCloseDetail_GivenPullRequestsFocus_WhenHandlingProgramActions_ThenCurrentViewFollowsTheModel(t *testing.T) {
	model := given_model()
	model.NextSideView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	actualErr = subject.closeDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
}

func TestOpenDetailAndCloseDetail_GivenPullRequestsFocus_WhenPressingQ_ThenCurrentViewReturnsToPullRequests(t *testing.T) {
	model := given_model()
	model.NextSideView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'q')
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
}

func TestOpenDetailAndCloseDetail_GivenRequestedPullRequestsTab_WhenHandlingProgramActions_ThenCurrentViewReturnsToPullRequestsWithTheRequestedTabSelected(t *testing.T) {
	model := given_model()
	model.NextSideView()
	model.NextPullRequestTab()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	actualErr = subject.closeDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	then_tabsAre(t, pullRequestsView, []string{"My PRs", "Requested"}, 1)
}

func TestSideViewCycling_GivenDetailFocus_WhenHandlingProgramActions_ThenCurrentViewStaysOnDetail(t *testing.T) {
	model := given_model()
	model.NextSideView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	actualErr = subject.nextSideView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	actualErr = subject.previousSideView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
}

func TestPaging_GivenDetailFocus_WhenHandlingProgramActions_ThenTheDetailViewMovesHalfAPageAndRecenters(t *testing.T) {
	model := NewModel(SeedData{
		Users: []Item{{
			Title:  "dummy-user-1",
			Detail: strings.TrimSpace(strings.Repeat("detail line\n", 80)),
		}},
	})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	step := maxInt(1, detailView.InnerHeight()/2)

	actualErr = subject.pageDown(gui, detailView)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewIsCenteredOnCursor(t, detailView, step, 80)

	actualErr = subject.pageUp(gui, detailView)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewIsCenteredOnCursor(t, detailView, 0, 80)
}

func TestLineNavigation_GivenDetailFocus_WhenHandlingProgramActions_ThenTheDetailCursorMovesByLineBeforeTheViewportScrolls(t *testing.T) {
	model := NewModel(SeedData{
		Users: []Item{{
			Title:  "dummy-user-1",
			Detail: strings.TrimSpace(strings.Repeat("detail line\n", 80)),
		}},
	})
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
	originX, originY := detailView.Origin()
	cursorX, cursorY := detailView.Cursor()
	if originX != 0 || originY != 0 {
		t.Fatalf("expected detail origin 0,0 after one downward move, actual %d,%d", originX, originY)
	}
	if cursorX != 0 || cursorY != 1 {
		t.Fatalf("expected detail cursor 0,1 after one downward move, actual %d,%d", cursorX, cursorY)
	}

	actualErr = subject.moveSelectionUp(gui, detailView)
	then_noError(t, actualErr)
	originX, originY = detailView.Origin()
	cursorX, cursorY = detailView.Cursor()
	if originX != 0 || originY != 0 {
		t.Fatalf("expected detail origin 0,0 after moving back up, actual %d,%d", originX, originY)
	}
	if cursorX != 0 || cursorY != 0 {
		t.Fatalf("expected detail cursor 0,0 after moving back up, actual %d,%d", cursorX, cursorY)
	}
}

func given_headlessGui(t *testing.T) *gocui.Gui {
	t.Helper()

	return given_headlessGuiWithSize(t, 120, 30)
}

func given_headlessGuiWithSize(t *testing.T, width int, height int) *gocui.Gui {
	t.Helper()

	gui, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
		Headless:   true,
		Width:      width,
		Height:     height,
	})
	if err != nil {
		t.Fatalf("expected no error, actual %v", err)
	}

	return gui
}

func then_viewExists(t *testing.T, gui *gocui.Gui, name string) {
	t.Helper()

	_, actualErr := gui.View(name)
	then_noError(t, actualErr)
}

func then_tabsAre(t *testing.T, view *gocui.View, expected []string, expectedIndex int) {
	t.Helper()

	if strings.Join(view.Tabs, "|") != strings.Join(expected, "|") {
		t.Fatalf("expected tabs %v, actual %v", expected, view.Tabs)
	}
	if view.TabIndex != expectedIndex {
		t.Fatalf("expected tab index %d, actual %d", expectedIndex, view.TabIndex)
	}
	if view.SelFgColor&gocui.AttrBold == 0 {
		t.Fatalf("expected selected tab color to include bold, actual %v", view.SelFgColor)
	}
}

func then_viewDoesNotExist(t *testing.T, gui *gocui.Gui, name string) {
	t.Helper()

	_, actualErr := gui.View(name)
	if !isUnknownViewError(actualErr) {
		t.Fatalf("expected view %q to be absent, actual error %v", name, actualErr)
	}
}

func then_currentViewNameIs(t *testing.T, gui *gocui.Gui, expected string) {
	t.Helper()

	actual := gui.CurrentView()
	if actual == nil || actual.Name() != expected {
		t.Fatalf("expected current view %q, actual %v", expected, actual)
	}
}

func then_noError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Fatalf("expected no error, actual %v", actual)
	}
}
