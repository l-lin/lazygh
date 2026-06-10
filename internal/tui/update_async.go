package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyPullRequestsCacheHydrated(message MsgPullRequestsCacheHydrated) {
	program.applyLoadedPullRequestRows(message.Tab, message.PullRequests)
	program.selectOpenedPullRequestRow(message.Tab)
}

func (program *Program) applyNotificationsCacheHydrated(message MsgNotificationsCacheHydrated) {
	program.model.SetNotificationRows(notificationRows(program.filterDoneNotifications(message.Notifications)))
}

func (program *Program) applyConnectedUserLoaded(message MsgConnectedUserLoaded) {
	connectedUserLogin := ""
	connectedUserName := ""
	if message.Err == nil {
		connectedUserLogin = strings.TrimSpace(message.User.Login)
		connectedUserName = strings.TrimSpace(message.User.Name)
	}
	if program.setConnectedUser(connectedUserLogin, connectedUserName) {
		program.invalidatePullRequestDetailDocumentCache()
		program.invalidateReviewDiffRenderCache()
	}
	program.model.SetUsers([]Item{connectedUserStateItem(message.User, message.Err)})
}

func (program *Program) applyManualRefreshCompletion(err error) []Cmd {
	completion := program.completeManualRefreshOperation(err)
	if completion.successMessage != "" {
		program.setFeedback(FocusDetailView, completion.successMessage)
	}
	if completion.popupError != "" {
		return program.applyErrorReportedMessage(completion.popupError)
	}
	return nil
}

func (program *Program) applyPullRequestsLoaded(message MsgPullRequestsLoaded) []Cmd {
	program.setPullRequestsLoading(message.Tab, false)
	manualRefresh := program.consumeManualPullRequestListRefresh(message.Tab)
	if message.Err == nil {
		program.cachePullRequests(message.Tab, message.PullRequests)
		program.applyLoadedPullRequestRows(message.Tab, message.PullRequests)
		program.selectOpenedPullRequestRow(message.Tab)
		if manualRefresh {
			return program.applyManualRefreshCompletion(nil)
		}
		return nil
	}

	if !program.shouldPreservePullRequestRowsOnRefreshError(message.Tab) {
		program.setPullRequestsCount(message.Tab, 0, false)
		program.model.SetPullRequestRows(message.Tab, pullRequestStateRows(program.pullRequestListState(message.Tab), nil, message.Err))
	}
	if manualRefresh {
		return program.applyManualRefreshCompletion(message.Err)
	}
	return nil
}

func (program *Program) selectOpenedPullRequestRow(tab PullRequestTab) {
	openedSummary, ok := program.openedPullRequestSummaryForTab(tab)
	if !ok {
		return
	}

	rows := program.model.PullRequestRows(tab)
	for index, row := range rows {
		if row.Summary == nil || !samePullRequestIdentity(*row.Summary, openedSummary) {
			continue
		}
		program.model.SelectPullRequestIndex(tab, index)
		return
	}
}

func (program *Program) applyNotificationsLoaded(message MsgNotificationsLoaded) []Cmd {
	program.finishNotificationsLoading()
	manualRefresh := program.consumeManualNotificationRefresh()
	if message.Err == nil {
		filteredNotifications := program.filterDoneNotifications(message.Notifications)
		program.cacheNotifications(filteredNotifications)
		program.model.SetNotificationRows(notificationRows(filteredNotifications))
		if manualRefresh {
			return program.applyManualRefreshCompletion(nil)
		}
		return nil
	}

	if !program.shouldPreserveNotificationRowsOnRefreshError() {
		program.model.SetNotificationRows(notificationsStateRows(nil, message.Err))
	}
	if manualRefresh {
		return program.applyManualRefreshCompletion(message.Err)
	}
	return nil
}

func (program *Program) applyPullRequestDetailLoaded(message MsgPullRequestDetailLoaded) []Cmd {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return nil
	}

	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withPullRequestDetailLoadCleared(key)
	})
	manualRefresh := program.consumeManualPullRequestDetailRefresh(key)
	if message.PendingReviewStateKnown {
		program.updateReviewStore(func(store reviewStore) reviewStore {
			return store.withPendingPullRequestReviewCached(key, message.PendingReviewState)
		})
	}
	if message.Err == nil {
		clonedDetail := clonePullRequestDetail(message.Detail)
		program.cachePullRequestDetail(message.Summary, clonedDetail)
		program.updateDetailStore(func(store detailStore) detailStore {
			return store.withPullRequestDetailCached(key, pullRequestDetailResult{detail: clonedDetail, sourceUpdatedAt: pullRequestSummaryVersion(message.Summary)})
		})
		program.refreshLoadedPullRequestSummaryFromDetail(message.Summary, clonedDetail)
		program.invalidatePullRequestDetailDocumentCache()
		if manualRefresh {
			return program.applyManualRefreshCompletion(nil)
		}
		return nil
	}

	if !program.canKeepPullRequestDetailOnRefreshError(key) {
		program.updateDetailStore(func(store detailStore) detailStore {
			return store.withPullRequestDetailCached(key, pullRequestDetailResult{err: message.Err, sourceUpdatedAt: pullRequestSummaryVersion(message.Summary)})
		})
		program.invalidatePullRequestDetailDocumentCache()
		if manualRefresh {
			return program.applyManualRefreshCompletion(message.Err)
		}
		return nil
	}

	cachedResult := program.pullRequestDetailCache[key]
	cachedResult.sourceUpdatedAt = pullRequestSummaryVersion(message.Summary)
	cachedResult.needsRefresh = false
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withPullRequestDetailCached(key, cachedResult)
	})
	if manualRefresh {
		return program.applyManualRefreshCompletion(message.Err)
	}
	return nil
}

