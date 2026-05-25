package tui

import "strings"

func (program *Program) applyConnectedUserLoaded(message MsgConnectedUserLoaded) {
	connectedUserLogin := ""
	connectedUserName := ""
	if message.Err == nil {
		connectedUserLogin = strings.TrimSpace(message.User.Login)
		connectedUserName = strings.TrimSpace(message.User.Name)
	}
	if program.connectedUserLogin != connectedUserLogin {
		program.connectedUserLogin = connectedUserLogin
		program.invalidatePullRequestDetailDocumentCache()
		program.invalidateReviewDiffRenderCache()
	}
	program.connectedUserName = connectedUserName
	program.model.SetUsers([]Item{connectedUserStateItem(message.User, message.Err)})
}

func (program *Program) applyPullRequestsLoaded(message MsgPullRequestsLoaded) {
	program.setPullRequestsLoading(message.Tab, false)
	manualRefresh := program.consumeManualPullRequestListRefresh(message.Tab)
	if message.Err == nil {
		program.cachePullRequests(message.Tab, message.PullRequests)
		rows := program.pullRequestRowsForTab(message.Tab, message.PullRequests, nil)
		program.setPullRequestsCount(message.Tab, pullRequestSummaryRowCount(rows), true)
		program.model.SetPullRequestRows(message.Tab, rows)
		program.selectOpenedPullRequestRow(message.Tab)
		if manualRefresh {
			program.completeManualRefreshOperation(program.gui, nil)
		}
		return
	}

	if manualRefresh {
		program.completeManualRefreshOperation(program.gui, message.Err)
	}
	if !program.shouldPreservePullRequestRowsOnRefreshError(message.Tab) {
		program.setPullRequestsCount(message.Tab, 0, false)
		program.model.SetPullRequestRows(message.Tab, program.pullRequestRowsForTab(message.Tab, nil, message.Err))
	}
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

func (program *Program) applyNotificationsLoaded(message MsgNotificationsLoaded) {
	program.notificationsLoading = false
	program.notificationsLoadingDetailMessage = ""
	manualRefresh := program.consumeManualNotificationRefresh()
	if message.Err == nil {
		filteredNotifications := program.filterDoneNotifications(message.Notifications)
		program.cacheNotifications(filteredNotifications)
		program.model.SetNotificationRows(notificationRows(filteredNotifications))
		if manualRefresh {
			program.completeManualRefreshOperation(program.gui, nil)
		}
		return
	}

	if manualRefresh {
		program.completeManualRefreshOperation(program.gui, message.Err)
	}
	if !program.shouldPreserveNotificationRowsOnRefreshError() {
		program.model.SetNotificationRows(notificationsStateRows(nil, message.Err))
	}
}

func (program *Program) applyPullRequestDetailLoaded(message MsgPullRequestDetailLoaded) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}

	delete(program.pullRequestDetailLoadInFlight, key)
	manualRefresh := program.consumeManualPullRequestDetailRefresh(key)
	if message.PendingReviewStateKnown {
		program.pendingPullRequestReviewCache[key] = message.PendingReviewState
	}
	if message.Err == nil {
		clonedDetail := clonePullRequestDetail(message.Detail)
		program.cachePullRequestDetail(message.Summary, clonedDetail)
		program.pullRequestDetailCache[key] = pullRequestDetailResult{detail: clonedDetail, sourceUpdatedAt: pullRequestSummaryVersion(message.Summary)}
		program.invalidatePullRequestDetailDocumentCache()
		if manualRefresh {
			program.completeManualRefreshOperation(program.gui, nil)
		}
		return
	}

	if !program.canKeepPullRequestDetailOnRefreshError(key) {
		program.pullRequestDetailCache[key] = pullRequestDetailResult{err: message.Err, sourceUpdatedAt: pullRequestSummaryVersion(message.Summary)}
		program.invalidatePullRequestDetailDocumentCache()
		if manualRefresh {
			program.completeManualRefreshOperation(program.gui, message.Err)
		}
		return
	}

	cachedResult := program.pullRequestDetailCache[key]
	cachedResult.sourceUpdatedAt = pullRequestSummaryVersion(message.Summary)
	cachedResult.needsRefresh = false
	program.pullRequestDetailCache[key] = cachedResult
	if manualRefresh {
		program.completeManualRefreshOperation(program.gui, message.Err)
	}
}

