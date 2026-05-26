package tui

import "github.com/jesseduffield/gocui"

type navigationCommandRuntime struct {
	dispatch                   func(*gocui.Gui, Msg) error
	resolveView                func(*gocui.Gui, *gocui.View, string) *gocui.View
	configureGUI               func(*gocui.Gui)
	focusDetailRenderedLine    func(*gocui.Gui, *gocui.View, int)
	handleDetailLineNavigation func(*gocui.Gui, *gocui.View, int)
	handlePageNavigation       func(*gocui.Gui, *gocui.View, pageNavigationKind)
	handleSideListViewport     func(*gocui.Gui, *gocui.View, viewportPlacement)
	handleActionsPopupViewport func(*gocui.Gui, *gocui.View, viewportPlacement)
	handleDetailViewport       func(*gocui.Gui, *gocui.View, detailViewportOperation)
}

func newNavigationCommandRuntime(program *Program) navigationCommandRuntime {
	if program == nil {
		return navigationCommandRuntime{}
	}
	return navigationCommandRuntime{
		dispatch:     program.dispatch,
		resolveView:  program.resolveView,
		configureGUI: program.configureGUI,
		focusDetailRenderedLine: func(gui *gocui.Gui, view *gocui.View, line int) {
			_ = program.mutateDetailViewStateWithoutRefresh(gui, view, func(document detailDocument, viewportHeight int) {
				program.focusDetailLine(document, viewportHeight, line)
			})
		},
		handleDetailLineNavigation: func(gui *gocui.Gui, view *gocui.View, delta int) {
			actualView := program.resolveView(gui, view, viewDetailName)
			steps := delta
			if steps < 0 {
				steps = -steps
			}
			_ = program.mutateDetailViewStateForYankMotion(gui, actualView, detailYankMotionLinewise, func(document detailDocument, viewportHeight int) {
				for range steps {
					if delta > 0 {
						program.detailState.viewState.moveDown(document, viewportHeight)
						continue
					}
					program.detailState.viewState.moveUp(document, viewportHeight)
				}
			})
		},
		handlePageNavigation: func(gui *gocui.Gui, view *gocui.View, kind pageNavigationKind) {
			actualView := program.resolveView(gui, view, program.currentViewName())
			pageSize := viewPageSize(actualView)
			delta := pageNavigationDelta(kind, pageSize)
			if program.model.Focus() == FocusDetailView {
				_ = program.mutateDetailViewStateForYankMotion(gui, actualView, detailYankMotionLinewise, func(document detailDocument, viewportHeight int) {
					switch kind {
					case pageNavigationKindHalfDown:
						program.detailState.viewState.pageDown(document, viewportHeight)
					case pageNavigationKindHalfUp:
						program.detailState.viewState.pageUp(document, viewportHeight)
					case pageNavigationKindFullDown:
						program.detailState.viewState.fullPageDown(document, viewportHeight)
					case pageNavigationKindFullUp:
						program.detailState.viewState.fullPageUp(document, viewportHeight)
					}
				})
				return
			}
			if program.actionContext().IsReviewContext() {
				if program.model.Focus() != FocusPullRequestsView {
					return
				}
				program.applyMoveReviewSelection(MsgMoveReviewSelection{Delta: delta})
			} else {
				program.applyMoveSideSelection(MsgMoveSideSelection{Delta: delta})
			}
			viewName, selectedVisibleLine, lineCount := program.currentSideListState()
			_ = program.recenterListSelection(gui, actualView, viewName, selectedVisibleLine, lineCount)
		},
		handleSideListViewport: func(gui *gocui.Gui, view *gocui.View, placement viewportPlacement) {
			viewName, selectedVisibleLine, lineCount := program.currentSideListState()
			if placement == viewportPlacementCenter {
				_ = program.recenterListSelection(gui, view, viewName, selectedVisibleLine, lineCount)
				return
			}
			_ = program.placeListSelection(gui, view, viewName, selectedVisibleLine, lineCount, placement)
		},
		handleActionsPopupViewport: func(gui *gocui.Gui, view *gocui.View, placement viewportPlacement) {
			selectedLine, lineCount := program.actionsPopupSelectionLineState()
			if placement == viewportPlacementCenter {
				_ = program.recenterListSelection(gui, view, viewActionsPopupName, selectedLine, lineCount)
				return
			}
			_ = program.placeListSelection(gui, view, viewActionsPopupName, selectedLine, lineCount, placement)
		},
		handleDetailViewport: func(gui *gocui.Gui, view *gocui.View, operation detailViewportOperation) {
			switch operation {
			case detailViewportOperationScrollDown:
				_ = program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.scrollDown(document, viewportHeight)
				})
			case detailViewportOperationScrollUp:
				_ = program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.scrollUp(document, viewportHeight)
				})
			case detailViewportOperationRecenter:
				_ = program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.recenter(document, viewportHeight)
					program.detailState.viewState.preserveViewportSyncCount++
				})
			case detailViewportOperationPlaceTop:
				_ = program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.placeCursorAtViewportTop(document, viewportHeight)
					program.detailState.viewState.preserveViewportSyncCount++
				})
			case detailViewportOperationPlaceBottom:
				_ = program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
					program.detailState.viewState.placeCursorAtViewportBottom(document, viewportHeight)
					program.detailState.viewState.preserveViewportSyncCount++
				})
			}
		},
	}
}

