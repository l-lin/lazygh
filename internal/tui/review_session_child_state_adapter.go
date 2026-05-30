package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) updateReviewSession(transition func(reviewSessionState) reviewSessionState) {
	program.updateNavigationState(func(state navigationStateModel) navigationStateModel {
		return state.withReviewSession(transition(state.reviewSession))
	})
}

func (program *Program) startReviewSessionState(start reviewSessionStartDescriptor) {
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.started(start)
	})
}

func (program *Program) clearReviewSession() {
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.cleared()
	})
}

func (program *Program) setReviewSessionSummary(summary githubdomain.PullRequest) {
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.withSummary(summary)
	})
}

func (program *Program) setReviewSessionSelectedFileTreeRow(row int) {
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.withSelectedFileTreeRow(row)
	})
}

func (program *Program) clampReviewSessionSelection() {
	if program == nil {
		return
	}

	readModel := program.reviewSessionReadModel()
	selectableRows, ok := readModel.selectableRows()
	if !ok {
		program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
			return state.clampedSelection(nil, nil)
		})
		return
	}

	fileRows, fileRowsOK := readModel.fileRows()
	if !fileRowsOK {
		fileRows = nil
	}
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.clampedSelection(selectableRows, fileRows)
	})
}

func (program *Program) adjustReviewSessionSelection(change int) {
	if program == nil {
		return
	}

	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
			return state.adjustedSelection(nil, change)
		})
		return
	}

	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.adjustedSelection(selectableRows, change)
	})
}

func (program *Program) adjustReviewSessionFileSelection(change int) {
	if program == nil {
		return
	}

	fileRows, ok := program.reviewSessionReadModel().fileRows()
	if !ok {
		program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
			return state.adjustedSelection(nil, change)
		})
		return
	}

	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.adjustedSelection(fileRows, change)
	})
}

func (program *Program) moveReviewSessionSelectionToTop() {
	if program == nil {
		return
	}

	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
			return state.selectionAtTop(nil)
		})
		return
	}

	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.selectionAtTop(selectableRows)
	})
}

func (program *Program) moveReviewSessionSelectionToBottom() {
	if program == nil {
		return
	}

	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
			return state.selectionAtBottom(nil)
		})
		return
	}

	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.selectionAtBottom(selectableRows)
	})
}

func (program *Program) setReviewSessionTreeRowCollapsed(rowID string, collapsed bool) {
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.withTreeRowCollapsed(rowID, collapsed)
	})
}

func (program *Program) setAllReviewSessionTreeRowsCollapsed(tree reviewDiffTree, collapsed bool) bool {
	if program == nil {
		return false
	}

	changed := false
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		updatedReviewSession, actualChanged := state.withAllTreeRowsCollapsed(tree, collapsed)
		changed = actualChanged
		if !actualChanged {
			return state
		}
		return updatedReviewSession
	})
	return changed
}

func (program *Program) setReviewSessionThreadCollapsed(threadID string, collapsed bool) {
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		return state.withThreadCollapsed(threadID, collapsed)
	})
}

func (program *Program) setAllReviewSessionThreadsCollapsed(threads []reviewDiffThread, collapsed bool) bool {
	if program == nil {
		return false
	}

	changed := false
	program.updateReviewSession(func(state reviewSessionState) reviewSessionState {
		updatedReviewSession, actualChanged := state.withAllThreadsCollapsed(threads, collapsed)
		changed = actualChanged
		if !actualChanged {
			return state
		}
		return updatedReviewSession
	})
	return changed
}
