package tui

import "strings"

func newBrowserConversationDetailDocument(text string, width int, wordWrapEnabled bool) detailDocument {
	styledLines := splitStyledTextLines(text)
	lines := make([]detailDocumentLine, 0, len(styledLines))
	for _, styledLine := range styledLines {
		lines = append(lines, browserConversationDetailDocumentLine(styledLine))
	}
	return newDetailDocumentFromLines(lines, width, wordWrapEnabled)
}

func browserConversationDetailDocumentLine(styledLine styledTextLine) detailDocumentLine {
	if splitColumn := browserConversationDiffPreviewBodyStartColumn(styledLine); splitColumn > 0 {
		prefix, body := splitStyledTextLine(styledLine, splitColumn)
		return detailDocumentLine{prefix: prefix, body: body}
	}
	return detailDocumentLine{body: cloneStyledTextLine(styledLine), preserveSingleRow: true}
}

func browserConversationDiffPreviewBodyStartColumn(line styledTextLine) int {
	separatorColumn := -1
	for index, character := range line.runes {
		if character == '│' {
			separatorColumn = index
			break
		}
	}
	if separatorColumn <= 0 {
		return 0
	}
	if !browserConversationUsesDiffPreviewPrefix(string(line.runes[:separatorColumn])) {
		return 0
	}
	return minInt(separatorColumn+2, len(line.runes))
}

func browserConversationUsesDiffPreviewPrefix(prefix string) bool {
	leftText, rightText, ok := strings.Cut(prefix, ":")
	if !ok {
		return false
	}
	leftNumber := strings.TrimSpace(leftText)
	rightNumber := strings.TrimSpace(rightText)
	if leftNumber == "" && rightNumber == "" {
		return false
	}
	return browserConversationUsesDiffPreviewLineNumber(leftNumber) && browserConversationUsesDiffPreviewLineNumber(rightNumber)
}

func browserConversationUsesDiffPreviewLineNumber(text string) bool {
	if text == "" {
		return true
	}
	for _, character := range text {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
