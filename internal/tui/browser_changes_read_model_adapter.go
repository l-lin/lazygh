package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) browserCollapsedChangesFileIDs(summary githubdomain.PullRequest, files []reviewDiffFile) map[string]bool {
	if program == nil {
		return browserCollapsedChangesFileIDs(nil, summary, files)
	}
	return browserCollapsedChangesFileIDs(program.browserCollapsedSectionStates, summary, files)
}

func (program *Program) browserCollapsedChangesThreadIDs(summary githubdomain.PullRequest, files []reviewDiffFile) map[string]bool {
	if program == nil {
		return browserCollapsedChangesThreadIDs(nil, summary, files)
	}
	return browserCollapsedChangesThreadIDs(program.browserCollapsedSectionStates, summary, files)
}

func (program *Program) browserCollapsedCommitDiffFileIDs(pullRequestKey string, commitOID string, files []reviewDiffFile) map[string]bool {
	sectionScopeKey := commitDiffCacheKey(pullRequestKey, commitOID)
	if program == nil {
		return browserCollapsedChangesFileIDsForScope(nil, sectionScopeKey, files)
	}
	return browserCollapsedChangesFileIDsForScope(program.browserCollapsedSectionStates, sectionScopeKey, files)
}

func (program *Program) browserCollapsedCommitDiffThreadIDs(pullRequestKey string, commitOID string, files []reviewDiffFile) map[string]bool {
	sectionScopeKey := commitDiffCacheKey(pullRequestKey, commitOID)
	if program == nil {
		return browserCollapsedChangesThreadIDsForScope(nil, sectionScopeKey, files)
	}
	return browserCollapsedChangesThreadIDsForScope(program.browserCollapsedSectionStates, sectionScopeKey, files)
}

func (program *Program) currentBrowserChangesReadModel(summary any, detailDocument detailDocument) (browserChangesReadModel, bool) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok || program == nil {
		return browserChangesReadModel{}, false
	}
	result, ok := program.pullRequestDiffForSummary(summaryValue)
	if !ok || result.err != nil {
		return browserChangesReadModel{}, false
	}

	selection := currentDetailCursorSelectionFor(detailDocument, program.detailState.viewState)
	model := browserChangesReadModel{
		sectionScopeKey:    pullRequestDetailKey(summaryValue.Repository, summaryValue.Number),
		files:              append([]reviewDiffFile(nil), result.data.Files...),
		selection:          selection,
		renderedRows:       copyReviewDiffRenderedRows(program.currentPullRequestChangesRenderedRows(summaryValue, result.data.Files, maxInt(selection.document.width, 1))),
		renderer:           program.markdownRenderer,
		wordWrapEnabled:    program.detailWordWrapEnabled(),
		connectedUserLogin: program.currentConnectedUserLogin(),
		collapsedFileIDs:   program.browserCollapsedChangesFileIDs(summaryValue, result.data.Files),
		collapsedThreadIDs: program.browserCollapsedChangesThreadIDs(summaryValue, result.data.Files),
	}
	return model, true
}

func (program *Program) currentCommitDiffReadModel(pullRequestKey string, commitOID string, detailDocument detailDocument) (browserChangesReadModel, bool) {
	if program == nil {
		return browserChangesReadModel{}, false
	}
	result, ok := program.commitDiffResultForTarget(pullRequestKey, commitOID)
	if !ok || result.err != nil {
		return browserChangesReadModel{}, false
	}

	selection := currentDetailCursorSelectionFor(detailDocument, program.detailState.viewState)
	model := browserChangesReadModel{
		sectionScopeKey:    commitDiffCacheKey(pullRequestKey, commitOID),
		files:              append([]reviewDiffFile(nil), result.data.Files...),
		selection:          selection,
		renderedRows:       copyReviewDiffRenderedRows(program.currentCommitDiffRenderedRows(pullRequestKey, commitOID, result.data.Files, maxInt(selection.document.width, 1))),
		renderer:           program.markdownRenderer,
		wordWrapEnabled:    program.detailWordWrapEnabled(),
		connectedUserLogin: program.currentConnectedUserLogin(),
		collapsedFileIDs:   program.browserCollapsedCommitDiffFileIDs(pullRequestKey, commitOID, result.data.Files),
		collapsedThreadIDs: program.browserCollapsedCommitDiffThreadIDs(pullRequestKey, commitOID, result.data.Files),
	}
	return model, true
}

func (program *Program) toggleChangesVisibility(model browserChangesReadModel) (detailViewSyncPlan, bool) {
	if thread, ok := model.threadAtCursor(); ok {
		sectionID := browserChangesThreadSectionIDForScope(model.sectionScopeKey, thread)
		collapsed := browserDetailSectionCollapsed(program.browserCollapsedSectionStates, sectionID, thread.IsResolved)
		program.setBrowserDetailSectionCollapsed(sectionID, !collapsed)
		return model.threadVisibilityPlan(thread)
	}

	filePath, ok := model.filePathAtCursor()
	if !ok {
		return detailViewSyncPlan{}, false
	}
	sectionID := browserChangesFileSectionIDForScope(model.sectionScopeKey, filePath)
	collapsed := browserDetailSectionCollapsed(program.browserCollapsedSectionStates, sectionID, false)
	program.setBrowserDetailSectionCollapsed(sectionID, !collapsed)
	return model.fileVisibilityPlan(filePath)
}

func (program *Program) toggleBrowserChangesVisibility(summary githubdomain.PullRequest, detailDocument detailDocument) (detailViewSyncPlan, bool) {
	model, ok := program.currentBrowserChangesReadModel(summary, detailDocument)
	if !ok {
		return detailViewSyncPlan{}, false
	}
	return program.toggleChangesVisibility(model)
}

func (program *Program) toggleCommitDiffVisibility(pullRequestKey string, commitOID string, detailDocument detailDocument) (detailViewSyncPlan, bool) {
	model, ok := program.currentCommitDiffReadModel(pullRequestKey, commitOID, detailDocument)
	if !ok {
		return detailViewSyncPlan{}, false
	}
	return program.toggleChangesVisibility(model)
}
