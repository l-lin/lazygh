package tui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

type reviewDiffRenderedRowKind int

const (
	reviewDiffRenderedRowKindSpacer reviewDiffRenderedRowKind = iota
	reviewDiffRenderedRowKindFileHeader
	reviewDiffRenderedRowKindTeamOwners
	reviewDiffRenderedRowKindHunkHeader
	reviewDiffRenderedRowKindDiffLine
	reviewDiffRenderedRowKindInlineCommentDecoration
	reviewDiffRenderedRowKindInlineCommentHeader
	reviewDiffRenderedRowKindNote
)

type reviewDiffRenderedRow struct {
	Kind     reviewDiffRenderedRowKind
	Text     string
	FilePath string
	Anchor   *reviewDiffRenderedRowAnchor
	Thread   *reviewDiffThread
	Comment  *githubdomain.PullRequestComment
}

func renderReviewDiffFile(file reviewDiffFile, renderer MarkdownRenderer, width int) string {
	return renderReviewDiffFileForViewer(file, renderer, width, "")
}

func renderReviewDiffFileForViewer(file reviewDiffFile, renderer MarkdownRenderer, width int, connectedUserLogin string) string {
	return renderReviewDiffFileWithCollapsedThreadsForViewer(file, renderer, width, nil, connectedUserLogin)
}

func renderReviewDiffFileWithCollapsedThreadsForViewer(file reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool, connectedUserLogin string) string {
	return reviewDiffRenderedRowsText(buildReviewDiffRenderedRowsWithCollapsedThreadsForViewer(file, renderer, width, collapsedThreadIDs, connectedUserLogin))
}

func buildReviewDiffRenderedRows(file reviewDiffFile, renderer MarkdownRenderer, width int) []reviewDiffRenderedRow {
	return buildReviewDiffRenderedRowsForViewer(file, renderer, width, "")
}

func buildReviewDiffRenderedRowsForViewer(file reviewDiffFile, renderer MarkdownRenderer, width int, connectedUserLogin string) []reviewDiffRenderedRow {
	return buildReviewDiffRenderedRowsWithCollapsedThreadsForViewer(file, renderer, width, nil, connectedUserLogin)
}

func buildReviewDiffRenderedRowsWithCollapsedThreadsForViewer(file reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool, connectedUserLogin string) []reviewDiffRenderedRow {
	filePath := strings.TrimSpace(file.Path)
	rows := reviewDiffFileHeaderRows(file, renderReviewDiffFileHeader(file))
	contentRows := reviewDiffRowsWithFilePath(buildReviewDiffFileContentRowsForViewer(file, renderer, width, collapsedThreadIDs, connectedUserLogin), filePath)
	if len(contentRows) == 0 {
		return rows
	}
	if reviewDiffHeaderRowsNeedContentSpacer(rows) {
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: "", FilePath: filePath})
	}
	rows = append(rows, contentRows...)
	return rows
}

func buildReviewDiffFileContentRowsForViewer(file reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool, connectedUserLogin string) []reviewDiffRenderedRow {
	rows := make([]reviewDiffRenderedRow, 0)
	placeholder := reviewDiffPlaceholderText(file)
	if placeholder != "" {
		rows = append(rows, reviewDiffRowsFromText(placeholder, reviewDiffRenderedRowKindNote)...)
	}

	numberWidth := reviewDiffLineNumberWidth(file.Hunks)
	matchedThreadIndexes := make([]bool, len(file.Threads))
	for hunkIndex, hunk := range file.Hunks {
		if hunkIndex > 0 || len(rows) > 0 {
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
		}
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindHunkHeader, Text: renderReviewDiffHunkHeader(hunk.Header)})
		changedRangesByLine := reviewDiffChangedStyleRanges(hunk.Lines)
		for lineIndex, line := range hunk.Lines {
			rows = append(rows, reviewDiffRenderedRow{
				Kind: reviewDiffRenderedRowKindDiffLine,
				Text: renderReviewDiffLine(file.Path, line, numberWidth, changedRangesByLine[lineIndex]),
				Anchor: &reviewDiffRenderedRowAnchor{
					Path: strings.TrimSpace(file.Path),
					Line: line,
				},
			})
			for threadIndex, thread := range file.Threads {
				if matchedThreadIndexes[threadIndex] || !reviewDiffThreadMatchesLine(thread, line) {
					continue
				}
				matchedThreadIndexes[threadIndex] = true
				rows = append(rows, renderReviewDiffThreadRowsForViewer(thread, renderer, width, numberWidth, reviewDiffThreadCollapsed(thread, collapsedThreadIDs), connectedUserLogin)...)
			}
		}
	}

	unmatchedThreads := reviewDiffUnmatchedThreads(file.Threads, matchedThreadIndexes)
	if len(unmatchedThreads) == 0 {
		return rows
	}
	if len(rows) > 0 {
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
	}
	rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindNote, Text: styleText("Inline discussion without visible diff context.", foregroundColorEscape(theme.DiffHunkHeaderHex))})
	for _, thread := range unmatchedThreads {
		rows = append(rows, renderReviewDiffThreadRowsForViewer(thread, renderer, width, numberWidth, reviewDiffThreadCollapsed(thread, collapsedThreadIDs), connectedUserLogin)...)
	}
	return rows
}

