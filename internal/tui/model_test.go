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

func TestNextSideView_GivenDetailFocus_WhenCycling_ThenFocusStaysOnDetail(t *testing.T) {
	subject := given_model()
	subject.OpenDetail()

	when_movingToNextSideView(subject)

	actual := subject.Focus()
	if actual != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, actual)
	}
}

func TestPreviousSideView_GivenDetailFocus_WhenCycling_ThenFocusStaysOnDetail(t *testing.T) {
	subject := given_model()
	subject.OpenDetail()

	when_movingToPreviousSideView(subject)

	actual := subject.Focus()
	if actual != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, actual)
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

func TestCloseDetail_GivenRequestedPullRequestsDetail_WhenPressingEscape_ThenFocusReturnsToPullRequestsWithTheSameTabAndSelection(t *testing.T) {
	subject := given_model()
	subject.NextSideView()
	subject.NextPullRequestTab()
	subject.MoveSelectionDown()
	subject.OpenDetail()

	when_closingDetail(subject)

	actualFocus := subject.Focus()
	if actualFocus != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actualFocus)
	}
	if subject.ActivePullRequestTab() != RequestedPullRequestsTab {
		t.Fatalf("expected tab %v, actual %v", RequestedPullRequestsTab, subject.ActivePullRequestTab())
	}
	if subject.SelectedPullRequestIndex(RequestedPullRequestsTab) != 1 {
		t.Fatalf("expected requested selection 1, actual %d", subject.SelectedPullRequestIndex(RequestedPullRequestsTab))
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

func TestPageSelection_GivenPullRequestsFocus_WhenPagingDownAndUp_ThenSelectionMovesByThePageSize(t *testing.T) {
	subject := NewModel(SeedData{
		Users: []Item{{Title: "user-1", Detail: "user detail 1"}},
		MyPullRequests: []Item{
			{Title: "my-pr-1", Detail: "My PR detail 1"},
			{Title: "my-pr-2", Detail: "My PR detail 2"},
			{Title: "my-pr-3", Detail: "My PR detail 3"},
			{Title: "my-pr-4", Detail: "My PR detail 4"},
			{Title: "my-pr-5", Detail: "My PR detail 5"},
		},
	})
	subject.FocusPullRequestsView()

	when_pagingDown(subject, 3)
	actual := subject.SelectedPullRequestIndex(MyPullRequestsTab)
	if actual != 3 {
		t.Fatalf("expected selection 3, actual %d", actual)
	}

	when_pagingUp(subject, 2)
	actual = subject.SelectedPullRequestIndex(MyPullRequestsTab)
	if actual != 1 {
		t.Fatalf("expected selection 1, actual %d", actual)
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

func TestPageSelection_GivenUserFocus_WhenPagingDownAndUp_ThenSelectionMovesByThePageSize(t *testing.T) {
	subject := NewModel(SeedData{
		Users: []Item{
			{Title: "user-1", Detail: "User detail 1"},
			{Title: "user-2", Detail: "User detail 2"},
			{Title: "user-3", Detail: "User detail 3"},
			{Title: "user-4", Detail: "User detail 4"},
			{Title: "user-5", Detail: "User detail 5"},
		},
	})

	when_pagingDown(subject, 4)
	actual := subject.SelectedUserIndex()
	if actual != 4 {
		t.Fatalf("expected selection 4, actual %d", actual)
	}

	when_pagingUp(subject, 2)
	actual = subject.SelectedUserIndex()
	if actual != 2 {
		t.Fatalf("expected selection 2, actual %d", actual)
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

	subject.NextPullRequestTab()
	subject.MoveSelectionDown()
	actual = subject.DetailContent()
	if actual != "Requested PR detail 2" {
		t.Fatalf("expected detail %q, actual %q", "Requested PR detail 2", actual)
	}
}

func TestFocusDetailView_GivenPullRequestsFocus_WhenJumpingToViewZero_ThenFocusMovesToDetailAndKeepsThePullRequestSource(t *testing.T) {
	subject := given_model()
	subject.NextSideView()
	subject.MoveSelectionDown()

	when_focusingDetailView(subject)

	actual := subject.Focus()
	if actual != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, actual)
	}

	actualDetail := subject.DetailContent()
	if actualDetail != "My PR detail 2" {
		t.Fatalf("expected detail %q, actual %q", "My PR detail 2", actualDetail)
	}
}

func TestFocusUserView_GivenDetailFocus_WhenJumpingToViewOne_ThenFocusMovesToTheUserView(t *testing.T) {
	subject := given_model()
	subject.NextSideView()
	subject.OpenDetail()

	when_focusingUserView(subject)

	actual := subject.Focus()
	if actual != FocusUserView {
		t.Fatalf("expected focus %v, actual %v", FocusUserView, actual)
	}
}

func TestFocusPullRequestsView_GivenDetailFocus_WhenJumpingToViewTwo_ThenFocusMovesToThePullRequestsView(t *testing.T) {
	subject := given_model()
	subject.OpenDetail()

	when_focusingPullRequestsView(subject)

	actual := subject.Focus()
	if actual != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actual)
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

func when_pagingDown(subject *Model, pageSize int) {
	subject.PageDown(pageSize)
}

func when_pagingUp(subject *Model, pageSize int) {
	subject.PageUp(pageSize)
}

func when_switchingToNextTab(subject *Model) {
	subject.NextPullRequestTab()
}

func when_switchingToPreviousTab(subject *Model) {
	subject.PreviousPullRequestTab()
}

func when_focusingDetailView(subject *Model) {
	subject.FocusDetailView()
}

func when_focusingUserView(subject *Model) {
	subject.FocusUserView()
}

func when_focusingPullRequestsView(subject *Model) {
	subject.FocusPullRequestsView()
}
