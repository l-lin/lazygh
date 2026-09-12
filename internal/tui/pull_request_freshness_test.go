package tui

import (
	"errors"
	"strings"
	"testing"

	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestApplyLoadedPullRequestRows_GivenASelectedPullRequestMovesOnRefresh_WhenApplyingRows_ThenSelectionFollowsItsIdentity(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.model.FocusPullRequestsView()
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{
		givenFreshnessTestPullRequest(10, "2026-05-05T12:00:00Z"),
		givenFreshnessTestPullRequest(11, "2026-05-05T11:00:00Z"),
	})
	subject.model.SelectPullRequestIndex(MyPullRequestsTab, 1)

	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{
		givenFreshnessTestPullRequest(11, "2026-05-05T13:00:00Z"),
		givenFreshnessTestPullRequest(10, "2026-05-05T12:00:00Z"),
	})

	actualIndex := subject.model.SelectedPullRequestIndex(MyPullRequestsTab)
	if actualIndex != 0 {
		t.Fatalf("expected selected PR to follow its identity to index %d, actual %d", 0, actualIndex)
	}
	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok || actualRow.Summary == nil || actualRow.Summary.Number != 11 {
		t.Fatalf("expected PR #11 to remain selected, actual %+v", actualRow)
	}
}

func TestApplyLoadedPullRequestRows_GivenUnsortedPullRequests_WhenApplyingRows_ThenItSortsByUpdatedAtAndIdentity(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))

	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{
		givenFreshnessTestPullRequest(11, "2026-05-05T11:00:00Z"),
		givenFreshnessTestPullRequest(10, "2026-05-05T12:00:00Z"),
		givenFreshnessTestPullRequest(9, "2026-05-05T12:00:00Z"),
	})

	actualRows := subject.model.PullRequestRows(MyPullRequestsTab)
	actualNumbers := []int{actualRows[0].Summary.Number, actualRows[1].Summary.Number, actualRows[2].Summary.Number}
	expectedNumbers := []int{9, 10, 11}
	for index, expected := range expectedNumbers {
		if actualNumbers[index] != expected {
			t.Fatalf("expected row %d to contain PR #%d, actual numbers %v", index, expected, actualNumbers)
		}
	}
}

func TestPullRequestsLoaded_GivenAnExistingListAndARefreshError_WhenApplyingTheResult_ThenItPreservesTheListAndReportsTheErrorInTheStatusLine(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.model.FocusPullRequestsView()
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z")})

	subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{Tab: MyPullRequestsTab, Err: errors.New("network unavailable")})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok || actualRow.Summary == nil || actualRow.Summary.Number != 42 {
		t.Fatalf("expected the existing PR to remain selected, actual %+v", actualRow)
	}
	if subject.feedbackMessage != pullRequestListRefreshErrorPrefix+"network unavailable" {
		t.Fatalf("expected refresh error feedback, actual %q", subject.feedbackMessage)
	}
}

func TestApplyLoadedPullRequestRows_GivenANewPullRequest_WhenApplyingRows_ThenItShowsTheNeutralUnreadMarker(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z")})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok {
		t.Fatal("expected a selected pull request")
	}
	if !actualRow.Unread {
		t.Fatal("expected a newly loaded pull request to be unread")
	}
	expectedPrefix := iconPullRequestUnread + " "
	if len(actualRow.Item.Title) < len(expectedPrefix) || actualRow.Item.Title[:len(expectedPrefix)] != expectedPrefix {
		t.Fatalf("expected unread marker prefix %q, actual title %q", expectedPrefix, actualRow.Item.Title)
	}
}

func TestPullRequestFreshness_GivenANewPRInViewTwo_WhenItsDetailFinishesLoading_ThenItRemainsUnreadUntilViewZeroReceivesFocus(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.model.FocusPullRequestsView()
	summary := givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z")
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{summary})

	subject.refreshLoadedPullRequestSummaryFromDetail(summary, githubdomain.PullRequestDetail{
		Title:     "Detailed PR",
		UpdatedAt: summary.UpdatedAt,
	})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok {
		t.Fatal("expected a selected pull request")
	}
	if !actualRow.Unread {
		t.Fatal("expected detail loading to preserve the unread marker while view two is focused")
	}

	subject.model.FocusDetailView()
	subject.markCurrentPullRequestSeen()
	actualRow, ok = subject.model.SelectedPullRequestRow()
	if !ok {
		t.Fatal("expected a selected pull request after focusing view zero")
	}
	if actualRow.Unread {
		t.Fatal("expected focusing view zero to clear the unread marker")
	}
}

