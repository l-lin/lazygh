package tui

import (
	"fmt"
	"strings"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestDefaultSeedData_GivenAFreshModel_WhenReadingMyPullRequests_ThenItStartsInALoadingState(t *testing.T) {
	subject := NewModel(DefaultSeedData())

	actualPullRequests := subject.PullRequests(MyPullRequestsTab)
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	if actualPullRequests[0].Title != myPullRequestsLoadingTitle {
		t.Fatalf("expected title %q, actual %q", myPullRequestsLoadingTitle, actualPullRequests[0].Title)
	}
	if actualPullRequests[0].Detail != myPullRequestsLoadingDetail {
		t.Fatalf("expected detail %q, actual %q", myPullRequestsLoadingDetail, actualPullRequests[0].Detail)
	}
}

func TestSetPullRequests_GivenMyPullRequests_WhenSelectingThePullRequestsView_ThenDetailContentShowsMetadataAndBody(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetPullRequests(MyPullRequestsTab, []Item{myPullRequestItem(githubcli.PullRequest{
		Title:      "fix(P3C-6986): exclude dependencies bump PRs + bump GHA",
		Number:     422,
		Repository: githubcli.Repository{NameWithOwner: "acme/foobar"},
		URL:        "https://github.com/acme/foobar/pull/422",
		Body:       "No need to trigger Claude review for PRs that only bump dependencies.",
		State:      "open",
		IsDraft:    false,
		UpdatedAt:  "2026-04-17T10:39:35Z",
	})})
	subject.FocusPullRequestsView()

	actualPullRequests := subject.PullRequests(MyPullRequestsTab)
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	if actualPullRequests[0].Title != " acme/foobar#422 fix(P3C-6986): exclude dependencies bump PRs + bump GHA" {
		t.Fatalf("expected title %q, actual %q", " acme/foobar#422 fix(P3C-6986): exclude dependencies bump PRs + bump GHA", actualPullRequests[0].Title)
	}

	actualDetail := subject.DetailContent()
	expectedFragments := []string{
		"Repository: acme/foobar",
		"Number: #422",
		"State: open",
		"Draft: no",
		"Updated: 2026-04-17T10:39:35Z",
		"URL: https://github.com/acme/foobar/pull/422",
		"No need to trigger Claude review for PRs that only bump dependencies.",
	}
	for _, expected := range expectedFragments {
		if !strings.Contains(actualDetail, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actualDetail)
		}
	}
}

func TestSetPullRequests_GivenDefaultRepositoryStyle_WhenSelectingThePullRequestsView_ThenItShowsTheFullRepositoryReferenceInTheVisibleTitle(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetPullRequests(MyPullRequestsTab, []Item{myPullRequestItem(githubcli.PullRequest{
		Title:      "Open PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
	})})
	subject.FocusPullRequestsView()

	actualPullRequests := subject.PullRequests(MyPullRequestsTab)
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	expected := " acme/widgets#42 Open PR"
	if actualPullRequests[0].Title != expected {
		t.Fatalf("expected title %q, actual %q", expected, actualPullRequests[0].Title)
	}
}

func TestSetPullRequests_GivenASelectedMyPullRequest_WhenRefreshingTheList_ThenTheSelectionIndexIsPreserved(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetPullRequests(MyPullRequestsTab, []Item{
		{Title: "pr-1", Detail: "detail-1"},
		{Title: "pr-2", Detail: "detail-2"},
		{Title: "pr-3", Detail: "detail-3"},
	})
	subject.FocusPullRequestsView()
	subject.MoveSelectionDown()

	subject.SetPullRequests(MyPullRequestsTab, []Item{
		{Title: "pr-1", Detail: "detail-1"},
		{Title: "pr-2", Detail: "detail-2"},
		{Title: "pr-3", Detail: "detail-3"},
	})

	actual := subject.SelectedPullRequestIndex(MyPullRequestsTab)
	if actual != 1 {
		t.Fatalf("expected selection 1, actual %d", actual)
	}
}

