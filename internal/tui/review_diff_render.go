package tui

import (
	"fmt"
	"strconv"
	"strings"

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
	Kind reviewDiffRenderedRowKind
	Text string
}

func renderReviewDiffFile(file reviewDiffFile, renderer MarkdownRenderer, width int) string {
	rows := buildReviewDiffRenderedRows(file, renderer, width)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, row.Text)
	}
	return strings.Join(lines, "\n")
}

func buildReviewDiffRenderedRows(file reviewDiffFile, renderer MarkdownRenderer, width int) []reviewDiffRenderedRow {
	rows := []reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindFileHeader, Text: renderReviewDiffFileHeader(file)}}
	contentRows := buildReviewDiffFileContentRows(file, renderer, width)
	if len(contentRows) == 0 {
		return rows
	}
	rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindSpacer, Text: ""})
	rows = append(rows, contentRows...)
	return rows
}

func buildReviewDiffFileContentRows(file reviewDiffFile, renderer MarkdownRenderer, width int) []reviewDiffRenderedRow {
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
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindDiffLine, Text: renderReviewDiffLine(line, numberWidth)})
			for threadIndex, thread := range file.Threads {
				if matchedThreadIndexes[threadIndex] || !reviewDiffThreadMatchesLine(thread, line) {
					continue
				}
				matchedThreadIndexes[threadIndex] = true
				rows = append(rows, renderReviewDiffThreadRows(thread, renderer, width, numberWidth)...)
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
		rows = append(rows, renderReviewDiffThreadRows(thread, renderer, width, numberWidth)...)
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

func renderReviewDiffThreadRows(thread reviewDiffThread, renderer MarkdownRenderer, width int, numberWidth int) []reviewDiffRenderedRow {
	rows := make([]reviewDiffRenderedRow, 0, len(thread.Comments)*6)
	gutter := renderReviewDiffInlineCommentGutter(numberWidth)
	rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: gutter + renderReviewDiffThreadStatus(thread)})

	threadWidth := reviewDiffInlineCommentWidth(width, numberWidth)
	commentBodyWidth := commentBoxInnerWidth(threadWidth)
	for _, comment := range thread.Comments {
		body := renderMarkdownWithFallback(comment.Body, renderer, commentBodyWidth, "No comment body.")
		header := detailCommentsIcon + " " + pullRequestCommentAuthorLogin(comment.Author) + " · " + formatTimestamp(comment.CreatedAt)
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: gutter + header})
		for _, boxLine := range strings.Split(renderRoundedCommentBox(body, threadWidth), "\n") {
			rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: gutter + boxLine})
		}
	}
	if len(thread.Comments) == 0 {
		rows = append(rows, reviewDiffRenderedRow{Kind: reviewDiffRenderedRowKindInlineCommentDecoration, Text: gutter + "No comments in thread."})
	}
	return rows
}

func renderReviewDiffInlineCommentGutter(numberWidth int) string {
	prefix := fmt.Sprintf("%s : %s │ ", strings.Repeat(" ", numberWidth), strings.Repeat(" ", numberWidth))
	return styleText(prefix, foregroundColorEscape(theme.DiffLineNumberHex))
}

func reviewDiffInlineCommentWidth(width int, numberWidth int) int {
	if width < minimumMarkdownRenderWidth {
		width = defaultDetailWrapWidth
	}
	availableWidth := width - runeCountInt(fmt.Sprintf("%s : %s │ ", strings.Repeat(" ", numberWidth), strings.Repeat(" ", numberWidth)))
	if availableWidth < minimumMarkdownRenderWidth {
		return minimumMarkdownRenderWidth
	}
	return availableWidth
}

func renderReviewDiffThreadStatus(thread reviewDiffThread) string {
	status := "Conversation"
	switch {
	case thread.IsOutdated:
		status = "Conversation · outdated"
	case thread.IsResolved:
		status = "Conversation · resolved"
	}
	return styleText("↪ "+status, foregroundColorEscape(theme.DiffHunkHeaderHex))
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
