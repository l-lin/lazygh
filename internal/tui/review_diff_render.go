package tui

import (
	"fmt"
	"strconv"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

type reviewDiffRenderedRowKind int

const (
	reviewDiffRenderedRowKindSpacer reviewDiffRenderedRowKind = iota
	reviewDiffRenderedRowKindFileHeader
	reviewDiffRenderedRowKindHunkHeader
	reviewDiffRenderedRowKindDiffLine
	reviewDiffRenderedRowKindInlineCommentDecoration
	reviewDiffRenderedRowKindNote
)

type reviewDiffRenderedRow struct {
	Kind    reviewDiffRenderedRowKind
	Text    string
	Anchor  *reviewDiffRenderedRowAnchor
	Thread  *reviewDiffThread
	Comment *githubcli.PullRequestComment
}

func renderReviewDiffFile(file reviewDiffFile, renderer MarkdownRenderer, width int) string {
	return renderReviewDiffFileWithCollapsedThreads(file, renderer, width, nil)
}

func renderReviewDiffFileWithCollapsedThreads(file reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool) string {
	rows := buildReviewDiffRenderedRowsWithCollapsedThreads(file, renderer, width, collapsedThreadIDs)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, row.Text)
	}
	return strings.Join(lines, "\n")
}

func buildReviewDiffRenderedRows(file reviewDiffFile, renderer MarkdownRenderer, width int) []reviewDiffRenderedRow {
	return buildReviewDiffRenderedRowsWithCollapsedThreads(file, renderer, width, nil)
}

func buildReviewDiffRenderedRowsWithCollapsedThreads(file reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool) []reviewDiffRenderedRow {
	rows := []reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindFileHeader, Text: renderReviewDiffFileHeader(file)}}
	contentRows := buildReviewDiffFileContentRows(file, renderer, width, collapsedThreadIDs)
	if len(contentRows) == 0 {
		return rows
	}
	rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
	rows = append(rows, contentRows...)
	return rows
}

func buildReviewDiffFileContentRows(file reviewDiffFile, renderer MarkdownRenderer, width int, collapsedThreadIDs map[string]bool) []reviewDiffRenderedRow {
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
		for _, line := range hunk.Lines {
			rows = append(rows, reviewDiffRenderedRow{
				Kind: reviewDiffRenderedRowKindDiffLine,
				Text: renderReviewDiffLine(line, numberWidth),
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
				rows = append(rows, renderReviewDiffThreadRows(thread, renderer, width, numberWidth, reviewDiffThreadCollapsed(thread, collapsedThreadIDs))...)
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
		rows = append(rows, renderReviewDiffThreadRows(thread, renderer, width, numberWidth, reviewDiffThreadCollapsed(thread, collapsedThreadIDs))...)
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
	for _, candidateLine := range thread.anchorLineNumbers() {
		if lineNumber == candidateLine {
			return true
		}
	}
	return false
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

func renderReviewDiffThreadRows(thread reviewDiffThread, renderer MarkdownRenderer, width int, _ int, collapsed bool) []reviewDiffRenderedRow {
	rows := make([]reviewDiffRenderedRow, 0, len(thread.Comments)*8)
	threadWidth := effectiveMarkdownWidth(width)
	threadCopy := thread

	rows = append(rows,
		reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""},
		reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: renderReviewDiffThreadHorizontalBorder(width), Thread: &threadCopy},
		reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: renderReviewDiffThreadStatus(thread, collapsed), Thread: &threadCopy},
	)
	if collapsed {
		rows = append(rows,
			reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: renderReviewDiffThreadHorizontalBorder(width), Thread: &threadCopy},
			reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""},
		)
		return rows
	}

	commentBodyWidth := commentBoxInnerWidth(threadWidth)
	for _, comment := range thread.Comments {
		commentCopy := comment
		body := renderInlineCommentBody(comment.Body, renderer, commentBodyWidth)
		for _, boxLine := range strings.Split(renderCommentBoxWithMetadata(comment.Author, comment.CreatedAt, body, threadWidth), "\n") {
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: boxLine, Thread: &threadCopy, Comment: &commentCopy})
		}
	}
	if len(thread.Comments) == 0 {
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: "No comments in thread.", Thread: &threadCopy})
	}
	rows = append(rows,
		reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: renderReviewDiffThreadHorizontalBorder(width), Thread: &threadCopy},
		reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""},
	)
	return rows
}

