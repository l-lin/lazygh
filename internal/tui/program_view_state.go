package tui

import "github.com/jesseduffield/gocui"

type viewConfigurator func(*gocui.View)
type viewRenderer func(*gocui.View)

func (program *Program) refreshViews(gui *gocui.Gui) error {
	program.gui = gui
	if !program.reviewSession.active {
		program.maybeLoadSelectedNotificationDetail(gui)
	}
	program.maybeLoadSelectedPullRequestDetail(gui)
	program.maybeLoadSelectedPullRequestDiff(gui)

	if err := program.refreshExistingView(gui, viewUserName, program.configureUserView, program.renderUserView); err != nil {
		return err
	}
	if err := program.refreshExistingView(gui, viewPullRequestsName, program.configurePullRequestsView, program.renderPullRequestsView); err != nil {
		return err
	}
	if err := program.refreshExistingView(gui, viewNotificationsName, program.configureNotificationsView, program.renderNotificationsView); err != nil {
		return err
	}
	if err := program.refreshExistingView(gui, viewDetailName, program.configureDetailView, program.renderDetailView); err != nil {
		return err
	}

	if err := program.layoutPaneFooterViews(gui); err != nil {
		return err
	}
	if err := program.layoutStatusLineView(gui); err != nil {
		return err
	}
	if err := program.refreshOverlayView(gui, program.helpVisible, viewHelpName, program.configureHelpView, program.renderHelpView); err != nil {
		return err
	}
	if err := program.refreshOverlayView(gui, program.searchPromptVisible(), viewSearchName, program.configureSearchView, program.renderSearchView); err != nil {
		return err
	}
	if err := program.refreshOverlayView(gui, program.modalEditorVisible(), viewModalEditorName, program.configureModalEditorView, program.renderModalEditorView); err != nil {
		return err
	}
	if err := program.refreshOverlayView(gui, program.pullRequestBuildRunPopupVisible(), viewPullRequestBuildInfoName, program.configurePullRequestBuildRunPopupView, program.renderPullRequestBuildRunPopupView); err != nil {
		return err
	}
	if err := program.refreshActionsPopupViews(gui); err != nil {
		return err
	}
	if err := program.layoutStatusLineKeyHintsView(gui); err != nil {
		return err
	}

	return program.syncCurrentView(gui)
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
		return deleteViewsIfPresent(gui, viewActionsPopupSearchName, viewActionsPopupName)
	}

	program.syncActionsPopupSearch()
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
	if program.helpVisible {
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
