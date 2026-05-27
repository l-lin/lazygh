package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestPullRequestDetailMissingBrowserTabData_GivenCompletedBuildWithoutLink_WhenChecking_ThenItRequiresARefresh(t *testing.T) {
	actual := pullRequestDetailMissingBrowserTabData(githubcli.PullRequestDetail{
		Commits:           []githubcli.PullRequestCommit{{OID: "abc123", MessageHeadline: "hydrate cache"}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"}},
	})

	if !actual {
		t.Fatal("expected completed builds without links to require a refresh")
	}
}
func TestLayout_GivenCachedPullRequests_WhenRendering_ThenItShowsThemBeforeTheBackgroundRefreshFinishes(t *testing.T) {
	cachedPullRequests := []githubcli.PullRequest{{Title: "Cached PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Cached body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{{Title: "Fresh PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Fresh body", State: "OPEN", UpdatedAt: "2026-05-05T10:05:00Z"}}}}
	cache := &fakePersistentPullRequestCache{pullRequestsBySearchKey: map[string][]githubcli.PullRequest{fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0]): cachedPullRequests}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.model.FocusUserView()
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	expected := iconPullRequest + " acme/widgets#42 Cached PR"
	if actualLine, ok := pullRequestsView.Line(0); !ok || !strings.Contains(actualLine, expected) {
		t.Fatalf("expected pull requests line %q to contain %q", actualLine, expected)
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued pull request refresh, actual %d", len(asyncRunner.runs))
	}
}
func TestLayout_GivenCachedPullRequestsAndBackgroundRefreshFailure_WhenRendering_ThenItKeepsTheCachedRowsVisible(t *testing.T) {
	cachedPullRequests := []githubcli.PullRequest{{Title: "Cached PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Cached body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{}, listErr: errors.New("boom")}
	cache := &fakePersistentPullRequestCache{pullRequestsBySearchKey: map[string][]githubcli.PullRequest{fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0]): cachedPullRequests}}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.model.FocusUserView()
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "Cached PR") {
		t.Fatalf("expected pull requests buffer to contain %q after refresh failure, actual %q", "Cached PR", pullRequestsView.Buffer())
	}
	if strings.Contains(pullRequestsView.Buffer(), myPullRequestsGenericErrorTitle) {
		t.Fatalf("expected cached pull requests to stay visible instead of %q, actual %q", myPullRequestsGenericErrorTitle, pullRequestsView.Buffer())
	}
}
func TestLayout_GivenPersistentPullRequestInvalidationAndANewSession_WhenRendering_ThenItStillShowsTheCachedRowsBeforeRefreshing(t *testing.T) {
	cachedPullRequests := []githubcli.PullRequest{{Title: "Cached PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Cached body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}}
	cache := &fakePersistentPullRequestCache{pullRequestsBySearchKey: map[string][]githubcli.PullRequest{fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0]): cachedPullRequests}}
	invalidatingSubject := NewProgram()
	invalidatingSubject.pullRequestCache = cache

	invalidatingSubject.invalidatePullRequestDetail("acme/widgets", 42)

	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{{Title: "Fresh PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Fresh body", State: "OPEN", UpdatedAt: "2026-05-05T10:05:00Z"}}}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.model.FocusUserView()
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "Cached PR") {
		t.Fatalf("expected pull requests buffer to contain %q after reopening, actual %q", "Cached PR", pullRequestsView.Buffer())
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued pull request refresh, actual %d", len(asyncRunner.runs))
	}
}
func TestLayout_GivenCachedPullRequestsAndLiveResultsInSearchOrder_WhenRendering_ThenItKeepsTheLiveRowOrder(t *testing.T) {
	cachedPullRequests := []githubcli.PullRequest{
		{Title: "Cached older", Number: 41, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/41", Body: "Cached body", State: "OPEN", UpdatedAt: "2026-05-05T09:00:00Z"},
		{Title: "Cached newer", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Cached body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"},
	}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{
		{Title: "Older live PR", Number: 41, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/41", Body: "Fresh body", State: "OPEN", UpdatedAt: "2026-05-05T09:00:00Z"},
		{Title: "Newer live PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Fresh body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"},
	}}}
	cache := &fakePersistentPullRequestCache{pullRequestsBySearchKey: map[string][]githubcli.PullRequest{fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0]): cachedPullRequests}}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.model.FocusUserView()
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	actualBuffer := pullRequestsView.Buffer()
	olderIndex := strings.Index(actualBuffer, "Older live PR")
	newerIndex := strings.Index(actualBuffer, "Newer live PR")
	if olderIndex < 0 || newerIndex < 0 {
		t.Fatalf("expected pull requests buffer to contain both live rows, actual %q", actualBuffer)
	}
	if olderIndex > newerIndex {
		t.Fatalf("expected the live rows to keep their search order, actual %q", actualBuffer)
	}
}
func TestLoadPullRequestsCommand_GivenAFreshLiveResult_WhenExecuting_ThenItStoresTheResultInThePersistentCache(t *testing.T) {
	expected := []githubcli.PullRequest{{Title: "Fresh PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Fresh body", State: "OPEN", UpdatedAt: "2026-05-05T10:05:00Z"}}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{myPullRequests: expected}}
	cache := &fakePersistentPullRequestCache{}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	subject.executeWorkflowCommands(gui, []Cmd{loadPullRequestsCmd{tab: MyPullRequestsTab}})

	actual := cache.savedPullRequestsBySearchKey[fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0])]
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected cached pull requests %+v, actual %+v", expected, actual)
	}
}
func TestMaybeLoadSelectedPullRequestDetail_GivenACachedDetailWithAMatchingSummaryVersion_WhenCheckingTheSelection_ThenItUsesTheCachedDetailWithoutTriggeringAGhRefresh(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:00:00Z"}
	cachedDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Cached body", State: "OPEN", Commits: []githubcli.PullRequestCommit{{OID: "cached123", MessageHeadline: "Cached commit"}}}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "Fresh PR", Number: 42, Body: "Fresh body", State: "OPEN", Commits: []githubcli.PullRequestCommit{{OID: "fresh123", MessageHeadline: "Fresh commit"}}}}}}
	cache := &fakePersistentPullRequestCache{details: map[string]persistcache.CachedPullRequestDetail{"acme/widgets#42": given_cachedPersistentPullRequestDetail(cachedDetail, summary.UpdatedAt)}}
	asyncRunner := &capturingAsyncRunner{}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.startupState.appStarted = true
	subject.configureGUI(gui)

	subject.maybeLoadSelectedPullRequestDetail(gui)

	actual, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected a cached pull request detail")
	}
	if actual.detail.Body != "Cached body" {
		t.Fatalf("expected cached detail body %q, actual %q", "Cached body", actual.detail.Body)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string(nil)) {
		t.Fatalf("expected no detail refresh calls, actual %v", loader.detailCalls)
	}
	if len(asyncRunner.runs) != 0 {
		t.Fatalf("expected no queued detail refresh, actual %d", len(asyncRunner.runs))
	}
}
func TestMaybeLoadSelectedPullRequestDetail_GivenACachedDetailMissingCommitData_WhenCheckingTheSelection_ThenItRefreshesItInBackground(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:00:00Z"}
	cachedDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Cached body", State: "OPEN"}
	freshDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Fresh body", State: "OPEN", Commits: []githubcli.PullRequestCommit{{OID: "abc1234", MessageHeadline: "Fresh commit"}}}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": freshDetail}}}
	cache := &fakePersistentPullRequestCache{details: map[string]persistcache.CachedPullRequestDetail{"acme/widgets#42": given_cachedPersistentPullRequestDetail(cachedDetail, summary.UpdatedAt)}}
	asyncRunner := &capturingAsyncRunner{}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.startupState.appStarted = true
	subject.configureGUI(gui)

	subject.maybeLoadSelectedPullRequestDetail(gui)
	actualBeforeRefresh, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected the cached detail to be available immediately")
	}
	if actualBeforeRefresh.detail.Body != "Cached body" {
		t.Fatalf("expected cached detail body %q before refresh, actual %q", "Cached body", actualBeforeRefresh.detail.Body)
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued detail refresh, actual %d", len(asyncRunner.runs))
	}

	asyncRunner.runs[0]()

	actualAfterRefresh, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected a refreshed pull request detail")
	}
	if actualAfterRefresh.detail.Body != "Fresh body" || len(actualAfterRefresh.detail.Commits) != 1 {
		t.Fatalf("expected refreshed detail body %q with commits, actual %+v", "Fresh body", actualAfterRefresh.detail)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected detail refresh calls %v, actual %v", []string{"acme/widgets#42"}, loader.detailCalls)
	}
}
func TestMaybeLoadSelectedPullRequestDetail_GivenACachedDetailWithAStaleSummaryVersion_WhenCheckingTheSelection_ThenItShowsTheCachedDetailAndRefreshesItInBackground(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:05:00Z"}
	cachedDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Cached body", State: "OPEN"}
	freshDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Fresh body", State: "OPEN"}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": freshDetail}}}
	cache := &fakePersistentPullRequestCache{details: map[string]persistcache.CachedPullRequestDetail{"acme/widgets#42": given_cachedPersistentPullRequestDetail(cachedDetail, "2026-05-05T10:00:00Z")}}
	asyncRunner := &capturingAsyncRunner{}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.startupState.appStarted = true
	subject.configureGUI(gui)

	subject.maybeLoadSelectedPullRequestDetail(gui)
	actualBeforeRefresh, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected the stale cached detail to be available immediately")
	}
	if actualBeforeRefresh.detail.Body != "Cached body" {
		t.Fatalf("expected cached detail body %q before refresh, actual %q", "Cached body", actualBeforeRefresh.detail.Body)
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued detail refresh, actual %d", len(asyncRunner.runs))
	}

	asyncRunner.runs[0]()

	actualAfterRefresh, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected a refreshed pull request detail")
	}
	if actualAfterRefresh.detail.Body != "Fresh body" {
		t.Fatalf("expected refreshed detail body %q, actual %q", "Fresh body", actualAfterRefresh.detail.Body)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected detail refresh calls %v, actual %v", []string{"acme/widgets#42"}, loader.detailCalls)
	}
	if actual := cache.savedDetails["acme/widgets#42"]; actual.Detail.Body != "Fresh body" || actual.SourceUpdatedAt != summary.UpdatedAt {
		t.Fatalf("expected saved cached detail to use body %q and version %q, actual %+v", "Fresh body", summary.UpdatedAt, actual)
	}
}
func TestMaybeLoadSelectedPullRequestDetail_GivenACachedDetailAndARefreshFailure_WhenCheckingTheSelection_ThenItKeepsTheCachedDetailVisible(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:05:00Z"}
	cachedDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Cached body", State: "OPEN"}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{detailErrors: map[string]error{"acme/widgets#42": errors.New("boom")}}}
	cache := &fakePersistentPullRequestCache{details: map[string]persistcache.CachedPullRequestDetail{"acme/widgets#42": given_cachedPersistentPullRequestDetail(cachedDetail, "2026-05-05T10:00:00Z")}}
	asyncRunner := &capturingAsyncRunner{}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.startupState.appStarted = true
	subject.configureGUI(gui)

	subject.maybeLoadSelectedPullRequestDetail(gui)
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued detail refresh, actual %d", len(asyncRunner.runs))
	}

	asyncRunner.runs[0]()

	actual, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected the cached detail to stay available")
	}
	if actual.err != nil {
		t.Fatalf("expected the cached detail to survive the refresh error, actual %v", actual.err)
	}
	if actual.detail.Body != "Cached body" {
		t.Fatalf("expected cached detail body %q after refresh failure, actual %q", "Cached body", actual.detail.Body)
	}
}
func TestMaybeLoadSelectedPullRequestDiff_GivenACachedDiffWithAMatchingSummaryVersion_WhenCheckingTheReviewSession_ThenItUsesTheCachedDiffWithoutTriggeringAGhRefresh(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:00:00Z"}
	cachedDiff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}, FileTeamOwnersAttempted: true}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": {UnifiedDiff: "diff --git a/main.go b/main.go\n+fresh"}}}}
	cache := &fakePersistentPullRequestCache{diffs: map[string]persistcache.CachedPullRequestDiff{"acme/widgets#42": given_cachedPersistentPullRequestDiff(cachedDiff, summary.UpdatedAt)}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.asyncRunner = asyncRunner
	subject.startReviewSession(summary, "PRR_cache")
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.startupState.appStarted = true
	subject.configureGUI(gui)

	subject.maybeLoadSelectedPullRequestDiff(gui)

	actual, ok := subject.pullRequestDiffForSummary(summary)
	if !ok {
		t.Fatal("expected a cached pull request diff")
	}
	if len(actual.data.Files) != 1 || actual.data.Files[0].Path != "main.go" {
		t.Fatalf("expected cached diff files %+v, actual %+v", []string{"main.go"}, actual.data.Files)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string(nil)) {
		t.Fatalf("expected no diff refresh calls, actual %v", loader.diffCalls)
	}
	if len(asyncRunner.runs) != 0 {
		t.Fatalf("expected no queued diff refresh, actual %d", len(asyncRunner.runs))
	}
}
func TestMaybeLoadSelectedPullRequestDiff_GivenBrowserChangesTabAndACachedDiffWithoutAttemptedTeamOwnershipLookup_WhenCheckingTheSelection_ThenItRefreshesInBackground(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:00:00Z"}
	cachedDiff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}}
	freshDiff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+fresh", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{
		diffs:          map[string]githubcli.PullRequestDiff{"acme/widgets#42": freshDiff},
		fileTeamOwners: map[string]map[string][]string{"acme/widgets#42": {"main.go": {"P3C"}}},
	}}
	cache := &fakePersistentPullRequestCache{diffs: map[string]persistcache.CachedPullRequestDiff{"acme/widgets#42": given_cachedPersistentPullRequestDiff(cachedDiff, summary.UpdatedAt)}}
	asyncRunner := &capturingAsyncRunner{}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.detailState.activeTab = ChangesDetailTab
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.startupState.appStarted = true
	subject.configureGUI(gui)

	subject.maybeLoadSelectedPullRequestDiff(gui)

	actualBeforeRefresh, ok := subject.pullRequestDiffForSummary(summary)
	if !ok {
		t.Fatal("expected the cached diff to be available immediately")
	}
	if len(actualBeforeRefresh.data.Files) != 1 || actualBeforeRefresh.data.Files[0].Path != "main.go" {
		t.Fatalf("expected cached diff files %+v, actual %+v", []string{"main.go"}, actualBeforeRefresh.data.Files)
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued diff refresh for team ownership, actual %d", len(asyncRunner.runs))
	}

	asyncRunner.runs[0]()

	actualAfterRefresh, ok := subject.pullRequestDiffForSummary(summary)
	if !ok {
		t.Fatal("expected a refreshed pull request diff")
	}
	expectedTeamOwners := []string{"P3C"}
	if !reflect.DeepEqual(actualAfterRefresh.data.Files[0].TeamOwners, expectedTeamOwners) {
		t.Fatalf("expected refreshed team owners %+v, actual %+v", expectedTeamOwners, actualAfterRefresh.data.Files[0].TeamOwners)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected diff refresh calls %v, actual %v", []string{"acme/widgets#42"}, loader.diffCalls)
	}
	if !reflect.DeepEqual(loader.fileTeamOwnerCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected team ownership calls %v, actual %v", []string{"acme/widgets#42"}, loader.fileTeamOwnerCalls)
	}

	expectedSavedDiff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+fresh", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1, TeamOwners: []string{"P3C"}}}, FileTeamOwnersAttempted: true}
	if actual := cache.savedDiffs["acme/widgets#42"]; !reflect.DeepEqual(actual.Diff, expectedSavedDiff) || actual.SourceUpdatedAt != summary.UpdatedAt {
		t.Fatalf("expected saved cached diff %+v with version %q, actual %+v", expectedSavedDiff, summary.UpdatedAt, actual)
	}
}
func TestMaybeLoadSelectedPullRequestDiff_GivenBrowserChangesTabAndAStaleCachedDiff_WhenCheckingTheSelection_ThenItShowsTheCachedDiffAndRefreshesItInBackground(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:05:00Z"}
	cachedDiff := githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/main.go b/main.go",
			"index 1111111..2222222 100644",
			"--- a/main.go",
			"+++ b/main.go",
			"@@ -1,1 +1,1 @@",
			"-old line",
			"+cached line",
		}, "\n"),
		Files:                   []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1, Deletions: 1}},
		FileTeamOwnersAttempted: true,
	}
	freshDiff := githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/main.go b/main.go",
			"index 1111111..2222222 100644",
			"--- a/main.go",
			"+++ b/main.go",
			"@@ -1,1 +1,1 @@",
			"-old line",
			"+fresh line",
		}, "\n"),
		Files:                   []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1, Deletions: 1}},
		FileTeamOwnersAttempted: true,
	}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": freshDiff}}}
	cache := &fakePersistentPullRequestCache{diffs: map[string]persistcache.CachedPullRequestDiff{"acme/widgets#42": given_cachedPersistentPullRequestDiff(cachedDiff, "2026-05-05T10:00:00Z")}}
	asyncRunner := &capturingAsyncRunner{}
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.detailState.activeTab = ChangesDetailTab
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.startupState.appStarted = true
	subject.configureGUI(gui)

	subject.maybeLoadSelectedPullRequestDiff(gui)
	actualBeforeRefresh, ok := subject.pullRequestDiffForSummary(summary)
	if !ok {
		t.Fatal("expected the stale cached diff to be available immediately")
	}
	if len(actualBeforeRefresh.data.Files) != 1 || actualBeforeRefresh.data.Files[0].Path != "main.go" {
		t.Fatalf("expected cached diff files %+v, actual %+v", []string{"main.go"}, actualBeforeRefresh.data.Files)
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued diff refresh, actual %d", len(asyncRunner.runs))
	}

	asyncRunner.runs[0]()

	actualAfterRefresh, ok := subject.pullRequestDiffForSummary(summary)
	if !ok {
		t.Fatal("expected a refreshed pull request diff")
	}
	if len(actualAfterRefresh.data.Files) != 1 || actualAfterRefresh.data.Files[0].Path != "main.go" || len(actualAfterRefresh.data.Files[0].Hunks) != 1 || len(actualAfterRefresh.data.Files[0].Hunks[0].Lines) != 2 || actualAfterRefresh.data.Files[0].Hunks[0].Lines[1].Text != "fresh line" {
		t.Fatalf("expected refreshed diff data to contain the fresh change, actual %+v", actualAfterRefresh.data.Files)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected diff refresh calls %v, actual %v", []string{"acme/widgets#42"}, loader.diffCalls)
	}
	if actual := cache.savedDiffs["acme/widgets#42"]; !reflect.DeepEqual(actual.Diff, freshDiff) || actual.SourceUpdatedAt != summary.UpdatedAt {
		t.Fatalf("expected saved cached diff %+v with version %q, actual %+v", freshDiff, summary.UpdatedAt, actual)
	}
}
func TestUpdate_GivenLoadedPullRequestDiffResult_WhenApplying_ThenItStoresTheResultInThePersistentCache(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:00:00Z"}
	expected := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+fresh", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}, FileTeamOwnersAttempted: true}
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": expected}}}
	cache := &fakePersistentPullRequestCache{}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache

	Update(subject, loadPullRequestDiffResult(newPullRequestDiffWorkflowRuntime(subject, nil), githubcli.ToDomainPullRequestSummary(summary)))

	actual := cache.savedDiffs["acme/widgets#42"]
	if !reflect.DeepEqual(actual.Diff, expected) || actual.SourceUpdatedAt != summary.UpdatedAt {
		t.Fatalf("expected saved diff %+v with version %q, actual %+v", expected, summary.UpdatedAt, actual)
	}
}
func TestInvalidatePullRequestDetail_GivenAPersistentCache_WhenInvalidating_ThenItInvalidatesTheStoredPullRequest(t *testing.T) {
	subject := NewProgram()
	cache := &fakePersistentPullRequestCache{}
	subject.pullRequestCache = cache

	subject.invalidatePullRequestDetail("acme/widgets", 42)

	if !reflect.DeepEqual(cache.invalidatedPullRequests, []string{"acme/widgets#42"}) {
		t.Fatalf("expected invalidated pull requests %v, actual %v", []string{"acme/widgets#42"}, cache.invalidatedPullRequests)
	}
}