func renderReviewDiffThreadStatus(thread reviewDiffThread, collapsed bool) string {
	chevron := ""
	if collapsed {
		chevron = ""
	}

	status := fmt.Sprintf("%s Comment on line %s%d", chevron, reviewDiffThreadSideLabel(thread), reviewDiffThreadDisplayLine(thread))
	if thread.IsResolved {
		status += " · resolved"
	}
	return styleText(status, foregroundColorEscape(theme.DiffHunkHeaderHex))
}

func renderReviewDiffThreadHorizontalBorder(width int) string {
	borderWidth := effectiveMarkdownWidth(width)
	if borderWidth < 1 {
		borderWidth = 1
	}
	return styleCommentBorder(strings.Repeat("─", borderWidth))
}

func reviewDiffThreadSideLabel(thread reviewDiffThread) string {
	switch thread.anchorSide() {
	case reviewDiffLineSideLeft:
		return "L"
	case reviewDiffLineSideRight:
		return "R"
	default:
		return "?"
	}
}

func reviewDiffThreadDisplayLine(thread reviewDiffThread) int {
	switch thread.anchorSide() {
	case reviewDiffLineSideLeft:
		return firstPositive(thread.OriginalStartLine, thread.OriginalLine, thread.StartLine, thread.Line)
	case reviewDiffLineSideRight:
		return firstPositive(thread.StartLine, thread.Line, thread.OriginalStartLine, thread.OriginalLine)
	default:
		return firstPositive(thread.StartLine, thread.Line, thread.OriginalStartLine, thread.OriginalLine)
	}
}

func renderReviewDiffFileHeader(file reviewDiffFile) string {
	parts := []string{
		styleText(reviewDiffHeaderPathIcon, foregroundColorEscape(theme.DiffLineNumberHex)) + " " + valueOrDash(strings.TrimSpace(file.Path)),
	}
	if file.ChangeType == reviewDiffChangeTypeRenamed && strings.TrimSpace(file.PreviousPath) != "" {
		parts = append(parts, fmt.Sprintf("renamed from %s", strings.TrimSpace(file.PreviousPath)))
	}
	parts = append(parts,
		styleText(fmt.Sprintf("+%d", file.Additions), foregroundColorEscape(theme.DiffAdditionForegroundHex)),
		styleText(fmt.Sprintf("-%d", file.Deletions), foregroundColorEscape(theme.DiffDeletionForegroundHex)),
	)
	return strings.Join(parts, "  ")
}

func renderReviewDiffHunkHeader(header string) string {
	return styleText(header, foregroundColorEscape(theme.DiffHunkHeaderHex))
}

func renderReviewDiffLine(line reviewDiffLine, numberWidth int) string {
	numberPrefix := foregroundColorEscape(theme.DiffLineNumberHex)
	prefix := styleText(
		fmt.Sprintf("%s : %s │ ", diffPreviewLineNumberText(line.LeftLine, numberWidth), diffPreviewLineNumberText(line.RightLine, numberWidth)),
		numberPrefix,
	)
	content := " " + line.Text
	switch line.Kind {
	case reviewDiffDeletionLine:
		return prefix + styleText("-"+line.Text, foregroundColorEscape(theme.DiffDeletionForegroundHex), backgroundColorEscape(theme.DiffDeletionBackgroundHex))
	case reviewDiffAdditionLine:
		return prefix + styleText("+"+line.Text, foregroundColorEscape(theme.DiffAdditionForegroundHex), backgroundColorEscape(theme.DiffAdditionBackgroundHex))
	default:
		return prefix + content
	}
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