func (program *Program) refreshLoadedPullRequestSummaryFromDetail(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) {
	program.mutateLoadedPullRequestSummaries(summary, func(current *githubdomain.PullRequest) {
		if current == nil {
			return
		}
		current.Title = firstNonEmpty(strings.TrimSpace(detail.Title), current.Title)
		current.Body = firstNonEmpty(strings.TrimSpace(detail.Body), current.Body)
		current.URL = firstNonEmpty(strings.TrimSpace(detail.URL), current.URL)
		current.State = firstNonEmpty(strings.TrimSpace(detail.State), current.State)
		current.IsDraft = detail.IsDraft
		current.UpdatedAt = firstNonEmpty(strings.TrimSpace(detail.UpdatedAt), current.UpdatedAt)
		current.ReviewDecision = firstNonEmpty(strings.TrimSpace(detail.ReviewDecision), current.ReviewDecision)
		current.MergeStateStatus = firstNonEmpty(strings.TrimSpace(detail.MergeStateStatus), current.MergeStateStatus)
		current.Mergeable = firstNonEmpty(strings.TrimSpace(detail.Mergeable), current.Mergeable)
	})
}

func (program *Program) applyPullRequestDiffLoaded(message MsgPullRequestDiffLoaded) []Cmd {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return nil
	}

	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withPullRequestDiffLoadCleared(key)
	})
	manualRefresh := program.consumeManualPullRequestDiffRefresh(key)
	if message.Err == nil {
		program.cachePullRequestDiff(message.Summary, message.RawDiff)
		program.updateReviewStore(func(store reviewStore) reviewStore {
			return store.withPullRequestDiffCached(key, pullRequestDiffResult{
				data:                    buildReviewDiffData(message.RawDiff),
				sourceUpdatedAt:         pullRequestSummaryVersion(message.Summary),
				fileTeamOwnersAttempted: message.RawDiff.FileTeamOwnersAttempted,
			})
		})
		program.invalidateReviewDiffRenderCache()
		program.invalidatePullRequestDetailDocumentCache()
		program.clampReviewSessionSelection()
		if manualRefresh {
			return program.applyManualRefreshCompletion(nil)
		}
		return nil
	}

	if !program.canKeepPullRequestDiffOnRefreshError(key) {
		program.updateReviewStore(func(store reviewStore) reviewStore {
			return store.withPullRequestDiffCached(key, pullRequestDiffResult{err: message.Err, sourceUpdatedAt: pullRequestSummaryVersion(message.Summary)})
		})
		program.invalidateReviewDiffRenderCache()
		program.invalidatePullRequestDetailDocumentCache()
		program.clampReviewSessionSelection()
		if manualRefresh {
			return program.applyManualRefreshCompletion(message.Err)
		}
		return nil
	}

	cachedResult := program.pullRequestDiffCache[key]
	cachedResult.sourceUpdatedAt = pullRequestSummaryVersion(message.Summary)
	cachedResult.needsRefresh = false
	cachedResult.fileTeamOwnersAttempted = cachedResult.fileTeamOwnersAttempted || program.shouldLoadPullRequestDiffTeamOwners()
	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withPullRequestDiffCached(key, cachedResult)
	})
	program.invalidatePullRequestDetailDocumentCache()
	if manualRefresh {
		return program.applyManualRefreshCompletion(message.Err)
	}
	return nil
}

func (program *Program) applyCommitDiffLoaded(message MsgCommitDiffLoaded) {
	key := commitDiffCacheKey(message.PullRequestKey, message.CommitOID)
	if key == "" {
		return
	}

	program.updateReviewStore(func(store reviewStore) reviewStore {
		store = store.withCommitDiffLoadCleared(key)
		if message.Err == nil {
			return store.withCommitDiffCached(key, commitDiffResult{data: buildCommitDiffReviewData(message.Diff)})
		}
		return store.withCommitDiffCached(key, commitDiffResult{err: message.Err})
	})
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyIssueDetailLoaded(message MsgIssueDetailLoaded) {
	key := notificationDetailKey(message.Repository, message.Number)
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withIssueDetailLoaded(key, issueDetailResult{detail: message.Detail, err: message.Err})
	})
}

func (program *Program) applyReleaseDetailLoaded(message MsgReleaseDetailLoaded) {
	key := notificationDetailKey(message.Repository, message.ID)
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withReleaseDetailLoaded(key, releaseDetailResult{detail: message.Detail, err: message.Err})
	})
}

func (program *Program) applyCurrentDetailImageHTMLLoaded(message MsgCurrentDetailImageHTMLLoaded) {
	loadFailed := message.Err != nil || strings.TrimSpace(message.RenderedHTML) == ""
	program.recordDetailImageHTMLLoadFinished(message.Source.key, loadFailed)
	if loadFailed {
		return
	}

	if !program.applyDetailImageHTMLRendered(message.Source.applyTarget, message.RenderedHTML) {
		return
	}
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyCurrentDetailImageLoaded(message MsgCurrentDetailImageLoaded) {
	loadFailed := message.Err != nil
	program.recordDetailImageLoadFinished(message.ImageURL, loadFailed)
	if loadFailed {
		return
	}

	program.detailImageStore.Store(message.ImageURL, message.Image)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyLoadingSpinnerTick() {
	program.clearExpiredYankHighlights()
	if !program.shouldAnimateLoadingSpinner() {
		return
	}
	program.advanceLoadingSpinnerFrame()
}

func (program *Program) applyTransientErrorPopupExpired(message MsgTransientErrorPopupExpired) {
	if program.overlayState.transientErrorPopup.generation != message.Generation {
		return
	}
	program.clearExpiredTransientErrorPopup(program.currentTime())
}
