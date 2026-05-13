package tui

import (
	"errors"
	"reflect"
	"testing"

	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestSubmitPullRequestTitleEdit_GivenLoadedSummaryAndDetail_WhenEditing_ThenItKeepsTheVisibleCachesHot(t *testing.T) {
	summary := githubcli.PullRequest{Title: "Old title", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Old body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}
	loader := &fakePullRequestDetailLoader{}
	cache := &fakePersistentPullRequestCache{}
	subject := given_workflowProgram(summary, loader)
	subject.pullRequestCache = cache
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "Old title", Number: 42, URL: summary.URL, Body: "Old body", State: "OPEN"}, sourceUpdatedAt: summary.UpdatedAt}

	actualErr := subject.submitPullRequestTitleEdit(pullRequestActionTarget{repository: "acme/widgets", number: 42}, "New title")

	then_noError(t, actualErr)
	actualRows := subject.model.PullRequestRows(MyPullRequestsTab)
	if len(actualRows) != 1 || actualRows[0].Summary == nil || actualRows[0].Summary.Title != "New title" {
		t.Fatalf("expected the visible pull request title %q, actual %+v", "New title", actualRows)
	}
	actualDetail, ok := subject.pullRequestDetailCache["acme/widgets#42"]
	if !ok {
		t.Fatal("expected the detail cache entry to stay loaded")
	}
	if actualDetail.detail.Title != "New title" {
		t.Fatalf("expected cached detail title %q, actual %+v", "New title", actualDetail.detail)
	}
	if !actualDetail.needsRefresh {
		t.Fatal("expected the cached detail to be marked for refresh")
	}
	if !reflect.DeepEqual(cache.invalidatedPullRequests, []string{"acme/widgets#42"}) {
		t.Fatalf("expected invalidated pull request keys %v, actual %v", []string{"acme/widgets#42"}, cache.invalidatedPullRequests)
	}
}

func TestSubmitPullRequestDescriptionEdit_GivenLoadedSummaryAndDetail_WhenEditing_ThenItKeepsTheVisibleCachesHot(t *testing.T) {
	summary := githubcli.PullRequest{Title: "Old title", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Old body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}
	loader := &fakePullRequestDetailLoader{}
	cache := &fakePersistentPullRequestCache{}
	subject := given_workflowProgram(summary, loader)
	subject.pullRequestCache = cache
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "Old title", Number: 42, URL: summary.URL, Body: "Old body", BodyHTML: "<p>Old body</p>", State: "OPEN"}, sourceUpdatedAt: summary.UpdatedAt}

	actualErr := subject.submitPullRequestDescriptionEdit(pullRequestActionTarget{repository: "acme/widgets", number: 42}, "New body")

	then_noError(t, actualErr)
	actualRows := subject.model.PullRequestRows(MyPullRequestsTab)
	if len(actualRows) != 1 || actualRows[0].Summary == nil || actualRows[0].Summary.Body != "New body" {
		t.Fatalf("expected the visible pull request body %q, actual %+v", "New body", actualRows)
	}
	actualDetail, ok := subject.pullRequestDetailCache["acme/widgets#42"]
	if !ok {
		t.Fatal("expected the detail cache entry to stay loaded")
	}
	if actualDetail.detail.Body != "New body" {
		t.Fatalf("expected cached detail body %q, actual %+v", "New body", actualDetail.detail)
	}
	if actualDetail.detail.BodyHTML != "" {
		t.Fatalf("expected cached detail HTML to be cleared after the optimistic edit, actual %+v", actualDetail.detail)
	}
	if !actualDetail.needsRefresh {
		t.Fatal("expected the cached detail to be marked for refresh")
	}
	if !reflect.DeepEqual(cache.invalidatedPullRequests, []string{"acme/widgets#42"}) {
		t.Fatalf("expected invalidated pull request keys %v, actual %v", []string{"acme/widgets#42"}, cache.invalidatedPullRequests)
	}
}