func TestPullRequestFreshness_GivenANewPRWhileViewZeroIsFocused_WhenApplyingRows_ThenItIsReadImmediately(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.model.FocusPullRequestsView()
	subject.model.FocusDetailView()
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z")})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok {
		t.Fatal("expected a selected pull request")
	}
	if actualRow.Unread {
		t.Fatal("expected a new pull request arriving in view zero to be read immediately")
	}
}

func TestPullRequestFreshness_GivenASeenVersion_WhenTheRemoteVersionIsNewer_ThenItMarksTheRowUnread(t *testing.T) {
	state := pullRequestFreshnessState{seen: true, seenUpdatedAt: "2026-05-05T10:00:00Z"}

	actual := state.isUnread("2026-05-05T10:05:00Z")

	if !actual {
		t.Fatal("expected a newer remote version to be unread")
	}
}

func TestPullRequestFreshness_GivenANewerSeenVersion_WhenAnOlderVersionIsMarkedSeen_ThenItKeepsTheNewerVersion(t *testing.T) {
	store := *newPullRequestListStore(nil)
	store = store.withPullRequestSeen("acme/widgets#42", "2026-05-05T11:00:00Z")
	store = store.withPullRequestSeen("acme/widgets#42", "2026-05-05T10:30:00Z")

	actual := store.pullRequestFreshnessFor("acme/widgets#42")

	if actual.seenUpdatedAt != "2026-05-05T11:00:00Z" {
		t.Fatalf("expected the newer seen version to remain in memory, actual %+v", actual)
	}
}

func TestPullRequestFreshness_GivenAPersistedSeenVersion_WhenApplyingTheSnapshot_ThenItRestoresTheReadState(t *testing.T) {
	subject := NewProgram()
	subject.applyPullRequestFreshnessSnapshot([]persistcache.PullRequestFreshness{{
		Repository:    "acme/widgets",
		Number:        42,
		Seen:          true,
		SeenUpdatedAt: "2026-05-05T10:30:00Z",
	}})

	actual := subject.pullRequestFreshnessFor("acme/widgets#42")
	if !actual.seen || actual.seenUpdatedAt != "2026-05-05T10:30:00Z" {
		t.Fatalf("expected the persisted read state to be restored, actual %+v", actual)
	}
}

func TestPullRequestFreshness_GivenAStalePersistedSnapshot_WhenApplyingItAfterLocalReads_ThenItKeepsTheNewestLocalVersion(t *testing.T) {
	subject := NewProgram()
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestFreshnessTrackingEnabled().withPullRequestSeen("acme/widgets#42", "2026-05-05T11:00:00Z")
	})

	subject.applyPullRequestFreshnessSnapshot([]persistcache.PullRequestFreshness{{
		Repository:    "acme/widgets",
		Number:        42,
		Seen:          true,
		SeenUpdatedAt: "2026-05-05T10:30:00Z",
	}})

	actual := subject.pullRequestFreshnessFor("acme/widgets#42")
	if actual.seenUpdatedAt != "2026-05-05T11:00:00Z" {
		t.Fatalf("expected the newest local version to survive a stale snapshot, actual %+v", actual)
	}
}

func TestPullRequestRow_GivenUnreadSuccessfulMergeChecks_WhenRenderingTheMarker_ThenItUsesTheRowBackgroundAndReadableForeground(t *testing.T) {
	row := pullRequestRowWithUnread(appconfig.RepositoryStyleOwnerName, githubcli.PullRequest{
		Title:                  "Approved PR",
		Number:                 42,
		Repository:             githubcli.Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "APPROVED",
		Mergeable:              "MERGEABLE",
		MergeStateStatus:       "CLEAN",
		StatusCheckRollupState: "SUCCESS",
	}, true)

	if len(row.Item.TitleSegments) == 0 || row.Item.TitleSegments[0].Text != iconPullRequestUnread+" " {
		t.Fatalf("expected the unread marker to be the first title segment, actual %+v", row.Item.TitleSegments)
	}
	if row.Item.TitleSegments[0].BackgroundHex != theme.SuccessBackgroundHex {
		t.Fatalf("expected the unread marker background %q, actual %q", theme.SuccessBackgroundHex, row.Item.TitleSegments[0].BackgroundHex)
	}
	actual := renderStyledItemTitle(row.Item.Title, row.Item.TitleSegments, "", false)
	if !strings.Contains(actual, backgroundColorEscape(theme.SuccessBackgroundHex)) {
		t.Fatalf("expected rendered unread marker to include the success background, actual %q", actual)
	}
	if !strings.Contains(actual, foregroundColorEscape(theme.ActiveTextHex)) && !strings.Contains(actual, foregroundColorEscape(theme.InactiveTextHex)) {
		t.Fatalf("expected rendered unread marker to include a readable foreground, actual %q", actual)
	}
}

