package tui

import (
	"strings"
	"unicode/utf8"

	"github.com/l-lin/lazygh/internal/theme"
)

const markdownCodeBlockPaddingSentinelRune = '\u2060'

type styledTextLine struct {
	runes            []rune
	stylePrefixes    []string
	hyperlinkTargets []string
	controls         []styledTextControl
}

func splitStyledTextLines(text string) []styledTextLine {
	parsedLines := parseStyledTextLines(text)
	normalizedLines := normalizeMarkdownCodeBlockAdjacentBlankLines(parsedLines)
	return addMarkdownCodeBlockPaddingLines(normalizedLines)
}

func parseStyledTextLines(text string) []styledTextLine {
	lines := []styledTextLine{{}}
	currentStylePrefix := ""
	currentHyperlinkTarget := ""

	for index := 0; index < len(text); {
		if text[index] == '\x1b' {
			if sequence, nextIndex, ok := consumeCSISequence(text, index); ok {
				if strings.HasSuffix(sequence, "m") {
					currentStylePrefix = updatedANSIStylePrefix(currentStylePrefix, sequence)
				}
				index = nextIndex
				continue
			}
			if imageSpec, nextIndex, ok := parseDetailImageMarkerSequence(text, index); ok {
				line := &lines[len(lines)-1]
				line.controls = append(line.controls, styledTextControl{column: len(line.runes), image: &imageSpec})
				index = nextIndex
				continue
			}
			if sequence, nextIndex, ok := consumeOSCSequence(text, index); ok {
				if target, ok := hyperlinkTargetFromOSCSequence(sequence); ok {
					currentHyperlinkTarget = target
				}
				index = nextIndex
				continue
			}
		}

		if text[index] == '\r' {
			index++
			continue
		}

		character, size := utf8.DecodeRuneInString(text[index:])
		if character == '\n' {
			lines = append(lines, styledTextLine{})
			index += size
			continue
		}

		line := &lines[len(lines)-1]
		line.runes = append(line.runes, character)
		line.stylePrefixes = append(line.stylePrefixes, currentStylePrefix)
		line.hyperlinkTargets = append(line.hyperlinkTargets, currentHyperlinkTarget)
		index += size
	}

	for index := range lines {
		lines[index] = trimTrailingStyledSpaces(lines[index])
	}

	return lines
}

func trimTrailingStyledSpaces(line styledTextLine) styledTextLine {
	trimmedLength := len(line.runes)
	for trimmedLength > 0 && line.runes[trimmedLength-1] == ' ' {
		trimmedLength--
	}

	line.runes = line.runes[:trimmedLength]
	line.stylePrefixes = line.stylePrefixes[:trimmedLength]
	line.hyperlinkTargets = line.hyperlinkTargets[:trimmedLength]
	for controlIndex := range line.controls {
		if line.controls[controlIndex].column > trimmedLength {
			line.controls[controlIndex].column = trimmedLength
		}
	}
	return line
}

func addMarkdownCodeBlockPaddingLines(lines []styledTextLine) []styledTextLine {
	if len(lines) == 0 {
		return lines
	}

	paddedLines := make([]styledTextLine, 0, len(lines)+6)
	for index, rawLine := range lines {
		if styledLineHasInlineCommentSuggestionSentinel(rawLine) {
			paddedLines = append(paddedLines, stripInlineCommentSuggestionSentinel(rawLine))
			continue
		}

		line := stripMarkdownCodeBlockPaddingSentinel(rawLine)
		isCodeBlockLine := styledLineUsesCommentBoxCodeBlockBackground(line)
		previousLineIsCodeBlock := index > 0 && styledLineUsesCommentBoxCodeBlockBackground(stripMarkdownCodeBlockPaddingSentinel(lines[index-1]))
		nextLineIsCodeBlock := index < len(lines)-1 && styledLineUsesCommentBoxCodeBlockBackground(stripMarkdownCodeBlockPaddingSentinel(lines[index+1]))
		previousLineIsBlank := index > 0 && styledLineIsBlank(stripMarkdownCodeBlockPaddingSentinel(lines[index-1]))
		nextLineIsBlank := index < len(lines)-1 && styledLineIsBlank(stripMarkdownCodeBlockPaddingSentinel(lines[index+1]))
		if isCodeBlockLine && !previousLineIsCodeBlock {
			if !previousLineIsBlank {
				paddedLines = append(paddedLines, styledTextLine{})
			}
			paddedLines = append(paddedLines, styledPaddingLine(line))
		}

		paddedLines = append(paddedLines, line)

		if isCodeBlockLine && !nextLineIsCodeBlock {
			paddedLines = append(paddedLines, styledPaddingLine(line))
			if !nextLineIsBlank {
				paddedLines = append(paddedLines, styledTextLine{})
			}
		}
	}

	return paddedLines
}

