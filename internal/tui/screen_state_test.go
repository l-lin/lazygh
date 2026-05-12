package tui

import "testing"

func TestScreenState_GivenBrowserMode_WhenFocusingPanelViewsByNumber_ThenTheActivePanelAndViewFollowTheRequestedNumber(t *testing.T) {
	subject := given_browserScreenState()

	then_screenFocusIs(t, subject, ActivePanelSide, sidePanelViewOne, FocusUserView)

	subject = subject.FocusViewNumber(int(mainPanelViewZero))
	then_screenFocusIs(t, subject, ActivePanelMain, mainPanelViewZero, FocusDetailView)

	subject = subject.FocusViewNumber(int(sidePanelViewTwo))
	then_screenFocusIs(t, subject, ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)

	subject = subject.FocusViewNumber(int(sidePanelViewThree))
	then_screenFocusIs(t, subject, ActivePanelSide, sidePanelViewThree, FocusNotificationsView)
}

func TestScreenState_GivenBrowserMode_WhenCyclingTheSidePanel_ThenTheOrderedViewsWrapWithoutMovingTheMainPanel(t *testing.T) {
	subject := given_browserScreenState()

	subject = subject.NextSideView()
	then_screenFocusIs(t, subject, ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)

	subject = subject.NextSideView()
	then_screenFocusIs(t, subject, ActivePanelSide, sidePanelViewThree, FocusNotificationsView)

	subject = subject.NextSideView()
	then_screenFocusIs(t, subject, ActivePanelSide, sidePanelViewOne, FocusUserView)

	subject = subject.FocusViewNumber(int(mainPanelViewZero))
	subject = subject.NextSideView()
	then_screenFocusIs(t, subject, ActivePanelMain, mainPanelViewZero, FocusDetailView)
}

func TestScreenState_GivenTabbedViews_WhenSwitchingTabs_ThenOnlyTheActiveViewChangesItsLocalTab(t *testing.T) {
	subject := given_browserScreenState().FocusViewNumber(int(sidePanelViewTwo))

	subject = subject.NextTab()
	then_viewTabLabelIs(t, subject, sidePanelViewTwo, "Requested")

	subject = subject.PreviousTab()
	then_viewTabLabelIs(t, subject, sidePanelViewTwo, "My PRs")

	subject = subject.FocusViewNumber(int(mainPanelViewZero))
	subject = subject.NextTab()
	then_viewTabLabelIs(t, subject, mainPanelViewZero, CommentsDetailTab.Label())
	then_viewTabLabelIs(t, subject, sidePanelViewTwo, "My PRs")

	subject = subject.FocusViewNumber(int(sidePanelViewOne))
	subject = subject.NextTab()
	then_viewTabLabelIs(t, subject, sidePanelViewOne, "")
}

func TestScreenState_GivenMainAndSidePanels_WhenCheckingCursorEligibility_ThenOnlyViewZeroCanOwnTheCursor(t *testing.T) {
	subject := given_browserScreenState()
	if subject.AllowsMainCursor() {
		t.Fatal("expected side-panel focus to hide the main cursor")
	}

	subject = subject.FocusViewNumber(int(mainPanelViewZero))
	if !subject.AllowsMainCursor() {
		t.Fatal("expected view 0 to own the main cursor")
	}

	subject = subject.FocusViewNumber(int(sidePanelViewThree))
	if subject.AllowsMainCursor() {
		t.Fatal("expected side-panel focus to hide the main cursor again")
	}
}

func TestScreenState_GivenPanelsAndOverlays_WhenResolvingMainViewAndKeyHints_ThenPureQueriesFollowTheTopContext(t *testing.T) {
	subject := given_browserScreenState().FocusViewNumber(int(sidePanelViewThree))

	resolver := subject.MainViewResolver()
	if resolver.SourceView.Number != int(sidePanelViewThree) {
		t.Fatalf("expected view 0 to resolve from side view %d, actual %d", sidePanelViewThree, resolver.SourceView.Number)
	}
	if actual := subject.KeyHintContext(); actual != KeyHintContextSidePanel {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextSidePanel, actual)
	}

	subject = subject.FocusViewNumber(int(mainPanelViewZero))
	if actual := subject.KeyHintContext(); actual != KeyHintContextMainPanel {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextMainPanel, actual)
	}

	subject = subject.WithOverlay(OverlayState{Kind: OverlayKindActionsPopup})
	if actual := subject.KeyHintContext(); actual != KeyHintContextActionsPopup {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextActionsPopup, actual)
	}

	subject = subject.WithOverlay(OverlayState{Kind: OverlayKindModalEditor})
	if actual := subject.KeyHintContext(); actual != KeyHintContextModalEditor {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextModalEditor, actual)
	}
}

func given_browserScreenState() ScreenState {
	return newBrowserScreenState(
		FocusUserView,
		MyPullRequestsTab,
		[]TabState{{Label: "Description"}, {Label: CommentsDetailTab.Label()}, {Label: CommitsDetailTab.Label()}, {Label: ChangesDetailTab.Label()}},
		[]TabState{{Label: "My PRs"}, {Label: "Requested"}},
	)
}

func then_screenFocusIs(t *testing.T, subject ScreenState, expectedPanel ActivePanel, expectedView panelViewNumber, expectedFocus Focus) {
	t.Helper()

	activeView := subject.ActiveView()
	if actual := subject.ActivePanel; actual != expectedPanel {
		t.Fatalf("expected active panel %v, actual %v", expectedPanel, actual)
	}
	if activeView.Number != int(expectedView) {
		t.Fatalf("expected active view number %d, actual %d", expectedView, activeView.Number)
	}
	if activeView.Focus != expectedFocus {
		t.Fatalf("expected active focus %v, actual %v", expectedFocus, activeView.Focus)
	}
}

func then_viewTabLabelIs(t *testing.T, subject ScreenState, viewNumber panelViewNumber, expected string) {
	t.Helper()

	view, ok := subject.ViewByNumber(int(viewNumber))
	if !ok {
		t.Fatalf("expected view %d to exist", viewNumber)
	}
	if expected == "" {
		if len(view.Tabs) != 0 {
			t.Fatalf("expected view %d to have no tabs, actual %+v", viewNumber, view.Tabs)
		}
		return
	}
	if len(view.Tabs) == 0 {
		t.Fatalf("expected view %d to have tabs", viewNumber)
	}
	if view.ActiveTab < 0 || view.ActiveTab >= len(view.Tabs) {
		t.Fatalf("expected active tab index inside %+v", view)
	}
	if actual := view.Tabs[view.ActiveTab].Label; actual != expected {
		t.Fatalf("expected active tab label %q for view %d, actual %q", expected, viewNumber, actual)
	}
}
