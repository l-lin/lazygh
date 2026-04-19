package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

var (
	ErrNoPullRequestURL     = errors.New("no pull request url")
	ErrClipboardUnavailable = errors.New("clipboard is unavailable")
)

const (
	yankSuccessMessage     = "󰆏 URL copied"
	yankFailureMessage     = "󰅚 Copy failed"
	yankUnavailableMessage = "󰌑 No PR URL"
)

func (program *Program) copyPullRequestURL(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() {
		return nil
	}

	err := program.copySelectedPullRequestURL()
	switch {
	case err == nil:
		program.setFeedback(program.model.Focus(), yankSuccessMessage)
	case errors.Is(err, ErrNoPullRequestURL):
		program.setFeedback(program.model.Focus(), yankUnavailableMessage)
	default:
		program.setFeedback(program.model.Focus(), yankFailureMessage)
	}

	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) copySelectedPullRequestURL() error {
	url, ok := program.selectedPullRequestURL()
	if !ok {
		return ErrNoPullRequestURL
	}
	if program.clipboardWriter == nil {
		return ErrClipboardUnavailable
	}

	return program.clipboardWriter.WriteText(url)
}

func (program *Program) selectedPullRequestURL() (string, bool) {
	if !program.isPullRequestContext() {
		return "", false
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return "", false
	}

	if result, ok := program.pullRequestDetailForSummary(summary); ok {
		url := strings.TrimSpace(result.detail.URL)
		if url != "" {
			return url, true
		}
	}

	url := strings.TrimSpace(summary.URL)
	if url == "" {
		return "", false
	}

	return url, true
}

func (program *Program) isPullRequestContext() bool {
	switch program.model.Focus() {
	case FocusPullRequestsView:
		return true
	case FocusDetailView:
		return program.model.currentSideFocus() == FocusPullRequestsView
	default:
		return false
	}
}
