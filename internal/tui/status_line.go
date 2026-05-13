package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/theme"
)

func (program *Program) configureStatusLineView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
}

func (program *Program) configureStatusLineKeyHintsView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
	view.FgColor = gocui.GetColor(theme.InactiveTitleHex)
}

func (program *Program) renderStatusLineView(view *gocui.View) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(program.statusLineText()))
}

func (program *Program) renderStatusLineKeyHintsView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}

func (program *Program) statusLineText() string {
	if message := strings.TrimSpace(program.feedbackMessage); message != "" {
		return message
	}

	if message := strings.TrimSpace(program.loadingStatusText()); message != "" {
		return message
	}

	return ""
}

func (program *Program) loadingStatusText() string {
	if program.storyReviewLoading {
		return program.loadingSpinnerFrame()
	}

	if message := strings.TrimSpace(program.assigneePickerLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	if message := strings.TrimSpace(program.pullRequestBuildRunLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	if message := strings.TrimSpace(program.selectedPullRequestDetailLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	if message := strings.TrimSpace(program.selectedNotificationDetailLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	if message := strings.TrimSpace(program.activePullRequestsLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	if message := strings.TrimSpace(program.notificationsLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	return ""
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
	if program.assigneePickerLoad == nil {
		return ""
	}
	trimmedCommand := strings.TrimSpace(program.assigneePickerLoad.command)
	if trimmedCommand == "" {
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