func (program *Program) applyPullRequestDiffLoaded(message MsgPullRequestDiffLoaded) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}

	delete(program.pullRequestDiffLoadInFlight, key)
	manualRefresh := program.consumeManualPullRequestDiffRefresh(key)
	if message.Err == nil {
		program.cachePullRequestDiff(message.Summary, message.RawDiff)
		program.pullRequestDiffCache[key] = pullRequestDiffResult{
			data:                    buildReviewDiffData(message.RawDiff),
			sourceUpdatedAt:         pullRequestSummaryVersion(message.Summary),
			fileTeamOwnersAttempted: message.RawDiff.FileTeamOwnersAttempted,
		}
		program.invalidateReviewDiffRenderCache()
		program.invalidatePullRequestDetailDocumentCache()
		program.clampReviewSessionSelection()
		if manualRefresh {
			program.completeManualRefreshOperation(program.gui, nil)
		}
		return
	}

	if !program.canKeepPullRequestDiffOnRefreshError(key) {
		program.pullRequestDiffCache[key] = pullRequestDiffResult{err: message.Err, sourceUpdatedAt: pullRequestSummaryVersion(message.Summary)}
		program.invalidateReviewDiffRenderCache()
		program.invalidatePullRequestDetailDocumentCache()
		program.clampReviewSessionSelection()
		if manualRefresh {
			program.completeManualRefreshOperation(program.gui, message.Err)
		}
		return
	}

	cachedResult := program.pullRequestDiffCache[key]
	cachedResult.sourceUpdatedAt = pullRequestSummaryVersion(message.Summary)
	cachedResult.needsRefresh = false
	cachedResult.fileTeamOwnersAttempted = cachedResult.fileTeamOwnersAttempted || program.shouldLoadPullRequestDiffTeamOwners()
	program.pullRequestDiffCache[key] = cachedResult
	program.invalidatePullRequestDetailDocumentCache()
	if manualRefresh {
		program.completeManualRefreshOperation(program.gui, message.Err)
	}
}

func (program *Program) applyIssueDetailLoaded(message MsgIssueDetailLoaded) {
	key := notificationDetailKey(message.Repository, message.Number)
	if key == "" {
		return
	}
	delete(program.issueDetailLoadInFlight, key)
	program.issueDetailCache[key] = issueDetailResult{detail: message.Detail, err: message.Err}
}

func (program *Program) applyReleaseDetailLoaded(message MsgReleaseDetailLoaded) {
	key := notificationDetailKey(message.Repository, message.ID)
	if key == "" {
		return
	}
	delete(program.releaseDetailLoadInFlight, key)
	program.releaseDetailCache[key] = releaseDetailResult{detail: message.Detail, err: message.Err}
}

func (program *Program) applyCurrentDetailImageHTMLLoaded(message MsgCurrentDetailImageHTMLLoaded) {
	delete(program.detailImageHTMLLoadInFlight, message.Source.key)
	if message.Err != nil || strings.TrimSpace(message.RenderedHTML) == "" {
		program.detailImageHTMLLoadFailed[message.Source.key] = true
		return
	}

	delete(program.detailImageHTMLLoadFailed, message.Source.key)
	message.Source.applyRenderedHTML(program, message.RenderedHTML)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyCurrentDetailImageLoaded(message MsgCurrentDetailImageLoaded) {
	delete(program.detailImageLoadInFlight, message.ImageURL)
	if message.Err != nil {
		program.detailImageLoadFailed[message.ImageURL] = true
		return
	}

	delete(program.detailImageLoadFailed, message.ImageURL)
	program.detailImageStore.Store(message.ImageURL, message.Image)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyLoadingSpinnerTick() {
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
