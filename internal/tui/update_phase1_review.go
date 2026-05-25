package tui

func (program *Program) applyMoveReviewFile(message MsgMoveReviewFile) {
	if !program.reviewModeActive() {
		return
	}

	selectableRows, ok := program.reviewSessionFileRows()
	if !ok || len(selectableRows) == 0 {
		return
	}

	originalRow := program.reviewSession.selectedFileTreeRow
	program.reviewSession.selectedFileTreeRow = adjustVisibleSelection(program.reviewSession.selectedFileTreeRow, selectableRows, message.Delta)
	if program.reviewSession.selectedFileTreeRow == originalRow {
		return
	}
}

func (program *Program) applyMoveReviewComment(message MsgMoveReviewComment) {
	if !program.reviewModeActive() {
		return
	}

	detailView := program.resolveView(program.gui, nil, viewDetailName)
	currentFileTreeRow, currentRenderedLine := program.currentReviewCommentPosition(detailView)
	target, ok := program.reviewSessionCommentTarget(detailView, currentFileTreeRow, currentRenderedLine, message.Direction)
	if !ok {
		return
	}

	program.detailViewState.clearPendingPrefix()
	program.reviewSession.selectedFileTreeRow = target.fileTreeRow
	_ = program.mutateDetailViewStateWithoutRefresh(program.gui, detailView, func(document detailDocument, viewportHeight int) {
		program.detailViewState.cursor = document.clampPosition(detailPosition{line: target.renderedLine, column: 0})
		program.detailViewState.preferredColumn = 0
		program.detailViewState.sync(document, viewportHeight)
	})
}

func (program *Program) applyToggleReviewTreeRowVisibility() {
	if !program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView || program.reviewTreeFoldBlocked() {
		return
	}

	visibleTree, _, ok := program.reviewSessionCurrentTree()
	if !ok || len(visibleTree.Rows) == 0 {
		return
	}
	selectedRowIndex := clampIndex(program.reviewSession.selectedFileTreeRow, len(visibleTree.Rows))
	selectedRow := visibleTree.Rows[selectedRowIndex]

	rawTree, _, rawTreeOK := program.reviewSessionRawTree()
	if !rawTreeOK {
		return
	}
	targetRow, ok := reviewDiffTreeNearestFoldableRow(rawTree, selectedRow.ID)
	if !ok {
		return
	}
	program.setReviewTreeRowCollapsed(targetRow.ID, !reviewDiffTreeRowCollapsed(targetRow, program.reviewSession.collapsedTreeRowIDs))
	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.reviewSession.collapsedTreeRowIDs)
	program.reviewSession.selectedFileTreeRow = reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, targetRow.ID)
}

func (program *Program) applySetAllReviewTreeFolds(message MsgSetAllReviewTreeFolds) {
	if !program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView || program.reviewTreeFoldBlocked() {
		program.clearPendingSelectionPrefix()
		return
	}

	rawTree, _, ok := program.reviewSessionRawTree()
	if !ok || len(rawTree.Rows) == 0 {
		return
	}
	currentVisibleTree, _, visibleTreeOK := program.reviewSessionCurrentTree()
	if !visibleTreeOK || len(currentVisibleTree.Rows) == 0 {
		return
	}
	selectedRowID := currentVisibleTree.Rows[clampIndex(program.reviewSession.selectedFileTreeRow, len(currentVisibleTree.Rows))].ID
	if !program.setAllReviewTreeRowsCollapsed(rawTree, message.Collapsed) {
		return
	}

	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.reviewSession.collapsedTreeRowIDs)
	program.reviewSession.selectedFileTreeRow = reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, selectedRowID)
}

func (program *Program) applyToggleInlineConversationVisibility(message MsgToggleInlineConversationVisibility) {
	_ = program.toggleInlineConversationVisibilityState(message.View)
}

func (program *Program) applySetAllDetailFolds(message MsgSetAllDetailFolds) {
	_ = program.setAllDetailFolds(nil, message.View, message.Collapsed)
}
