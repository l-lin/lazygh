package tui

func (program *Program) applyMoveReviewFile(message MsgMoveReviewFile) {
	if !program.reviewModeActive() {
		return
	}

	selectableRows, ok := program.reviewSessionFileRows()
	if !ok || len(selectableRows) == 0 {
		return
	}

	originalRow := program.navigationState.reviewSession.selectedFileTreeRow
	program.navigationState.reviewSession = program.navigationState.reviewSession.adjustedSelection(selectableRows, message.Delta)
	if program.navigationState.reviewSession.selectedFileTreeRow == originalRow {
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

	program.detailState.viewState.clearPendingPrefix()
	program.navigationState.reviewSession = program.navigationState.reviewSession.withSelectedFileTreeRow(target.fileTreeRow)
	_ = program.mutateDetailViewStateWithoutRefresh(program.gui, detailView, func(document detailDocument, viewportHeight int) {
		program.focusDetailLine(document, viewportHeight, target.renderedLine)
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
	selectedRowIndex := clampIndex(program.navigationState.reviewSession.selectedFileTreeRow, len(visibleTree.Rows))
	selectedRow := visibleTree.Rows[selectedRowIndex]

	rawTree, _, rawTreeOK := program.reviewSessionRawTree()
	if !rawTreeOK {
		return
	}
	targetRow, ok := reviewDiffTreeNearestFoldableRow(rawTree, selectedRow.ID)
	if !ok {
		return
	}
	program.navigationState.reviewSession = program.navigationState.reviewSession.withTreeRowCollapsed(targetRow.ID, !reviewDiffTreeRowCollapsed(targetRow, program.navigationState.reviewSession.collapsedTreeRowIDs))
	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.navigationState.reviewSession.collapsedTreeRowIDs)
	program.navigationState.reviewSession = program.navigationState.reviewSession.withSelectedFileTreeRow(reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, targetRow.ID))
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
	selectedRowID := currentVisibleTree.Rows[clampIndex(program.navigationState.reviewSession.selectedFileTreeRow, len(currentVisibleTree.Rows))].ID
	updatedReviewSession, changed := program.navigationState.reviewSession.withAllTreeRowsCollapsed(rawTree, message.Collapsed)
	if !changed {
		return
	}
	program.navigationState.reviewSession = updatedReviewSession

	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.navigationState.reviewSession.collapsedTreeRowIDs)
	program.navigationState.reviewSession = program.navigationState.reviewSession.withSelectedFileTreeRow(reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, selectedRowID))
}

func (program *Program) applyToggleInlineConversationVisibility(message MsgToggleInlineConversationVisibility) {
	_ = program.toggleInlineConversationVisibilityState(message.View)
}

func (program *Program) applySetAllDetailFolds(message MsgSetAllDetailFolds) {
	_ = program.setAllDetailFolds(nil, message.View, message.Collapsed)
}
