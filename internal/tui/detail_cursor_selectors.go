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

type pullRequestCommitsCursorContext struct {
	summary   githubdomain.PullRequest
	commits   []githubdomain.PullRequestCommit
	selection detailCursorSelection
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
	pullRequestCommitsSummary  githubdomain.PullRequest
	pullRequestCommits         []githubdomain.PullRequestCommit
	pullRequestCommitsKnown    bool
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

func (model detailCursorReadModel) pullRequestCommitsContext() (pullRequestCommitsCursorContext, bool) {
	if !model.pullRequestCommitsKnown {
		return pullRequestCommitsCursorContext{}, false
	}
	return pullRequestCommitsCursorContext{
		summary:   model.pullRequestCommitsSummary,
		commits:   append([]githubdomain.PullRequestCommit(nil), model.pullRequestCommits...),
		selection: model.selection,
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

func commitAtCursor(context pullRequestCommitsCursorContext) (githubdomain.PullRequestCommit, bool) {
	orderedCommits := sortedPullRequestCommitsDescending(context.commits)
	if len(orderedCommits) == 0 || len(context.selection.document.lines) == 0 {
		return githubdomain.PullRequestCommit{}, false
	}

	lineIndex := context.selection.document.clampPosition(context.selection.state.cursor).line
	expectedCommitIndex := 0
	for documentLineIndex, line := range context.selection.document.lines {
		if expectedCommitIndex >= len(orderedCommits) {
			break
		}
		if string(line) != renderPullRequestCommitHeaderLineText(orderedCommits[expectedCommitIndex]) {
			continue
		}
		if documentLineIndex == lineIndex {
			return orderedCommits[expectedCommitIndex], true
		}
		expectedCommitIndex++
	}
	return githubdomain.PullRequestCommit{}, false
}
