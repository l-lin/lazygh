package tui

import (
	"fmt"
	"strconv"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func renderReviewDiffFile(file reviewDiffFile) string {
	header := renderReviewDiffFileHeader(file)
	if strings.TrimSpace(file.Placeholder) != "" {
		return renderPullRequestDetailContent(header, file.Placeholder)
	}
	if len(file.Hunks) == 0 {
		return renderPullRequestDetailContent(header, reviewDiffPlaceholder(file))
	}

	lines := make([]string, 0, len(file.Hunks)*4)
	numberWidth := reviewDiffLineNumberWidth(file.Hunks)
	for hunkIndex, hunk := range file.Hunks {
		if hunkIndex > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, renderReviewDiffHunkHeader(hunk.Header))
		for _, line := range hunk.Lines {
			lines = append(lines, renderReviewDiffLine(line, numberWidth))
		}
	}
	return renderPullRequestDetailContent(header, strings.Join(lines, "\n"))
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
