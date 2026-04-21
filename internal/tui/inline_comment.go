package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

var diffHunkHeaderPattern = regexp.MustCompile(`^@@ -([0-9]+)(?:,[0-9]+)? \+([0-9]+)(?:,[0-9]+)? @@`)

type diffPreviewLineKind int

const (
	diffPreviewHunkHeaderLine diffPreviewLineKind = iota
	diffPreviewContextLine
	diffPreviewDeletionLine
	diffPreviewAdditionLine
	diffPreviewNoteLine
)

type diffPreviewLine struct {
	kind      diffPreviewLineKind
	text      string
	leftLine  int
	rightLine int
	target    bool
}

func renderPullRequestInlineCommentSection(comment githubcli.PullRequestInlineComment, body string, width int) string {
	lines := []string{detailCommentsIcon + " " + pullRequestCommentAuthorLogin(comment.Author) + " · " + formatTimestamp(comment.CreatedAt)}
	lines = append(lines, renderPullRequestInlineCommentLocationLine(comment))

	diffPreview := renderPullRequestInlineCommentDiffPreview(comment)
	if diffPreview != "" {
		lines = append(lines, diffPreview)
	}
	lines = append(lines, renderRoundedCommentBox(body, width))
	return strings.Join(lines, "\n")
}

func renderPullRequestInlineCommentLocationLine(comment githubcli.PullRequestInlineComment) string {
	additions, deletions := diffHunkChangeCounts(comment.DiffHunk)
	location := pullRequestInlineCommentLocation(comment)
	additionText := styleText(fmt.Sprintf("+%d", additions), foregroundColorEscape(theme.DiffAdditionForegroundHex))
	deletionText := styleText(fmt.Sprintf("-%d", deletions), foregroundColorEscape(theme.DiffDeletionForegroundHex))

	segments := []string{styleText(detailInlineCommentLocationIcon, foregroundColorEscape(theme.DiffLineNumberHex))}
	if location != "" {
		segments = append(segments, location)
	}
	segments = append(segments, additionText, deletionText)
	if len(segments) == 0 {
		return ""
	}

	return segments[0] + " " + strings.Join(segments[1:], "  ")
}

func renderPullRequestInlineCommentDiffPreview(comment githubcli.PullRequestInlineComment) string {
	previewLines := parseDiffPreviewLines(comment.DiffHunk)
	if len(previewLines) == 0 {
		return styleText("No diff preview available.", foregroundColorEscape(theme.DiffHunkHeaderHex))
	}

	markTargetDiffPreviewLines(previewLines, comment)
	numberWidth := diffPreviewLineNumberWidth(previewLines)
	renderedLines := make([]string, 0, len(previewLines))
	for _, previewLine := range previewLines {
		renderedLines = append(renderedLines, renderDiffPreviewLine(previewLine, numberWidth))
	}
	return strings.Join(renderedLines, "\n")
}

func renderDiffPreviewLine(previewLine diffPreviewLine, numberWidth int) string {
	if previewLine.kind == diffPreviewHunkHeaderLine || previewLine.kind == diffPreviewNoteLine {
		return styleText(previewLine.text, diffPreviewHunkHeaderPrefix(previewLine.target))
	}

	numberPrefix := diffPreviewLineNumberPrefix(previewLine.target)
	prefix := styleText(diffPreviewLinePrefixText(previewLine, numberWidth), numberPrefix)

	contentPrefix := diffPreviewLineContentPrefix(previewLine)
	if contentPrefix == "" {
		return prefix + previewLine.text
	}
	return prefix + styleText(previewLine.text, contentPrefix)
}

func diffPreviewLinePrefixText(previewLine diffPreviewLine, numberWidth int) string {
	leftLine := diffPreviewLineNumberText(previewLine.leftLine, numberWidth)
	rightLine := diffPreviewLineNumberText(previewLine.rightLine, numberWidth)
	return fmt.Sprintf("%s : %s │ ", leftLine, rightLine)
}

func diffPreviewLineNumberText(lineNumber int, width int) string {
	if width < 1 {
		width = 1
	}
	if lineNumber <= 0 {
		return strings.Repeat(" ", width)
	}
	return fmt.Sprintf("%*d", width, lineNumber)
}

func diffPreviewLineContentPrefix(previewLine diffPreviewLine) string {
	boldPrefix := ""
	if previewLine.target {
		boldPrefix = ansiBold
	}

	switch previewLine.kind {
	case diffPreviewAdditionLine:
		return boldPrefix + foregroundColorEscape(theme.DiffAdditionForegroundHex) + backgroundColorEscape(theme.DiffAdditionBackgroundHex)
	case diffPreviewDeletionLine:
		return boldPrefix + foregroundColorEscape(theme.DiffDeletionForegroundHex) + backgroundColorEscape(theme.DiffDeletionBackgroundHex)
	case diffPreviewContextLine:
		if previewLine.target {
			return boldPrefix
		}
	}

	return ""
}

func diffPreviewLineNumberPrefix(target bool) string {
	prefix := foregroundColorEscape(theme.DiffLineNumberHex)
	if target {
		return ansiBold + prefix
	}
	return prefix
}

func diffPreviewHunkHeaderPrefix(target bool) string {
	prefix := foregroundColorEscape(theme.DiffHunkHeaderHex)
	if target {
		return ansiBold + prefix
	}
	return prefix
}

