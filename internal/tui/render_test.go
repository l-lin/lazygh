package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

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
	for _, expected := range []string{"--- Local ---", "--- Global ---", "j", "Scroll down", "k", "Scroll up", "?", "Toggle help", "tab/l", "Switch side view"} {
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

func TestPaging_GivenDetailFocus_WhenHandlingProgramActions_ThenTheDetailViewScrollsByPage(t *testing.T) {
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

	actualErr = subject.pageDown(gui, detailView)
	then_noError(t, actualErr)
	_, originY := detailView.Origin()
	if originY < 1 {
		t.Fatalf("expected detail origin to move down, actual %d", originY)
	}

	actualErr = subject.pageUp(gui, detailView)
	then_noError(t, actualErr)
	_, originY = detailView.Origin()
	if originY != 0 {
		t.Fatalf("expected detail origin to return to 0, actual %d", originY)
	}
}

func TestLineNavigation_GivenDetailFocus_WhenHandlingProgramActions_ThenTheDetailViewScrollsByLine(t *testing.T) {
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
	_, originY := detailView.Origin()
	if originY != 1 {
		t.Fatalf("expected detail origin 1 after moving down, actual %d", originY)
	}

	actualErr = subject.moveSelectionUp(gui, detailView)
	then_noError(t, actualErr)
	_, originY = detailView.Origin()
	if originY != 0 {
		t.Fatalf("expected detail origin 0 after moving up, actual %d", originY)
	}
}

func given_headlessGui(t *testing.T) *gocui.Gui {
	t.Helper()

	gui, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
		Headless:   true,
		Width:      120,
		Height:     30,
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
