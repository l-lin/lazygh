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
	if message.Err != nil {
		return program.applyErrorReportedMessage(message.Err.Error())
	}

	commands := program.applyActionsPopupAsyncCompletion(message.Completion)
	if program.model != nil && program.model.ActionsPopupVisible() {
		program.clearPendingSelectionPrefix()
		program.closeActionsPopupState()
	}
	return commands
}

func (program *Program) applyNotificationMutationStarted(message MsgNotificationMutationStarted) {
	program.model.SetNotificationRows(message.OptimisticRows)
	program.clearFeedbackMessage()
	program.notificationsLoading = true
	program.notificationsLoadingDetailMessage = strings.TrimSpace(message.LoadingMessage)
}

func (program *Program) restoreNotificationMutationSnapshot(snapshot notificationMutationSnapshot) {
	program.model.SetNotificationRows(snapshot.rows)
	program.model.SelectNotificationIndex(snapshot.selectedIndex)
}

func (program *Program) applyNotificationMutationFinished(message MsgNotificationMutationFinished) []Cmd {
	program.notificationsLoading = false
	program.notificationsLoadingDetailMessage = ""
	if message.Err != nil {
		program.restoreNotificationMutationSnapshot(message.Snapshot)
		return program.applyErrorReportedMessage(message.Err.Error())
	}

	program.cacheNotifications(program.loadedNotifications())
	program.setFeedback(program.model.Focus(), message.SuccessFeedbackMessage)
	return nil
}

func (program *Program) applyStoryReviewPrepared(message MsgStoryReviewPrepared) []Cmd {
	program.finishStoryReviewLoading()
	if message.Err != nil {
		if popupMessage, ok := transientErrorPopupActionMessage(message.Err); ok {
			return program.applyErrorReportedMessage(popupMessage)
		}
		program.setFeedback(program.model.Focus(), strings.TrimSpace(message.Err.Error()))
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
	program.markAssigneePickerSearchLoading(message.Query)
}

func (program *Program) applyAssigneePickerSearchLoaded(message MsgAssigneePickerSearchLoaded) []Cmd {
	if !program.assigneePickerSearchRequestCurrent(message.RequestID, message.Query) {
		return nil
	}

	if message.Err != nil {
		program.applyAssigneePickerSearchLoadedState(message.Query, nil)
		program.clearActionsPopupErrorMessage()
		program.updateActionsPopupSearch(program.model.ActionsPopupSearchQuery())
		return program.applyErrorReportedMessage(normalizedAssigneePickerError(message.Err).Error())
	}

	program.applyAssigneePickerSearchLoadedState(message.Query, message.Results)
	program.clearActionsPopupErrorMessage()
	program.updateActionsPopupSearch(program.model.ActionsPopupSearchQuery())
	return nil
}

func (program *Program) applyPullRequestBuildRunLoaded(message MsgPullRequestBuildRunLoaded) []Cmd {
	program.pullRequestBuildRunLoad = nil
	if message.Err != nil {
		return program.applyErrorReportedMessage(normalizeGHCommandError(message.Err).Error())
	}

	popupContent := message.Target.popupContent
	popupContent.body = message.RawRunOutput
	popupContent.jobs = append([]githubdomain.PullRequestBuildRunJob(nil), message.Jobs...)
	program.openPullRequestBuildRunPopupState(popupContent)
	if message.JobsErr != nil {
		return program.applyErrorReportedMessage(normalizeGHCommandError(message.JobsErr).Error())
	}
	return nil
}

func (program *Program) applyPullRequestBuildRunJobLogLoaded(message MsgPullRequestBuildRunJobLogLoaded) []Cmd {
	program.pullRequestBuildRunLoad = nil
	if message.Err != nil {
		return program.applyErrorReportedMessage(normalizeGHCommandError(message.Err).Error())
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
