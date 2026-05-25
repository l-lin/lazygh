package tui

import "github.com/jesseduffield/gocui"

func (program *Program) setDetailWrapWidth(width int) {
	if program == nil {
		return
	}
	program.detailState.wrapWidth = width
}

func (program *Program) resetDetailViewState() {
	if program == nil {
		return
	}
	program.detailState.viewState.reset()
}

func (program *Program) syncCurrentDetailViewport(detailDocument detailDocument, viewportHeight int) {
	if program == nil {
		return
	}
	program.detailState.viewState.sync(detailDocument, viewportHeight)
}

func (program *Program) placeDetailCursorAtLine(detailDocument detailDocument, line int) {
	if program == nil {
		return
	}
	program.detailState.viewState.cursor = detailDocument.clampPosition(detailPosition{line: line, column: 0})
	program.detailState.viewState.preferredColumn = 0
}

func (program *Program) focusDetailLine(detailDocument detailDocument, viewportHeight int, line int) {
	if program == nil {
		return
	}
	program.placeDetailCursorAtLine(detailDocument, line)
	program.syncCurrentDetailViewport(detailDocument, viewportHeight)
}

func (program *Program) syncDetailViewState(detailDocument detailDocument, viewportHeight int) {
	identity := program.currentDetailIdentity()
	if identity != program.detailState.lastIdentity {
		program.detailState.lastIdentity = identity
		program.resetDetailViewState()
	}

	program.syncCurrentDetailViewport(detailDocument, viewportHeight)
	program.detailState.viewState.syncSearch(detailDocument, program.model.DetailSearchQuery())
}

func (program *Program) syncDetailViewRenderState(view *gocui.View) {
	if program == nil || view == nil {
		return
	}

	program.setDetailWrapWidth(effectiveMarkdownWidth(view.InnerWidth()))
	detailDocument := program.currentDetailDocument(view)
	program.syncDetailViewState(detailDocument, view.InnerHeight())
}
