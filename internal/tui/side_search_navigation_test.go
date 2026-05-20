package tui

import "testing"

func TestSideViews_GivenSubmittedSearch_WhenMovingWithJAndK_ThenSelectionMovesAcrossItemsInsteadOfMatches(t *testing.T) {
	testCases := []struct {
		name                 string
		focus                Focus
		viewName             string
		expectedAfterSubmit  string
		expectedAfterFirstJ  string
		expectedAfterSecondJ string
		expectedAfterK       string
	}{
		{name: "users", focus: FocusUserView, viewName: viewUserName, expectedAfterSubmit: "beta entry", expectedAfterFirstJ: "gamma row", expectedAfterSecondJ: "delta entry", expectedAfterK: "gamma row"},
		{name: "pull requests", focus: FocusPullRequestsView, viewName: viewPullRequestsName, expectedAfterSubmit: "beta entry", expectedAfterFirstJ: "gamma row", expectedAfterSecondJ: "delta entry", expectedAfterK: "gamma row"},
		{name: "notifications", focus: FocusNotificationsView, viewName: viewNotificationsName, expectedAfterSubmit: "beta entry", expectedAfterFirstJ: "gamma row", expectedAfterSecondJ: "delta entry", expectedAfterK: "gamma row"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			model := given_sideSearchNavigationModel()
			focusSideView(model, testCase.focus)
			subject := NewProgramWithModel(model)
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)

			then_noError(t, subject.layout(gui))
			then_noError(t, subject.openSearch(gui, nil))
			searchView, actualErr := gui.View(viewSearchName)
			then_noError(t, actualErr)
			when_typingSearchQuery(t, subject, searchView, "entry")
			then_noError(t, subject.submitSearch(gui, searchView))

			then_sideViewSelectionIs(t, subject, testCase.focus, testCase.expectedAfterSubmit)

			view, actualErr := gui.View(testCase.viewName)
			then_noError(t, actualErr)
			then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), testCase.viewName, 'j')(gui, view))
			then_sideViewSelectionIs(t, subject, testCase.focus, testCase.expectedAfterFirstJ)

			then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), testCase.viewName, 'j')(gui, view))
			then_sideViewSelectionIs(t, subject, testCase.focus, testCase.expectedAfterSecondJ)

			then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), testCase.viewName, 'k')(gui, view))
			then_sideViewSelectionIs(t, subject, testCase.focus, testCase.expectedAfterK)
		})
	}
}

func TestSideViews_GivenSubmittedSearch_WhenRepeatingWithNAndN_ThenItMovesBetweenMatchesWithWraparound(t *testing.T) {
	testCases := []struct {
		name         string
		focus        Focus
		viewName     string
		expectedNext string
		expectedPrev string
	}{
		{name: "users", focus: FocusUserView, viewName: viewUserName, expectedNext: "delta entry", expectedPrev: "delta entry"},
		{name: "notifications", focus: FocusNotificationsView, viewName: viewNotificationsName, expectedNext: "delta entry", expectedPrev: "delta entry"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			model := given_sideSearchNavigationModel()
			focusSideView(model, testCase.focus)
			subject := NewProgramWithModel(model)
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)

			then_noError(t, subject.layout(gui))
			then_noError(t, subject.openSearch(gui, nil))
			searchView, actualErr := gui.View(viewSearchName)
			then_noError(t, actualErr)
			when_typingSearchQuery(t, subject, searchView, "entry")
			then_noError(t, subject.submitSearch(gui, searchView))
			then_sideViewSelectionIs(t, subject, testCase.focus, "beta entry")

			view, actualErr := gui.View(testCase.viewName)
			then_noError(t, actualErr)
			then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), testCase.viewName, 'n')(gui, view))
			then_sideViewSelectionIs(t, subject, testCase.focus, testCase.expectedNext)

			then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), testCase.viewName, 'n')(gui, view))
			then_sideViewSelectionIs(t, subject, testCase.focus, "beta entry")

			then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), testCase.viewName, 'N')(gui, view))
			then_sideViewSelectionIs(t, subject, testCase.focus, testCase.expectedPrev)
		})
	}
}

func given_sideSearchNavigationModel() *Model {
	return NewModel(SeedData{
		Users: []Item{
			{Title: "alpha row", Detail: "user alpha"},
			{Title: "beta entry", Detail: "user beta"},
			{Title: "gamma row", Detail: "user gamma"},
			{Title: "delta entry", Detail: "user delta"},
		},
		MyPullRequests: []Item{
			{Title: "alpha row", Detail: "pr alpha"},
			{Title: "beta entry", Detail: "pr beta"},
			{Title: "gamma row", Detail: "pr gamma"},
			{Title: "delta entry", Detail: "pr delta"},
		},
		Notifications: []Item{
			{Title: "alpha row", Detail: "notification alpha"},
			{Title: "beta entry", Detail: "notification beta"},
			{Title: "gamma row", Detail: "notification gamma"},
			{Title: "delta entry", Detail: "notification delta"},
		},
	})
}

func focusSideView(model *Model, focus Focus) {
	switch focus {
	case FocusPullRequestsView:
		model.FocusPullRequestsView()
	case FocusNotificationsView:
		model.FocusNotificationsView()
	default:
		model.FocusUserView()
	}
}

func then_sideViewSelectionIs(t *testing.T, subject *Program, focus Focus, expectedTitle string) {
	t.Helper()

	switch focus {
	case FocusPullRequestsView:
		actual := subject.model.PullRequests(subject.model.ActivePullRequestTab())[subject.model.SelectedPullRequestIndex(subject.model.ActivePullRequestTab())].Title
		if actual != expectedTitle {
			t.Fatalf("expected selected pull request %q, actual %q", expectedTitle, actual)
		}
	case FocusNotificationsView:
		actual := subject.model.Notifications()[subject.model.SelectedNotificationIndex()].Title
		if actual != expectedTitle {
			t.Fatalf("expected selected notification %q, actual %q", expectedTitle, actual)
		}
	default:
		actual := subject.model.Users()[subject.model.SelectedUserIndex()].Title
		if actual != expectedTitle {
			t.Fatalf("expected selected user %q, actual %q", expectedTitle, actual)
		}
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingSideSearchFollowBindings_ThenUserAndNotificationsViewsSupportNAndN(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'n', handler: subject.nextUserSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'N', handler: subject.previousUserSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewNotificationsName, key: 'n', handler: subject.nextNotificationsSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewNotificationsName, key: 'N', handler: subject.previousNotificationsSearchMatch})
}

func TestSearchPrompt_GivenOpenSearchInNotificationsView_WhenSubmittingThenMovingSelection_ThenThePromptClosesBackToTheNotificationsPane(t *testing.T) {
	model := given_sideSearchNavigationModel()
	model.FocusNotificationsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openSearch(gui, nil))
	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	when_typingSearchQuery(t, subject, searchView, "entry")
	then_noError(t, subject.submitSearch(gui, searchView))
	then_currentViewNameIs(t, gui, viewNotificationsName)

	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, 'j')(gui, notificationsView))
	then_sideViewSelectionIs(t, subject, FocusNotificationsView, "gamma row")
}