func normalizeMarkdownCodeBlockAdjacentBlankLines(lines []styledTextLine) []styledTextLine {
	normalizedLines := make([]styledTextLine, 0, len(lines))
	for index, line := range lines {
		if !styledLineIsBlank(line) {
			normalizedLines = append(normalizedLines, line)
			continue
		}

		previousLineIsCodeBlock := index > 0 && styledLineUsesCommentBoxCodeBlockBackground(lines[index-1])
		nextLineIsCodeBlock := index < len(lines)-1 && styledLineUsesCommentBoxCodeBlockBackground(lines[index+1])
		switch {
		case previousLineIsCodeBlock && nextLineIsCodeBlock:
			normalizedLines = append(normalizedLines, markedStyledPaddingLine(lines[index-1]))
		default:
			normalizedLines = append(normalizedLines, line)
		}
	}
	return normalizedLines
}

func styledLineHasInlineCommentSuggestionSentinel(line styledTextLine) bool {
	return len(line.runes) > 0 && line.runes[0] == inlineCommentSuggestionLineSentinelRune
}

func stripInlineCommentSuggestionSentinel(line styledTextLine) styledTextLine {
	if !styledLineHasInlineCommentSuggestionSentinel(line) {
		return line
	}

	strippedLine := styledTextLine{
		runes:            append([]rune(nil), line.runes[1:]...),
		stylePrefixes:    append([]string(nil), line.stylePrefixes[1:]...),
		hyperlinkTargets: append([]string(nil), line.hyperlinkTargets[1:]...),
		controls:         make([]styledTextControl, 0, len(line.controls)),
	}
	for _, control := range line.controls {
		if control.column == 0 {
			continue
		}
		strippedLine.controls = append(strippedLine.controls, styledTextControl{column: control.column - 1, image: control.image})
	}
	return strippedLine
}

func styledLineHasMarkdownCodeBlockPaddingSentinel(line styledTextLine) bool {
	return len(line.runes) > 0 && line.runes[0] == markdownCodeBlockPaddingSentinelRune
}

func stripMarkdownCodeBlockPaddingSentinel(line styledTextLine) styledTextLine {
	if !styledLineHasMarkdownCodeBlockPaddingSentinel(line) {
		return line
	}

	strippedLine := styledTextLine{
		runes:            append([]rune(nil), line.runes[1:]...),
		stylePrefixes:    append([]string(nil), line.stylePrefixes[1:]...),
		hyperlinkTargets: append([]string(nil), line.hyperlinkTargets[1:]...),
		controls:         make([]styledTextControl, 0, len(line.controls)),
	}
	for _, control := range line.controls {
		if control.column == 0 {
			continue
		}
		strippedLine.controls = append(strippedLine.controls, styledTextControl{column: control.column - 1, image: control.image})
	}
	return strippedLine
}

func markedStyledPaddingLine(line styledTextLine) styledTextLine {
	paddingLine := styledPaddingLine(line)
	if len(paddingLine.stylePrefixes) == 0 {
		return paddingLine
	}

	paddingTarget := ""
	if len(paddingLine.hyperlinkTargets) > 0 {
		paddingTarget = paddingLine.hyperlinkTargets[0]
	}

	return styledTextLine{
		runes:            []rune{markdownCodeBlockPaddingSentinelRune, ' '},
		stylePrefixes:    []string{paddingLine.stylePrefixes[0], paddingLine.stylePrefixes[0]},
		hyperlinkTargets: []string{paddingTarget, paddingTarget},
	}
}

func styledPaddingLine(line styledTextLine) styledTextLine {
	if len(line.stylePrefixes) == 0 {
		return styledTextLine{}
	}

	paddingTarget := ""
	if len(line.hyperlinkTargets) > 0 {
		paddingTarget = line.hyperlinkTargets[0]
	}

	return styledTextLine{runes: []rune{' '}, stylePrefixes: []string{line.stylePrefixes[0]}, hyperlinkTargets: []string{paddingTarget}}
}