func TestApplyLoadedPullRequestRows_GivenTheSelectedPullRequestDisappears_WhenApplyingRows_ThenItSelectsTheFirstRemainingRow(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.model.FocusPullRequestsView()
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{
		givenFreshnessTestPullRequest(10, "2026-05-05T12:00:00Z"),
		givenFreshnessTestPullRequest(11, "2026-05-05T11:00:00Z"),
	})
	subject.model.SelectPullRequestIndex(MyPullRequestsTab, 1)

	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{
		givenFreshnessTestPullRequest(12, "2026-05-05T13:00:00Z"),
	})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok || actualRow.Summary == nil || actualRow.Summary.Number != 12 {
		t.Fatalf("expected the first remaining PR to be selected, actual %+v", actualRow)
	}
}

func TestPullRequestsLoaded_GivenAnOlderRefreshCompletesAfterANewerRefresh_WhenApplyingTheOlderResult_ThenItLeavesTheNewerRowsUntouched(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.applyPullRequestsLoadPlanned(MsgPullRequestsLoadPlanned{Tab: MyPullRequestsTab})
	firstGeneration := subject.pullRequestLoadGeneration(MyPullRequestsTab)
	subject.applyPullRequestsLoadPlanned(MsgPullRequestsLoadPlanned{Tab: MyPullRequestsTab})
	secondGeneration := subject.pullRequestLoadGeneration(MyPullRequestsTab)
	if secondGeneration <= firstGeneration {
		t.Fatalf("expected refresh generation to increase, first %d second %d", firstGeneration, secondGeneration)
	}

	subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{
		Tab:        MyPullRequestsTab,
		Generation: secondGeneration,
		PullRequests: []githubdomain.PullRequest{
			givenFreshnessTestPullRequest(12, "2026-05-05T13:00:00Z"),
		},
	})
	subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{
		Tab:        MyPullRequestsTab,
		Generation: firstGeneration,
		PullRequests: []githubdomain.PullRequest{
			givenFreshnessTestPullRequest(10, "2026-05-05T10:00:00Z"),
		},
	})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok || actualRow.Summary == nil || actualRow.Summary.Number != 12 {
		t.Fatalf("expected the stale result to be ignored, actual %+v", actualRow)
	}
}

func TestApplyLoadedPullRequestRows_GivenTheSelectedPullRequestIsVisibleInViewZero_WhenItsRemoteVersionChanges_ThenItRemainsSeen(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.model.FocusPullRequestsView()
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z")})
	subject.model.FocusDetailView()
	subject.markCurrentPullRequestSeen()

	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(42, "2026-05-05T13:00:00Z")})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok {
		t.Fatal("expected a selected pull request")
	}
	if actualRow.Unread {
		t.Fatal("expected a remotely updated PR visible in view zero to remain seen")
	}
}

func TestPullRequestsLoaded_GivenAListLoadIsInvalidated_WhenItsOldResultArrives_ThenItLeavesTheCurrentRowsUntouched(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z")})
	subject.applyPullRequestsLoadPlanned(MsgPullRequestsLoadPlanned{Tab: MyPullRequestsTab})
	staleGeneration := subject.pullRequestLoadGeneration(MyPullRequestsTab)
	subject.resetPullRequestListLoadState()

	subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{
		Tab:        MyPullRequestsTab,
		Generation: staleGeneration,
		PullRequests: []githubdomain.PullRequest{
			givenFreshnessTestPullRequest(99, "2026-05-05T13:00:00Z"),
		},
	})

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok || actualRow.Summary == nil || actualRow.Summary.Number != 42 {
		t.Fatalf("expected the invalidated result to be ignored, actual %+v", actualRow)
	}
}

func TestPullRequestFreshness_GivenExistingState_WhenReplacingThePersistentCache_ThenItDoesNotLeakIntoTheNewCache(t *testing.T) {
	subject := NewProgram()
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestFreshnessTrackingEnabled().withPullRequestSeen("acme/widgets#42", "2026-05-05T12:00:00Z")
	})

	subject.applyCacheConfigApplied(MsgCacheConfigApplied{})

	actual := subject.pullRequestFreshnessFor("acme/widgets#42")
	if actual.seen || actual.seenUpdatedAt != "" {
		t.Fatalf("expected the old cache freshness to be cleared, actual %+v", actual)
	}
}

