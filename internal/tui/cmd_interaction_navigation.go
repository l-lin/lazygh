package tui

import "github.com/jesseduffield/gocui"

type navigationCommandRuntime struct {
	executeMessage             func(*gocui.Gui, Msg) error
	resolveView                func(*gocui.Gui, *gocui.View, string) *gocui.View
	configureGUI               func(*gocui.Gui)
	currentViewName            func() string
	currentDetailDocument      func(*gocui.View) detailDocument
	handleSideListViewport     func(*gocui.Gui, *gocui.View, viewportPlacement)
	handleActionsPopupViewport func(*gocui.Gui, *gocui.View, viewportPlacement)
}

func newNavigationCommandRuntime(program *Program) navigationCommandRuntime {
	if program == nil {
		return navigationCommandRuntime{}
	}
	return navigationCommandRuntime{
		executeMessage:        program.executeRuntimeMessage,
		resolveView:           program.resolveView,
		configureGUI:          program.configureGUI,
		currentViewName:       program.currentViewName,
		currentDetailDocument: program.currentDetailDocument,
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
	if runtime.executeMessage == nil || runtime.currentDetailDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	_ = runtime.executeMessage(gui, MsgFocusDetailRenderedLineResolved{RenderedLine: command.RenderedLine, Document: runtime.currentDetailDocument(actualView), ViewportHeight: viewPageSize(actualView)})
}

type pageNavigationCmd struct {
	Kind pageNavigationKind
}

func (command pageNavigationCmd) execute(program *Program, gui *gocui.Gui) {
	executePageNavigationCommand(newNavigationCommandRuntime(program), gui, command)
}

func executePageNavigationCommand(runtime navigationCommandRuntime, gui *gocui.Gui, command pageNavigationCmd) {
	if runtime.executeMessage == nil {
		return
	}

	fallbackName := ""
	if runtime.currentViewName != nil {
		fallbackName = runtime.currentViewName()
	}
	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, fallbackName)
	}
	_ = runtime.executeMessage(gui, MsgPageNavigationResolved{Kind: command.Kind, PageSize: viewPageSize(actualView)})
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
	if runtime.executeMessage == nil || runtime.resolveView == nil {
		return
	}
	actualView := runtime.resolveView(gui, nil, viewActionsPopupName)
	_ = runtime.executeMessage(gui, MsgActionsPopupPageResolved{Kind: command.Kind, PageSize: viewPageSize(actualView)})
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
	if runtime.executeMessage == nil || runtime.currentDetailDocument == nil {
		return
	}

	actualView := (*gocui.View)(nil)
	if runtime.resolveView != nil {
		actualView = runtime.resolveView(gui, nil, viewDetailName)
	}
	_ = runtime.executeMessage(gui, MsgDetailViewportResolved{Operation: command.Operation, Document: runtime.currentDetailDocument(actualView), ViewportHeight: viewPageSize(actualView)})
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
