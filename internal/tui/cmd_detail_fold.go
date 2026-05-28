package tui

import "github.com/jesseduffield/gocui"

type detailFoldCommandRuntime struct {
	dispatch              func(*gocui.Gui, Msg) error
	resolveView           func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailDocument func(*gocui.View) detailDocument
}

func newDetailFoldCommandRuntime(program *Program) detailFoldCommandRuntime {
	if program == nil {
		return detailFoldCommandRuntime{}
	}
	return detailFoldCommandRuntime{
		dispatch:              program.dispatch,
		resolveView:           program.resolveView,
		currentDetailDocument: program.currentDetailDocument,
	}
}

type toggleInlineConversationVisibilityCmd struct{}

func (command toggleInlineConversationVisibilityCmd) execute(program *Program, gui *gocui.Gui) {
	executeToggleInlineConversationVisibilityCommand(newDetailFoldCommandRuntime(program), gui, command)
}

func executeToggleInlineConversationVisibilityCommand(runtime detailFoldCommandRuntime, gui *gocui.Gui, command toggleInlineConversationVisibilityCmd) {
	if runtime.dispatch == nil || runtime.currentDetailDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	_ = runtime.dispatch(gui, MsgToggleInlineConversationVisibilityResolved{Document: runtime.currentDetailDocument(actualView), ViewportHeight: viewPageSize(actualView)})
}

type setAllDetailFoldsCmd struct {
	Collapsed bool
}

func (command setAllDetailFoldsCmd) execute(program *Program, gui *gocui.Gui) {
	executeSetAllDetailFoldsCommand(newDetailFoldCommandRuntime(program), gui, command)
}

func executeSetAllDetailFoldsCommand(runtime detailFoldCommandRuntime, gui *gocui.Gui, command setAllDetailFoldsCmd) {
	if runtime.dispatch == nil || runtime.currentDetailDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	_ = runtime.dispatch(gui, MsgSetAllDetailFoldsResolved{Collapsed: command.Collapsed, Document: runtime.currentDetailDocument(actualView), ViewportHeight: viewPageSize(actualView)})
}
