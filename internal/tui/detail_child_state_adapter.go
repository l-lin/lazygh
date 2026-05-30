package tui

import "github.com/jesseduffield/gocui"

func (program *Program) updateDetailState(transition func(detailStateModel) detailStateModel) {
	if program == nil {
		return
	}
	program.detailState = transition(program.detailState)
}

func (program *Program) setDetailWrapWidth(width int) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withWrapWidth(width)
	})
}

func (program *Program) toggleDetailWordWrapState() {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withWordWrapToggled()
	})
}

func (program *Program) setDetailActiveTab(tab DetailTab) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withActiveTab(tab)
	})
}

func (program *Program) applyProjectedDetailStateApplication(application projectedScreenStateApplication) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withProjectedScreenStateApplication(application)
	})
}

func (program *Program) advanceDetailActiveTab(delta int, count int) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withAdvancedActiveTab(delta, count)
	})
}

func (program *Program) clearDetailPendingPrefix() {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withPendingPrefixCleared()
	})
}

func (program *Program) exitDetailVisualMode() {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withVisualModeExited()
	})
}

func (program *Program) resetDetailViewState() {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withResetViewState()
	})
}

func (program *Program) syncCurrentDetailViewport(detailDocument detailDocument, viewportHeight int) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withSyncedViewport(detailDocument, viewportHeight)
	})
}

func (program *Program) placeDetailCursorAtLine(detailDocument detailDocument, line int) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withCursorAtLine(detailDocument, line)
	})
}

func (program *Program) focusDetailLine(detailDocument detailDocument, viewportHeight int, line int) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withFocusedLine(detailDocument, viewportHeight, line)
	})
}

func (program *Program) syncDetailViewState(detailDocument detailDocument, viewportHeight int) {
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.synced(program.currentDetailIdentity(), detailDocument, viewportHeight, program.model.DetailSearchQuery())
	})
}

func (program *Program) prepareSelectedDetailClipboard(detailDocument detailDocument, viewportHeight int) detailClipboardResult {
	prepared := detailClipboardResult{}
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		prepared = state.preparedClipboard(program.currentDetailIdentity(), detailDocument, viewportHeight, program.model.DetailSearchQuery())
		return prepared.state
	})
	return prepared
}

func (program *Program) syncDetailViewShellState(view *gocui.View) {
	if program == nil || view == nil {
		return
	}

	detailDocument := program.currentDetailDocument(view)
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.syncedForRender(
			program.currentDetailIdentity(),
			detailDocument,
			effectiveMarkdownWidth(view.InnerWidth()),
			view.InnerHeight(),
			program.model.DetailSearchQuery(),
		)
	})
}
