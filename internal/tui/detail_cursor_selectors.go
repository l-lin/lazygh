package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type detailCursorSelection struct {
	document detailDocument
	state    detailViewState
}

type pullRequestDescriptionCursorContext struct {
	summary   githubdomain.PullRequest
	detail    githubdomain.PullRequestDetail
	selection detailCursorSelection
}

type browserChangesCursorContext struct {
	summary      githubdomain.PullRequest
	renderedRows []reviewDiffRenderedRow
	selection    detailCursorSelection
}

type reviewDiffCursorContext struct {
	summary      githubdomain.PullRequest
	renderedRows []reviewDiffRenderedRow
	selection    detailCursorSelection
}

type detailCursorReadModel struct {
	selection                  detailCursorSelection
	descriptionSummary         githubdomain.PullRequest
	descriptionDetail          githubdomain.PullRequestDetail
	descriptionKnown           bool
	browserChangesSummary      githubdomain.PullRequest
	browserChangesRenderedRows []reviewDiffRenderedRow
	browserChangesKnown        bool
	reviewDiffSummary          githubdomain.PullRequest
	reviewDiffRenderedRows     []reviewDiffRenderedRow
	reviewDiffKnown            bool
}

func (model detailCursorReadModel) currentSelection() detailCursorSelection {
	return model.selection
}

func (model detailCursorReadModel) pullRequestDescriptionContext() (pullRequestDescriptionCursorContext, bool) {
	if !model.descriptionKnown {
		return pullRequestDescriptionCursorContext{}, false
	}
	return pullRequestDescriptionCursorContext{
		summary:   model.descriptionSummary,
		detail:    model.descriptionDetail,
		selection: model.selection,
	}, true
}

func (model detailCursorReadModel) browserChangesContext() (browserChangesCursorContext, bool) {
	if !model.browserChangesKnown {
		return browserChangesCursorContext{}, false
	}
	return browserChangesCursorContext{
		summary:      model.browserChangesSummary,
		renderedRows: append([]reviewDiffRenderedRow(nil), model.browserChangesRenderedRows...),
		selection:    model.selection,
	}, true
}

func (model detailCursorReadModel) reviewDiffContext() (reviewDiffCursorContext, bool) {
	if !model.reviewDiffKnown {
		return reviewDiffCursorContext{}, false
	}
	return reviewDiffCursorContext{
		summary:      model.reviewDiffSummary,
		renderedRows: append([]reviewDiffRenderedRow(nil), model.reviewDiffRenderedRows...),
		selection:    model.selection,
	}, true
}
