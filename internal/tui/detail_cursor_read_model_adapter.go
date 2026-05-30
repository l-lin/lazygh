package tui

func currentDetailCursorSelectionFor(document detailDocument, state detailViewState) detailCursorSelection {
	state.sync(document, 1)
	return detailCursorSelection{document: document, state: state}
}

func (program *Program) currentDetailCursorReadModel() detailCursorReadModel {
	if program == nil {
		return detailCursorReadModel{}
	}

	selection := currentDetailCursorSelectionFor(program.currentDetailDocument(nil), program.detailState.viewState)
	model := detailCursorReadModel{selection: selection}
	if summary, detail, ok := program.currentPullRequestDescriptionSummaryAndDetail(); ok {
		model.descriptionSummary = summary
		model.descriptionDetail = detail
		model.descriptionKnown = true
	}
	if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
		if result, ok := program.pullRequestDiffForSummary(summary); ok && result.err == nil {
			model.browserChangesSummary = summary
			model.browserChangesRenderedRows = program.currentPullRequestChangesRenderedRows(summary, result.data.Files, maxInt(selection.document.width, 1))
			model.browserChangesKnown = true
		}
	}

	reviewModel := program.reviewSessionReadModel()
	if reviewModel.isActive() && !reviewModel.showsDescription() && !reviewModel.showsStoryChapter() {
		if selectedFile, ok := reviewModel.selectedDiffFile(); ok {
			model.reviewDiffSummary = reviewModel.summary
			model.reviewDiffRenderedRows = program.currentReviewDiffRenderedRows(selectedFile, maxInt(selection.document.width, 1))
			model.reviewDiffKnown = true
		}
	}
	return model
}

func (program *Program) currentDetailCursorSelection() detailCursorSelection {
	return program.currentDetailCursorReadModel().currentSelection()
}

func (program *Program) currentPullRequestDescriptionCursorContext() (pullRequestDescriptionCursorContext, bool) {
	return program.currentDetailCursorReadModel().pullRequestDescriptionContext()
}

func (program *Program) currentBrowserChangesCursorContext() (browserChangesCursorContext, bool) {
	return program.currentDetailCursorReadModel().browserChangesContext()
}

func (program *Program) currentReviewDiffCursorContext() (reviewDiffCursorContext, bool) {
	return program.currentDetailCursorReadModel().reviewDiffContext()
}
