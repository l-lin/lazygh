package tui

import (
	"errors"
	"reflect"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDetailProgress_GivenCachedContent_WhenDescriptionArrives_ThenItKeepsEnrichmentAndLoadOwnership(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	expectedVersion := "previous version"
	subject.pullRequestDetailCache[key] = pullRequestDetailResult{detail: githubdomain.PullRequestDetail{Body: "cached", InlineComments: []githubdomain.PullRequestInlineComment{{Body: "cached comment"}}}, sourceUpdatedAt: expectedVersion}
	subject.detailState.viewState.originRow = 5
	subject.browserCollapsedSectionStates["description"] = true
	expectedViewState := subject.detailState.viewState
	subject.applyPullRequestDetailLoadPlanned(MsgPullRequestDetailLoadPlanned{Key: key})
	if actual := subject.pullRequestDetailCache[key].detail.Body; actual != "cached" {
		t.Fatalf("expected cached description during refresh, actual %q", actual)
	}
	actualCommands := Update(subject, MsgPullRequestDetailProgress{Summary: summary, Detail: githubdomain.PullRequestDetail{Body: "fresh", Title: "fresh title"}})
	actual := subject.pullRequestDetailCache[key]
	if actual.detail.Body != "fresh" || len(actual.detail.InlineComments) != 1 || actual.sourceUpdatedAt != expectedVersion || !actual.needsRefresh || !subject.pullRequestDetailLoadInFlight[key] || len(actualCommands) != 0 {
		t.Fatalf("expected fresh description with retained enrichment and unfinished load, actual %+v, commands %+v", actual, actualCommands)
	}
	if !reflect.DeepEqual(subject.detailState.viewState, expectedViewState) || !subject.browserCollapsedSectionStates["description"] {
		t.Fatal("expected cursor, scroll and folds to survive progress")
	}
	subject.applyPullRequestDetailLoaded(MsgPullRequestDetailLoaded{Summary: summary, Err: errors.New("enrichment failed")})
	actual = subject.pullRequestDetailCache[key]
	if actual.detail.Body != "fresh" || actual.sourceUpdatedAt != expectedVersion || actual.retry.failureCount != 1 {
		t.Fatalf("expected retained progress with stale version and retry after failure, actual %+v", actual)
	}
}

func TestDiffProgress_GivenCachedReview_WhenDiffArrives_ThenItKeepsCursorFoldsAndLoadOwnership(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	rawDiff := githubdomain.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go\n@@ -1 +1 @@\n-old\n+fresh", Files: []githubdomain.PullRequestDiffFile{{Path: "main.go"}}}
	expectedVersion := "previous version"
	subject.pullRequestDiffCache[key] = pullRequestDiffResult{data: buildReviewDiffData(rawDiff), sourceUpdatedAt: expectedVersion}
	subject.detailState.viewState.originRow = 2
	subject.browserCollapsedSectionStates["main.go"] = true
	subject.startReviewSession(summary, "")
	subject.navigationState.reviewSession.collapsedTreeRowIDs["main.go"] = true
	subject.navigationState.reviewSession.collapsedThreadIDs["thread-1"] = true
	expectedSession := subject.navigationState.reviewSession
	expectedViewState := subject.detailState.viewState
	subject.applyPullRequestDiffLoadPlanned(MsgPullRequestDiffLoadPlanned{Key: key})
	actualCommands := Update(subject, MsgPullRequestDiffProgress{Summary: summary, RawDiff: rawDiff})
	actual := subject.pullRequestDiffCache[key]
	if len(actual.data.Files) != 1 || actual.sourceUpdatedAt != expectedVersion || !actual.needsRefresh || !subject.pullRequestDiffLoadInFlight[key] || len(actualCommands) != 0 {
		t.Fatalf("expected published diff with unfinished load, actual %+v, commands %+v", actual, actualCommands)
	}
	if !reflect.DeepEqual(subject.detailState.viewState, expectedViewState) || !reflect.DeepEqual(subject.navigationState.reviewSession, expectedSession) || !subject.browserCollapsedSectionStates["main.go"] {
		t.Fatal("expected cursor, scroll and folds to survive progress")
	}
	subject.applyPullRequestDiffLoaded(MsgPullRequestDiffLoaded{Summary: summary, Err: errors.New("threads failed")})
	actual = subject.pullRequestDiffCache[key]
	if len(actual.data.Files) != 1 || actual.sourceUpdatedAt != expectedVersion || actual.retry.failureCount != 1 {
		t.Fatalf("expected retained diff with stale version and retry after failure, actual %+v", actual)
	}
}

func TestProgress_GivenRefreshOwnership_WhenPublishing_ThenItDoesNotFinishBatchesOrPersistPartialContent(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	batchID := subject.beginScheduledPullRequestRefreshBatch(7, time.Now())
	subject.addScheduledPullRequestRefreshWork(batchID, 2)
	subject.finishScheduledPullRequestRefreshPlanning(batchID)
	subject.beginManualRefresh("refreshed", 2)
	subject.markManualPullRequestDetailRefresh(summary)
	subject.markManualPullRequestDiffRefresh(summary)
	expectedBatch := subject.scheduledPullRequestRefreshBatches[batchID]
	expectedManual := subject.manualRefreshState
	Update(subject, MsgPullRequestDetailProgress{Summary: summary, Detail: githubdomain.PullRequestDetail{Body: "fresh"}})
	Update(subject, MsgPullRequestDiffProgress{Summary: summary})
	actualBatch := subject.scheduledPullRequestRefreshBatches[batchID]
	actualManual := subject.manualRefreshState
	if !reflect.DeepEqual(actualBatch, expectedBatch) || !reflect.DeepEqual(actualManual, expectedManual) {
		t.Fatalf("expected retained batch and manual ownership, actual batch=%+v manual=%+v", actualBatch, actualManual)
	}
	if actual := len(subject.persistentCacheRuntime.pending); actual != 0 {
		t.Fatalf("expected no persistence of partial content, actual %d actions", actual)
	}
}

func TestProgress_GivenNoCache_WhenBaseContentArrives_ThenItDisplaysContentWithoutMarkingItSuccessfullyFetched(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	Update(subject, MsgPullRequestDetailProgress{Summary: summary, Detail: githubdomain.PullRequestDetail{Body: "fresh"}})
	Update(subject, MsgPullRequestDiffProgress{Summary: summary, RawDiff: githubdomain.PullRequestDiff{Files: []githubdomain.PullRequestDiffFile{{Path: "main.go"}}}})
	actualDetail := subject.pullRequestDetailCache[key]
	actualDiff := subject.pullRequestDiffCache[key]
	if actualDetail.detail.Body != "fresh" || len(actualDiff.data.Files) != 1 || actualDetail.sourceUpdatedAt != "" || actualDiff.sourceUpdatedAt != "" || !actualDetail.needsRefresh || !actualDiff.needsRefresh {
		t.Fatalf("expected visible partial content without successful versions, actual detail=%+v diff=%+v", actualDetail, actualDiff)
	}
}
