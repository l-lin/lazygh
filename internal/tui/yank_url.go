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
	if program.model.Focus() == FocusDetailView && program.detailViewState.mode.isVisual() {
		return program.copySelectedDetailText(gui, view)
	}

	program.detailViewState.clearPendingPrefix()
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

func (program *Program) copySelectedDetailText(gui *gocui.Gui, view *gocui.View) error {
	actualView := view
	if actualView == nil && gui != nil {
		if detailView, actualErr := gui.View(viewDetailName); actualErr == nil {
			actualView = detailView
		}
	}

	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewPageSize(actualView))
	selectedText := program.detailViewState.selectedText(detailDocument)

	var err error
	switch {
	case program.clipboardWriter == nil:
		err = ErrClipboardUnavailable
	default:
		err = program.clipboardWriter.WriteText(selectedText)
	}

	program.detailViewState.exitVisualMode()
	switch {
	case err == nil:
		program.setFeedback(program.model.Focus(), detailYankSuccessMessage)
	default:
		program.setFeedback(program.model.Focus(), detailYankFailureMessage)
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
	if program.reviewSession.active {
		return true
	}

	switch program.model.Focus() {
	case FocusPullRequestsView:
		return true
	case FocusDetailView:
		return program.model.currentSideFocus() == FocusPullRequestsView
	default:
		return false
	}
}
