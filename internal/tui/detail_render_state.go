package tui

import "github.com/jesseduffield/gocui"

func (program *Program) prepareViewRenderState(viewName string, view *gocui.View) {
	if view == nil {
		return
	}
	if viewName == viewDetailName {
		program.syncDetailViewRenderState(view)
		return
	}
	if viewName == viewPullRequestBuildInfoName {
		program.syncPullRequestBuildRunPopupRenderState(view)
	}
}
