package tui

import (
	"fmt"
	"strings"
)

func (program *Program) statusLinePresenter() statusLinePresenter {
	if program == nil {
		return statusLinePresenter{}
	}

	return statusLinePresenter{
		feedbackMessage:                       program.feedbackMessage,
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

func (program *Program) activePullRequestsLoadingStatus() string {
	tab := program.model.ActivePullRequestTab()
	switch tab {
	case MyPullRequestsTab:
		if program.myPullRequestsLoading {
			return program.pullRequestListState(tab).loadingDetail
		}
	case RequestedPullRequestsTab:
		if program.requestedPullRequestsLoading {
			return program.pullRequestListState(tab).loadingDetail
		}
	default:
		if program.additionalPullRequestsLoading[tab] {
			return program.pullRequestListState(tab).loadingDetail
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

	return fmt.Sprintf("Running `gh pr view %d -R %s --json ...`.", summary.Number, pullRequestRepositoryName(summary.Repository))
}

func (program *Program) selectedPullRequestDiffLoadingStatus() string {
	summary, ok := program.selectedPullRequestSummaryForDiff()
	if !ok {
		return ""
	}
	if !program.pullRequestDiffLoadInFlight[pullRequestDetailKey(summary.Repository, summary.Number)] {
		return ""
	}

	return fmt.Sprintf("Running `gh api repos/%s/pulls/%d -H 'Accept: application/vnd.github.v3.diff'`.", pullRequestRepositoryName(summary.Repository), summary.Number)
}

func (program *Program) notificationsLoadingStatus() string {
	if !program.notificationsLoading {
		return ""
	}
	if message := strings.TrimSpace(program.notificationsLoadingDetailMessage); message != "" {
		return message
	}
	return notificationsLoadingDetail
}

func (program *Program) assigneePickerLoadingStatus() string {
	if !program.assigneePickerVisible() {
		return ""
	}
	trimmedCommand := strings.TrimSpace(program.actionsPopupWidget.assigneePicker.searchCommand)
	if !program.actionsPopupWidget.assigneePicker.searchLoading || trimmedCommand == "" {
		return ""
	}
	return fmt.Sprintf("Running `%s`.", trimmedCommand)
}

func (program *Program) pullRequestBuildRunLoadingStatus() string {
	if program.pullRequestBuildRunLoad == nil {
		return ""
	}
	trimmedCommand := strings.TrimSpace(program.pullRequestBuildRunLoad.command)
	if trimmedCommand == "" {
		return ""
	}
	return fmt.Sprintf("Running `%s`.", trimmedCommand)
}

func (program *Program) ghCommandLoadingStatus() string {
	return strings.TrimSpace(program.ghCommandLoadingMessage)
}
