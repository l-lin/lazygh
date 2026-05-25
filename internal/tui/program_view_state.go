package tui

import "github.com/jesseduffield/gocui"

type viewConfigurator func(*gocui.View)
type viewRenderer func(*gocui.View)

func (program *Program) afterStateChange(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	program.gui = gui
	program.clearExpiredYankHighlights()
	program.clearExpiredTransientErrorPopup(program.currentTime())
	if program.model.ActionsPopupVisible() {
		program.syncActionsPopupSearch()
	}
	if program.startupState.appStarted {
		program.executeCmds(gui, program.plannedCommands(gui))
	}
	return program.refreshViews(gui)
}

func (program *Program) refreshViews(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	program.gui = gui
	maxX, maxY := gui.Size()
	if actualErr := program.applyScreenComposition(gui, program.screenCompositionForSize(maxX, maxY)); actualErr != nil {
		return actualErr
	}
	return program.syncShellState(gui)
}

func (program *Program) refreshExistingView(gui *gocui.Gui, viewName string, configure viewConfigurator, render viewRenderer) error {
	view, err := gui.View(viewName)
	if err != nil {
		if isUnknownViewError(err) {
			return nil
		}
		return err
	}

	configure(view)
	program.prepareViewRenderState(viewName, view)
	render(view)
	return nil
}

func (program *Program) refreshOverlayView(gui *gocui.Gui, visible bool, viewName string, configure viewConfigurator, render viewRenderer) error {
	if !visible {
		return deleteViewIfPresent(gui, viewName)
	}
	if err := program.refreshExistingView(gui, viewName, configure, render); err != nil {
		return err
	}

	_, err := gui.SetViewOnTop(viewName)
	if err != nil && !isUnknownViewError(err) {
		return err
	}

	return nil
}

func (program *Program) refreshActionsPopupViews(gui *gocui.Gui) error {
	if !program.model.ActionsPopupVisible() {
		return deleteViewsIfPresent(gui, viewActionsPopupSearchName, viewActionsPopupName, viewActionsPopupChromeName)
	}

	if err := program.layoutActionsPopupViews(gui); err != nil {
		return err
	}
	if program.model.ActionsPopupSearchActive() {
		return program.layoutActionsPopupSearchView(gui)
	}

	return deleteViewIfPresent(gui, viewActionsPopupSearchName)
}

func (program *Program) syncCurrentView(gui *gocui.Gui) error {
	gui.Cursor = program.shouldShowCursor()
	if program.overlayState.helpVisible {
		gui.Cursor = false
		return program.setCurrentViewIfPresent(gui, viewHelpName)
	}

	return program.setCurrentViewIfPresent(gui, program.currentViewName())
}

func (program *Program) shouldShowCursor() bool {
	switch {
	case program.modalEditorVisible():
		return true
	case program.model.ActionsPopupSearchActive():
		return true
	case program.searchPromptVisible():
		return true
	case program.model.ActionsPopupVisible():
		return false
	case program.pullRequestBuildRunPopupVisible():
		return true
	default:
		return program.screenState().AllowsMainCursor()
	}
}

func (program *Program) setCurrentViewIfPresent(gui *gocui.Gui, viewName string) error {
	_, err := gui.SetCurrentView(viewName)
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

func (program *Program) currentViewName() string {
	if program.modalEditorVisible() {
		return viewModalEditorName
	}
	if program.model.ActionsPopupVisible() {
		if program.model.ActionsPopupSearchActive() {
			return viewActionsPopupSearchName
		}
		return viewActionsPopupName
	}
	if program.searchPromptVisible() {
		return viewSearchName
	}
	if program.pullRequestBuildRunPopupVisible() {
		return viewPullRequestBuildInfoName
	}

	focus := program.screenState().ActiveView().Focus
	if !program.model.PaneVisible(focus) {
		return paneViewName(program.model.FullscreenPane())
	}
	return paneViewName(focus)
}