func TestPullRequestListStore_GivenCachedRowsAndALiveReload_WhenPlanningTheLoad_ThenItHydratesTheModelAndQueuesOneCommand(t *testing.T) {
	cachedSummary := githubcli.PullRequest{Title: "Cached PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Cached body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}
	loader := &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{{Title: "Fresh PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Fresh body", State: "OPEN", UpdatedAt: "2026-05-05T10:05:00Z"}}}
	cache := &fakePersistentPullRequestCache{pullRequestsBySearchKey: map[string][]githubcli.PullRequest{fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0]): {cachedSummary}}}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.notificationsLoadStarted = true
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actual := subject.pullRequestListStore.planLoad(subject, gui, MyPullRequestsTab)

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	actualRows := subject.model.PullRequestRows(MyPullRequestsTab)
	if len(actualRows) != 1 || actualRows[0].Summary == nil || actualRows[0].Summary.Title != "Cached PR" {
		t.Fatalf("expected cached pull request rows %+v, actual %+v", []string{"Cached PR"}, actualRows)
	}
	if actual := subject.pullRequestListStore.planLoad(subject, gui, MyPullRequestsTab); len(actual) != 0 {
		t.Fatalf("expected no second load command while the first is already planned, actual %d", len(actual))
	}
}

func TestExecuteWorkflowCommands_GivenASelectedPullRequestDetailLoadPlan_WhenRunningTheCommand_ThenItFeedsTheLoadedDetailBackIntoTheStores(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Summary body", State: "OPEN", UpdatedAt: "2026-05-05T10:05:00Z"}
	cachedDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Cached body", State: "OPEN"}
	freshDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Fresh body", State: "OPEN", Commits: []githubcli.PullRequestCommit{{OID: "abc123", MessageHeadline: "Fresh commit"}}}
	loader := &fakePullRequestDetailLoader{
		details:              map[string]githubcli.PullRequestDetail{"acme/widgets#42": freshDetail},
		reviewKeyByPendingID: map[string]string{"PRR_pending": "acme/widgets#42"},
	}
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
	subject.configureGUI(gui)

	planned := subject.detailStore.planSelectedPullRequestDetailLoad(subject, gui)

	if len(planned) != 1 {
		t.Fatalf("expected one planned detail load command, actual %d", len(planned))
	}
	actualBeforeRefresh, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected the cached detail to be hydrated before the refresh runs")
	}
	if actualBeforeRefresh.detail.Body != "Cached body" {
		t.Fatalf("expected cached detail body %q before refresh, actual %+v", "Cached body", actualBeforeRefresh.detail)
	}

	subject.executeWorkflowCommands(gui, planned)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued async run, actual %d", len(asyncRunner.runs))
	}
	if actual := subject.detailStore.planSelectedPullRequestDetailLoad(subject, gui); len(actual) != 0 {
		t.Fatalf("expected no second detail load command while the first is still in flight, actual %d", len(actual))
	}

	asyncRunner.runs[0]()

	actualAfterRefresh, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected the refreshed detail to be cached")
	}
	if actualAfterRefresh.detail.Body != "Fresh body" || len(actualAfterRefresh.detail.Commits) != 1 {
		t.Fatalf("expected the refreshed detail body %q with commits, actual %+v", "Fresh body", actualAfterRefresh.detail)
	}
	pendingState, ok := subject.pendingPullRequestReviewCache["acme/widgets#42"]
	if !ok {
		t.Fatal("expected the pending review state to be cached for review startup")
	}
	if pendingState.id != "PRR_pending" {
		t.Fatalf("expected cached pending review id %q, actual %+v", "PRR_pending", pendingState)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42"}, loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.getPendingReviewCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected pending review lookup calls %v, actual %v", []string{"acme/widgets#42"}, loader.getPendingReviewCalls)
	}
}

