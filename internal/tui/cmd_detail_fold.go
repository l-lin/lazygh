package tui

import "github.com/jesseduffield/gocui"

type detailFoldCommandRuntime struct {
	resolveView                        func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailDocument              func(*gocui.View) detailDocument
	syncDetailViewState                func(detailDocument, int)
	placeDetailCursorAtLine            func(detailDocument, int)
	toggleInlineConversationVisibility func(detailDocument) (detailViewSyncPlan, bool)
	setAllDetailFolds                  func(detailDocument, bool) (detailViewSyncPlan, bool)
}

func newDetailFoldCommandRuntime(program *Program) detailFoldCommandRuntime {
	if program == nil {
		return detailFoldCommandRuntime{}
	}
	return detailFoldCommandRuntime{
		resolveView:                        program.resolveView,
		currentDetailDocument:              program.currentDetailDocument,
		syncDetailViewState:                program.syncDetailViewState,
		placeDetailCursorAtLine:            program.placeDetailCursorAtLine,
		toggleInlineConversationVisibility: program.toggleInlineConversationVisibilityState,
		setAllDetailFolds:                  program.setAllDetailFolds,
	}
}

type toggleInlineConversationVisibilityCmd struct {
	View *gocui.View
}

func (command toggleInlineConversationVisibilityCmd) execute(program *Program, gui *gocui.Gui) {
	executeToggleInlineConversationVisibilityCommand(newDetailFoldCommandRuntime(program), gui, command)
}

func executeToggleInlineConversationVisibilityCommand(runtime detailFoldCommandRuntime, gui *gocui.Gui, command toggleInlineConversationVisibilityCmd) {
	if runtime.currentDetailDocument == nil || runtime.toggleInlineConversationVisibility == nil || runtime.syncDetailViewState == nil {
		return
	}

	actualView := command.View
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, command.View, viewDetailName)
	}
	detailDocument := runtime.currentDetailDocument(actualView)
	plan, ok := runtime.toggleInlineConversationVisibility(detailDocument)
	if !ok {
		return
	}
	applyDetailViewSyncPlan(runtime, plan, viewPageSize(actualView))
}

type setAllDetailFoldsCmd struct {
	View      *gocui.View
	Collapsed bool
}

func (command setAllDetailFoldsCmd) execute(program *Program, gui *gocui.Gui) {
	executeSetAllDetailFoldsCommand(newDetailFoldCommandRuntime(program), gui, command)
}

func executeSetAllDetailFoldsCommand(runtime detailFoldCommandRuntime, gui *gocui.Gui, command setAllDetailFoldsCmd) {
	if runtime.currentDetailDocument == nil || runtime.setAllDetailFolds == nil || runtime.syncDetailViewState == nil {
		return
	}

	actualView := command.View
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, command.View, viewDetailName)
	}
	detailDocument := runtime.currentDetailDocument(actualView)
	plan, ok := runtime.setAllDetailFolds(detailDocument, command.Collapsed)
	if !ok {
		return
	}
	applyDetailViewSyncPlan(runtime, plan, viewPageSize(actualView))
}

func applyDetailViewSyncPlan(runtime detailFoldCommandRuntime, plan detailViewSyncPlan, viewportHeight int) {
	if plan.focusLineKnown && runtime.placeDetailCursorAtLine != nil {
		runtime.placeDetailCursorAtLine(plan.document, plan.focusLine)
	}
	runtime.syncDetailViewState(plan.document, viewportHeight)
}
