package tui

import "github.com/jesseduffield/gocui"

func (program *Program) setDetailWrapWidth(width int) {
	if program == nil {
		return
	}
	program.detailState = program.detailState.withWrapWidth(width)
}

func (program *Program) resetDetailViewState() {
	if program == nil {
		return
	}
	program.detailState = program.detailState.withResetViewState()
}

func (program *Program) syncCurrentDetailViewport(detailDocument detailDocument, viewportHeight int) {
	if program == nil {
		return
	}
	program.detailState = program.detailState.withSyncedViewport(detailDocument, viewportHeight)
}

func (program *Program) placeDetailCursorAtLine(detailDocument detailDocument, line int) {
	if program == nil {
		return
	}
	program.detailState = program.detailState.withCursorAtLine(detailDocument, line)
}

func (program *Program) focusDetailLine(detailDocument detailDocument, viewportHeight int, line int) {
	if program == nil {
		return
	}
	program.detailState = program.detailState.withFocusedLine(detailDocument, viewportHeight, line)
}

func (program *Program) syncDetailViewState(detailDocument detailDocument, viewportHeight int) {
	if program == nil {
		return
	}
	program.detailState = program.detailState.synced(program.currentDetailIdentity(), detailDocument, viewportHeight, program.model.DetailSearchQuery())
}

func (program *Program) syncDetailViewShellState(view *gocui.View) {
	if program == nil || view == nil {
		return
	}

	detailDocument := program.currentDetailDocument(view)
	program.detailState = program.detailState.syncedForRender(
		program.currentDetailIdentity(),
		detailDocument,
		effectiveMarkdownWidth(view.InnerWidth()),
		view.InnerHeight(),
		program.model.DetailSearchQuery(),
	)
}
