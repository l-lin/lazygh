package tui

import (
	"errors"

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
	return program.dispatch(gui, MsgReadPullRequestURLFromClipboardRequested{})
}

func (program *Program) openPullRequestByURLEditor(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgOpenPullRequestByURLPromptRequested{})
}

func openPullRequestByURLPromptDescriptor() modalEditorOpenDescriptor {
	return newLineModalEditorOpenDescriptorWithHeightAndSubmitDescriptor(openPullRequestByURLActionTitle, "", newOpenPullRequestByURLSubmitDescriptor(), openPullRequestByURLEditorHeight)
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
		id:        "open-pull-request-by-url",
		title:     openPullRequestByURLActionTitle,
		icon:      actionsPopupOpenPullRequestByURLIcon,
		requested: MsgModalEditorOpened{Descriptor: openPullRequestByURLPromptDescriptor()},
	}
}
