package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func renderPullRequestChangesRows(rows []reviewDiffRenderedRow) string {
	if len(rows) == 0 {
		return "No changes yet."
	}
	return reviewDiffRenderedRowsText(rows)
}

func buildPullRequestChangesRenderedRows(files []reviewDiffFile, renderer MarkdownRenderer, width int) []reviewDiffRenderedRow {
	return buildPullRequestChangesRenderedRowsForViewer(files, renderer, width, nil, nil, "")
}

func buildPullRequestChangesRenderedRowsForViewer(files []reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool, collapsedFileIDs map[string]bool, connectedUserLogin string) []reviewDiffRenderedRow {
	return buildPullRequestChangesRenderedRowsForViewerWithWordWrap(files, renderer, width, true, collapsedThreadIDs, collapsedFileIDs, connectedUserLogin)
}

func buildPullRequestChangesRenderedRowsForViewerWithWordWrap(files []reviewDiffFile, renderer MarkdownRenderer, width int, wordWrapEnabled bool, collapsedThreadIDs map[string]bool, collapsedFileIDs map[string]bool, connectedUserLogin string) []reviewDiffRenderedRow {
	rows := make([]reviewDiffRenderedRow, 0, len(files)*8)
	for index, file := range files {
		filePath := strings.TrimSpace(file.Path)
		collapsed := filePath != "" && collapsedFileIDs != nil && collapsedFileIDs[filePath]
		if index > 0 {
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
		}
		rows = append(rows, reviewDiffFileHeaderRows(file, renderFoldableReviewDiffFileHeader(file, collapsed))...)
		if collapsed {
			continue
		}
		contentRows := reviewDiffRowsWithFilePath(buildReviewDiffFileContentRowsForViewerAndWordWrap(file, renderer, width, wordWrapEnabled, collapsedThreadIDs, connectedUserLogin), filePath)
		if len(contentRows) == 0 {
			continue
		}
		if reviewDiffHeaderRowsNeedContentSpacer(rows) {
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: "", FilePath: filePath})
		}
		rows = append(rows, contentRows...)
	}
	return rows
}

func browserChangesFileSectionID(summary githubdomain.PullRequest, filePath string) string {
	return browserDetailSectionID(pullRequestDetailKey(summary.Repository, summary.Number), "changes-file", 0, filePath)
}

func browserChangesThreadSectionID(summary githubdomain.PullRequest, thread reviewDiffThread) string {
	return browserDetailSectionID(pullRequestDetailKey(summary.Repository, summary.Number), "changes-thread", 0, thread.ID)
}

func browserCollapsedChangesFileIDs(collapsedSectionStates map[string]bool, summary githubdomain.PullRequest, files []reviewDiffFile) map[string]bool {
	collapsedFileIDs := map[string]bool{}
	for _, file := range files {
		filePath := strings.TrimSpace(file.Path)
		if filePath == "" {
			continue
		}
		collapsedFileIDs[filePath] = browserDetailSectionCollapsed(collapsedSectionStates, browserChangesFileSectionID(summary, filePath), false)
	}
	if len(collapsedFileIDs) == 0 {
		return nil
	}
	return collapsedFileIDs
}

func browserCollapsedChangesThreadIDs(collapsedSectionStates map[string]bool, summary githubdomain.PullRequest, files []reviewDiffFile) map[string]bool {
	collapsedThreadIDs := map[string]bool{}
	for _, file := range files {
		for _, thread := range file.Threads {
			threadID := strings.TrimSpace(thread.ID)
			if threadID == "" {
				continue
			}
			collapsedThreadIDs[threadID] = browserDetailSectionCollapsed(collapsedSectionStates, browserChangesThreadSectionID(summary, thread), thread.IsResolved)
		}
	}
	if len(collapsedThreadIDs) == 0 {
		return nil
	}
	return collapsedThreadIDs
}

func reviewDiffFilePathAtCursor(renderedRows []reviewDiffRenderedRow, document detailDocument, state detailViewState) (string, bool) {
	row, ok := reviewDiffRenderedRowAtCursor(renderedRows, document, state)
	if !ok || strings.TrimSpace(row.FilePath) == "" {
		return "", false
	}
	return strings.TrimSpace(row.FilePath), true
}

func reviewDiffRenderedRowAtCursor(renderedRows []reviewDiffRenderedRow, document detailDocument, state detailViewState) (reviewDiffRenderedRow, bool) {
	if len(renderedRows) == 0 || len(document.rows) == 0 {
		return reviewDiffRenderedRow{}, false
	}

	rowIndex := document.rowIndexForPosition(state.cursor)
	if rowIndex < 0 || rowIndex >= len(document.rows) {
		return reviewDiffRenderedRow{}, false
	}
	renderedRowIndex := document.rows[rowIndex].line
	if renderedRowIndex < 0 || renderedRowIndex >= len(renderedRows) {
		return reviewDiffRenderedRow{}, false
	}
	return renderedRows[renderedRowIndex], true
}

func reviewDiffFileHeaderLineIndex(renderedRows []reviewDiffRenderedRow, filePath string) int {
	trimmedFilePath := strings.TrimSpace(filePath)
	for index, row := range renderedRows {
		if row.Kind != reviewDiffRenderedRowKindFileHeader || strings.TrimSpace(row.FilePath) != trimmedFilePath {
			continue
		}
		return index
	}
	return -1
}
