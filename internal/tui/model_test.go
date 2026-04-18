package tui

import "testing"

func TestNextSideView_GivenUserFocus_WhenCyclingForward_ThenFocusMovesToPullRequestsAndBack(t *testing.T) {
	subject := given_model()

	when_movingToNextSideView(subject)
	actual := subject.Focus()
	if actual != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actual)
	}

	when_movingToNextSideView(subject)
	actual = subject.Focus()
	if actual != FocusUserView {
		t.Fatalf("expected focus %v, actual %v", FocusUserView, actual)
	}
}

func TestPreviousSideView_GivenPullRequestsFocus_WhenCyclingBackward_ThenFocusMovesToUserAndBack(t *testing.T) {
	subject := given_model()
	subject.NextSideView()

	when_movingToPreviousSideView(subject)
	actual := subject.Focus()
	if actual != FocusUserView {
		t.Fatalf("expected focus %v, actual %v", FocusUserView, actual)
	}

	when_movingToPreviousSideView(subject)
	actual = subject.Focus()
	if actual != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actual)
	}
}

func TestNextSideView_GivenDetailFocus_WhenCycling_ThenFocusNeverStaysOnDetail(t *testing.T) {
	subject := given_model()
	subject.OpenDetail()

	when_movingToNextSideView(subject)

	actual := subject.Focus()
	if actual != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actual)
	}
}

func TestOpenDetail_GivenUserFocus_WhenPressingEnter_ThenFocusMovesToDetailAndShowsUserContent(t *testing.T) {
	subject := given_model()

	when_openingDetail(subject)

	actual := subject.Focus()
	if actual != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, actual)
	}

	expected := "User detail 1"
	actualDetail := subject.DetailContent()
	if actualDetail != expected {
		t.Fatalf("expected detail %q, actual %q", expected, actualDetail)
	}
}

func TestOpenDetail_GivenPullRequestsFocus_WhenPressingEnter_ThenFocusMovesToDetailAndShowsPullRequestContent(t *testing.T) {
	subject := given_model()
	subject.NextSideView()

	when_movingSelectionDown(subject)
	when_openingDetail(subject)

	actual := subject.Focus()
	if actual != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, actual)
	}

	expected := "My PR detail 2"
	actualDetail := subject.DetailContent()
	if actualDetail != expected {
		t.Fatalf("expected detail %q, actual %q", expected, actualDetail)
	}
}

func TestCloseDetail_GivenDetailFocus_WhenPressingEscape_ThenFocusReturnsToPreviousSideView(t *testing.T) {
	subject := given_model()
	subject.NextSideView()
	subject.OpenDetail()

	when_closingDetail(subject)

	actual := subject.Focus()
	if actual != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actual)
	}
}

func TestMoveSelection_GivenPullRequestsFocus_WhenMovingDownAndUp_ThenSelectionChangesWithinBounds(t *testing.T) {
	subject := given_model()
	subject.NextSideView()

	when_movingSelectionDown(subject)
	when_movingSelectionDown(subject)
	when_movingSelectionDown(subject)

	actual := subject.SelectedPullRequestIndex(MyPullRequestsTab)
	if actual != 1 {
		t.Fatalf("expected selection 1, actual %d", actual)
	}

	when_movingSelectionUp(subject)
	when_movingSelectionUp(subject)

	actual = subject.SelectedPullRequestIndex(MyPullRequestsTab)
	if actual != 0 {
		t.Fatalf("expected selection 0, actual %d", actual)
	}
}

func TestMoveSelection_GivenUserFocus_WhenMovingDownAndUp_ThenSelectionChangesWithinBounds(t *testing.T) {
	subject := given_model()

	when_movingSelectionDown(subject)
	when_movingSelectionDown(subject)

	actual := subject.SelectedUserIndex()
	if actual != 1 {
		t.Fatalf("expected selection 1, actual %d", actual)
	}

	when_movingSelectionUp(subject)
	when_movingSelectionUp(subject)

	actual = subject.SelectedUserIndex()
	if actual != 0 {
		t.Fatalf("expected selection 0, actual %d", actual)
	}
}

func TestNextTab_GivenPullRequestsFocus_WhenSwitchingTabs_ThenActiveTabChangesAndPreservesSelectionPerTab(t *testing.T) {
	subject := given_model()
	subject.NextSideView()
	subject.MoveSelectionDown()

	when_switchingToNextTab(subject)
	when_movingSelectionDown(subject)
	when_switchingToPreviousTab(subject)

	actual := subject.ActivePullRequestTab()
	if actual != MyPullRequestsTab {
		t.Fatalf("expected tab %v, actual %v", MyPullRequestsTab, actual)
	}

	actualSelection := subject.SelectedPullRequestIndex(MyPullRequestsTab)
	if actualSelection != 1 {
		t.Fatalf("expected my PR selection 1, actual %d", actualSelection)
	}

	subject.NextPullRequestTab()
	actualSelection = subject.SelectedPullRequestIndex(RequestedPullRequestsTab)
	if actualSelection != 1 {
		t.Fatalf("expected requested PR selection 1, actual %d", actualSelection)
	}
}

func TestNextTab_GivenUserFocus_WhenSwitchingTabs_ThenTabDoesNotChange(t *testing.T) {
	subject := given_model()

	when_switchingToNextTab(subject)

	actual := subject.ActivePullRequestTab()
	if actual != MyPullRequestsTab {
		t.Fatalf("expected tab %v, actual %v", MyPullRequestsTab, actual)
	}
}

func TestDetailContent_GivenFocusedSourceSelectionChanges_WhenRenderingDetail_ThenContentTracksTheSourceView(t *testing.T) {
	subject := given_model()

	when_movingSelectionDown(subject)
	actual := subject.DetailContent()
	if actual != "User detail 2" {
		t.Fatalf("expected detail %q, actual %q", "User detail 2", actual)
	}

	subject.NextSideView()
	subject.MoveSelectionDown()
	actual = subject.DetailContent()
	if actual != "My PR detail 2" {
		t.Fatalf("expected detail %q, actual %q", "My PR detail 2", actual)
	}
}

func given_model() *Model {
	return NewModel(SeedData{
		Users: []Item{
			{Title: "dummy-user-1", Detail: "User detail 1"},
			{Title: "dummy-user-2", Detail: "User detail 2"},
		},
		MyPullRequests: []Item{
			{Title: "my-pr-1", Detail: "My PR detail 1"},
			{Title: "my-pr-2", Detail: "My PR detail 2"},
		},
		RequestedPullRequests: []Item{
			{Title: "requested-pr-1", Detail: "Requested PR detail 1"},
			{Title: "requested-pr-2", Detail: "Requested PR detail 2"},
		},
	})
}

func when_movingToNextSideView(subject *Model) {
	subject.NextSideView()
}

func when_movingToPreviousSideView(subject *Model) {
	subject.PreviousSideView()
}

func when_openingDetail(subject *Model) {
	subject.OpenDetail()
}

func when_closingDetail(subject *Model) {
	subject.CloseDetail()
}

func when_movingSelectionDown(subject *Model) {
	subject.MoveSelectionDown()
}

func when_movingSelectionUp(subject *Model) {
	subject.MoveSelectionUp()
}

func when_switchingToNextTab(subject *Model) {
	subject.NextPullRequestTab()
}

func when_switchingToPreviousTab(subject *Model) {
	subject.PreviousPullRequestTab()
}