type configureGUICmd struct{}

func (configureGUICmd) execute(program *Program, gui *gocui.Gui) {
	executeConfigureGUICommand(newNavigationCommandRuntime(program), gui)
}

func executeConfigureGUICommand(runtime navigationCommandRuntime, gui *gocui.Gui) {
	if gui == nil || runtime.configureGUI == nil {
		return
	}
	runtime.configureGUI(gui)
}

type focusReviewCommentCmd struct {
	RenderedLine int
}

func (command focusReviewCommentCmd) execute(program *Program, gui *gocui.Gui) {
	executeFocusReviewCommentCommand(newNavigationCommandRuntime(program), gui, command)
}

func executeFocusReviewCommentCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command focusReviewCommentCmd) {
	if runtime.resolveView == nil || runtime.focusDetailRenderedLine == nil {
		return
	}
	actualView := runtime.resolveView(gui, nil, viewDetailName)
	runtime.focusDetailRenderedLine(gui, actualView, command.RenderedLine)
}

type detailLineNavigationCmd struct {
	Delta int
}

func (command detailLineNavigationCmd) execute(program *Program, gui *gocui.Gui) {
	executeDetailLineNavigationCommand(newNavigationCommandRuntime(program), gui, command)
}

func executeDetailLineNavigationCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command detailLineNavigationCmd) {
	if runtime.handleDetailLineNavigation == nil {
		return
	}
	runtime.handleDetailLineNavigation(gui, nil, command.Delta)
}

type pageNavigationCmd struct {
	Kind pageNavigationKind
}

func (command pageNavigationCmd) execute(program *Program, gui *gocui.Gui) {
	executePageNavigationCommand(newNavigationCommandRuntime(program), gui, command)
}

func executePageNavigationCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command pageNavigationCmd) {
	if runtime.handlePageNavigation == nil {
		return
	}
	runtime.handlePageNavigation(gui, nil, command.Kind)
}

type readOnlyScrollCmd struct {
	FallbackName string
	Kind         pageNavigationKind
}

func (command readOnlyScrollCmd) execute(program *Program, gui *gocui.Gui) {
	executeReadOnlyScrollCommand(newNavigationCommandRuntime(program), gui, command)
}

func executeReadOnlyScrollCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command readOnlyScrollCmd) {
	if runtime.resolveView == nil {
		return
	}

	actualView := runtime.resolveView(gui, nil, command.FallbackName)
	if actualView == nil {
		return
	}

	pageSize := viewPageSize(actualView)
	delta := pageNavigationDelta(command.Kind, pageSize)
	originX, originY := actualView.Origin()
	maxOriginY := maxInt(0, len(actualView.BufferLines())-pageSize)
	actualView.SetOrigin(originX, clampInt(originY+delta, 0, maxOriginY))
}

type sideListViewportCmd struct {
	Placement viewportPlacement
}

func (command sideListViewportCmd) execute(program *Program, gui *gocui.Gui) {
	executeSideListViewportCommand(newNavigationCommandRuntime(program), gui, command)
}

func executeSideListViewportCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command sideListViewportCmd) {
	if runtime.handleSideListViewport == nil {
		return
	}
	runtime.handleSideListViewport(gui, nil, command.Placement)
}

type resolveActionsPopupPageSizeCmd struct {
	Kind pageNavigationKind
}

func (command resolveActionsPopupPageSizeCmd) execute(program *Program, gui *gocui.Gui) {
	executeResolveActionsPopupPageSizeCommand(newNavigationCommandRuntime(program), gui, command)
}

func executeResolveActionsPopupPageSizeCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command resolveActionsPopupPageSizeCmd) {
	if runtime.dispatch == nil || runtime.resolveView == nil {
		return
	}
	actualView := runtime.resolveView(gui, nil, viewActionsPopupName)
	_ = runtime.dispatch(gui, MsgActionsPopupPageResolved{Kind: command.Kind, PageSize: viewPageSize(actualView)})
}

type actionsPopupViewportCmd struct {
	Placement viewportPlacement
}

func (command actionsPopupViewportCmd) execute(program *Program, gui *gocui.Gui) {
	executeActionsPopupViewportCommand(newNavigationCommandRuntime(program), gui, command)
}

func executeActionsPopupViewportCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command actionsPopupViewportCmd) {
	if runtime.handleActionsPopupViewport == nil {
		return
	}
	runtime.handleActionsPopupViewport(gui, nil, command.Placement)
}

type detailViewportCmd struct {
	Operation detailViewportOperation
}

func (command detailViewportCmd) execute(program *Program, gui *gocui.Gui) {
	executeDetailViewportCommand(newNavigationCommandRuntime(program), gui, command)
}

func executeDetailViewportCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command detailViewportCmd) {
	if runtime.handleDetailViewport == nil {
		return
	}
	runtime.handleDetailViewport(gui, nil, command.Operation)
}

func pageNavigationDelta(kind pageNavigationKind, pageSize int) int {
	switch kind {
	case pageNavigationKindHalfDown:
		return pageDelta(pageSize)
	case pageNavigationKindHalfUp:
		return -pageDelta(pageSize)
	case pageNavigationKindFullDown:
		return fullPageDelta(pageSize)
	case pageNavigationKindFullUp:
		return -fullPageDelta(pageSize)
	default:
		return 0
	}
}
