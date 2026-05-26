package tui

import "github.com/jesseduffield/gocui"

type detailSearchCommandRuntime struct {
	dispatch              func(*gocui.Gui, Msg) error
	resolveView           func(*gocui.Gui, *gocui.View, string) *gocui.View
	currentDetailDocument func(*gocui.View) detailDocument
	syncDetailViewState   func(detailDocument, int)
	currentDetailCursor   func() detailPosition
	repeatDetailSearch    func(*gocui.Gui, *gocui.View, bool)
	followDetailSearch    func(*gocui.Gui, bool)
}

func newDetailSearchCommandRuntime(program *Program) detailSearchCommandRuntime {
	if program == nil {
		return detailSearchCommandRuntime{}
	}
	return detailSearchCommandRuntime{
		dispatch:              program.dispatch,
		resolveView:           program.resolveView,
		currentDetailDocument: program.currentDetailDocument,
		syncDetailViewState:   program.syncDetailViewState,
		currentDetailCursor: func() detailPosition {
			return program.detailState.viewState.cursor
		},
		repeatDetailSearch: func(gui *gocui.Gui, view *gocui.View, reverse bool) {
			_ = program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
				if reverse {
					program.detailState.viewState.followPreviousSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
					return
				}
				program.detailState.viewState.followNextSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
			})
		},
		followDetailSearch: func(gui *gocui.Gui, reverse bool) {
			_ = program.mutateDetailViewStateWithoutRefresh(gui, nil, func(document detailDocument, viewportHeight int) {
				if reverse {
					program.detailState.viewState.followPreviousSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
					return
				}
				program.detailState.viewState.followSubmittedSearch(document, program.model.DetailSearchQuery(), viewportHeight)
			})
		},
	}
}

type resolveDetailSearchWordCmd struct {
	Reverse bool
}

func (command resolveDetailSearchWordCmd) execute(program *Program, gui *gocui.Gui) {
	executeResolveDetailSearchWordCommand(newDetailSearchCommandRuntime(program), gui, command)
}

func executeResolveDetailSearchWordCommand(runtime detailSearchCommandRuntime, gui *gocui.Gui, command resolveDetailSearchWordCmd) {
	if runtime.dispatch == nil || runtime.resolveView == nil || runtime.currentDetailDocument == nil || runtime.syncDetailViewState == nil || runtime.currentDetailCursor == nil {
		return
	}

	actualView := runtime.resolveView(gui, nil, viewDetailName)
	document := runtime.currentDetailDocument(actualView)
	runtime.syncDetailViewState(document, viewPageSize(actualView))
	query, ok := document.wordAt(runtime.currentDetailCursor())
	if !ok {
		return
	}
	_ = runtime.dispatch(gui, MsgDetailSearchWordResolved{Query: query, Reverse: command.Reverse})
}

type followDetailSearchCmd struct {
	Reverse bool
}

func (command followDetailSearchCmd) execute(program *Program, gui *gocui.Gui) {
	executeFollowDetailSearchCommand(newDetailSearchCommandRuntime(program), gui, command)
}

func executeFollowDetailSearchCommand(runtime detailSearchCommandRuntime, gui *gocui.Gui, command followDetailSearchCmd) {
	if runtime.followDetailSearch == nil {
		return
	}
	runtime.followDetailSearch(gui, command.Reverse)
}

type repeatDetailSearchCmd struct {
	Reverse bool
}

func (command repeatDetailSearchCmd) execute(program *Program, gui *gocui.Gui) {
	executeRepeatDetailSearchCommand(newDetailSearchCommandRuntime(program), gui, command)
}

func executeRepeatDetailSearchCommand(runtime detailSearchCommandRuntime, gui *gocui.Gui, command repeatDetailSearchCmd) {
	if runtime.repeatDetailSearch == nil {
		return
	}
	runtime.repeatDetailSearch(gui, nil, command.Reverse)
}