func trimStyledBlankLines(text string) string {
	lines := parseStyledTextLines(text)
	if len(lines) == 0 {
		return ""
	}

	firstVisibleLineIndex := 0
	for firstVisibleLineIndex < len(lines) && styledLineIsBlank(lines[firstVisibleLineIndex]) {
		firstVisibleLineIndex++
	}
	if firstVisibleLineIndex >= len(lines) {
		return ""
	}

	lastVisibleLineIndex := len(lines) - 1
	for lastVisibleLineIndex >= firstVisibleLineIndex && styledLineIsBlank(lines[lastVisibleLineIndex]) {
		lastVisibleLineIndex--
	}

	renderedLines := make([]string, 0, lastVisibleLineIndex-firstVisibleLineIndex+1)
	for _, line := range lines[firstVisibleLineIndex : lastVisibleLineIndex+1] {
		renderedLines = append(renderedLines, renderStyledTextLine(line))
	}
	return strings.Join(renderedLines, "\n")
}

func styledLineIsBlank(line styledTextLine) bool {
	if len(line.controls) > 0 {
		return false
	}
	return strings.TrimSpace(string(line.runes)) == ""
}

func renderStyledTextLineWithWidth(line styledTextLine, width int) string {
	if width <= len(line.runes) {
		return renderStyledTextLine(line)
	}

	return renderStyledTextLine(line) + renderStyledPadding(styledTextLinePaddingPrefix(line, width), width-len(line.runes))
}

func styledTextLinePaddingPrefix(line styledTextLine, width int) string {
	if paddingPrefix := markdownFullWidthLinePaddingPrefix(width, line.stylePrefixes, 0, len(line.runes)-1); paddingPrefix != "" {
		return paddingPrefix
	}
	if styledLineUsesCommentBoxCodeBlockBackground(line) {
		return backgroundColorEscape(theme.SelectedLineBackgroundHex)
	}
	return ""
}

func renderStyledPadding(prefix string, width int) string {
	if width <= 0 {
		return ""
	}

	padding := strings.Repeat(" ", width)
	if prefix == "" {
		return padding
	}
	return prefix + padding + ansiReset
}

func renderStyledTextLine(line styledTextLine) string {
	if len(line.runes) == 0 && len(line.controls) == 0 {
		return ""
	}

	var builder strings.Builder
	currentPrefix := ""
	for index := 0; index <= len(line.runes); index++ {
		for _, control := range line.controls {
			if control.column != index || control.image == nil {
				continue
			}
			builder.WriteString(encodeDetailImageMarker(*control.image))
		}
		if index >= len(line.runes) {
			break
		}

		prefix := ""
		if index < len(line.stylePrefixes) {
			prefix = line.stylePrefixes[index]
		}
		if prefix != currentPrefix {
			if currentPrefix != "" {
				builder.WriteString(ansiReset)
			}
			if prefix != "" {
				builder.WriteString(prefix)
			}
			currentPrefix = prefix
		}
		builder.WriteRune(line.runes[index])
	}
	if currentPrefix != "" {
		builder.WriteString(ansiReset)
	}

	return builder.String()
}

func markdownFullWidthLinePaddingPrefix(width int, prefixes []string, startColumn int, endColumn int) string {
	if width <= 0 || endColumn < startColumn || (endColumn-startColumn+1) >= width {
		return ""
	}

	for _, backgroundHex := range []string{theme.MarkdownHeadingBackgroundHex, theme.SelectedLineBackgroundHex} {
		if paddingPrefix := fullWidthLinePaddingPrefixForBackground(prefixes, startColumn, endColumn, backgroundHex); paddingPrefix != "" {
			return paddingPrefix
		}
	}

	return ""
}

func fullWidthLinePaddingPrefixForBackground(prefixes []string, startColumn int, endColumn int, backgroundHex string) string {
	backgroundSequence := trueColorANSIParameterSequence(48, backgroundHex)
	if backgroundSequence == "" {
		return ""
	}

	paddingPrefix := ""
	for column := startColumn; column <= endColumn && column < len(prefixes); column++ {
		if !strings.Contains(prefixes[column], backgroundSequence) {
			return ""
		}
		paddingPrefix = prefixes[column]
	}

	return paddingPrefix
}

