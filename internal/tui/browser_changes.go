package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func renderPullRequestChangesRows(rows []reviewDiffRenderedRow) string {
	if len(rows) == 0 {
		return "No changes yet."
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, row.Text)
	}
	return strings.Join(lines, "\n")
}

func buildPullRequestChangesRenderedRows(files []reviewDiffFile, renderer MarkdownRenderer, width int) []reviewDiffRenderedRow {
	return buildPullRequestChangesRenderedRowsForViewer(files, renderer, width, nil, "")
}

func buildPullRequestChangesRenderedRowsForViewer(files []reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool, connectedUserLogin string) []reviewDiffRenderedRow {
	rows := make([]reviewDiffRenderedRow, 0, len(files)*8)
	for index, file := range files {
		if index > 0 {
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
		}
		rows = append(rows, buildReviewDiffRenderedRowsWithCollapsedThreadsForViewer(file, renderer, width, collapsedThreadIDs, connectedUserLogin)...)
	}
	return rows
}

func browserChangesThreadSectionID(summary githubcli.PullRequest, thread reviewDiffThread) string {
	return browserDetailSectionID(pullRequestDetailKey(summary.Repository, summary.Number), "changes-thread", 0, thread.ID)
}

func (program *Program) browserCollapsedChangesThreadIDs(summary githubcli.PullRequest, files []reviewDiffFile) map[string]bool {
	collapsedThreadIDs := map[string]bool{}
	for _, file := range files {
		for _, thread := range file.Threads {
			threadID := strings.TrimSpace(thread.ID)
			if threadID == "" {
				continue
			}
			collapsedThreadIDs[threadID] = program.browserDetailSectionCollapsed(browserChangesThreadSectionID(summary, thread), thread.IsResolved)
		}
	}
	if len(collapsedThreadIDs) == 0 {
		return nil
	}
	return collapsedThreadIDs
}

func (program *Program) currentPullRequestChangesRenderedRows(summary githubcli.PullRequest, files []reviewDiffFile, width int) []reviewDiffRenderedRow {
	return buildPullRequestChangesRenderedRowsForViewer(files, program.markdownRenderer, width, program.browserCollapsedChangesThreadIDs(summary, files), program.currentConnectedUserLogin())
}

func (program *Program) toggleBrowserChangesThreadVisibility(gui *gocui.Gui, summary githubcli.PullRequest, detailDocument detailDocument) error {
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok || result.err != nil {
		return nil
	}

	renderedRows := program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width)
	thread, ok := reviewDiffThreadAtCursor(renderedRows, detailDocument, program.detailViewState)
	if !ok {
		return nil
	}

	sectionID := browserChangesThreadSectionID(summary, thread)
	collapsed := program.browserDetailSectionCollapsed(sectionID, thread.IsResolved)
	program.setBrowserDetailSectionCollapsed(sectionID, !collapsed)

	updatedRows := program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width)
	headerLineIndex := reviewDiffThreadHeaderLineIndex(updatedRows, thread.ID)
	if headerLineIndex >= 0 {
		program.detailViewState.cursor = detailPosition{line: headerLineIndex, column: 0}
		program.detailViewState.preferredColumn = 0
	}
	if gui == nil {
		return nil
	}
	return program.refreshViews(gui)
}
