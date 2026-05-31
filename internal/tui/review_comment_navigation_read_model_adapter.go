package tui

func (program *Program) currentReviewCommentNavigationReadModel() reviewCommentNavigationReadModel {
	if program == nil {
		return reviewCommentNavigationReadModel{}
	}

	reviewModel := program.reviewSessionReadModel()
	model := reviewCommentNavigationReadModel{active: reviewModel.isActive(), selectedFileTreeRow: reviewModel.selectedFileTreeRow}
	if !reviewModel.isActive() {
		return model
	}

	selection := currentDetailCursorSelectionFor(program.currentDetailDocument(nil), program.detailState.viewState)
	model.selection = selection
	tree, files, ok := reviewModel.currentTree()
	if !ok {
		return model
	}

	width := maxInt(selection.document.width, 1)
	model.files = make([]reviewCommentNavigationFile, 0, len(tree.Rows))
	for _, row := range tree.Rows {
		if row.FileIndex < 0 || row.FileIndex >= len(files) {
			continue
		}
		model.files = append(model.files, reviewCommentNavigationFile{
			fileTreeRow:  row.VisibleRowIndex,
			renderedRows: copyReviewDiffRenderedRows(program.currentReviewDiffRenderedRows(files[row.FileIndex], width)),
		})
	}
	return model
}
