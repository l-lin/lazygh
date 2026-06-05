package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDetailStore_GivenCacheAndLoadState_WhenApplyingWorkflowTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newDetailStore(nil)
	subject.pullRequestDetailCache["acme/widgets#1"] = pullRequestDetailResult{detail: githubdomain.PullRequestDetail{Title: "old detail"}}
	subject.pullRequestDetailLoadInFlight["acme/widgets#1"] = true
	subject.issueDetailLoadInFlight["acme/widgets#7"] = true
	subject.releaseDetailLoadInFlight["acme/widgets#8"] = true

	loadPlanned := subject.withPullRequestDetailLoadPlanned("acme/widgets#42")
	cached := loadPlanned.withPullRequestDetailCached("acme/widgets#42", pullRequestDetailResult{detail: githubdomain.PullRequestDetail{Title: "new detail"}})
	cleared := cached.withPullRequestDetailLoadCleared("acme/widgets#1")
	invalidated := cleared.withoutPullRequestDetail("acme/widgets#1")
	issueLoaded := invalidated.withIssueDetailLoaded("acme/widgets#7", issueDetailResult{detail: githubdomain.IssueDetail{Body: "issue body"}})
	releaseLoaded := issueLoaded.withReleaseDetailLoaded("acme/widgets#8", releaseDetailResult{detail: githubdomain.ReleaseDetail{Body: "release body"}})
	reset := releaseLoaded.withWorkflowStateReset()

	if !loadPlanned.pullRequestDetailLoadInFlight["acme/widgets#42"] {
		t.Fatal("expected the planned state to track the new detail load in flight")
	}
	if actual := cached.pullRequestDetailCache["acme/widgets#42"].detail.Title; actual != "new detail" {
		t.Fatalf("expected cached detail title %q, actual %q", "new detail", actual)
	}
	if cleared.pullRequestDetailLoadInFlight["acme/widgets#1"] {
		t.Fatal("expected the cleared state to forget the old detail load")
	}
	if _, ok := invalidated.pullRequestDetailCache["acme/widgets#1"]; ok {
		t.Fatal("expected the invalidated state to drop the old detail cache entry")
	}
	if issueLoaded.issueDetailLoadInFlight["acme/widgets#7"] {
		t.Fatal("expected issue detail loading to clear after storing the result")
	}
	if actual := issueLoaded.issueDetailCache["acme/widgets#7"].detail.Body; actual != "issue body" {
		t.Fatalf("expected cached issue body %q, actual %q", "issue body", actual)
	}
	if releaseLoaded.releaseDetailLoadInFlight["acme/widgets#8"] {
		t.Fatal("expected release detail loading to clear after storing the result")
	}
	if actual := releaseLoaded.releaseDetailCache["acme/widgets#8"].detail.Body; actual != "release body" {
		t.Fatalf("expected cached release body %q, actual %q", "release body", actual)
	}
	if len(reset.pullRequestDetailCache) != 0 || len(reset.pullRequestDetailLoadInFlight) != 0 || len(reset.issueDetailCache) != 0 || len(reset.issueDetailLoadInFlight) != 0 || len(reset.releaseDetailCache) != 0 || len(reset.releaseDetailLoadInFlight) != 0 {
		t.Fatalf("expected workflow reset to clear detail and notification caches, actual detail=%d detailLoads=%d issue=%d issueLoads=%d release=%d releaseLoads=%d", len(reset.pullRequestDetailCache), len(reset.pullRequestDetailLoadInFlight), len(reset.issueDetailCache), len(reset.issueDetailLoadInFlight), len(reset.releaseDetailCache), len(reset.releaseDetailLoadInFlight))
	}

	if actual := subject.pullRequestDetailCache["acme/widgets#1"].detail.Title; actual != "old detail" {
		t.Fatalf("expected the original detail title %q, actual %q", "old detail", actual)
	}
	if !subject.pullRequestDetailLoadInFlight["acme/widgets#1"] {
		t.Fatal("expected the original detail load-in-flight state to stay true")
	}
	if _, ok := subject.pullRequestDetailLoadInFlight["acme/widgets#42"]; ok {
		t.Fatal("expected the original state to stay free of the new detail load key")
	}
	if subject.issueDetailCache["acme/widgets#7"].detail.Body != "" {
		t.Fatalf("expected the original issue cache to stay empty, actual %+v", subject.issueDetailCache["acme/widgets#7"])
	}
	if subject.releaseDetailCache["acme/widgets#8"].detail.Body != "" {
		t.Fatalf("expected the original release cache to stay empty, actual %+v", subject.releaseDetailCache["acme/widgets#8"])
	}
}

