package tui

import "github.com/jesseduffield/gocui"

type detailSearchCommandRuntime struct {
	executeMessage        func(*gocui.Gui, Msg) error
	resolveView           func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailDocument func(*gocui.View) detailDocument
}

func newDetailSearchCommandRuntime(program *Program) detailSearchCommandRuntime {
	if program == nil {
		return detailSearchCommandRuntime{}
	}
	return detailSearchCommandRuntime{
		executeMessage:        program.executeRuntimeMessage,
		resolveView:           program.resolveView,
		currentDetailDocument: program.currentDetailDocument,
	}
}

type resolveDetailSearchWordCmd struct {
	Reverse bool
}

func (command resolveDetailSearchWordCmd) execute(program *Program, gui *gocui.Gui) {
	executeResolveDetailSearchWordCommand(newDetailSearchCommandRuntime(program), gui, command)
}

func executeResolveDetailSearchWordCommand(runtime detailSearchCommandRuntime, gui *gocui.Gui, command resolveDetailSearchWordCmd) {
	if runtime.executeMessage == nil || runtime.currentDetailDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	_ = runtime.executeMessage(gui, MsgDetailSearchWordResolved{
		Document:       runtime.currentDetailDocument(actualView),
		ViewportHeight: viewPageSize(actualView),
		Reverse:        command.Reverse,
	})
}
