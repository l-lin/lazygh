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
	yankSuccessMessage     = iconCopy + " URL copied"
	yankFailureMessage     = iconWarning + " Copy failed"
	yankUnavailableMessage = iconUnavailable + " No PR URL"
)

func (program *Program) copyPullRequestURL(gui *gocui.Gui, view *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.helpVisible || program.model.SearchActive() {
		return nil
	}
	return program.dispatch(gui, MsgCopyPullRequestURLRequested{View: view})
}

func (program *Program) copySelectedDetailText(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgCopySelectedDetailTextRequested{View: view})
}

func (program *Program) copySelectedPullRequestURL() error {
	url, ok := program.selectedPullRequestURL()
	if !ok {
		return ErrNoPullRequestURL
	}
	return program.writeTextToClipboard(url)
}

func (program *Program) selectedPullRequestURL() (string, bool) {
	summary, ok := program.currentPullRequestSummary()
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
	_, ok := program.currentPullRequestSummary()
	return ok
}
