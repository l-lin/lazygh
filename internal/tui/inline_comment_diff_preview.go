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

func renderDiffPreviewLine(path string, previewLine diffPreviewLine, numberWidth int, changedRanges []styledRuneRange) string {
	if previewLine.kind == diffPreviewHunkHeaderLine || previewLine.kind == diffPreviewNoteLine {
		return styleText(previewLine.text, diffPreviewHunkHeaderPrefix(previewLine.target))
	}

	numberPrefix := diffPreviewLineNumberPrefix(previewLine.target)
	prefix := styleText(diffPreviewLinePrefixText(previewLine, numberWidth), numberPrefix)
	return prefix + renderSyntaxHighlightedCode(path, previewLine.text, diffPreviewLineContentPrefix(previewLine), changedRanges)
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
		return boldPrefix + foregroundColorEscape(theme.DiffAdditionHex) + backgroundColorEscape(theme.DiffAdditionBackgroundHex)
	case diffPreviewDeletionLine:
		return boldPrefix + foregroundColorEscape(theme.DiffDeletionHex) + backgroundColorEscape(theme.DiffDeletionBackgroundHex)
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

func diffPreviewChangedStyleRanges(previewLines []diffPreviewLine) [][]styledRuneRange {
	rangesByLine := make([][]styledRuneRange, len(previewLines))
	for groupStart := 0; groupStart < len(previewLines); {
		if !diffPreviewLineSupportsIntralineHighlight(previewLines[groupStart].kind) {
			groupStart++
			continue
		}

		groupEnd := groupStart
		deletionIndexes := make([]int, 0)
		additionIndexes := make([]int, 0)
		for groupEnd < len(previewLines) && diffPreviewLineSupportsIntralineHighlight(previewLines[groupEnd].kind) {
			switch previewLines[groupEnd].kind {
			case diffPreviewDeletionLine:
				deletionIndexes = append(deletionIndexes, groupEnd)
			case diffPreviewAdditionLine:
				additionIndexes = append(additionIndexes, groupEnd)
			}
			groupEnd++
		}

		pairCount := minInt(len(deletionIndexes), len(additionIndexes))
		for pairIndex := range pairCount {
			deletionLineIndex := deletionIndexes[pairIndex]
			additionLineIndex := additionIndexes[pairIndex]
			deletionRanges, additionRanges := reviewDiffLineChangedStyleRanges(previewLines[deletionLineIndex].text, previewLines[additionLineIndex].text)
			rangesByLine[deletionLineIndex] = append(rangesByLine[deletionLineIndex], deletionRanges...)
			rangesByLine[additionLineIndex] = append(rangesByLine[additionLineIndex], additionRanges...)
		}

		groupStart = groupEnd
	}
	return rangesByLine
}

func diffPreviewLineSupportsIntralineHighlight(kind diffPreviewLineKind) bool {
	return kind == diffPreviewDeletionLine || kind == diffPreviewAdditionLine
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

func trimDiffPreviewLinesForConversation(previewLines []diffPreviewLine, comment githubcli.PullRequestInlineComment, minimumVisibleLines int) []diffPreviewLine {
	if len(previewLines) == 0 {
		return nil
	}
	if minimumVisibleLines < 1 {
		minimumVisibleLines = 1
	}

	startLine, endLine, side := pullRequestInlineCommentTargetRange(comment)
	if startLine <= 0 && endLine <= 0 {
		return append([]diffPreviewLine(nil), previewLines...)
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

	targetLineCount := endLine - startLine + 1
	leadingContextLineCount := maxInt(minimumVisibleLines-targetLineCount, 0)
	minimumPreviewLine := firstPositiveConversationPreviewLineNumber(previewLines, side)
	if minimumPreviewLine <= 0 {
		return append([]diffPreviewLine(nil), previewLines...)
	}
	previewStartLine := maxInt(minimumPreviewLine, startLine-leadingContextLineCount)

	startIndex := -1
	endIndex := -1
	for index, previewLine := range previewLines {
		lineNumber := conversationPreviewLineNumber(previewLine, side)
		if lineNumber <= 0 {
			continue
		}
		if lineNumber < previewStartLine {
			continue
		}
		if lineNumber > endLine {
			break
		}
		if startIndex < 0 {
			startIndex = index
		}
		endIndex = index
	}
	if startIndex < 0 || endIndex < 0 {
		return append([]diffPreviewLine(nil), previewLines...)
	}

	for startIndex > 0 && conversationPreviewLineNumber(previewLines[startIndex-1], side) == 0 && previewLines[startIndex-1].kind != diffPreviewHunkHeaderLine && previewLines[startIndex-1].kind != diffPreviewNoteLine {
		startIndex--
	}
	for endIndex+1 < len(previewLines) && conversationPreviewLineNumber(previewLines[endIndex+1], side) == 0 && previewLines[endIndex+1].kind != diffPreviewHunkHeaderLine && previewLines[endIndex+1].kind != diffPreviewNoteLine {
		endIndex++
	}

	prefixLineCount := 0
	for prefixLineCount < len(previewLines) && (previewLines[prefixLineCount].kind == diffPreviewHunkHeaderLine || previewLines[prefixLineCount].kind == diffPreviewNoteLine) {
		prefixLineCount++
	}
	trimmedPreviewLines := make([]diffPreviewLine, 0, prefixLineCount+(endIndex-startIndex+1))
	trimmedPreviewLines = append(trimmedPreviewLines, previewLines[:prefixLineCount]...)
	trimmedPreviewLines = append(trimmedPreviewLines, previewLines[startIndex:endIndex+1]...)
	return trimmedPreviewLines
}

func firstPositiveConversationPreviewLineNumber(previewLines []diffPreviewLine, side string) int {
	for _, previewLine := range previewLines {
		lineNumber := conversationPreviewLineNumber(previewLine, side)
		if lineNumber > 0 {
			return lineNumber
		}
	}
	return 0
}

func conversationPreviewLineNumber(previewLine diffPreviewLine, side string) int {
	if strings.EqualFold(strings.TrimSpace(side), "LEFT") {
		return previewLine.leftLine
	}
	return previewLine.rightLine
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
