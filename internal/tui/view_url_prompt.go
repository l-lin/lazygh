package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	openPullRequestByURLActionTitle              = "Open PR from URL"
	openPullRequestByURLEditorHeight             = lineModalEditorTotalHeight
	openPullRequestByClipboardInvalidMessage     = "Clipboard does not contain a GitHub pull request URL"
	openPullRequestByClipboardUnavailableMessage = "Clipboard is unavailable"
	openPullRequestByClipboardFailureMessage     = "Failed to read clipboard"
)

func (program *Program) openPullRequestByClipboardShortcut(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.mainPaneActionBlocked() || program.actionContext().IsReviewContext() {
		return nil
	}

	clipboardURL, err := program.clipboardPullRequestURL()
	if err != nil {
		program.setFeedback(program.model.Focus(), openPullRequestByClipboardFeedbackMessage(err))
		return program.refreshViewsIfGUI(gui)
	}
	if err := program.OpenPullRequestByURL(clipboardURL); err != nil {
		program.setFeedback(program.model.Focus(), strings.TrimSpace(err.Error()))
		return program.refreshViewsIfGUI(gui)
	}

	return program.refreshViewsIfGUI(gui)
}

func (program *Program) openPullRequestByURLEditor(gui *gocui.Gui) error {
	if err := program.openLineModalEditorWithHeight(gui, openPullRequestByURLActionTitle, "", program.OpenPullRequestByURL, openPullRequestByURLEditorHeight); err != nil {
		return err
	}
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) clipboardPullRequestURL() (string, error) {
	if program == nil || program.clipboardReader == nil {
		return "", ErrClipboardUnavailable
	}

	clipboardText, err := program.clipboardReader.ReadText()
	if err != nil {
		return "", err
	}
	pullRequest, err := githubdomain.ParsePullRequestURL(clipboardText)
	if err != nil {
		return "", err
	}
	return pullRequest.URL, nil
}

func openPullRequestByClipboardFeedbackMessage(err error) string {
	switch {
	case errors.Is(err, ErrClipboardUnavailable):
		return openPullRequestByClipboardUnavailableMessage
	case errors.Is(err, githubdomain.ErrInvalidPullRequestURL):
		return openPullRequestByClipboardInvalidMessage
	default:
		return openPullRequestByClipboardFailureMessage
	}
}

func (program *Program) openPullRequestByURLActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "open-pull-request-by-url",
		title:   openPullRequestByURLActionTitle,
		icon:    actionsPopupOpenPullRequestByURLIcon,
		execute: program.executeOpenPullRequestByURLAction,
	}
}

func (program *Program) executeOpenPullRequestByURLAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.openModalEditorFromActionsPopup(gui, program.openPullRequestByURLEditor)
}
