package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) layoutStatusLineView(gui *gocui.Gui) error {
	view, err := program.layoutBottomPromptView(gui, viewStatusLineName)
	if err != nil {
		return err
	}

	program.configureStatusLineView(view)
	program.renderStatusLineView(view)
	_, err = gui.SetViewOnTop(viewStatusLineName)
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

func (program *Program) configureStatusLineView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
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
	if message := strings.TrimSpace(program.selectedPullRequestDetailLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	if message := strings.TrimSpace(program.activePullRequestsLoadingStatus()); message != "" {
		return program.loadingSpinnerStatus(message)
	}

	return ""
}

func (program *Program) activePullRequestsLoadingStatus() string {
	switch program.model.ActivePullRequestTab() {
	case RequestedPullRequestsTab:
		if program.requestedPullRequestsLoading {
			return requestedPullRequestsLoadingDetail
		}
	default:
		if program.myPullRequestsLoading {
			return myPullRequestsLoadingDetail
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