func TestPullRequestFreshness_GivenAConfiguredSearchBeforeTheCacheIsApplied_WhenTheCacheIsApplied_ThenItQueuesMembershipReconciliation(t *testing.T) {
	subject := NewProgram()
	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{{Label: "Mine", Command: []string{"search", "prs", "--author", "@me"}}})
	subject.persistentCacheRuntime.pending = nil

	subject.applyCacheConfigApplied(MsgCacheConfigApplied{PullRequestCache: &fakePersistentPullRequestCache{}})

	found := false
	for _, action := range subject.persistentCacheRuntime.pending {
		if _, ok := action.(reconcilePullRequestSearchesPersistentCacheAction); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected cache configuration to queue membership reconciliation")
	}
}

func TestPullRequestCustomSearch_GivenAnExistingCustomSearch_WhenReplacingIt_ThenItQueuesMembershipReconciliation(t *testing.T) {
	subject := NewProgram()
	searches := append(appconfig.DefaultPullRequestSearches(), appconfig.PullRequestSearch{
		Label:   pullRequestCustomSearchLabel,
		Command: []string{"search", "prs", "--state", "open"},
	})
	subject.setRuntimePullRequestSearches(searches)
	subject.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(searches))
	subject.persistentCacheRuntime.pending = nil

	subject.upsertPullRequestCustomSearch(appconfig.PullRequestSearch{
		Label:   pullRequestCustomSearchLabel,
		Command: []string{"search", "prs", "--state", "closed"},
	})

	found := false
	for _, action := range subject.persistentCacheRuntime.pending {
		if _, ok := action.(reconcilePullRequestSearchesPersistentCacheAction); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected replacing a custom search to queue membership reconciliation")
	}
}

func TestPastedPullRequestFreshness_GivenASeenPullRequest_WhenRemovingIt_ThenItRemovesTheInMemoryFreshness(t *testing.T) {
	subject := NewProgram()
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(10, "2026-05-05T10:00:00Z")})
	subject.applyLoadedPullRequestRows(RequestedPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(11, "2026-05-05T11:00:00Z")})
	subject.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestAdded(givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z"))
	})
	subject.syncPastedPullRequestTab()
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestSeen("acme/widgets#42", "2026-05-05T12:00:00Z")
	})
	subject.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestRemoved(givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z"))
	})

	subject.syncPastedPullRequestTab()

	actual := subject.pullRequestFreshnessFor("acme/widgets#42")
	if actual.seen || actual.seenUpdatedAt != "" {
		t.Fatalf("expected removed pasted PR freshness to be cleared, actual %+v", actual)
	}
}

func TestPullRequestsLoaded_GivenOneTabRefreshFailsAndAnotherSucceeds_WhenApplyingBothResults_ThenItKeepsTheFailedTabErrorVisible(t *testing.T) {
	subject := NewProgram()
	subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{Tab: MyPullRequestsTab, Err: errors.New("network unavailable")})
	subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{
		Tab: RequestedPullRequestsTab,
		PullRequests: []githubdomain.PullRequest{
			givenFreshnessTestPullRequest(11, "2026-05-05T11:00:00Z"),
		},
	})

	if subject.feedbackMessage != pullRequestListRefreshErrorPrefix+"network unavailable" {
		t.Fatalf("expected the failed tab error to remain visible, actual %q", subject.feedbackMessage)
	}

	subject.applyPullRequestsLoaded(MsgPullRequestsLoaded{
		Tab: MyPullRequestsTab,
		PullRequests: []githubdomain.PullRequest{
			givenFreshnessTestPullRequest(10, "2026-05-05T10:00:00Z"),
		},
	})
	if subject.feedbackMessage != "" {
		t.Fatalf("expected the failed tab error to clear after its tab succeeds, actual %q", subject.feedbackMessage)
	}
}

func TestPullRequestFreshness_GivenViewZeroFocus_WhenMarkingTheSelectedPullRequestSeen_ThenItClearsTheMarker(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "Mine"}}}))
	subject.model.FocusPullRequestsView()
	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{givenFreshnessTestPullRequest(42, "2026-05-05T12:00:00Z")})
	subject.model.FocusDetailView()

	subject.markCurrentPullRequestSeen()

	actualRow, ok := subject.model.SelectedPullRequestRow()
	if !ok {
		t.Fatal("expected a selected pull request")
	}
	if actualRow.Unread {
		t.Fatal("expected focusing view zero to clear the unread marker")
	}
}

func givenFreshnessTestPullRequest(number int, updatedAt string) githubdomain.PullRequest {
	return githubdomain.PullRequest{
		Number:     number,
		Title:      "PR",
		Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
		State:      "OPEN",
		UpdatedAt:  updatedAt,
	}
}
