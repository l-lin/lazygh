package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type detailCursorSelection struct {
	document detailDocument
	state    detailViewState
}

func (program *Program) currentDetailCursorSelection() detailCursorSelection {
	if program == nil {
		return detailCursorSelection{}
	}

	document := program.currentDetailDocument(nil)
	state := program.detailState.viewState
	state.sync(document, 1)
	return detailCursorSelection{
		document: document,
		state:    state,
	}
}

type pullRequestDescriptionCursorContext struct {
	summary   githubdomain.PullRequest
	detail    githubdomain.PullRequestDetail
	selection detailCursorSelection
}

func (program *Program) currentPullRequestDescriptionCursorContext() (pullRequestDescriptionCursorContext, bool) {
	summary, detail, ok := program.currentPullRequestDescriptionSummaryAndDetail()
	if !ok {
		return pullRequestDescriptionCursorContext{}, false
	}
	return pullRequestDescriptionCursorContext{
		summary:   summary,
		detail:    detail,
		selection: program.currentDetailCursorSelection(),
	}, true
}

type browserChangesCursorContext struct {
	summary      githubdomain.PullRequest
	renderedRows []reviewDiffRenderedRow
	selection    detailCursorSelection
}

func (program *Program) currentBrowserChangesCursorContext() (browserChangesCursorContext, bool) {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return browserChangesCursorContext{}, false
	}
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok || result.err != nil {
		return browserChangesCursorContext{}, false
	}

	selection := program.currentDetailCursorSelection()
	return browserChangesCursorContext{
		summary:      summary,
		renderedRows: program.currentPullRequestChangesRenderedRows(summary, result.data.Files, selection.document.width),
		selection:    selection,
	}, true
}

type reviewDiffCursorContext struct {
	summary      githubdomain.PullRequest
	renderedRows []reviewDiffRenderedRow
	selection    detailCursorSelection
}

func (program *Program) currentReviewDiffCursorContext() (reviewDiffCursorContext, bool) {
	if !program.reviewModeActive() {
		return reviewDiffCursorContext{}, false
	}

	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return reviewDiffCursorContext{}, false
	}

	selection := program.currentDetailCursorSelection()
	return reviewDiffCursorContext{
		summary:      program.navigationState.reviewSession.summary,
		renderedRows: program.currentReviewDiffRenderedRows(selectedFile, selection.document.width),
		selection:    selection,
	}, true
}
