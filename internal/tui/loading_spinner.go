package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/jesseduffield/gocui"
)

const loadingSpinnerTickInterval = 120 * time.Millisecond

var loadingSpinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

func (program *Program) startLoadingSpinner(gui *gocui.Gui) func() {
	if gui == nil {
		return func() {}
	}

	program.captureGUI(gui)
	program.publishLoadingSpinnerAnimating()
	done := make(chan struct{})
	ticker := time.NewTicker(loadingSpinnerTickInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				program.tickLoadingSpinner(program.dispatchAsyncMessage)
			}
		}
	}()

	return func() {
		close(done)
	}
}

func (program *Program) publishLoadingSpinnerAnimating() {
	if program == nil {
		return
	}

	program.loadingSpinnerAnimating.Store(program.shouldAnimateLoadingSpinner())
}

func (program *Program) tickLoadingSpinner(dispatch func(Msg)) {
	if program == nil || dispatch == nil || !program.loadingSpinnerAnimating.Load() {
		return
	}

	dispatch(MsgLoadingSpinnerTick{})
}

func (program *Program) shouldAnimateLoadingSpinner() bool {
	return program.storyReviewLoading || program.assigneePickerLoading() || program.activePullRequestsLoading() || program.notificationsLoading || program.selectedPullRequestDetailLoading() || program.selectedNotificationDetailLoading() || program.pullRequestBuildRunLoading() || program.ghCommandLoading() || program.hasYankHighlights()
}

func (program *Program) activePullRequestsLoading() bool {
	tab := program.model.ActivePullRequestTab()
	switch tab {
	case MyPullRequestsTab:
		return program.myPullRequestsLoading
	case RequestedPullRequestsTab:
		return program.requestedPullRequestsLoading
	default:
		return program.additionalPullRequestsLoading[tab]
	}
}

func (program *Program) selectedPullRequestDetailLoading() bool {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return false
	}

	return program.pullRequestDetailLoadInFlight[pullRequestDetailKey(summary.Repository, summary.Number)]
}

func (program *Program) pullRequestBuildRunLoading() bool {
	return program != nil && program.pullRequestBuildRunLoad != nil
}

func (program *Program) ghCommandLoading() bool {
	return program != nil && strings.TrimSpace(program.ghCommandLoadingMessage) != ""
}

func (program *Program) loadingSpinnerFrame() string {
	if len(loadingSpinnerFrames) == 0 {
		return ""
	}

	return string(loadingSpinnerFrames[program.startupState.loadingSpinnerFrameIndex%len(loadingSpinnerFrames)])
}

func (program *Program) advanceLoadingSpinnerFrame() {
	if len(loadingSpinnerFrames) == 0 {
		return
	}

	program.advanceStartupLoadingSpinnerFrame(len(loadingSpinnerFrames))
}

func (program *Program) loadingSpinnerStatus(label string) string {
	spinner := program.loadingSpinnerFrame()
	trimmedLabel := strings.TrimSpace(label)
	switch {
	case spinner == "":
		return trimmedLabel
	case trimmedLabel == "":
		return spinner
	default:
		return fmt.Sprintf("%s %s", spinner, trimmedLabel)
	}
}

func (program *Program) displayItemTitle(item Item) string {
	if program.isPullRequestLoadingItem(item) || program.isNotificationLoadingItem(item) {
		return program.loadingSpinnerFrame()
	}

	return item.Title
}
