package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) currentDetailDocument(view *gocui.View) detailDocument {
	width := program.detailState.wrapWidth
	if view != nil && view.InnerWidth() > 0 {
		width = view.InnerWidth()
	}
	if width < 1 {
		width = 1
	}

	if program.reviewModeActive() && !program.reviewSessionShowsDescription() && !program.reviewSessionShowsStoryChapter() {
		if selectedFile, ok := program.selectedReviewSessionDiffFile(); ok {
			return program.currentReviewDiffDocument(selectedFile, width)
		}
	}

	if cacheKey, ok := program.currentPullRequestDetailDocumentCacheKey(width); ok {
		if document, ok := program.pullRequestDetailDocumentForKey(cacheKey); ok {
			return document
		}

		document := program.buildCurrentDetailDocument(width)
		program.cachePullRequestDetailDocument(cacheKey, document)
		return document
	}

	return program.buildCurrentDetailDocument(width)
}

func (program *Program) currentPullRequestConversationDocument(summary any, detail any, width int) browserConversationDocument {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return browserConversationDocument{}
	}
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		return browserConversationDocument{}
	}
	if cacheKey, ok := pullRequestConversationDocumentCacheKey(summaryValue, width); ok {
		if document, ok := program.pullRequestConversationDocumentForKey(cacheKey); ok {
			return document
		}

		document := buildBrowserConversationDocument(program.buildPullRequestConversationSections(summaryValue, detailValue, width))
		program.cachePullRequestConversationDocument(cacheKey, document)
		return document
	}

	return buildBrowserConversationDocument(program.buildPullRequestConversationSections(summaryValue, detailValue, width))
}

func (program *Program) currentPullRequestChangesRenderedRows(summary githubdomain.PullRequest, files []reviewDiffFile, width int) []reviewDiffRenderedRow {
	if cacheKey, ok := pullRequestChangesRenderedRowsCacheKey(summary, width); ok {
		if rows, ok := program.pullRequestChangesRenderedRowsForKey(cacheKey); ok {
			return rows
		}

		rows := buildPullRequestChangesRenderedRowsForViewer(files, program.markdownRenderer, width, program.browserCollapsedChangesThreadIDs(summary, files), program.browserCollapsedChangesFileIDs(summary, files), program.currentConnectedUserLogin())
		program.cachePullRequestChangesRenderedRows(cacheKey, rows)
		return rows
	}

	return buildPullRequestChangesRenderedRowsForViewer(files, program.markdownRenderer, width, program.browserCollapsedChangesThreadIDs(summary, files), program.browserCollapsedChangesFileIDs(summary, files), program.currentConnectedUserLogin())
}

func (program *Program) currentReviewDiffRenderedRows(file reviewDiffFile, width int) []reviewDiffRenderedRow {
	cacheKey := program.reviewDiffRenderKey(file, width)
	if entry, ok := program.cachedReviewDiffRenderEntry(cacheKey); ok && len(entry.rows) > 0 {
		return entry.rows
	}

	rows := buildReviewDiffRenderedRowsWithCollapsedThreadsForViewer(file, program.markdownRenderer, width, program.navigationState.reviewSession.collapsedThreadIDs, program.currentConnectedUserLogin())
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
	document := newReviewDiffDetailDocument(rows, width)
	entry, _ := program.cachedReviewDiffRenderEntry(cacheKey)
	entry.document = document
	program.storeReviewDiffRenderEntry(cacheKey, entry)
	return document
}