func TestDetailStore_GivenBrowserCollapsedSectionStates_WhenApplyingCollapseTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newDetailStore(nil)
	subject.browserCollapsedSectionStates["acme/widgets#1:overview"] = true

	singleCollapsed, actualSingleChanged := subject.withBrowserDetailSectionCollapsed(" acme/widgets#42:checks ", true)
	bulkCollapsed, actualBulkChanged := singleCollapsed.withBrowserDetailSectionsCollapsed([]string{"acme/widgets#43:reviews", "", " acme/widgets#44:threads "}, true)
	unchanged, actualUnchanged := bulkCollapsed.withBrowserDetailSectionsCollapsed([]string{"acme/widgets#43:reviews", "acme/widgets#44:threads"}, true)

	expectedSingleKey := "acme/widgets#42:checks"
	if !actualSingleChanged {
		t.Fatal("expected the single-section transition to report a change")
	}
	if actual := singleCollapsed.browserCollapsedSectionStates[expectedSingleKey]; !actual {
		t.Fatalf("expected collapsed state for %q to be true", expectedSingleKey)
	}
	if !actualBulkChanged {
		t.Fatal("expected the bulk transition to report a change")
	}
	for _, expected := range []string{"acme/widgets#43:reviews", "acme/widgets#44:threads"} {
		if actual := bulkCollapsed.browserCollapsedSectionStates[expected]; !actual {
			t.Fatalf("expected collapsed state for %q to be true", expected)
		}
	}
	if actualUnchanged {
		t.Fatal("expected the unchanged bulk transition to report no change")
	}
	if _, ok := unchanged.browserCollapsedSectionStates[""]; ok {
		t.Fatal("expected blank section ids to stay absent from the collapsed-state map")
	}

	if actual := subject.browserCollapsedSectionStates["acme/widgets#1:overview"]; !actual {
		t.Fatal("expected the original collapsed overview state to stay true")
	}
	if _, ok := subject.browserCollapsedSectionStates[expectedSingleKey]; ok {
		t.Fatal("expected the original state to stay free of the new single-section entry")
	}
	if _, ok := subject.browserCollapsedSectionStates["acme/widgets#43:reviews"]; ok {
		t.Fatal("expected the original state to stay free of the new bulk-section entry")
	}
}

func TestReviewStore_GivenDiffCacheAndLoadState_WhenApplyingWorkflowTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newReviewStore(nil)
	subject.pullRequestDiffCache["acme/widgets#1"] = pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{{Path: "old.go"}}}}
	subject.pullRequestDiffLoadInFlight["acme/widgets#1"] = true

	loadPlanned := subject.withPullRequestDiffLoadPlanned("acme/widgets#42")
	cached := loadPlanned.withPullRequestDiffCached("acme/widgets#42", pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{{Path: "main.go"}}}})
	cleared := cached.withPullRequestDiffLoadCleared("acme/widgets#1")
	invalidated := cleared.withoutPullRequestDiff("acme/widgets#1")
	reset := invalidated.withDiffWorkflowStateReset()

	if !loadPlanned.pullRequestDiffLoadInFlight["acme/widgets#42"] {
		t.Fatal("expected the planned state to track the new diff load in flight")
	}
	if actual := cached.pullRequestDiffCache["acme/widgets#42"].data.Files[0].Path; actual != "main.go" {
		t.Fatalf("expected cached diff path %q, actual %q", "main.go", actual)
	}
	if cleared.pullRequestDiffLoadInFlight["acme/widgets#1"] {
		t.Fatal("expected the cleared state to forget the old diff load")
	}
	if _, ok := invalidated.pullRequestDiffCache["acme/widgets#1"]; ok {
		t.Fatal("expected the invalidated state to drop the old diff cache entry")
	}
	if len(reset.pullRequestDiffCache) != 0 || len(reset.pullRequestDiffLoadInFlight) != 0 {
		t.Fatalf("expected diff workflow reset to clear caches and in-flight state, actual cache=%d inFlight=%d", len(reset.pullRequestDiffCache), len(reset.pullRequestDiffLoadInFlight))
	}

	if actual := subject.pullRequestDiffCache["acme/widgets#1"].data.Files[0].Path; actual != "old.go" {
		t.Fatalf("expected the original diff path %q, actual %q", "old.go", actual)
	}
	if !subject.pullRequestDiffLoadInFlight["acme/widgets#1"] {
		t.Fatal("expected the original diff load-in-flight state to stay true")
	}
	if _, ok := subject.pullRequestDiffLoadInFlight["acme/widgets#42"]; ok {
		t.Fatal("expected the original state to stay free of the new diff load key")
	}
}

