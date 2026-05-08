package tui

import (
	"strings"
	"testing"

	persistcache "codeberg.org/l-lin/lazygh/internal/cache"
	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenClearCacheActionSelected_WhenExecutingOnce_ThenItAsksForConfirmationBeforeClearing(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	cache := &fakePersistentPullRequestCache{
		pullRequestsBySearchKey: map[string][]githubcli.PullRequest{
			fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0]): {given_actionsPopupPullRequest()},
		},
	}
	subject.pullRequestCache = cache
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "Cached PR", Number: 42, Body: "Cached body"}}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}})}
	subject.pullRequestDetailDocumentCache[pullRequestDetailDocumentCacheKey{pullRequestKey: "acme/widgets#42", tab: DescriptionDetailTab, width: 80}] = detailDocument{}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{identity: "acme/widgets#42:main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("clear cache", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "clear cache"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	if cache.clearCalls != 0 {
		t.Fatalf("expected cache clear calls %d, actual %d", 0, cache.clearCalls)
	}
	if len(subject.pullRequestDetailCache) != 1 {
		t.Fatalf("expected detail cache to stay intact before confirmation, actual %d entries", len(subject.pullRequestDetailCache))
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "Press Enter again to clear cache") {
		t.Fatalf("expected popup title to contain %q, actual %q", "Press Enter again to clear cache", popupView.Title)
	}
}

func TestActionsPopup_GivenConfirmedClearCacheAction_WhenExecuting_ThenItClearsPersistentAndInMemoryCachesReloadsAndShowsTheEmptyState(t *testing.T) {
	loader := &cacheAwarePullRequestLoader{fakePullRequestDetailLoader: &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{}}}
	subject := NewProgramWithModelAndLoader(given_pullRequestCommentModel(), loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	cache := &fakePersistentPullRequestCache{
		pullRequestsBySearchKey: map[string][]githubcli.PullRequest{
			fakePersistentPullRequestSearchKey(appconfig.DefaultPullRequestSearches()[0]): {given_actionsPopupPullRequest()},
		},
		details: map[string]persistcache.CachedPullRequestDetail{
			"acme/widgets#42": {Detail: githubcli.PullRequestDetail{Title: "Cached PR", Number: 42, Body: "Cached body"}},
		},
		diffs: map[string]persistcache.CachedPullRequestDiff{
			"acme/widgets#42": {Diff: githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}}},
		},
	}
	subject.pullRequestCache = cache
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "Cached PR", Number: 42, Body: "Cached body"}}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}})}
	subject.pullRequestDetailDocumentCache[pullRequestDetailDocumentCacheKey{pullRequestKey: "acme/widgets#42", tab: DescriptionDetailTab, width: 80}] = detailDocument{}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{identity: "acme/widgets#42:main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("clear cache", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "clear cache"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	if cache.clearCalls != 1 {
		t.Fatalf("expected cache clear calls %d, actual %d", 1, cache.clearCalls)
	}
	if len(subject.pullRequestDetailCache) != 0 {
		t.Fatalf("expected detail cache to be empty, actual %d entries", len(subject.pullRequestDetailCache))
	}
	if len(subject.pullRequestDiffCache) != 0 {
		t.Fatalf("expected diff cache to be empty, actual %d entries", len(subject.pullRequestDiffCache))
	}
	if len(subject.pullRequestDetailDocumentCache) != 0 {
		t.Fatalf("expected detail document cache to be empty, actual %d entries", len(subject.pullRequestDetailDocumentCache))
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected review diff render cache to be empty, actual %d entries", len(subject.reviewDiffRenderCache))
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), myPullRequestsEmptyTitle) {
		t.Fatalf("expected pull requests buffer to contain %q after clearing the cache, actual %q", myPullRequestsEmptyTitle, pullRequestsView.Buffer())
	}
	then_statusLineContains(t, gui, "Cache cleared.")
}
