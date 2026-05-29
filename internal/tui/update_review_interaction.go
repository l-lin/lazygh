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
	program.adjustReviewSessionSelection(message.Delta)
	if program.navigationState.reviewSession.selectedFileTreeRow == originalRow {
		return
	}
}

func (program *Program) applyMoveReviewComment(message MsgMoveReviewComment) []Cmd {
	if !program.reviewModeActive() {
		return nil
	}

	currentFileTreeRow, currentRenderedLine := program.currentReviewCommentPosition()
	target, ok := program.reviewSessionCommentTarget(currentFileTreeRow, currentRenderedLine, message.Direction)
	if !ok {
		return nil
	}

	program.clearDetailPendingPrefix()
	program.setReviewSessionSelectedFileTreeRow(target.fileTreeRow)
	return []Cmd{focusReviewCommentCmd{RenderedLine: target.renderedLine}}
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
	program.setReviewSessionTreeRowCollapsed(targetRow.ID, !reviewDiffTreeRowCollapsed(targetRow, program.navigationState.reviewSession.collapsedTreeRowIDs))
	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.navigationState.reviewSession.collapsedTreeRowIDs)
	program.setReviewSessionSelectedFileTreeRow(reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, targetRow.ID))
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
	if !program.setAllReviewSessionTreeRowsCollapsed(rawTree, message.Collapsed) {
		return
	}

	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.navigationState.reviewSession.collapsedTreeRowIDs)
	program.setReviewSessionSelectedFileTreeRow(reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, selectedRowID))
}

func (program *Program) applyToggleInlineConversationVisibility(message MsgToggleInlineConversationVisibility) []Cmd {
	return []Cmd{toggleInlineConversationVisibilityCmd{}}
}

func (program *Program) applySetAllDetailFolds(message MsgSetAllDetailFolds) []Cmd {
	return []Cmd{setAllDetailFoldsCmd{Collapsed: message.Collapsed}}
}