func reviewDiffPlaceholderText(file reviewDiffFile) string {
	if strings.TrimSpace(file.Placeholder) != "" {
		return file.Placeholder
	}
	if len(file.Hunks) == 0 {
		return reviewDiffPlaceholder(file)
	}
	return ""
}

func reviewDiffRowsFromText(text string, kind reviewDiffRenderedRowKind) []reviewDiffRenderedRow {
	lines := strings.Split(strings.ReplaceAll(strings.TrimRight(text, "\n"), "\r", ""), "\n")
	rows := make([]reviewDiffRenderedRow, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, reviewDiffRenderedRow{Kind: kind, Text: line})
	}
	return rows
}

func reviewDiffRowsWithFilePath(rows []reviewDiffRenderedRow, filePath string) []reviewDiffRenderedRow {
	trimmedFilePath := strings.TrimSpace(filePath)
	if trimmedFilePath == "" || len(rows) == 0 {
		return rows
	}

	updatedRows := make([]reviewDiffRenderedRow, 0, len(rows))
	for _, row := range rows {
		updatedRow := row
		updatedRow.FilePath = trimmedFilePath
		updatedRows = append(updatedRows, updatedRow)
	}
	return updatedRows
}

func reviewDiffRenderedRowsText(rows []reviewDiffRenderedRow) string {
	if len(rows) == 0 {
		return ""
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, row.Text)
	}
	return strings.Join(lines, "\n")
}

func reviewDiffFileHeaderRows(file reviewDiffFile, headerText string) []reviewDiffRenderedRow {
	filePath := strings.TrimSpace(file.Path)
	rows := []reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindFileHeader, Text: headerText, FilePath: filePath}}
	if teamOwnersRow, ok := reviewDiffTeamOwnersRenderedRow(file); ok {
		rows = append(rows, teamOwnersRow)
	}
	return rows
}

func reviewDiffHeaderRowsNeedContentSpacer(rows []reviewDiffRenderedRow) bool {
	if len(rows) == 0 {
		return false
	}
	return rows[len(rows)-1].Kind != reviewDiffRenderedRowKindTeamOwners
}

func reviewDiffTeamOwnersRenderedRow(file reviewDiffFile) (reviewDiffRenderedRow, bool) {
	normalizedTeamOwners := normalizeReviewDiffTeamOwners(file.TeamOwners)
	if len(normalizedTeamOwners) == 0 {
		return reviewDiffRenderedRow{}, false
	}

	return reviewDiffRenderedRow{
		Kind:     reviewDiffRenderedRowKindTeamOwners,
		Text:     renderReviewDiffTeamOwners(normalizedTeamOwners),
		FilePath: strings.TrimSpace(file.Path),
	}, true
}

func renderReviewDiffTeamOwners(teamOwners []string) string {
	normalizedTeamOwners := normalizeReviewDiffTeamOwners(teamOwners)
	if len(normalizedTeamOwners) == 0 {
		return ""
	}
	return styleText("  "+reviewDiffTeamOwnershipIcon+" "+strings.Join(normalizedTeamOwners, ", "), foregroundColorEscape(theme.TeamOwnershipHex))
}

func reviewDiffUnmatchedThreads(threads []reviewDiffThread, matchedThreadIndexes []bool) []reviewDiffThread {
	unmatchedThreads := make([]reviewDiffThread, 0, len(threads))
	for threadIndex, thread := range threads {
		if threadIndex < len(matchedThreadIndexes) && matchedThreadIndexes[threadIndex] {
			continue
		}
		unmatchedThreads = append(unmatchedThreads, thread)
	}
	return unmatchedThreads
}

func reviewDiffThreadMatchesLine(thread reviewDiffThread, line reviewDiffLine) bool {
	side := thread.anchorSide()
	if side == reviewDiffLineSideNone || !line.supportsSide(side) {
		return false
	}
	lineNumber := reviewDiffLineNumberForSide(line, side)
	return slices.Contains(thread.anchorLineNumbers(), lineNumber)
}

func (thread reviewDiffThread) anchorSide() reviewDiffLineSide {
	if thread.Side != reviewDiffLineSideNone {
		return thread.Side
	}
	return thread.StartSide
}

func (thread reviewDiffThread) anchorLineNumbers() []int {
	lineNumbers := make([]int, 0, 4)
	for _, lineNumber := range []int{thread.Line, thread.OriginalLine, thread.StartLine, thread.OriginalStartLine} {
		if lineNumber <= 0 || indexOfInt(lineNumbers, lineNumber) >= 0 {
			continue
		}
		lineNumbers = append(lineNumbers, lineNumber)
	}
	return lineNumbers
}

