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
	if pullRequestsView.Title != "[2]-[My PRs] - Requested" {
		t.Fatalf("expected pull requests title %q, actual %q", "[2]-[My PRs] - Requested", pullRequestsView.Title)
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
	inactiveTextColor := gocui.GetColor(theme.InactiveTextHex)
	activeBorderColor := gocui.GetColor(theme.ActiveBorderHex)
	inactiveBorderColor := gocui.GetColor(theme.InactiveBorderHex)

	if gui.SelFrameColor != activeBorderColor {
		t.Fatalf("expected active border color %v, actual %v", activeBorderColor, gui.SelFrameColor)
	}
	if gui.FrameColor != inactiveBorderColor {
		t.Fatalf("expected inactive border color %v, actual %v", inactiveBorderColor, gui.FrameColor)
	}
	if gui.SelBgColor != gocui.ColorDefault {
		t.Fatalf("expected no active selection background, actual %v", gui.SelBgColor)
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
	if detailView.FgColor != inactiveTextColor {
		t.Fatalf("expected inactive detail text color %v, actual %v", inactiveTextColor, detailView.FgColor)
	}
	if pullRequestsView.FgColor != inactiveTextColor {
		t.Fatalf("expected inactive pull request text color %v, actual %v", inactiveTextColor, pullRequestsView.FgColor)
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
	if userView.SelBgColor != gocui.ColorDefault {
		t.Fatalf("expected no active line background color, actual %v", userView.SelBgColor)
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
