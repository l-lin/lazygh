package tui

import "github.com/jesseduffield/gocui"

func (program *Program) syncViewShellState(viewName string, view *gocui.View) {
	if view == nil {
		return
	}
	if viewName == viewDetailName {
		program.syncDetailViewShellState(view)
		return
	}
	if viewName == viewPullRequestBuildInfoName {
		program.syncPullRequestBuildRunPopupShellState(view)
	}
}