func TestPullRequestRow_GivenDefaultRepositoryStyle_WhenBuildingTheListRow_ThenItUsesTheFullRepositoryReference(t *testing.T) {
	actual := pullRequestRow(githubcli.PullRequest{
		Title:      "Open PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
	}).Item

	expected := " acme/widgets#42 Open PR"
	if actual.Title != expected {
		t.Fatalf("expected title %q, actual %q", expected, actual.Title)
	}
	if len(actual.TitleSegments) != 3 {
		t.Fatalf("expected 3 title segments, actual %d", len(actual.TitleSegments))
	}
	if actual.TitleSegments[1].Text != "acme/widgets#42" {
		t.Fatalf("expected reference segment %q, actual %q", "acme/widgets#42", actual.TitleSegments[1].Text)
	}
}

func TestPullRequestRow_GivenShortRepositoryStyle_WhenBuildingTheListRow_ThenItUsesTheShortRepositoryReference(t *testing.T) {
	actual := pullRequestRowWithRepositoryStyle(appconfig.RepositoryStyleName, githubcli.PullRequest{
		Title:      "Open PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
	}).Item

	expected := " widgets#42 Open PR"
	if actual.Title != expected {
		t.Fatalf("expected title %q, actual %q", expected, actual.Title)
	}
	if len(actual.TitleSegments) != 3 {
		t.Fatalf("expected 3 title segments, actual %d", len(actual.TitleSegments))
	}
	if actual.TitleSegments[1].Text != "widgets#42" {
		t.Fatalf("expected reference segment %q, actual %q", "widgets#42", actual.TitleSegments[1].Text)
	}
}

func TestPullRequestRow_GivenPullRequestStatuses_WhenBuildingTheListRow_ThenItPrependsAStateColoredIcon(t *testing.T) {
	testCases := []struct {
		name                    string
		pullRequest             githubcli.PullRequest
		expectedIconText        string
		expectedIconPrefix      string
		expectedVisibleTitle    string
		expectedReferencePrefix string
		expectedTitlePrefix     string
	}{
		{
			name: "open",
			pullRequest: githubcli.PullRequest{
				Title:      "Open PR",
				Number:     42,
				Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
				State:      "OPEN",
			},
			expectedIconText:        " ",
			expectedIconPrefix:      foregroundColorEscape(theme.PullRequestStatusOpenHex),
			expectedVisibleTitle:    " acme/widgets#42 Open PR",
			expectedReferencePrefix: foregroundColorEscape(theme.PullRequestReferenceHex),
			expectedTitlePrefix:     foregroundColorEscape(theme.PullRequestTitleHex),
		},
		{
			name: "draft",
			pullRequest: githubcli.PullRequest{
				Title:      "Draft PR",
				Number:     43,
				Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
				State:      "OPEN",
				IsDraft:    true,
			},
			expectedIconText:        " ",
			expectedIconPrefix:      foregroundColorEscape(theme.PullRequestStatusDraftHex),
			expectedVisibleTitle:    " acme/widgets#43 Draft PR",
			expectedReferencePrefix: foregroundColorEscape(theme.PullRequestReferenceHex),
			expectedTitlePrefix:     foregroundColorEscape(theme.PullRequestTitleHex),
		},
		{
			name: "closed",
			pullRequest: githubcli.PullRequest{
				Title:      "Closed PR",
				Number:     44,
				Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
				State:      "CLOSED",
			},
			expectedIconText:        " ",
			expectedIconPrefix:      foregroundColorEscape(theme.PullRequestStatusClosedHex),
			expectedVisibleTitle:    " acme/widgets#44 Closed PR",
			expectedReferencePrefix: foregroundColorEscape(theme.PullRequestReferenceHex),
			expectedTitlePrefix:     foregroundColorEscape(theme.PullRequestTitleHex),
		},
		{
			name: "closed draft",
			pullRequest: githubcli.PullRequest{
				Title:      "Closed draft PR",
				Number:     46,
				Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
				State:      "CLOSED",
				IsDraft:    true,
			},
			expectedIconText:        " ",
			expectedIconPrefix:      foregroundColorEscape(theme.PullRequestStatusClosedHex),
			expectedVisibleTitle:    " acme/widgets#46 Closed draft PR",
			expectedReferencePrefix: foregroundColorEscape(theme.PullRequestReferenceHex),
			expectedTitlePrefix:     foregroundColorEscape(theme.PullRequestTitleHex),
		},
		{
			name: "merged",
			pullRequest: githubcli.PullRequest{
				Title:      "Merged PR",
				Number:     45,
				Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
				State:      "MERGED",
			},
			expectedIconText:        " ",
			expectedIconPrefix:      foregroundColorEscape(theme.PullRequestStatusMergedHex),
			expectedVisibleTitle:    " acme/widgets#45 Merged PR",
			expectedReferencePrefix: foregroundColorEscape(theme.PullRequestReferenceHex),
			expectedTitlePrefix:     foregroundColorEscape(theme.PullRequestTitleHex),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := pullRequestRow(testCase.pullRequest).Item

			if actual.Title != testCase.expectedVisibleTitle {
				t.Fatalf("expected title %q, actual %q", testCase.expectedVisibleTitle, actual.Title)
			}
			if len(actual.TitleSegments) != 3 {
				t.Fatalf("expected 3 title segments, actual %d", len(actual.TitleSegments))
			}
			if actual.TitleSegments[0].Text != testCase.expectedIconText {
				t.Fatalf("expected icon segment text %q, actual %q", testCase.expectedIconText, actual.TitleSegments[0].Text)
			}
			if actual.TitleSegments[0].Prefix != testCase.expectedIconPrefix {
				t.Fatalf("expected icon segment prefix %q, actual %q", testCase.expectedIconPrefix, actual.TitleSegments[0].Prefix)
			}
			if actual.TitleSegments[1].Text != fmt.Sprintf("acme/widgets#%d", testCase.pullRequest.Number) {
				t.Fatalf("expected reference segment %q, actual %q", fmt.Sprintf("acme/widgets#%d", testCase.pullRequest.Number), actual.TitleSegments[1].Text)
			}
			if actual.TitleSegments[1].Prefix != testCase.expectedReferencePrefix {
				t.Fatalf("expected reference segment prefix %q, actual %q", testCase.expectedReferencePrefix, actual.TitleSegments[1].Prefix)
			}
			if actual.TitleSegments[2].Text != " "+testCase.pullRequest.Title {
				t.Fatalf("expected title segment text %q, actual %q", " "+testCase.pullRequest.Title, actual.TitleSegments[2].Text)
			}
			if actual.TitleSegments[2].Prefix != testCase.expectedTitlePrefix {
				t.Fatalf("expected title segment prefix %q, actual %q", testCase.expectedTitlePrefix, actual.TitleSegments[2].Prefix)
			}
		})
	}
}