func TestReviewStore_GivenPendingReviewCacheState_WhenApplyingPendingReviewTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newReviewStore(nil)
	subject.pendingPullRequestReviewCache["acme/widgets#1"] = pendingPullRequestReviewState{id: "review-1"}

	stored := subject.withPendingPullRequestReviewCached("acme/widgets#42", pendingPullRequestReviewState{id: "review-42"})
	forgotten := stored.withoutPendingPullRequestReview("acme/widgets#1")
	reset := forgotten.withPendingReviewCacheReset()

	if actual := stored.pendingPullRequestReviewCache["acme/widgets#42"].id; actual != "review-42" {
		t.Fatalf("expected stored pending review id %q, actual %q", "review-42", actual)
	}
	if _, ok := forgotten.pendingPullRequestReviewCache["acme/widgets#1"]; ok {
		t.Fatal("expected the forgotten state to drop the original pending review entry")
	}
	if len(reset.pendingPullRequestReviewCache) != 0 {
		t.Fatalf("expected the pending-review reset to clear the cache, actual %d entries", len(reset.pendingPullRequestReviewCache))
	}

	if actual := subject.pendingPullRequestReviewCache["acme/widgets#1"].id; actual != "review-1" {
		t.Fatalf("expected the original pending review id %q, actual %q", "review-1", actual)
	}
	if _, ok := subject.pendingPullRequestReviewCache["acme/widgets#42"]; ok {
		t.Fatal("expected the original cache to stay free of the new pending review entry")
	}
}

func TestReviewStore_GivenStoryReviewCacheState_WhenApplyingStoryReviewTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newReviewStore(nil)
	subject.storyReviewCache["acme/widgets#1"] = storyReviewResult{story: reviewStoryData{Summary: "old story"}, pendingReviewID: "PRR_old"}

	stored := subject.withStoryReviewCached("acme/widgets#42", storyReviewResult{story: reviewStoryData{Summary: "new story"}, pendingReviewID: "PRR_new"})
	forgotten := stored.withoutStoryReview("acme/widgets#1")
	reset := forgotten.withStoryReviewCacheReset()

	if actual := stored.storyReviewCache["acme/widgets#42"].story.Summary; actual != "new story" {
		t.Fatalf("expected stored story summary %q, actual %q", "new story", actual)
	}
	if actual := stored.storyReviewCache["acme/widgets#42"].pendingReviewID; actual != "PRR_new" {
		t.Fatalf("expected stored pending review id %q, actual %q", "PRR_new", actual)
	}
	if _, ok := forgotten.storyReviewCache["acme/widgets#1"]; ok {
		t.Fatal("expected the forgotten state to drop the original story review entry")
	}
	if len(reset.storyReviewCache) != 0 {
		t.Fatalf("expected the story-review reset to clear the cache, actual %d entries", len(reset.storyReviewCache))
	}

	if actual := subject.storyReviewCache["acme/widgets#1"].story.Summary; actual != "old story" {
		t.Fatalf("expected the original story summary %q, actual %q", "old story", actual)
	}
	if _, ok := subject.storyReviewCache["acme/widgets#42"]; ok {
		t.Fatal("expected the original cache to stay free of the new story review entry")
	}
}