func TestReviewStore_GivenACachedDiffWithoutTeamOwners_WhenPlanningTheSelectedLoad_ThenItHydratesTheDiffAndQueuesOneCommand(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}
	cachedDiff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n@@ -1 +1 @@\n-old\n+new", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1, Deletions: 1}}}
	freshDiff := githubcli.PullRequestDiff{UnifiedDiff: cachedDiff.UnifiedDiff, Files: cachedDiff.Files}
	loader := &fakePullRequestDetailLoader{
		diffs:          map[string]githubcli.PullRequestDiff{"acme/widgets#42": freshDiff},
		fileTeamOwners: map[string]map[string][]string{"acme/widgets#42": {"main.go": {"@acme/platform"}}},
	}
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
	subject.activeDetailTab = ChangesDetailTab
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	planned := subject.reviewStore.planSelectedPullRequestDiffLoad(subject, gui)

	if len(planned) != 1 {
		t.Fatalf("expected one planned diff load command, actual %d", len(planned))
	}
	actualBeforeRefresh, ok := subject.pullRequestDiffForSummary(summary)
	if !ok {
		t.Fatal("expected the cached diff to be hydrated before the refresh runs")
	}
	if len(actualBeforeRefresh.data.Files) != 1 || actualBeforeRefresh.data.Files[0].Path != "main.go" {
		t.Fatalf("expected the cached diff files to stay visible, actual %+v", actualBeforeRefresh.data.Files)
	}

	subject.executeWorkflowCommands(gui, planned)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued async run, actual %d", len(asyncRunner.runs))
	}
	if actual := subject.reviewStore.planSelectedPullRequestDiffLoad(subject, gui); len(actual) != 0 {
		t.Fatalf("expected no second diff load command while the first is still in flight, actual %d", len(actual))
	}

	asyncRunner.runs[0]()

	actualAfterRefresh, ok := subject.pullRequestDiffForSummary(summary)
	if !ok {
		t.Fatal("expected the refreshed diff to be cached")
	}
	if !actualAfterRefresh.fileTeamOwnersAttempted {
		t.Fatal("expected the refreshed diff to remember the file owner lookup")
	}
	expectedTeamOwners := []string{"@acme/platform"}
	if !reflect.DeepEqual(actualAfterRefresh.data.Files[0].TeamOwners, expectedTeamOwners) {
		t.Fatalf("expected team owners %v, actual %+v", expectedTeamOwners, actualAfterRefresh.data.Files[0].TeamOwners)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42"}, loader.diffCalls)
	}
	if !reflect.DeepEqual(loader.fileTeamOwnerCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected file team owner lookup calls %v, actual %v", []string{"acme/widgets#42"}, loader.fileTeamOwnerCalls)
	}
}

func TestOptimisticMutationCoordinator_GivenLoadedPullRequestDetail_WhenAppendingAComment_ThenItMarksTheVisibleCacheForRefresh(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Body", State: "OPEN"}
	cache := &fakePersistentPullRequestCache{}
	subject := given_workflowProgram(summary, &fakePullRequestDetailLoader{})
	subject.pullRequestCache = cache
	subject.connectedUserLogin = "octocat"
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, URL: summary.URL, Body: summary.Body, State: summary.State}}

	subject.optimisticallyAppendPullRequestComment(pullRequestCommentTarget{repository: "acme/widgets", number: 42}, "Ship it")

	actual := subject.pullRequestDetailCache["acme/widgets#42"]
	if len(actual.detail.Comments) != 1 {
		t.Fatalf("expected one optimistic comment, actual %+v", actual.detail.Comments)
	}
	if actual.detail.Comments[0].Body != "Ship it" || !actual.detail.Comments[0].ViewerDidAuthor || actual.detail.Comments[0].Author == nil || actual.detail.Comments[0].Author.Login != "octocat" {
		t.Fatalf("expected the optimistic comment to belong to the connected user, actual %+v", actual.detail.Comments[0])
	}
	if !actual.needsRefresh {
		t.Fatal("expected the detail cache to be marked for refresh")
	}
	if !reflect.DeepEqual(cache.invalidatedPullRequests, []string{"acme/widgets#42"}) {
		t.Fatalf("expected invalidated pull request keys %v, actual %v", []string{"acme/widgets#42"}, cache.invalidatedPullRequests)
	}
}