func TestPullRequestRow_GivenSuccessfulMergeChecks_WhenBuildingTheListRow_ThenItUsesTheSuccessBackgroundForEachTitleSegment(t *testing.T) {
	actual := pullRequestRow(githubcli.PullRequest{
		Title:                  "Approved PR",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "APPROVED",
		Mergeable:              "MERGEABLE",
		MergeStateStatus:       "CLEAN",
		StatusCheckRollupState: "SUCCESS",
	}).Item

	if actual.Title != " acme/widgets#42 Approved PR" {
		t.Fatalf("expected title %q, actual %q", " acme/widgets#42 Approved PR", actual.Title)
	}
	if len(actual.TitleSegments) != 3 {
		t.Fatalf("expected 3 title segments, actual %d", len(actual.TitleSegments))
	}

	expectedBackground := backgroundColorEscape(theme.SuccessBackgroundHex)
	for index, segment := range actual.TitleSegments {
		if !strings.Contains(segment.Prefix, expectedBackground) {
			t.Fatalf("expected title segment %d prefix %q to contain %q", index, segment.Prefix, expectedBackground)
		}
	}
}

func TestPullRequestRow_GivenApprovedReviewsWithoutPassingMergeChecks_WhenBuildingTheListRow_ThenItKeepsTheDefaultBackground(t *testing.T) {
	actual := pullRequestRow(githubcli.PullRequest{
		Title:                  "Waiting PR",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "APPROVED",
		Mergeable:              "MERGEABLE",
		MergeStateStatus:       "CLEAN",
		StatusCheckRollupState: "PENDING",
	}).Item

	if actual.Title != " acme/widgets#42 Waiting PR" {
		t.Fatalf("expected title %q, actual %q", " acme/widgets#42 Waiting PR", actual.Title)
	}
	unexpectedBackground := backgroundColorEscape(theme.SuccessBackgroundHex)
	for index, segment := range actual.TitleSegments {
		if strings.Contains(segment.Prefix, unexpectedBackground) {
			t.Fatalf("expected title segment %d prefix %q to avoid %q", index, segment.Prefix, unexpectedBackground)
		}
	}
}

