package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingResizeBindings_ThenPlusAndMinusExistInTheMainViewsOnly(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: '+', handler: subject.growFocusedPane})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: '-', handler: subject.shrinkFocusedPane})
	}

	for _, viewName := range []string{viewSearchName, viewHelpName, viewActionsPopupName, viewActionsPopupSearchName, viewModalEditorName} {
		then_bindingDoesNotExist(t, actual, viewName, '+')
		then_bindingDoesNotExist(t, actual, viewName, '-')
	}
}

func TestPaneResize_GivenPullRequestsFocusWithAnAppliedSearch_WhenCyclingPlusThroughFullscreenAndBack_ThenTheHiddenPanesReappearWithTheSameContext(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.NextPullRequestTab()
	model.StartSearch()
	model.UpdateSearchDraft("2")
	model.SubmitSearch()

	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.growFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	actualErr = subject.growFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_viewDoesNotExist(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewDoesNotExist(t, gui, viewDetailName)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	then_tabsAre(t, pullRequestsView, []string{"My PRs", "Requested"}, 1)
	if !strings.Contains(pullRequestsView.Buffer(), "requested-pr-2") {
		t.Fatalf("expected the pull requests buffer to keep the filtered row, actual %q", pullRequestsView.Buffer())
	}

	pullRequestsFooterView, actualErr := gui.View(viewPullRequestsFooterName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(pullRequestsFooterView.Buffer()); actual != "/2 (1 match)  •  ? Help  / Search  a Actions" {
		t.Fatalf("expected pull requests footer %q, actual %q", "/2 (1 match)  •  ? Help  / Search  a Actions", actual)
	}

	actualErr = subject.growFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Requested PR detail 2") {
		t.Fatalf("expected detail buffer to contain %q after restoring the layout, actual %q", "Requested PR detail 2", detailView.Buffer())
	}

	pullRequestsFooterView, actualErr = gui.View(viewPullRequestsFooterName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(pullRequestsFooterView.Buffer()); actual != "/2 (1 match)  •  ? Help  / Search  a Actions" {
		t.Fatalf("expected pull requests footer %q after restoring the layout, actual %q", "/2 (1 match)  •  ? Help  / Search  a Actions", actual)
	}
}

func TestPaneResize_GivenDetailFocusWithAnAppliedSearch_WhenTogglingFullscreen_ThenTheDetailPaneHidesTheSidePanesAndRestoresThemWithTheSameContext(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	model.StartSearch()
	model.UpdateSearchDraft("detail 1")
	model.SubmitSearch()

	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialDetailX0, initialDetailY0, initialDetailX1, initialDetailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)

	actualErr = subject.growFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewUserName)
	then_viewDoesNotExist(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	detailX0, detailY0, detailX1, detailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	if detailX0 != 0 || detailY0 != 0 || detailX1 != 119 || detailY1 != 28 {
		t.Fatalf("expected fullscreen detail frame (%d,%d)-(%d,%d), actual (%d,%d)-(%d,%d)", 0, 0, 119, 28, detailX0, detailY0, detailX1, detailY1)
	}

	detailFooterView, actualErr := gui.View(viewDetailFooterName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(detailFooterView.Buffer()); actual != "/detail 1 (1 match)  •  ? Help  / Search" {
		t.Fatalf("expected detail footer %q, actual %q", "/detail 1 (1 match)  •  ? Help  / Search", actual)
	}

	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	if subject.model.Focus() != FocusDetailView {
		t.Fatalf("expected focus %v while detail is fullscreen, actual %v", FocusDetailView, subject.model.Focus())
	}

	actualErr = subject.shrinkFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	detailX0, detailY0, detailX1, detailY1, actualErr = gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	if detailX0 != initialDetailX0 || detailY0 != initialDetailY0 || detailX1 != initialDetailX1 || detailY1 != initialDetailY1 {
		t.Fatalf("expected restored detail frame (%d,%d)-(%d,%d), actual (%d,%d)-(%d,%d)", initialDetailX0, initialDetailY0, initialDetailX1, initialDetailY1, detailX0, detailY0, detailX1, detailY1)
	}

	detailFooterView, actualErr = gui.View(viewDetailFooterName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(detailFooterView.Buffer()); actual != "/detail 1 (1 match)  •  ? Help  / Search" {
		t.Fatalf("expected detail footer %q after restoring the layout, actual %q", "/detail 1 (1 match)  •  ? Help  / Search", actual)
	}
}

func TestPaneResize_GivenUserFocusWithAnAppliedSearch_WhenCyclingMinusThroughFullscreenAndBack_ThenTheDefaultLayoutReturnsExactly(t *testing.T) {
	model := given_model()
	model.FocusUserView()
	model.StartSearch()
	model.UpdateSearchDraft("2")
	model.SubmitSearch()

	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialUserX0, initialUserY0, initialUserX1, initialUserY1, actualErr := gui.ViewPosition(viewUserName)
	then_noError(t, actualErr)
	initialPullRequestsX0, initialPullRequestsY0, initialPullRequestsX1, initialPullRequestsY1, actualErr := gui.ViewPosition(viewPullRequestsName)
	then_noError(t, actualErr)
	initialDetailX0, initialDetailY0, initialDetailX1, initialDetailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)

	actualErr = subject.shrinkFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
	then_viewExists(t, gui, viewUserName)
	then_viewDoesNotExist(t, gui, viewPullRequestsName)
	then_viewDoesNotExist(t, gui, viewDetailName)

	userFooterView, actualErr := gui.View(viewUserFooterName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(userFooterView.Buffer()); actual != "/2 (1 match)  •  ? Help  / Search" {
		t.Fatalf("expected user footer %q, actual %q", "/2 (1 match)  •  ? Help  / Search", actual)
	}

	actualErr = subject.focusPullRequestsView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
	if subject.model.Focus() != FocusUserView {
		t.Fatalf("expected focus %v while the user pane is fullscreen, actual %v", FocusUserView, subject.model.Focus())
	}

	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)
	if subject.model.Focus() != FocusUserView {
		t.Fatalf("expected focus %v while the user pane is fullscreen, actual %v", FocusUserView, subject.model.Focus())
	}

	actualErr = subject.shrinkFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	actualErr = subject.shrinkFocusedPane(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewUserName)

	userX0, userY0, userX1, userY1, actualErr := gui.ViewPosition(viewUserName)
	then_noError(t, actualErr)
	pullRequestsX0, pullRequestsY0, pullRequestsX1, pullRequestsY1, actualErr := gui.ViewPosition(viewPullRequestsName)
	then_noError(t, actualErr)
	detailX0, detailY0, detailX1, detailY1, actualErr := gui.ViewPosition(viewDetailName)
	then_noError(t, actualErr)
	if userX0 != initialUserX0 || userY0 != initialUserY0 || userX1 != initialUserX1 || userY1 != initialUserY1 {
		t.Fatalf("expected the restored user pane frame (%d,%d)-(%d,%d), actual (%d,%d)-(%d,%d)", initialUserX0, initialUserY0, initialUserX1, initialUserY1, userX0, userY0, userX1, userY1)
	}
	if pullRequestsX0 != initialPullRequestsX0 || pullRequestsY0 != initialPullRequestsY0 || pullRequestsX1 != initialPullRequestsX1 || pullRequestsY1 != initialPullRequestsY1 {
		t.Fatalf("expected the restored pull requests pane frame (%d,%d)-(%d,%d), actual (%d,%d)-(%d,%d)", initialPullRequestsX0, initialPullRequestsY0, initialPullRequestsX1, initialPullRequestsY1, pullRequestsX0, pullRequestsY0, pullRequestsX1, pullRequestsY1)
	}
	if detailX0 != initialDetailX0 || detailY0 != initialDetailY0 || detailX1 != initialDetailX1 || detailY1 != initialDetailY1 {
		t.Fatalf("expected the restored detail pane frame (%d,%d)-(%d,%d), actual (%d,%d)-(%d,%d)", initialDetailX0, initialDetailY0, initialDetailX1, initialDetailY1, detailX0, detailY0, detailX1, detailY1)
	}

	userFooterView, actualErr = gui.View(viewUserFooterName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(userFooterView.Buffer()); actual != "/2 (1 match)  •  ? Help  / Search" {
		t.Fatalf("expected user footer %q after restoring the layout, actual %q", "/2 (1 match)  •  ? Help  / Search", actual)
	}
}

func TestPaneResize_GivenSearchHelpOrAnotherModal_WhenHandlingResizeActions_ThenTheLayoutStateDoesNotChange(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(*Program, *gocui.Gui) error
	}{
		{name: "search", setup: func(subject *Program, gui *gocui.Gui) error {
			return subject.openSearch(gui, nil)
		}},
		{name: "help", setup: func(subject *Program, gui *gocui.Gui) error {
			return subject.toggleHelp(gui, nil)
		}},
		{name: "actions popup", setup: func(subject *Program, gui *gocui.Gui) error {
			return subject.openActionsPopup(gui, nil)
		}},
		{name: "modal editor", setup: func(subject *Program, gui *gocui.Gui) error {
			return subject.openLineModalEditor(gui, "Resize blocker", "", nil)
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			model := given_model()
			model.FocusPullRequestsView()
			subject := NewProgramWithModel(model)
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)

			actualErr := subject.layout(gui)
			then_noError(t, actualErr)
			actualErr = testCase.setup(subject, gui)
			then_noError(t, actualErr)

			initialLayoutSize := subject.model.PaneLayoutSize()
			initialView := gui.CurrentView().Name()

			actualErr = subject.growFocusedPane(gui, nil)
			then_noError(t, actualErr)
			actualErr = subject.shrinkFocusedPane(gui, nil)
			then_noError(t, actualErr)

			if subject.model.PaneLayoutSize() != initialLayoutSize {
				t.Fatalf("expected layout size %v, actual %v", initialLayoutSize, subject.model.PaneLayoutSize())
			}
			then_currentViewNameIs(t, gui, initialView)
		})
	}
}