type cacheAwarePullRequestLoader struct {
	*fakePullRequestDetailLoader
	listErr error
}

func (loader *cacheAwarePullRequestLoader) ListPullRequests(commandArguments []string) ([]githubdomain.PullRequestSummary, error) {
	if loader.listErr != nil {
		return nil, loader.listErr
	}
	return loader.fakePullRequestDetailLoader.ListPullRequests(commandArguments)
}

type fakeSavedPullRequestDetail struct {
	Detail          githubcli.PullRequestDetail
	SourceUpdatedAt string
}

type fakeSavedPullRequestDiff struct {
	Diff            githubcli.PullRequestDiff
	SourceUpdatedAt string
}

type fakePersistentPullRequestCache struct {
	pullRequestsBySearchKey      map[string][]githubcli.PullRequest
	notifications                []githubcli.Notification
	details                      map[string]persistcache.CachedPullRequestDetail
	diffs                        map[string]persistcache.CachedPullRequestDiff
	savedPullRequestsBySearchKey map[string][]githubcli.PullRequest
	savedNotifications           []githubcli.Notification
	savedDetails                 map[string]fakeSavedPullRequestDetail
	savedDiffs                   map[string]fakeSavedPullRequestDiff
	invalidatedPullRequests      []string
	clearCalls                   int
}

