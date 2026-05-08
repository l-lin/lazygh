package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestLayout_GivenBrowserMode_WhenRendering_ThenViewThreeShowsNotificationsAndStartsCollapsed(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	then_viewExists(t, gui, viewNotificationsName)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	if notificationsView.TitlePrefix != "[3]" {
		t.Fatalf("expected notifications title prefix %q, actual %q", "[3]", notificationsView.TitlePrefix)
	}
	if notificationsView.InnerHeight() != 3 {
		t.Fatalf("expected collapsed notifications inner height %d, actual %d", 3, notificationsView.InnerHeight())
	}
	if pullRequestsView.InnerHeight() <= notificationsView.InnerHeight() {
		t.Fatalf("expected pull requests inner height %d to exceed notifications inner height %d", pullRequestsView.InnerHeight(), notificationsView.InnerHeight())
	}
}

func TestFocusNotificationsView_GivenBrowserMode_WhenJumpingToViewThree_ThenTheNotificationsPaneExpandsAndTakesFocus(t *testing.T) {
	model := given_model()
	model.FocusNotificationsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewNotificationsName)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	if pullRequestsView.InnerHeight() != 3 {
		t.Fatalf("expected collapsed pull requests inner height %d, actual %d", 3, pullRequestsView.InnerHeight())
	}
	if notificationsView.InnerHeight() <= pullRequestsView.InnerHeight() {
		t.Fatalf("expected notifications inner height %d to exceed pull requests inner height %d", notificationsView.InnerHeight(), pullRequestsView.InnerHeight())
	}
}

func TestNotificationsPane_GivenNotificationsFocus_WhenOpeningDetailWithEnter_ThenItKeepsTheExpandedHeight(t *testing.T) {
	model := given_model()
	model.FocusNotificationsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialNotificationsHeight := given_viewInnerHeight(t, gui, viewNotificationsName)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, gocui.KeyEnter)
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	actualNotificationsHeight := given_viewInnerHeight(t, gui, viewNotificationsName)
	if actualNotificationsHeight != initialNotificationsHeight {
		t.Fatalf("expected notifications height %d after pressing enter, actual %d", initialNotificationsHeight, actualNotificationsHeight)
	}
}

func TestNotificationsPane_GivenNotificationsFocus_WhenJumpingToViewZero_ThenItKeepsTheExpandedHeight(t *testing.T) {
	model := given_model()
	model.FocusNotificationsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialNotificationsHeight := given_viewInnerHeight(t, gui, viewNotificationsName)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, '0')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	actualNotificationsHeight := given_viewInnerHeight(t, gui, viewNotificationsName)
	if actualNotificationsHeight != initialNotificationsHeight {
		t.Fatalf("expected notifications height %d after pressing 0, actual %d", initialNotificationsHeight, actualNotificationsHeight)
	}
}

func TestReviewMode_GivenNotificationsFocusShortcut_WhenHandlingTheBrowserOnlyPane_ThenReviewModeKeepsItsExistingViews(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	then_viewDoesNotExist(t, gui, viewNotificationsName)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, '3')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	if subject.model.Focus() != FocusPullRequestsView {
		t.Fatalf("expected focus %v in review mode, actual %v", FocusPullRequestsView, subject.model.Focus())
	}
}

func given_viewInnerHeight(t *testing.T, gui *gocui.Gui, viewName string) int {
	t.Helper()

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	return view.InnerHeight()
}
