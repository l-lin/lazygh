package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestMergeQueueMutationSnapshot struct {
	summary   githubdomain.PullRequest
	detail    pullRequestDetailResult
	hasDetail bool
}

type pullRequestMergeQueueAsyncError struct {
	err                error
	snapshot           pullRequestMergeQueueMutationSnapshot
	rollbackQueueState bool
}

func effectivePullRequestMergeQueueEnabled(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) bool {
	if detail.IsMergeQueueEnabled {
		return true
	}
	return summary.IsMergeQueueEnabled
}

func effectivePullRequestMergeQueueEntry(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) *githubdomain.PullRequestMergeQueueEntry {
	if detail.MergeQueueEntry != nil {
		return detail.MergeQueueEntry
	}
	return summary.MergeQueueEntry
}

func effectivePullRequestInMergeQueue(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) bool {
	if detail.IsInMergeQueue || detail.MergeQueueEntry != nil {
		return true
	}
	return summary.IsInMergeQueue || summary.MergeQueueEntry != nil
}

func (program *Program) currentPullRequestMergeQueueEnabled() bool {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return false
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return effectivePullRequestMergeQueueEnabled(summary, result.detail)
	}
	return summary.IsMergeQueueEnabled
}

func (program *Program) currentPullRequestInMergeQueue() bool {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return false
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return effectivePullRequestInMergeQueue(summary, result.detail)
	}
	return summary.IsInMergeQueue || summary.MergeQueueEntry != nil
}

func optimisticPullRequestMergeQueueEntry(inQueue bool) *githubdomain.PullRequestMergeQueueEntry {
	if !inQueue {
		return nil
	}
	return &githubdomain.PullRequestMergeQueueEntry{State: "QUEUED"}
}

func (program *Program) applyVisiblePullRequestMergeQueueMutation(summary githubdomain.PullRequest, inQueue bool) {
	mergeQueueEntry := optimisticPullRequestMergeQueueEntry(inQueue)

	program.mutateLoadedPullRequestSummaries(summary, func(current *githubdomain.PullRequest) {
		current.IsMergeQueueEnabled = true
		current.IsInMergeQueue = inQueue
		current.MergeQueueEntry = clonePullRequestMergeQueueEntry(mergeQueueEntry)
		current.AutoMergeRequest = nil
	})

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}
	if result, ok := program.pullRequestDetailCache[key]; ok && result.err == nil {
		result.detail.IsMergeQueueEnabled = true
		result.detail.IsInMergeQueue = inQueue
		result.detail.MergeQueueEntry = clonePullRequestMergeQueueEntry(mergeQueueEntry)
		result.detail.AutoMergeRequest = nil
		result.sourceUpdatedAt = ""
		result.needsRefresh = true
		program.applyPullRequestDetailCacheResult(pullRequestRepositoryName(summary.Repository), summary.Number, result, pullRequestDetailCacheApplyOptions{clearInFlight: true, invalidateDocuments: true, invalidatePersistent: true})
		return
	}

	program.invalidatePullRequestDetailDocumentCache()
	program.invalidatePersistentPullRequest(pullRequestRepositoryName(summary.Repository), summary.Number)
}

func (program *Program) capturePullRequestMergeQueueMutationSnapshot(summary githubdomain.PullRequest) pullRequestMergeQueueMutationSnapshot {
	snapshot := pullRequestMergeQueueMutationSnapshot{summary: summary}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		snapshot.summary.IsMergeQueueEnabled = effectivePullRequestMergeQueueEnabled(summary, result.detail)
		snapshot.summary.IsInMergeQueue = effectivePullRequestInMergeQueue(summary, result.detail)
		snapshot.summary.MergeQueueEntry = clonePullRequestMergeQueueEntry(effectivePullRequestMergeQueueEntry(summary, result.detail))
		if result.detail.AutoMergeRequest != nil {
			snapshot.summary.AutoMergeRequest = clonePullRequestAutoMergeRequest(result.detail.AutoMergeRequest)
		}
		snapshot.detail = result
		snapshot.detail.detail = clonePullRequestDetail(result.detail)
		snapshot.hasDetail = true
		return snapshot
	}
	snapshot.summary.MergeQueueEntry = clonePullRequestMergeQueueEntry(summary.MergeQueueEntry)
	snapshot.summary.AutoMergeRequest = clonePullRequestAutoMergeRequest(summary.AutoMergeRequest)
	return snapshot
}

func (program *Program) restorePullRequestMergeQueueMutationSnapshot(snapshot pullRequestMergeQueueMutationSnapshot) {
	program.mutateLoadedPullRequestSummaries(snapshot.summary, func(current *githubdomain.PullRequest) {
		current.IsMergeQueueEnabled = snapshot.summary.IsMergeQueueEnabled
		current.IsInMergeQueue = snapshot.summary.IsInMergeQueue
		current.MergeQueueEntry = clonePullRequestMergeQueueEntry(snapshot.summary.MergeQueueEntry)
		current.AutoMergeRequest = clonePullRequestAutoMergeRequest(snapshot.summary.AutoMergeRequest)
	})
	if !snapshot.hasDetail {
		return
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(snapshot.summary.Repository))
	if repository == "" || snapshot.summary.Number <= 0 {
		return
	}
	result := snapshot.detail
	result.detail = clonePullRequestDetail(snapshot.detail.detail)
	program.applyPullRequestDetailCacheResult(repository, snapshot.summary.Number, result, pullRequestDetailCacheApplyOptions{clearInFlight: true, invalidateDocuments: true, invalidatePersistent: true})
}

func newPullRequestMergeQueueAsyncError(err error, snapshot pullRequestMergeQueueMutationSnapshot) error {
	if err == nil {
		return nil
	}
	return pullRequestMergeQueueAsyncError{err: newTransientErrorPopupActionError(err), snapshot: snapshot, rollbackQueueState: true}
}

func (err pullRequestMergeQueueAsyncError) Error() string {
	if err.err == nil {
		return ""
	}
	return err.err.Error()
}

func (err pullRequestMergeQueueAsyncError) Unwrap() error {
	return err.err
}

func pullRequestMergeQueueRollback(err error) (pullRequestMergeQueueMutationSnapshot, bool) {
	var actual pullRequestMergeQueueAsyncError
	if !errors.As(err, &actual) || !actual.rollbackQueueState {
		return pullRequestMergeQueueMutationSnapshot{}, false
	}
	return actual.snapshot, true
}