func (cache *fakePersistentPullRequestCache) PullRequests(search appconfig.PullRequestSearch) ([]githubdomain.PullRequestSummary, bool, error) {
	pullRequests, ok := cache.pullRequestsBySearchKey[fakePersistentPullRequestSearchKey(search)]
	if !ok {
		return nil, false, nil
	}
	return githubcli.ToDomainPullRequests(pullRequests), true, nil
}
func (cache *fakePersistentPullRequestCache) SavePullRequests(search appconfig.PullRequestSearch, pullRequests []githubdomain.PullRequestSummary) error {
	if cache.savedPullRequestsBySearchKey == nil {
		cache.savedPullRequestsBySearchKey = map[string][]githubcli.PullRequest{}
	}
	cache.savedPullRequestsBySearchKey[fakePersistentPullRequestSearchKey(search)] = githubcli.PullRequestsFromDomain(pullRequests)
	return nil
}
func (cache *fakePersistentPullRequestCache) Notifications() ([]githubdomain.Notification, bool, error) {
	if cache.notifications == nil {
		return nil, false, nil
	}
	return githubcli.ToDomainNotifications(cache.notifications), true, nil
}
func (cache *fakePersistentPullRequestCache) SaveNotifications(notifications []githubdomain.Notification) error {
	cache.savedNotifications = githubcli.NotificationsFromDomain(notifications)
	return nil
}
func (cache *fakePersistentPullRequestCache) PullRequestDetail(repository string, number int) (persistcache.CachedPullRequestDetail, bool, error) {
	detail, ok := cache.details[strings.TrimSpace(repository)+"#"+itoa(number)]
	return detail, ok, nil
}
func (cache *fakePersistentPullRequestCache) SavePullRequestDetail(summary githubdomain.PullRequestSummary, detail githubdomain.PullRequestDetail) error {
	if cache.savedDetails == nil {
		cache.savedDetails = map[string]fakeSavedPullRequestDetail{}
	}
	cache.savedDetails[pullRequestDetailKey(githubcli.RepositoryFromDomain(summary.Repository), summary.Number)] = fakeSavedPullRequestDetail{Detail: githubcli.PullRequestDetailFromDomain(detail), SourceUpdatedAt: summary.UpdatedAt}
	return nil
}
func (cache *fakePersistentPullRequestCache) PullRequestDiff(repository string, number int) (persistcache.CachedPullRequestDiff, bool, error) {
	diff, ok := cache.diffs[strings.TrimSpace(repository)+"#"+itoa(number)]
	return diff, ok, nil
}
func (cache *fakePersistentPullRequestCache) SavePullRequestDiff(summary githubdomain.PullRequestSummary, diff githubdomain.PullRequestDiff) error {
	if cache.savedDiffs == nil {
		cache.savedDiffs = map[string]fakeSavedPullRequestDiff{}
	}
	cache.savedDiffs[pullRequestDetailKey(githubcli.RepositoryFromDomain(summary.Repository), summary.Number)] = fakeSavedPullRequestDiff{Diff: githubcli.PullRequestDiffFromDomain(diff), SourceUpdatedAt: summary.UpdatedAt}
	return nil
}
func (cache *fakePersistentPullRequestCache) InvalidatePullRequest(repository string, number int) error {
	key := strings.TrimSpace(repository) + "#" + itoa(number)
	cache.invalidatedPullRequests = append(cache.invalidatedPullRequests, key)
	delete(cache.details, key)
	delete(cache.diffs, key)
	return nil
}
func (cache *fakePersistentPullRequestCache) Clear() error {
	cache.clearCalls++
	cache.pullRequestsBySearchKey = map[string][]githubcli.PullRequest{}
	cache.notifications = nil
	cache.details = map[string]persistcache.CachedPullRequestDetail{}
	cache.diffs = map[string]persistcache.CachedPullRequestDiff{}
	return nil
}
func (cache *fakePersistentPullRequestCache) Close() error {
	return nil
}
func fakePersistentPullRequestSearchKey(search appconfig.PullRequestSearch) string {
	return strings.TrimSpace(search.Label) + "|" + strings.Join(search.Command, "\x00")
}
func given_cachedPersistentPullRequestDetail(detail githubcli.PullRequestDetail, sourceUpdatedAt string) persistcache.CachedPullRequestDetail {
	return persistcache.CachedPullRequestDetail{Detail: githubcli.ToDomainPullRequestDetail(detail), SourceUpdatedAt: sourceUpdatedAt}
}
func given_cachedPersistentPullRequestDiff(diff githubcli.PullRequestDiff, sourceUpdatedAt string) persistcache.CachedPullRequestDiff {
	return persistcache.CachedPullRequestDiff{Diff: githubcli.ToDomainPullRequestDiff(diff), SourceUpdatedAt: sourceUpdatedAt}
}
