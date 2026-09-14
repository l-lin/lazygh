package tui

import "strings"

func (program *Program) statusLinePresenter() statusLinePresenter {
	if program == nil {
		return statusLinePresenter{}
	}

	return statusLinePresenter{
		feedbackMessage:                       program.feedbackMessage,
		statusLineOperationLoadingMessage:     program.statusLineOperationLoadingStatus(),
		statusLineOperationFailureMessage:     program.statusLineOperationFailureStatus(),
		statusLineOperationSuccessMessage:     program.statusLineOperationSuccessStatus(),
		statusLineOperationStarted:            program.statusLineOperationStarted(),
		loadingSpinner:                        program.loadingSpinnerFrame(),
		storyReviewLoading:                    program.storyReviewLoading,
		storyReviewLoadingMessage:             strings.TrimSpace(program.storyReviewLoadingMessage),
		assigneePickerLoadingMessage:          program.assigneePickerLoadingStatus(),
		pullRequestBuildRunLoadingMessage:     program.pullRequestBuildRunLoadingStatus(),
		ghCommandLoadingMessage:               program.ghCommandLoadingStatus(),
		selectedPullRequestDetailLoadingText:  program.selectedPullRequestDetailLoadingStatus(),
		selectedPullRequestDiffLoadingText:    program.selectedPullRequestDiffLoadingStatus(),
		selectedNotificationDetailLoadingText: program.selectedNotificationDetailLoadingStatus(),
		activePullRequestsLoadingText:         program.activePullRequestsLoadingStatus(),
		notificationsLoadingText:              program.notificationsLoadingStatus(),
	}
}

func (program *Program) pullRequestsLoading(tab PullRequestTab) bool {
	switch tab {
	case MyPullRequestsTab:
		return program.myPullRequestsLoading
	case RequestedPullRequestsTab:
		return program.requestedPullRequestsLoading
	default:
		return program.additionalPullRequestsLoading[tab]
	}
}

func (program *Program) activePullRequestsLoadingStatus() string {
	tab := program.model.ActivePullRequestTab()
	switch tab {
	case MyPullRequestsTab, RequestedPullRequestsTab:
		if program.pullRequestsLoading(tab) {
			return statusLineRefreshPullRequestListOperation().loadingLabel
		}
	default:
		if program.additionalPullRequestsLoading[tab] {
			return statusLineRefreshPullRequestListOperation().loadingLabel
		}
	}

	return ""
}

func (program *Program) selectedPullRequestDetailLoadingStatus() string {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return ""
	}
	if !program.pullRequestDetailLoadInFlight[pullRequestDetailKey(summary.Repository, summary.Number)] {
		return ""
	}

	return statusLinePullRequestOperation("Refreshing", summary).loadingLabel
}

func (program *Program) selectedPullRequestDiffLoadingStatus() string {
	summary, ok := program.selectedPullRequestSummaryForDiff()
	if !ok {
		return ""
	}
	if !program.pullRequestDiffLoadInFlight[pullRequestDetailKey(summary.Repository, summary.Number)] {
		return ""
	}

	return statusLinePullRequestOperation("Refreshing", summary).loadingLabel
}

func (program *Program) notificationsLoadingStatus() string {
	if !program.notificationsLoading {
		return ""
	}
	return statusLineRefreshNotificationsOperation().loadingLabel
}

func (program *Program) assigneePickerLoadingStatus() string {
	if !program.assigneePickerVisible() {
		return ""
	}
	if !program.actionsPopupWidget.assigneePicker.searchLoading {
		return ""
	}
	return "Searching assignees"
}

func (program *Program) pullRequestBuildRunLoadingStatus() string {
	if program.pullRequestBuildRunLoad == nil {
		return ""
	}
	return strings.TrimSpace(program.pullRequestBuildRunLoad.statusLineLabel)
}

func (program *Program) ghCommandLoadingStatus() string {
	return strings.TrimSpace(program.ghCommandLoadingMessage)
}
