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
	program.actionsPopupWidget.clearSearchEditor()
	program.clearActionsPopupPendingConfirmation()
	program.actionsPopupWidget.errorMessage = ""
	program.actionsPopupWidget.reactionPicker = nil
	program.actionsPopupWidget.themePicker = nil
	program.actionsPopupWidget.assigneePicker = nil
	program.actionsPopupWidget.assigneePickerLoad = nil
}

func (program *Program) applyActionsPopupAsyncGHCommandFinished(message MsgActionsPopupAsyncGHCommandFinished) []Cmd {
	program.clearGHCommandLoading()
	if message.Err != nil {
		return []Cmd{reportErrorCmd{Message: strings.TrimSpace(message.Err.Error())}}
	}

	var commands []Cmd
	if message.Success != nil {
		commands = message.Success.apply(program)
	}
	if program.model != nil && program.model.ActionsPopupVisible() {
		program.clearPendingSelectionPrefix()
		program.closeActionsPopupState()
	}
	return commands
}

func (program *Program) applyNotificationMutationStarted(message MsgNotificationMutationStarted) {
	program.model.SetNotificationRows(message.OptimisticRows)
	program.feedbackMessage = ""
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
		return []Cmd{reportErrorCmd{Message: strings.TrimSpace(message.Err.Error())}}
	}

	program.cacheNotifications(program.loadedNotifications())
	program.setFeedback(program.model.Focus(), message.SuccessFeedbackMessage)
	return nil
}

func (program *Program) applyStoryReviewPrepared(message MsgStoryReviewPrepared) []Cmd {
	program.storyReviewLoading = false
	if message.Err != nil {
		if popupMessage, ok := transientErrorPopupActionMessage(message.Err); ok {
			return []Cmd{reportErrorCmd{Message: popupMessage}}
		}
		program.setFeedback(program.model.Focus(), strings.TrimSpace(message.Err.Error()))
		return nil
	}

	program.feedbackMessage = ""
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

	trimmedQuery := strings.TrimSpace(message.Query)
	program.actionsPopupWidget.assigneePicker.searchLoading = false
	program.actionsPopupWidget.assigneePicker.searchCommand = ""
	program.actionsPopupWidget.assigneePicker.searchQuery = trimmedQuery
	if message.Err != nil {
		program.actionsPopupWidget.assigneePicker.searchResults = nil
		program.actionsPopupWidget.errorMessage = ""
		program.syncActionsPopupSearch()
		return []Cmd{reportErrorCmd{Message: strings.TrimSpace(normalizedAssigneePickerError(message.Err).Error())}}
	}

	program.actionsPopupWidget.assigneePicker.rememberCandidates(message.Results)
	program.actionsPopupWidget.assigneePicker.searchResults = append([]githubdomain.PullRequestAuthor(nil), message.Results...)
	program.actionsPopupWidget.errorMessage = ""
	program.syncActionsPopupSearch()
	return nil
}

func (program *Program) applyPullRequestBuildRunLoaded(message MsgPullRequestBuildRunLoaded) []Cmd {
	program.pullRequestBuildRunLoad = nil
	if message.Err != nil {
		return []Cmd{reportErrorCmd{Message: strings.TrimSpace(normalizeGHCommandError(message.Err).Error())}}
	}

	popupContent := message.Target.popupContent
	popupContent.body = message.RawRunOutput
	popupContent.jobs = append([]githubdomain.PullRequestBuildRunJob(nil), message.Jobs...)
	program.openPullRequestBuildRunPopupState(popupContent)
	if message.JobsErr != nil {
		return []Cmd{reportErrorCmd{Message: strings.TrimSpace(normalizeGHCommandError(message.JobsErr).Error())}}
	}
	return nil
}

func (program *Program) applyPullRequestBuildRunJobLogLoaded(message MsgPullRequestBuildRunJobLogLoaded) []Cmd {
	program.pullRequestBuildRunLoad = nil
	if message.Err != nil {
		return []Cmd{reportErrorCmd{Message: strings.TrimSpace(normalizeGHCommandError(message.Err).Error())}}
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