func TestOptimisticMutationCoordinator_GivenLoadedDetailAndDiff_WhenAddingAReaction_ThenItUpdatesBothCaches(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Body", State: "OPEN"}
	subject := given_workflowProgram(summary, &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{ID: "PR_kw42", InlineCommentThreads: []githubcli.PullRequestReviewThread{{ID: "thread-1", Path: "main.go", Comments: []githubcli.PullRequestComment{{ID: "PRRC_1"}}}}}}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{{Path: "main.go", Threads: []reviewDiffThread{{ID: "thread-1", Comments: []githubcli.PullRequestComment{{ID: "PRRC_1"}}}}}}}}

	subject.optimisticallyAddReaction(pullRequestReactionActionTarget{repository: "acme/widgets", number: 42, subjectID: "PRRC_1", invalidateDiff: true}, githubcli.ReactionContentHeart)

	actualDetail := subject.pullRequestDetailCache["acme/widgets#42"]
	actualDiff := subject.pullRequestDiffCache["acme/widgets#42"]
	if !actualDetail.needsRefresh {
		t.Fatal("expected the detail cache to be marked for refresh")
	}
	if !actualDiff.needsRefresh {
		t.Fatal("expected the diff cache to be marked for refresh")
	}
	expectedReactions := []githubcli.ReactionGroup{{Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}}
	if !reflect.DeepEqual(actualDiff.data.Files[0].Threads[0].Comments[0].ReactionGroups, expectedReactions) {
		t.Fatalf("expected reaction groups %+v, actual %+v", expectedReactions, actualDiff.data.Files[0].Threads[0].Comments[0].ReactionGroups)
	}
}

func TestOptimisticMutationCoordinator_GivenLoadedDetailAndDiff_WhenResolvingAReviewThread_ThenItUpdatesBothCaches(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "Body", State: "OPEN"}
	subject := given_workflowProgram(summary, &fakePullRequestDetailLoader{})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{InlineCommentThreads: []githubcli.PullRequestReviewThread{{ID: "thread-1", Path: "main.go", Comments: []githubcli.PullRequestComment{{ID: "PRRC_1"}}}}}}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{{Path: "main.go", Threads: []reviewDiffThread{{ID: "thread-1", Path: "main.go", Comments: []githubcli.PullRequestComment{{ID: "PRRC_1"}}}}}}}}

	subject.optimisticallySetReviewThreadResolved(pullRequestReviewThreadActionTarget{repository: "acme/widgets", number: 42, threadID: "thread-1"}, true)

	actualDetail := subject.pullRequestDetailCache["acme/widgets#42"]
	actualDiff := subject.pullRequestDiffCache["acme/widgets#42"]
	if !actualDetail.detail.InlineCommentThreads[0].IsResolved {
		t.Fatalf("expected the detail thread to be resolved, actual %+v", actualDetail.detail.InlineCommentThreads[0])
	}
	if !actualDiff.data.Files[0].Threads[0].IsResolved {
		t.Fatalf("expected the diff thread to be resolved, actual %+v", actualDiff.data.Files[0].Threads[0])
	}
}

func TestNotificationStore_GivenAFailedNotificationMutation_WhenFinishingTheAsyncWork_ThenItRestoresThePreviousRows(t *testing.T) {
	notifications := []githubcli.Notification{
		given_unsupportedNotification("n-push-1", true, "Push notification one"),
		given_unsupportedNotification("n-push-2", false, "Push notification two"),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...), markNotificationReadErr: errors.New("boom")}
	subject := given_notificationActionProgram(loader.notifications, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.markSelectedNotificationRead(gui)
	then_noError(t, actualErr)
	if len(asyncRunner.runs) == 0 {
		t.Fatal("expected an async notification mutation run")
	}
	if actualRows := subject.model.NotificationRows(); actualRows[0].Notification == nil || actualRows[0].Notification.Unread {
		t.Fatalf("expected the notification to become read optimistically, actual %+v", actualRows)
	}

	asyncRunner.runs[0]()

	actualRows := subject.model.NotificationRows()
	if actualRows[0].Notification == nil || !actualRows[0].Notification.Unread {
		t.Fatalf("expected the notification mutation rollback to restore the unread state, actual %+v", actualRows)
	}
	then_statusLineContains(t, gui, "boom")
}

func given_workflowProgram(summary githubcli.PullRequest, loader *fakePullRequestDetailLoader) *Program {
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.uiUpdater = immediateUIUpdater{}
	return subject
}
