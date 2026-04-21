package tui

import (
	"fmt"
	"strconv"
	"strings"
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
		lines = append(lines, hunk.Header)
		for _, line := range hunk.Lines {
			lines = append(lines, renderReviewDiffLine(line, numberWidth))
		}
	}
	return renderPullRequestDetailContent(header, strings.Join(lines, "\n"))
}

func renderReviewDiffFileHeader(file reviewDiffFile) string {
	parts := []string{valueOrDash(strings.TrimSpace(file.Path))}
	if file.ChangeType == reviewDiffChangeTypeRenamed && strings.TrimSpace(file.PreviousPath) != "" {
		parts = append(parts, fmt.Sprintf("renamed from %s", strings.TrimSpace(file.PreviousPath)))
	}
	parts = append(parts, fmt.Sprintf("+%d", file.Additions), fmt.Sprintf("-%d", file.Deletions))
	return strings.Join(parts, "  ·  ")
}

func renderReviewDiffLine(line reviewDiffLine, numberWidth int) string {
	prefix := fmt.Sprintf("%s : %s │ ", diffPreviewLineNumberText(line.LeftLine, numberWidth), diffPreviewLineNumberText(line.RightLine, numberWidth))
	marker := " "
	switch line.Kind {
	case reviewDiffDeletionLine:
		marker = "-"
	case reviewDiffAdditionLine:
		marker = "+"
	}
	return prefix + marker + line.Text
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
