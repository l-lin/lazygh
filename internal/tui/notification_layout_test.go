package tui

import "testing"

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
	if notificationsView.InnerHeight() != 1 {
		t.Fatalf("expected collapsed notifications inner height %d, actual %d", 1, notificationsView.InnerHeight())
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
	if pullRequestsView.InnerHeight() != 1 {
		t.Fatalf("expected collapsed pull requests inner height %d, actual %d", 1, pullRequestsView.InnerHeight())
	}
	if notificationsView.InnerHeight() <= pullRequestsView.InnerHeight() {
		t.Fatalf("expected notifications inner height %d to exceed pull requests inner height %d", notificationsView.InnerHeight(), pullRequestsView.InnerHeight())
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