func styledLineUsesCommentBoxCodeBlockBackground(line styledTextLine) bool {
	if len(line.runes) == 0 {
		return false
	}

	hasCodeBlockBackground := false
	for index := range line.runes {
		if index >= len(line.stylePrefixes) {
			return false
		}
		if stylePrefixUsesCommentBoxCodeBlockBackground(line.stylePrefixes[index]) {
			hasCodeBlockBackground = true
			continue
		}
		return false
	}

	return hasCodeBlockBackground
}

func stylePrefixUsesCommentBoxCodeBlockBackground(prefix string) bool {
	for _, backgroundHex := range []string{
		theme.SelectedLineBackgroundHex,
		theme.DiffAdditionBackgroundHex,
		theme.DiffAdditionHighlightBackgroundHex,
		theme.DiffDeletionBackgroundHex,
		theme.DiffDeletionHighlightBackgroundHex,
	} {
		if stylePrefixHasBackgroundHex(prefix, backgroundHex) {
			return true
		}
	}
	return false
}

func stylePrefixHasBackgroundHex(prefix string, backgroundHex string) bool {
	backgroundSequence := trueColorANSIParameterSequence(48, backgroundHex)
	if backgroundSequence == "" {
		return false
	}
	return strings.Contains(prefix, backgroundSequence)
}

func consumeCSISequence(text string, startIndex int) (string, int, bool) {
	if startIndex+1 >= len(text) || text[startIndex] != '\x1b' || text[startIndex+1] != '[' {
		return "", startIndex, false
	}

	for index := startIndex + 2; index < len(text); index++ {
		if text[index] >= 0x40 && text[index] <= 0x7e {
			return text[startIndex : index+1], index + 1, true
		}
	}

	return text[startIndex:], len(text), true
}

func consumeOSCSequence(text string, startIndex int) (string, int, bool) {
	if startIndex+1 >= len(text) || text[startIndex] != '\x1b' || text[startIndex+1] != ']' {
		return "", startIndex, false
	}

	for index := startIndex + 2; index < len(text); index++ {
		switch text[index] {
		case '\a':
			return text[startIndex : index+1], index + 1, true
		case '\x1b':
			if index+1 < len(text) && text[index+1] == '\\' {
				return text[startIndex : index+2], index + 2, true
			}
		}
	}

	return text[startIndex:], len(text), true
}

func hyperlinkTargetFromOSCSequence(sequence string) (string, bool) {
	if !strings.HasPrefix(sequence, "\x1b]8;") {
		return "", false
	}

	trimmedSequence := strings.TrimSuffix(sequence, "\a")
	trimmedSequence = strings.TrimSuffix(trimmedSequence, "\x1b\\")
	payload := strings.TrimPrefix(trimmedSequence, "\x1b]8;")
	_, after, ok := strings.Cut(payload, ";")
	if !ok {
		return "", false
	}

	return strings.TrimSpace(after), true
}

func updatedANSIStylePrefix(currentPrefix string, sequence string) string {
	parameters := ansiSequenceParameters(sequence)
	if len(parameters) == 0 {
		return ""
	}

	strippedParameters, containsReset := stripLeadingANSIReset(parameters)
	if containsReset {
		if len(strippedParameters) == 0 {
			return ""
		}
		return "\x1b[" + strings.Join(strippedParameters, ";") + "m"
	}
	if currentPrefix == "" {
		return sequence
	}

	return currentPrefix + sequence
}

func stripLeadingANSIReset(parameters []string) ([]string, bool) {
	firstNonResetIndex := 0
	containsReset := false
	for firstNonResetIndex < len(parameters) {
		if parameters[firstNonResetIndex] != "" && parameters[firstNonResetIndex] != "0" {
			break
		}
		containsReset = true
		firstNonResetIndex++
	}

	return parameters[firstNonResetIndex:], containsReset
}

func ansiSequenceParameters(sequence string) []string {
	if !strings.HasPrefix(sequence, "\x1b[") || !strings.HasSuffix(sequence, "m") {
		return nil
	}

	parameters := strings.TrimSuffix(strings.TrimPrefix(sequence, "\x1b["), "m")
	if parameters == "" {
		return nil
	}

	return strings.Split(parameters, ";")
}
