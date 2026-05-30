package tui

import "github.com/jesseduffield/gocui"

const (
	enableDetailWordWrapActionTitle  = "Enable word wrap"
	disableDetailWordWrapActionTitle = "Disable word wrap"
	detailWordWrapHelpLabel          = "Toggle word wrap"
	detailWordWrapEnabledMessage     = "Word wrap enabled"
	detailWordWrapDisabledMessage    = "Word wrap disabled"
)

func detailWordWrapActionTitle(enabled bool) string {
	if enabled {
		return disableDetailWordWrapActionTitle
	}
	return enableDetailWordWrapActionTitle
}

func (program *Program) detailWordWrapEnabled() bool {
	if program == nil {
		return true
	}
	return program.detailState.wordWrapEnabled()
}

func markdownRenderWidthForWordWrap(width int, wordWrapEnabled bool) int {
	if !wordWrapEnabled {
		return disabledMarkdownWordWrap
	}
	return width
}

func (program *Program) detailMarkdownRenderWidth(width int) int {
	return markdownRenderWidthForWordWrap(width, program.detailWordWrapEnabled())
}

func detailWordWrapFeedbackMessage(enabled bool) string {
	if enabled {
		return detailWordWrapEnabledMessage
	}
	return detailWordWrapDisabledMessage
}

func (program *Program) toggleDetailWordWrap(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgToggleDetailWordWrapRequested{})
}

func (program *Program) currentDetailWordWrapAction() (actionsPopupAction, bool) {
	if program == nil || program.model == nil || program.model.Focus() != FocusDetailView {
		return actionsPopupAction{}, false
	}
	return actionsPopupAction{
		id:        "toggle-word-wrap",
		title:     detailWordWrapActionTitle(program.detailWordWrapEnabled()),
		icon:      actionsPopupWordWrapIcon,
		keywords:  []string{"wrap", "nowrap", "soft wrap", "line wrap"},
		requested: MsgToggleDetailWordWrapRequested{},
	}, true
}