func TestPullRequestRow_GivenApprovedReviewsWithRequestedTeamReview_WhenBuildingTheListRow_ThenItKeepsTheDefaultBackground(t *testing.T) {
	actual := pullRequestRow(githubcli.PullRequest{
		Title:                  "Waiting on team",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "APPROVED",
		ReviewRequests:         []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "P3C", Slug: "p3c", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}}},
		Mergeable:              "MERGEABLE",
		MergeStateStatus:       "CLEAN",
		StatusCheckRollupState: "SUCCESS",
	}).Item

	if actual.Title != " acme/widgets#42 Waiting on team" {
		t.Fatalf("expected title %q, actual %q", " acme/widgets#42 Waiting on team", actual.Title)
	}
	for index, unexpectedBackground := range []string{backgroundColorEscape(theme.SuccessBackgroundHex), backgroundColorEscape(theme.FailureBackgroundHex)} {
		for segmentIndex, segment := range actual.TitleSegments {
			if strings.Contains(segment.Prefix, unexpectedBackground) {
				t.Fatalf("expected title segment %d prefix %q to avoid unexpected background %d %q", segmentIndex, segment.Prefix, index, unexpectedBackground)
			}
		}
	}
}

func TestPullRequestRow_GivenBlockedMergeStateWithPassingReviewsAndChecks_WhenBuildingTheListRow_ThenItKeepsTheDefaultBackground(t *testing.T) {
	actual := pullRequestRow(githubcli.PullRequest{
		Title:                  "Blocked by merge checks",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "APPROVED",
		Mergeable:              "MERGEABLE",
		MergeStateStatus:       "BLOCKED",
		StatusCheckRollupState: "SUCCESS",
	}).Item

	if actual.Title != " acme/widgets#42 Blocked by merge checks" {
		t.Fatalf("expected title %q, actual %q", " acme/widgets#42 Blocked by merge checks", actual.Title)
	}
	for index, unexpectedBackground := range []string{backgroundColorEscape(theme.SuccessBackgroundHex), backgroundColorEscape(theme.FailureBackgroundHex)} {
		for segmentIndex, segment := range actual.TitleSegments {
			if strings.Contains(segment.Prefix, unexpectedBackground) {
				t.Fatalf("expected title segment %d prefix %q to avoid unexpected background %d %q", segmentIndex, segment.Prefix, index, unexpectedBackground)
			}
		}
	}
}

func TestPullRequestRow_GivenFailingMergeChecks_WhenBuildingTheListRow_ThenItUsesTheFailureBackgroundForEachTitleSegment(t *testing.T) {
	actual := pullRequestRow(githubcli.PullRequest{
		Title:                  "Blocked PR",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "CHANGES_REQUESTED",
		MergeStateStatus:       "BLOCKED",
		StatusCheckRollupState: "FAILURE",
	}).Item

	if actual.Title != " acme/widgets#42 Blocked PR" {
		t.Fatalf("expected title %q, actual %q", " acme/widgets#42 Blocked PR", actual.Title)
	}
	if len(actual.TitleSegments) != 3 {
		t.Fatalf("expected 3 title segments, actual %d", len(actual.TitleSegments))
	}

	expectedBackground := backgroundColorEscape(theme.FailureBackgroundHex)
	for index, segment := range actual.TitleSegments {
		if !strings.Contains(segment.Prefix, expectedBackground) {
			t.Fatalf("expected title segment %d prefix %q to contain %q", index, segment.Prefix, expectedBackground)
		}
	}
}

