package tui

func (program *Program) setReviewSessionSelectedFileTreeRow(row int) {
	if program == nil {
		return
	}
	program.navigationState.reviewSession = program.navigationState.reviewSession.withSelectedFileTreeRow(row)
}

func (program *Program) clampReviewSessionSelection() {
	if program == nil {
		return
	}

	readModel := program.reviewSessionReadModel()
	selectableRows, ok := readModel.selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.clampedSelection(nil, nil)
		return
	}

	fileRows, fileRowsOK := readModel.fileRows()
	if !fileRowsOK {
		fileRows = nil
	}
	program.navigationState.reviewSession = program.navigationState.reviewSession.clampedSelection(selectableRows, fileRows)
}

func (program *Program) adjustReviewSessionSelection(change int) {
	if program == nil {
		return
	}

	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.adjustedSelection(nil, change)
		return
	}

	program.navigationState.reviewSession = program.navigationState.reviewSession.adjustedSelection(selectableRows, change)
}

func (program *Program) moveReviewSessionSelectionToTop() {
	if program == nil {
		return
	}

	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtTop(nil)
		return
	}

	program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtTop(selectableRows)
}

func (program *Program) moveReviewSessionSelectionToBottom() {
	if program == nil {
		return
	}

	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtBottom(nil)
		return
	}

	program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtBottom(selectableRows)
}

func (program *Program) setReviewSessionThreadCollapsed(threadID string, collapsed bool) {
	if program == nil {
		return
	}
	program.navigationState.reviewSession = program.navigationState.reviewSession.withThreadCollapsed(threadID, collapsed)
}

func (program *Program) setAllReviewSessionThreadsCollapsed(threads []reviewDiffThread, collapsed bool) bool {
	if program == nil {
		return false
	}
	updatedReviewSession, changed := program.navigationState.reviewSession.withAllThreadsCollapsed(threads, collapsed)
	if !changed {
		return false
	}
	program.navigationState.reviewSession = updatedReviewSession
	return true
}
