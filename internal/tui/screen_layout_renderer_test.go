package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestScreenLayout_GivenBrowserMode_WhenPlanningFrames_ThenItKeepsViewsZeroThroughThreeAndTheBottomStatusLine(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())

	actual := subject.screenLayoutForSize(100, 30)

	then_screenLayoutPanelFrameIs(t, actual, mainPanelViewZero, 35, 0, 99, 28)
	then_screenLayoutPanelFrameIs(t, actual, sidePanelViewOne, 0, 0, 34, 2)
	then_screenLayoutPanelFrameIs(t, actual, sidePanelViewTwo, 0, 3, 34, 23)
	then_screenLayoutPanelFrameIs(t, actual, sidePanelViewThree, 0, 24, 34, 28)
	then_screenLayoutViewFrameIs(t, actual.StatusLine, -1, 28, 100, 30)
}

func TestScreenLayout_GivenSearchAndKeyHints_WhenPlanningOverlays_ThenTheSearchPromptReusesTheStatusRowAndKeyHintsStayRightAligned(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGuiWithSize(t, 100, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openSearch(gui, nil))

	actual := subject.screenLayoutForSize(100, 30)
	searchFrame, ok := actual.OverlayFrame(viewSearchName)
	if !ok {
		t.Fatalf("expected overlay frame %q", viewSearchName)
	}

	then_screenLayoutViewFrameIs(t, searchFrame, -1, 28, 100, 30)
	if searchFrame.Frame != actual.StatusLine.Frame {
		t.Fatalf("expected search frame %+v to reuse the status-line frame %+v", searchFrame.Frame, actual.StatusLine.Frame)
	}
	if !actual.StatusLineKeyHints.Visible {
		t.Fatal("expected key hints frame to stay visible")
	}
	if actual.StatusLineKeyHints.Frame.x1 != actual.StatusLine.Frame.x1 {
		t.Fatalf("expected key hints right edge %d, actual %d", actual.StatusLine.Frame.x1, actual.StatusLineKeyHints.Frame.x1)
	}
}

func TestMainPanelRenderer_GivenBrowserPullRequestDetail_WhenConfiguring_ThenItUsesViewZeroTabsWithoutAStandaloneTitle(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	layout := subject.screenLayoutForSize(100, 30)
	frame, ok := layout.PanelFrameByViewNumber(int(mainPanelViewZero))
	if !ok {
		t.Fatalf("expected panel frame %d", mainPanelViewZero)
	}

	renderer, ok := subject.mainPanelRenderer().Renderer(frame)
	if !ok {
		t.Fatalf("expected a renderer for view %d", mainPanelViewZero)
	}

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	view, actualErr := gui.SetView("test-main", 0, 0, 60, 10, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}

	renderer.Configure(view)

	if view.Title != "" {
		t.Fatalf("expected no standalone title, actual %q", view.Title)
	}
	if view.TitlePrefix != "[0]" {
		t.Fatalf("expected title prefix %q, actual %q", "[0]", view.TitlePrefix)
	}
	then_tabsAre(t, view, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label(), CommitsDetailTab.Label(), ChangesDetailTab.Label()}, 0)
	if view.Highlight {
		t.Fatal("expected the read-only main panel to avoid list highlighting")
	}
}

func TestSidePanelRenderer_GivenBrowserPullRequestsView_WhenConfiguring_ThenItUsesTheVisibleNumberTabsAndSelectionHighlight(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	layout := subject.screenLayoutForSize(100, 30)
	frame, ok := layout.PanelFrameByViewNumber(int(sidePanelViewTwo))
	if !ok {
		t.Fatalf("expected panel frame %d", sidePanelViewTwo)
	}

	renderer, ok := subject.sidePanelRenderer().Renderer(frame)
	if !ok {
		t.Fatalf("expected a renderer for view %d", sidePanelViewTwo)
	}

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	view, actualErr := gui.SetView("test-side", 0, 0, 40, 10, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}

	renderer.Configure(view)

	if view.TitlePrefix != "[2]" {
		t.Fatalf("expected title prefix %q, actual %q", "[2]", view.TitlePrefix)
	}
	then_tabsAre(t, view, []string{"My PRs", "Requested"}, 0)
	if !view.Highlight {
		t.Fatal("expected the active side-panel list to keep selection highlighting")
	}
}