func TestPullRequestRow_GivenRequestedReviewTeams_WhenBuildingTheListRow_ThenItDoesNotAppendThemToTheTitle(t *testing.T) {
	actual := pullRequestRow(githubcli.PullRequest{
		Title:      "Need teams",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
		ReviewRequests: []githubcli.PullRequestReviewRequest{
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-one"}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "VIBE", Slug: "vibe", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "P3C", Slug: "p3c", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Name: "FYP", Slug: "fyp", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
		},
	}).Item

	if actual.Title != " acme/widgets#42 Need teams" {
		t.Fatalf("expected title %q, actual %q", " acme/widgets#42 Need teams", actual.Title)
	}
	if len(actual.TitleSegments) != 3 {
		t.Fatalf("expected 3 title segments, actual %d", len(actual.TitleSegments))
	}
	if strings.Contains(actual.Title, "VIBE") || strings.Contains(actual.Title, detailReviewRequestsIcon) {
		t.Fatalf("expected title to omit requested review teams, actual %q", actual.Title)
	}
}

func TestMyPullRequestsErrorItem_GivenAnAuthenticationError_WhenBuildingTheState_ThenItShowsTheRecoveryMessage(t *testing.T) {
	actual := myPullRequestsErrorItem(fmt.Errorf("wrap: %w", githubcli.ErrUnauthenticated))

	if actual.Title != myPullRequestsUnauthenticatedTitle {
		t.Fatalf("expected title %q, actual %q", myPullRequestsUnauthenticatedTitle, actual.Title)
	}
	if actual.Detail != myPullRequestsUnauthenticatedDetail {
		t.Fatalf("expected detail %q, actual %q", myPullRequestsUnauthenticatedDetail, actual.Detail)
	}
}

func TestDefaultSeedData_GivenAFreshModel_WhenReadingRequestedPullRequests_ThenItStartsInALoadingState(t *testing.T) {
	subject := NewModel(DefaultSeedData())

	actualPullRequests := subject.PullRequests(RequestedPullRequestsTab)
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	if actualPullRequests[0].Title != requestedPullRequestsLoadingTitle {
		t.Fatalf("expected title %q, actual %q", requestedPullRequestsLoadingTitle, actualPullRequests[0].Title)
	}
	if actualPullRequests[0].Detail != requestedPullRequestsLoadingDetail {
		t.Fatalf("expected detail %q, actual %q", requestedPullRequestsLoadingDetail, actualPullRequests[0].Detail)
	}
}

func TestSetPullRequests_GivenRequestedPullRequests_WhenSelectingTheRequestedTab_ThenDetailContentShowsMetadataAndBody(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetPullRequests(RequestedPullRequestsTab, []Item{requestedPullRequestItem(githubcli.PullRequest{
		Title:      "feat(postmortems): integrate post-mortem writing guide",
		Number:     845,
		Repository: githubcli.Repository{NameWithOwner: "acme/foobar"},
		URL:        "https://github.com/acme/foobar/pull/845",
		Body:       "## Summary\n\n- Adds new skill",
		State:      "open",
		IsDraft:    false,
		UpdatedAt:  "2026-04-17T20:35:05Z",
	})})
	subject.FocusPullRequestsView()
	subject.NextPullRequestTab()

	actualDetail := subject.DetailContent()
	expectedFragments := []string{
		"Repository: acme/foobar",
		"Number: #845",
		"State: open",
		"Draft: no",
		"Updated: 2026-04-17T20:35:05Z",
		"URL: https://github.com/acme/foobar/pull/845",
		"## Summary",
	}
	for _, expected := range expectedFragments {
		if !strings.Contains(actualDetail, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actualDetail)
		}
	}
}

func TestRequestedPullRequestsErrorItem_GivenAnAuthenticationError_WhenBuildingTheState_ThenItShowsTheRecoveryMessage(t *testing.T) {
	actual := requestedPullRequestsErrorItem(fmt.Errorf("wrap: %w", githubcli.ErrUnauthenticated))

	if actual.Title != requestedPullRequestsUnauthenticatedTitle {
		t.Fatalf("expected title %q, actual %q", requestedPullRequestsUnauthenticatedTitle, actual.Title)
	}
	if actual.Detail != requestedPullRequestsUnauthenticatedDetail {
		t.Fatalf("expected detail %q, actual %q", requestedPullRequestsUnauthenticatedDetail, actual.Detail)
	}
}
