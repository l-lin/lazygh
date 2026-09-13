package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) closeActionsPopupState() {
	if program == nil {
		return
	}
	program.model.CloseActionsPopup()
	program.resetActionsPopupWidgetChrome()
}

func (program *Program) applyActionsPopupAsyncGHCommandFinished(message MsgActionsPopupAsyncGHCommandFinished) []Cmd {
	program.clearGHCommandLoading()
	program.finishStatusLineOperation(message.StatusLineOperationID, message.Err)
	if message.Err != nil {
		if target, previousCollapsed, ok := inlineCommentResolutionRollback(message.Err); ok {
			program.applyInlineCommentResolutionCollapsed(target, previousCollapsed)
		}
		if snapshot, ok := pullRequestMergeQueueRollback(message.Err); ok {
			program.restorePullRequestMergeQueueMutationSnapshot(snapshot)
		}
		return nil
	}

	commands := program.applyActionsPopupAsyncCompletion(message.Completion)
	program.clearFeedbackMessage()
	if program.model != nil && program.model.ActionsPopupVisible() {
		program.clearPendingSelectionPrefix()
		program.closeActionsPopupState()
	}
	return commands
}

func (program *Program) applyNotificationMutationStarted(message MsgNotificationMutationStarted) uint64 {
	program.model.SetNotificationRows(message.OptimisticRows)
	program.clearFeedbackMessage()
	statusLabel := strings.TrimRight(strings.TrimSpace(message.LoadingMessage), ".")
	statusLineOperationID := program.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: statusLabel, failureLabel: statusLabel})
	program.startNotificationMutationLoading(message.LoadingMessage)
	program.setNotificationsStatusLineOperationID(statusLineOperationID)
	return statusLineOperationID
}

func (program *Program) restoreNotificationMutationSnapshot(snapshot notificationMutationSnapshot) {
	program.model.SetNotificationRows(snapshot.rows)
	program.model.SelectNotificationIndex(snapshot.selectedIndex)
}

func (program *Program) applyNotificationMutationFinished(message MsgNotificationMutationFinished) []Cmd {
	program.finishNotificationsLoading()
	program.finishStatusLineOperation(message.StatusLineOperationID, message.Err)
	if message.Err != nil {
		program.restoreNotificationMutationSnapshot(message.Snapshot)
		return nil
	}

	program.cacheNotifications(program.loadedNotifications())
	program.clearFeedbackMessage()
	return nil
}

func (program *Program) applyStoryReviewPrepared(message MsgStoryReviewPrepared) []Cmd {
	program.finishStoryReviewLoading()
	program.finishStatusLineOperation(message.StatusLineOperationID, message.Err)
	if message.Err != nil {
		return nil
	}

	program.clearFeedbackMessage()
	program.applyPreparedStoryReview(message.Prepared)
	return nil
}

func (program *Program) applyAssigneePickerSearchLoadingStarted(message MsgAssigneePickerSearchLoadingStarted) {
	if !program.assigneePickerSearchRequestCurrent(message.RequestID, message.Query) {
		return
	}
	statusLineOperationID := program.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: "Searching assignees", failureLabel: "Searching assignees"})
	program.markAssigneePickerSearchLoading(message.Query, statusLineOperationID)
}

func (program *Program) applyAssigneePickerSearchLoaded(message MsgAssigneePickerSearchLoaded) []Cmd {
	if !program.assigneePickerSearchRequestCurrent(message.RequestID, message.Query) {
		return nil
	}

	statusLineOperationID := program.actionsPopupWidget.assigneePicker.searchStatusOperationID
	program.finishStatusLineOperation(statusLineOperationID, message.Err)
	if message.Err != nil {
		program.applyAssigneePickerSearchLoadedState(message.Query, nil)
		program.clearActionsPopupErrorMessage()
		program.updateActionsPopupSearch(program.model.ActionsPopupSearchQuery())
		return nil
	}

	program.applyAssigneePickerSearchLoadedState(message.Query, message.Results)
	program.clearActionsPopupErrorMessage()
	program.updateActionsPopupSearch(program.model.ActionsPopupSearchQuery())
	return nil
}

func (program *Program) applyPullRequestBuildRunLoaded(message MsgPullRequestBuildRunLoaded) []Cmd {
	statusLineOperationID := uint64(0)
	if program.pullRequestBuildRunLoad != nil {
		statusLineOperationID = program.pullRequestBuildRunLoad.statusLineOperationID
	}
	operationErr := message.Err
	if operationErr == nil {
		operationErr = message.JobsErr
	}
	program.clearPullRequestBuildRunLoad()
	program.finishStatusLineOperation(statusLineOperationID, operationErr)
	if message.Err != nil {
		return nil
	}

	popupContent := message.Target.popupContent
	popupContent.body = message.RawRunOutput
	popupContent.jobs = append([]githubdomain.PullRequestBuildRunJob(nil), message.Jobs...)
	program.openPullRequestBuildRunPopupState(popupContent)
	return nil
}

func (program *Program) applyPullRequestBuildRunJobLogLoaded(message MsgPullRequestBuildRunJobLogLoaded) []Cmd {
	statusLineOperationID := uint64(0)
	if program.pullRequestBuildRunLoad != nil {
		statusLineOperationID = program.pullRequestBuildRunLoad.statusLineOperationID
	}
	program.clearPullRequestBuildRunLoad()
	program.finishStatusLineOperation(statusLineOperationID, message.Err)
	if message.Err != nil {
		return nil
	}

	program.openPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{
		title:         pullRequestBuildRunLogsPopupTitle(message.Job.Name),
		runURL:        strings.TrimSpace(message.Job.URL),
		repository:    message.Repository,
		body:          sanitizePullRequestBuildRunLog(message.RawLogOutput),
		widthPercent:  pullRequestBuildLogsPopupWidthPercent,
		heightPercent: pullRequestBuildLogsPopupHeightPercent,
	})
	return nil
}
