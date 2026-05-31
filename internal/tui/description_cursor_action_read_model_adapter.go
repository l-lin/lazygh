package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) currentPullRequestDescriptionSummaryAndDetail() (githubdomain.PullRequest, githubdomain.PullRequestDetail, bool) {
	actionContext := program.actionContext()
	if actionContext.IsReviewContext() {
		summary, detail, ok := program.reviewSessionDescriptionSummaryAndDetail()
		if !ok {
			return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
		}
		return summary, detail, true
	}
	if !actionContext.ShowsPullRequestDescription() {
		return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
	}
	return summary, result.detail, true
}

func (program *Program) detailCursorActionsAvailable() bool {
	actionContext := program.actionContext()
	if actionContext.ActiveView.Focus == FocusDetailView {
		return true
	}
	return actionContext.IsReviewContext() && actionContext.ActiveView.Focus == FocusUserView && actionContext.ShowsPullRequestDescription()
}

func (program *Program) currentDescriptionCursorActionReadModel() descriptionCursorActionReadModel {
	if program == nil {
		return descriptionCursorActionReadModel{}
	}

	model := descriptionCursorActionReadModel{actionsAvailable: program.detailCursorActionsAvailable()}
	context, ok := program.currentPullRequestDescriptionCursorContext()
	if !ok {
		return model
	}

	model.contextKnown = true
	model.selection = context.selection
	model.summary = context.summary
	model.detail = context.detail
	if sectionAtCursor, ok := program.browserOverviewSectionAtCursor(context.summary, context.detail, context.selection.document.width, context.selection.state.cursor.line); ok {
		model.overviewCursor = sectionAtCursor
		model.overviewCursorKnown = true
	}
	return model
}
