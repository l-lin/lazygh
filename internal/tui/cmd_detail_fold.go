package tui

import "github.com/jesseduffield/gocui"

type detailFoldCommandRuntime struct {
	dispatch                           func(*gocui.Gui, Msg) error
	resolveView                        func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailDocument              func(*gocui.View) detailDocument
	toggleInlineConversationVisibility func(detailDocument) (detailViewSyncPlan, bool)
	setAllDetailFolds                  func(detailDocument, bool) (detailViewSyncPlan, bool)
}

func newDetailFoldCommandRuntime(program *Program) detailFoldCommandRuntime {
	if program == nil {
		return detailFoldCommandRuntime{}
	}
	return detailFoldCommandRuntime{
		dispatch:                           program.dispatch,
		resolveView:                        program.resolveView,
		currentDetailDocument:              program.currentDetailDocument,
		toggleInlineConversationVisibility: program.toggleInlineConversationVisibilityState,
		setAllDetailFolds:                  program.setAllDetailFolds,
	}
}

type toggleInlineConversationVisibilityCmd struct{}

func (command toggleInlineConversationVisibilityCmd) execute(program *Program, gui *gocui.Gui) {
	executeToggleInlineConversationVisibilityCommand(newDetailFoldCommandRuntime(program), gui, command)
}

func executeToggleInlineConversationVisibilityCommand(runtime detailFoldCommandRuntime, gui *gocui.Gui, command toggleInlineConversationVisibilityCmd) {
	if runtime.dispatch == nil || runtime.currentDetailDocument == nil || runtime.toggleInlineConversationVisibility == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	detailDocument := runtime.currentDetailDocument(actualView)
	plan, ok := runtime.toggleInlineConversationVisibility(detailDocument)
	if !ok {
		return
	}
	_ = runtime.dispatch(gui, MsgDetailViewSyncPlanResolved{Plan: plan, ViewportHeight: viewPageSize(actualView)})
}

type setAllDetailFoldsCmd struct {
	Collapsed bool
}

func (command setAllDetailFoldsCmd) execute(program *Program, gui *gocui.Gui) {
	executeSetAllDetailFoldsCommand(newDetailFoldCommandRuntime(program), gui, command)
}

func executeSetAllDetailFoldsCommand(runtime detailFoldCommandRuntime, gui *gocui.Gui, command setAllDetailFoldsCmd) {
	if runtime.dispatch == nil || runtime.currentDetailDocument == nil || runtime.setAllDetailFolds == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	detailDocument := runtime.currentDetailDocument(actualView)
	plan, ok := runtime.setAllDetailFolds(detailDocument, command.Collapsed)
	if !ok {
		return
	}
	_ = runtime.dispatch(gui, MsgDetailViewSyncPlanResolved{Plan: plan, ViewportHeight: viewPageSize(actualView)})
}
