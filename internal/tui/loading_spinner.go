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

	done := make(chan struct{})
	ticker := time.NewTicker(loadingSpinnerTickInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
					if !program.shouldAnimateLoadingSpinner() {
						return nil
					}

					program.advanceLoadingSpinnerFrame()
					return program.refreshViews(gui)
				})
			}
		}
	}()

	return func() {
		close(done)
	}
}

func (program *Program) shouldAnimateLoadingSpinner() bool {
	return program.activePullRequestsLoading() || program.selectedPullRequestDetailLoading()
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

func (program *Program) loadingSpinnerFrame() string {
	if len(loadingSpinnerFrames) == 0 {
		return ""
	}

	return string(loadingSpinnerFrames[program.loadingSpinnerFrameIndex%len(loadingSpinnerFrames)])
}

func (program *Program) advanceLoadingSpinnerFrame() {
	if len(loadingSpinnerFrames) == 0 {
		return
	}

	program.loadingSpinnerFrameIndex = (program.loadingSpinnerFrameIndex + 1) % len(loadingSpinnerFrames)
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
	if program.isPullRequestLoadingItem(item) {
		return program.loadingSpinnerFrame()
	}

	return item.Title
}