func pullRequestInlineCommentLocation(comment githubcli.PullRequestInlineComment) string {
	path := strings.TrimSpace(comment.Path)
	if path == "" {
		return ""
	}

	startLine, endLine, _ := pullRequestInlineCommentTargetRange(comment)
	switch {
	case startLine > 0 && endLine > 0 && startLine != endLine:
		return fmt.Sprintf("%s:%d-%d", path, startLine, endLine)
	case endLine > 0:
		return fmt.Sprintf("%s:%d", path, endLine)
	case startLine > 0:
		return fmt.Sprintf("%s:%d", path, startLine)
	default:
		return path
	}
}

func diffHunkChangeCounts(diffHunk string) (int, int) {
	previewLines := parseDiffPreviewLines(diffHunk)
	additions := 0
	deletions := 0
	for _, previewLine := range previewLines {
		switch previewLine.kind {
		case diffPreviewAdditionLine:
			additions++
		case diffPreviewDeletionLine:
			deletions++
		}
	}
	return additions, deletions
}

func markTargetDiffPreviewLines(previewLines []diffPreviewLine, comment githubcli.PullRequestInlineComment) {
	startLine, endLine, side := pullRequestInlineCommentTargetRange(comment)
	if startLine <= 0 && endLine <= 0 {
		return
	}
	if endLine == 0 {
		endLine = startLine
	}
	if startLine == 0 {
		startLine = endLine
	}
	if startLine > endLine {
		startLine, endLine = endLine, startLine
	}

	for index := range previewLines {
		targetLineNumber := previewLines[index].rightLine
		if side == "LEFT" {
			targetLineNumber = previewLines[index].leftLine
		}
		if targetLineNumber >= startLine && targetLineNumber <= endLine {
			previewLines[index].target = true
		}
	}
}

func pullRequestInlineCommentTargetRange(comment githubcli.PullRequestInlineComment) (int, int, string) {
	side := strings.ToUpper(strings.TrimSpace(comment.Side))
	if side == "LEFT" {
		startLine := firstPositive(comment.OriginalStartLine, comment.OriginalLine, comment.StartLine, comment.Line)
		endLine := firstPositive(comment.OriginalLine, comment.OriginalStartLine, comment.Line, comment.StartLine)
		return startLine, endLine, side
	}

	startLine := firstPositive(comment.StartLine, comment.Line, comment.OriginalStartLine, comment.OriginalLine)
	endLine := firstPositive(comment.Line, comment.StartLine, comment.OriginalLine, comment.OriginalStartLine)
	return startLine, endLine, "RIGHT"
}

func parseDiffPreviewLines(diffHunk string) []diffPreviewLine {
	normalizedDiffHunk := strings.TrimSpace(strings.ReplaceAll(diffHunk, "\r", ""))
	if normalizedDiffHunk == "" {
		return nil
	}

	lines := strings.Split(normalizedDiffHunk, "\n")
	previewLines := make([]diffPreviewLine, 0, len(lines))
	leftLineNumber := 0
	rightLineNumber := 0
	haveHunkHeader := false
	for _, line := range lines {
		if leftStart, rightStart, ok := parseDiffHunkHeader(line); ok {
			leftLineNumber = leftStart
			rightLineNumber = rightStart
			haveHunkHeader = true
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewHunkHeaderLine, text: line})
			continue
		}
		if !haveHunkHeader {
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewNoteLine, text: line})
			continue
		}
		if line == `\ No newline at end of file` {
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewNoteLine, text: line})
			continue
		}
		if line == "" {
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewContextLine, text: "", leftLine: leftLineNumber, rightLine: rightLineNumber})
			leftLineNumber++
			rightLineNumber++
			continue
		}

		switch line[0] {
		case ' ':
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewContextLine, text: line[1:], leftLine: leftLineNumber, rightLine: rightLineNumber})
			leftLineNumber++
			rightLineNumber++
		case '-':
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewDeletionLine, text: line[1:], leftLine: leftLineNumber})
			leftLineNumber++
		case '+':
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewAdditionLine, text: line[1:], rightLine: rightLineNumber})
			rightLineNumber++
		default:
			previewLines = append(previewLines, diffPreviewLine{kind: diffPreviewNoteLine, text: line})
		}
	}

	return previewLines
}

func diffPreviewLineNumberWidth(previewLines []diffPreviewLine) int {
	width := 1
	for _, previewLine := range previewLines {
		width = maxInt(width, runeCountInt(strconv.Itoa(maxInt(previewLine.leftLine, previewLine.rightLine))))
	}
	return width
}

func parseDiffHunkHeader(line string) (int, int, bool) {
	matches := diffHunkHeaderPattern.FindStringSubmatch(strings.TrimSpace(line))
	if len(matches) != 3 {
		return 0, 0, false
	}

	leftLine, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, false
	}
	rightLine, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, false
	}
	return leftLine, rightLine, true
}

func styleText(text string, prefixes ...string) string {
	if text == "" {
		return ""
	}

	prefix := strings.Join(filterEmptyStrings(prefixes), "")
	if prefix == "" {
		return text
	}
	return prefix + text + ansiReset
}

func filterEmptyStrings(values []string) []string {
	filteredValues := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		filteredValues = append(filteredValues, value)
	}
	return filteredValues
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func runeCountInt(value string) int {
	return len([]rune(value))
}