func renderReviewDiffThreadRowsForViewer(thread reviewDiffThread, renderer MarkdownRenderer, width int, _ int, collapsed bool, connectedUserLogin string) []reviewDiffRenderedRow {
	rows := make([]reviewDiffRenderedRow, 0, len(thread.Comments)*8)
	threadWidth := effectiveMarkdownWidth(width)
	threadCopy := thread
	suggestionContext := pullRequestInlineCommentFromReviewDiffThread(thread)

	rows = append(rows,
		reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""},
		reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentHeader, Text: renderReviewDiffThreadStatus(thread, collapsed), Thread: &threadCopy},
	)
	if collapsed {
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
		return rows
	}

	for commentIndex, comment := range thread.Comments {
		commentCopy := comment
		renderedCommentBlock := renderInlineThreadCommentBlockForViewer(comment, suggestionContext, renderer, threadWidth, commentIndex, len(thread.Comments), connectedUserLogin)
		for boxLine := range strings.SplitSeq(renderedCommentBlock, "\n") {
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: boxLine, Thread: &threadCopy, Comment: &commentCopy})
		}
	}
	if len(thread.Comments) == 0 {
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: "No comments in thread.", Thread: &threadCopy})
	}
	rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
	return rows
}

func renderReviewDiffThreadStatus(thread reviewDiffThread, collapsed bool) string {
	comment := pullRequestInlineCommentFromReviewDiffThread(thread)
	return renderInlineThreadHeaderLine(
		pullRequestInlineCommentLocation(comment),
		collapsed,
		inlineThreadStatusBadges(thread.IsResolved, thread.IsOutdated),
	)
}

func renderReviewDiffFileHeader(file reviewDiffFile) string {
	return strings.Join(reviewDiffFileHeaderSegments(file), "  ")
}

func renderFoldableReviewDiffFileHeader(file reviewDiffFile, collapsed bool) string {
	chevron := browserDetailExpandedChevron
	if collapsed {
		chevron = browserDetailCollapsedChevron
	}

	headerSegments := reviewDiffFileHeaderSegments(file)
	if len(headerSegments) == 0 {
		return styleText(chevron, foregroundColorEscape(theme.DiffLineNumberHex))
	}
	header := styleText(chevron, foregroundColorEscape(theme.DiffLineNumberHex)) + " " + headerSegments[0]
	if len(headerSegments) == 1 {
		return header
	}
	return strings.Join(append([]string{header}, headerSegments[1:]...), "  ")
}

func reviewDiffFileHeaderSegments(file reviewDiffFile) []string {
	segments := []string{
		styleText(reviewDiffHeaderPathIcon, foregroundColorEscape(theme.DiffLineNumberHex)) + " " + valueOrDash(strings.TrimSpace(file.Path)),
	}
	if file.ChangeType == reviewDiffChangeTypeRenamed && strings.TrimSpace(file.PreviousPath) != "" {
		segments = append(segments, fmt.Sprintf("renamed from %s", strings.TrimSpace(file.PreviousPath)))
	}
	segments = append(segments,
		styleText(fmt.Sprintf("+%d", file.Additions), foregroundColorEscape(theme.DiffAdditionHex)),
		styleText(fmt.Sprintf("-%d", file.Deletions), foregroundColorEscape(theme.DiffDeletionHex)),
	)
	return segments
}

func renderReviewDiffHunkHeader(header string) string {
	return styleText(header, foregroundColorEscape(theme.DiffHunkHeaderHex))
}

func renderReviewDiffLine(path string, line reviewDiffLine, numberWidth int, changedRanges []styledRuneRange) string {
	numberPrefix := foregroundColorEscape(theme.DiffLineNumberHex)
	prefix := styleText(
		fmt.Sprintf("%s : %s │ ", diffPreviewLineNumberText(line.LeftLine, numberWidth), diffPreviewLineNumberText(line.RightLine, numberWidth)),
		numberPrefix,
	)
	basePrefix := ""
	sign := " "
	switch line.Kind {
	case reviewDiffDeletionLine:
		basePrefix = foregroundColorEscape(theme.DiffDeletionHex) + backgroundColorEscape(theme.DiffDeletionBackgroundHex)
		sign = "-"
	case reviewDiffAdditionLine:
		basePrefix = foregroundColorEscape(theme.DiffAdditionHex) + backgroundColorEscape(theme.DiffAdditionBackgroundHex)
		sign = "+"
	}

	return prefix + styleText(sign, basePrefix) + renderSyntaxHighlightedCode(path, line.Text, basePrefix, changedRanges)
}

func reviewDiffLineNumberWidth(hunks []reviewDiffHunk) int {
	width := 1
	for _, hunk := range hunks {
		for _, line := range hunk.Lines {
			width = maxInt(width, runeCountInt(strconv.Itoa(maxInt(line.LeftLine, line.RightLine))))
		}
	}
	return width
}
