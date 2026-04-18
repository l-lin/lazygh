package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestLayout_GivenFreshProgram_WhenRendering_ThenCreatesThreeViewsAndFocusesTheUserView(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	then_viewExists(t, gui, viewDetailName)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)

	actual := gui.CurrentView()
	if actual == nil || actual.Name() != viewUserName {
		t.Fatalf("expected current view %q, actual %v", viewUserName, actual)
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)

	actualBuffer := pullRequestsView.Buffer()
	if !strings.Contains(actualBuffer, "[My PRs]") || !strings.Contains(actualBuffer, "Requested") {
		t.Fatalf("expected pull request tabs in buffer, actual %q", actualBuffer)
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

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)

	actual := gui.CurrentView()
	if actual == nil || actual.Name() != viewDetailName {
		t.Fatalf("expected current view %q, actual %v", viewDetailName, actual)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	actualBuffer := detailView.Buffer()
	if !strings.Contains(actualBuffer, "My PR detail 2") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "My PR detail 2", actualBuffer)
	}
}

func TestOpenDetailAndCloseDetail_GivenPullRequestsFocus_WhenHandlingProgramActions_ThenCurrentViewFollowsTheModel(t *testing.T) {
	model := given_model()
	model.NextSideView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	actualErr = subject.closeDetail(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
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
