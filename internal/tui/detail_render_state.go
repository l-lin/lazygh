package tui

import "github.com/jesseduffield/gocui"

func (program *Program) prepareViewRenderState(viewName string, view *gocui.View) {
	if view == nil {
		return
	}
	if viewName == viewDetailName {
		program.syncDetailViewRenderState(view)
	}
}

func (program *Program) syncDetailViewRenderState(view *gocui.View) {
	if program == nil || view == nil {
		return
	}

	program.detailState.wrapWidth = effectiveMarkdownWidth(view.InnerWidth())
	detailDocument := program.currentDetailDocument(view)
	program.syncDetailViewState(detailDocument, view.InnerHeight())
}