func TestDetailStore_GivenDocumentAndRenderedRowCaches_WhenApplyingCacheTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newDetailStore(nil)
	originalKey := pullRequestDetailDocumentCacheKey{pullRequestKey: "acme/widgets#1", tab: DescriptionDetailTab, width: 72}
	cachedKey := pullRequestDetailDocumentCacheKey{pullRequestKey: "acme/widgets#42", tab: CommentsDetailTab, width: 88}
	subject.pullRequestDetailDocumentCache[originalKey] = detailDocument{width: 72}
	subject.pullRequestConversationDocumentCache[originalKey] = browserConversationDocument{text: "old conversation"}
	subject.pullRequestChangesRenderedRowsCache[originalKey] = []reviewDiffRenderedRow{{Text: "old rows"}}
	givenRows := []reviewDiffRenderedRow{{Text: "new rows"}}

	detailCached := subject.withPullRequestDetailDocumentCached(cachedKey, detailDocument{width: 88})
	conversationCached := detailCached.withPullRequestConversationDocumentCached(cachedKey, browserConversationDocument{text: "new conversation"})
	rowsCached := conversationCached.withPullRequestChangesRenderedRowsCached(cachedKey, givenRows)
	givenRows[0].Text = "mutated rows"
	reset := rowsCached.withDocumentRenderCachesReset()

	if actual := detailCached.pullRequestDetailDocumentCache[cachedKey].width; actual != 88 {
		t.Fatalf("expected cached detail document width %d, actual %d", 88, actual)
	}
	if actual := conversationCached.pullRequestConversationDocumentCache[cachedKey].text; actual != "new conversation" {
		t.Fatalf("expected cached conversation text %q, actual %q", "new conversation", actual)
	}
	if actual := rowsCached.pullRequestChangesRenderedRowsCache[cachedKey][0].Text; actual != "new rows" {
		t.Fatalf("expected cached rendered rows text %q, actual %q", "new rows", actual)
	}
	if len(reset.pullRequestDetailDocumentCache) != 0 || len(reset.pullRequestConversationDocumentCache) != 0 || len(reset.pullRequestChangesRenderedRowsCache) != 0 {
		t.Fatalf("expected document/render cache reset to clear all maps, actual detail=%d conversation=%d rows=%d", len(reset.pullRequestDetailDocumentCache), len(reset.pullRequestConversationDocumentCache), len(reset.pullRequestChangesRenderedRowsCache))
	}

	if actual := subject.pullRequestDetailDocumentCache[originalKey].width; actual != 72 {
		t.Fatalf("expected original detail document width %d, actual %d", 72, actual)
	}
	if actual := subject.pullRequestConversationDocumentCache[originalKey].text; actual != "old conversation" {
		t.Fatalf("expected original conversation text %q, actual %q", "old conversation", actual)
	}
	if actual := subject.pullRequestChangesRenderedRowsCache[originalKey][0].Text; actual != "old rows" {
		t.Fatalf("expected original rendered rows text %q, actual %q", "old rows", actual)
	}
	if _, ok := subject.pullRequestDetailDocumentCache[cachedKey]; ok {
		t.Fatal("expected the original detail document cache to stay free of the new entry")
	}
	if _, ok := subject.pullRequestConversationDocumentCache[cachedKey]; ok {
		t.Fatal("expected the original conversation document cache to stay free of the new entry")
	}
	if _, ok := subject.pullRequestChangesRenderedRowsCache[cachedKey]; ok {
		t.Fatal("expected the original rendered rows cache to stay free of the new entry")
	}
}

func TestReviewStore_GivenRenderCacheEntries_WhenApplyingRenderCacheTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newReviewStore(nil)
	originalKey := reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 1, filePath: "old.go", width: 72}
	cachedKey := reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "main.go", width: 88}
	subject.reviewDiffRenderCache[originalKey] = reviewDiffRenderCacheEntry{rows: []reviewDiffRenderedRow{{Text: "old rows"}}, document: detailDocument{width: 72}}
	givenRows := []reviewDiffRenderedRow{{Text: "new rows"}}

	stored := subject.withReviewDiffRenderEntryCached(cachedKey, reviewDiffRenderCacheEntry{rows: givenRows, document: detailDocument{width: 88}})
	givenRows[0].Text = "mutated rows"
	reset := stored.withReviewDiffRenderCacheReset()

	if actual := stored.reviewDiffRenderCache[cachedKey].rows[0].Text; actual != "new rows" {
		t.Fatalf("expected cached review rows text %q, actual %q", "new rows", actual)
	}
	if actual := stored.reviewDiffRenderCache[cachedKey].document.width; actual != 88 {
		t.Fatalf("expected cached review document width %d, actual %d", 88, actual)
	}
	if len(reset.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected review diff render cache reset to clear the map, actual %d entries", len(reset.reviewDiffRenderCache))
	}

	if actual := subject.reviewDiffRenderCache[originalKey].rows[0].Text; actual != "old rows" {
		t.Fatalf("expected original review rows text %q, actual %q", "old rows", actual)
	}
	if _, ok := subject.reviewDiffRenderCache[cachedKey]; ok {
		t.Fatal("expected the original review render cache to stay free of the new entry")
	}
}