func TestStatusLinePresenter_GivenFeedbackAndLoading_WhenPresenting_ThenFeedbackWinsOverTheSpinner(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	subject.feedbackMessage = "Saved"
	subject.storyReviewLoading = true

	actual := subject.statusLinePresenter().Text()

	if actual != "Saved" {
		t.Fatalf("expected status line text %q, actual %q", "Saved", actual)
	}
}

func TestScreenComposition_GivenThePlannedLayout_WhenApplyingItToTheGUI_ThenTheVisibleViewsMatchThePureLayoutPlan(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGuiWithSize(t, 100, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	expected := subject.screenCompositionForSize(100, 30)

	then_noError(t, subject.applyScreenComposition(gui, expected))
	then_viewMatchesScreenLayoutPanelFrame(t, gui, viewUserName, expected.Layout, int(sidePanelViewOne))
	then_viewMatchesScreenLayoutPanelFrame(t, gui, viewPullRequestsName, expected.Layout, int(sidePanelViewTwo))
	then_viewMatchesScreenLayoutPanelFrame(t, gui, viewNotificationsName, expected.Layout, int(sidePanelViewThree))
	then_viewMatchesScreenLayoutPanelFrame(t, gui, viewDetailName, expected.Layout, int(mainPanelViewZero))
	then_viewMatchesScreenLayoutFrame(t, gui, viewStatusLineName, expected.Layout.StatusLine)
}

func then_screenLayoutPanelFrameIs(t *testing.T, layout ScreenLayout, number panelViewNumber, expectedX0 int, expectedY0 int, expectedX1 int, expectedY1 int) {
	t.Helper()

	frame, ok := layout.PanelFrameByViewNumber(int(number))
	if !ok {
		t.Fatalf("expected panel frame %d", number)
	}
	then_paneFrameEqualsCoordinates(t, frame.Frame, expectedX0, expectedY0, expectedX1, expectedY1)
}

func then_screenLayoutViewFrameIs(t *testing.T, frame screenViewFrame, expectedX0 int, expectedY0 int, expectedX1 int, expectedY1 int) {
	t.Helper()
	then_paneFrameEqualsCoordinates(t, frame.Frame, expectedX0, expectedY0, expectedX1, expectedY1)
}

func then_paneFrameEqualsCoordinates(t *testing.T, actual paneFrame, expectedX0 int, expectedY0 int, expectedX1 int, expectedY1 int) {
	t.Helper()

	if actual.x0 != expectedX0 || actual.y0 != expectedY0 || actual.x1 != expectedX1 || actual.y1 != expectedY1 {
		t.Fatalf(
			"expected frame (%d,%d)-(%d,%d), actual (%d,%d)-(%d,%d)",
			expectedX0,
			expectedY0,
			expectedX1,
			expectedY1,
			actual.x0,
			actual.y0,
			actual.x1,
			actual.y1,
		)
	}
}

func then_viewMatchesScreenLayoutPanelFrame(t *testing.T, gui *gocui.Gui, viewName string, layout ScreenLayout, viewNumber int) {
	t.Helper()

	frame, ok := layout.PanelFrameByViewNumber(viewNumber)
	if !ok {
		t.Fatalf("expected panel frame %d", viewNumber)
	}
	then_viewMatchesScreenLayoutFrame(t, gui, viewName, screenViewFrame{ViewName: viewName, Frame: frame.Frame, Visible: frame.Visible})
}

func then_viewMatchesScreenLayoutFrame(t *testing.T, gui *gocui.Gui, viewName string, expected screenViewFrame) {
	t.Helper()

	x0, y0, x1, y1, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)
	then_paneFrameEqualsCoordinates(t, paneFrame{x0: x0, y0: y0, x1: x1, y1: y1}, expected.Frame.x0, expected.Frame.y0, expected.Frame.x1, expected.Frame.y1)
}
