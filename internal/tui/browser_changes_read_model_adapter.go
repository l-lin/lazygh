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
		summary:            summaryValue,
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

func (program *Program) toggleBrowserChangesVisibility(summary githubdomain.PullRequest, detailDocument detailDocument) (detailViewSyncPlan, bool) {
	model, ok := program.currentBrowserChangesReadModel(summary, detailDocument)
	if !ok {
		return detailViewSyncPlan{}, false
	}

	if thread, ok := model.threadAtCursor(); ok {
		sectionID := browserChangesThreadSectionID(model.summary, thread)
		collapsed := browserDetailSectionCollapsed(program.browserCollapsedSectionStates, sectionID, thread.IsResolved)
		program.setBrowserDetailSectionCollapsed(sectionID, !collapsed)
		return model.threadVisibilityPlan(thread)
	}

	filePath, ok := model.filePathAtCursor()
	if !ok {
		return detailViewSyncPlan{}, false
	}
	sectionID := browserChangesFileSectionID(model.summary, filePath)
	collapsed := browserDetailSectionCollapsed(program.browserCollapsedSectionStates, sectionID, false)
	program.setBrowserDetailSectionCollapsed(sectionID, !collapsed)
	return model.fileVisibilityPlan(filePath)
}
