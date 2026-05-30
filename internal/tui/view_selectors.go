package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) currentPullRequestChangesRenderedRows(summary githubdomain.PullRequest, files []reviewDiffFile, width int) []reviewDiffRenderedRow {
	if cacheKey, ok := pullRequestChangesRenderedRowsCacheKey(summary, width); ok {
		if rows, ok := program.pullRequestChangesRenderedRowsForKey(cacheKey); ok {
			return rows
		}

		rows := buildPullRequestChangesRenderedRowsForViewerWithWordWrap(files, program.markdownRenderer, width, program.detailWordWrapEnabled(), program.browserCollapsedChangesThreadIDs(summary, files), program.browserCollapsedChangesFileIDs(summary, files), program.currentConnectedUserLogin())
		program.cachePullRequestChangesRenderedRows(cacheKey, rows)
		return rows
	}

	return buildPullRequestChangesRenderedRowsForViewerWithWordWrap(files, program.markdownRenderer, width, program.detailWordWrapEnabled(), program.browserCollapsedChangesThreadIDs(summary, files), program.browserCollapsedChangesFileIDs(summary, files), program.currentConnectedUserLogin())
}

func (program *Program) currentReviewDiffRenderedRows(file reviewDiffFile, width int) []reviewDiffRenderedRow {
	cacheKey := program.reviewDiffRenderKey(file, width)
	if entry, ok := program.cachedReviewDiffRenderEntry(cacheKey); ok && len(entry.rows) > 0 {
		return entry.rows
	}

	rows := buildReviewDiffRenderedRowsWithCollapsedThreadsForViewerAndWordWrap(file, program.markdownRenderer, width, program.detailWordWrapEnabled(), program.navigationState.reviewSession.collapsedThreadIDs, program.currentConnectedUserLogin())
	entry, _ := program.cachedReviewDiffRenderEntry(cacheKey)
	entry.rows = rows
	program.storeReviewDiffRenderEntry(cacheKey, entry)
	return rows
}

func (program *Program) currentReviewDiffDocument(file reviewDiffFile, width int) detailDocument {
	cacheKey := program.reviewDiffRenderKey(file, width)
	if entry, ok := program.cachedReviewDiffRenderEntry(cacheKey); ok && len(entry.document.rows) > 0 {
		return entry.document
	}

	rows := program.currentReviewDiffRenderedRows(file, width)
	document := newReviewDiffDetailDocumentWithWordWrap(rows, width, program.detailWordWrapEnabled())
	entry, _ := program.cachedReviewDiffRenderEntry(cacheKey)
	entry.document = document
	program.storeReviewDiffRenderEntry(cacheKey, entry)
	return document
}
