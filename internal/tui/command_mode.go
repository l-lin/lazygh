package tui

import "github.com/jesseduffield/gocui"

const (
	commandPromptPrefix = ":"
	commandMessages     = "messages"
	commandHistory      = "history"
)

func (program *Program) commandModeActive() bool {
	return program != nil && program.searchWidget.commandModeActive()
}

func (program *Program) openCommandMode(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgOpenCommandMode{})
}

func (program *Program) submitCommand(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgSubmitCommand{})
}

func (program *Program) cancelCommand(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgCancelCommand{})
}

func (program *Program) searchSubmitKeybindingAction() keybindingAction {
	if program.commandModeActive() {
		return fixedKeybindingActionFor(keymapScopeSearch, "submit_command", []string{viewSearchName}, program.submitCommand, "enter", "ctrl+j")
	}
	return configuredKeybindingActionFor(keymapScopeSearch, "submit", []string{viewSearchName}, program.submitSearch)
}

func (program *Program) searchCancelKeybindingAction() keybindingAction {
	if program.commandModeActive() {
		return fixedKeybindingActionFor(keymapScopeSearch, "cancel_command", []string{viewSearchName}, program.cancelCommand, "esc")
	}
	return configuredKeybindingActionFor(keymapScopeSearch, "cancel", []string{viewSearchName}, program.cancelSearch)
}

func (program *Program) commandModeBlocked() bool {
	return program.overlayState.helpVisible ||
		program.model.SearchActive() ||
		program.model.ActionsPopupVisible() ||
		program.modalEditorVisible() ||
		program.pullRequestBuildRunPopupVisible()
}
