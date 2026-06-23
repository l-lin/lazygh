package tui

import (
	"fmt"
	"strings"
)

func reviewInlineCommentSuggestionBody(snippet string) string {
	if snippet == "" {
		return ""
	}

	return fmt.Sprintf("```suggestion\n%s\n```", strings.ReplaceAll(snippet, "\r", ""))
}

func reviewDiffSelectedSnippet(renderedRows []reviewDiffRenderedRow, document detailDocument, state detailViewState) string {
	if _, _, ok := state.visualRowSelection(document); ok {
		return reviewDiffSelectedLineSnippet(renderedRows, reviewDiffSelectedRenderedRowIndexes(document, state))
	}
	if start, end, ok := state.visualSelection(document); ok {
		return reviewDiffSelectedCharacterSnippet(renderedRows, document, start, end)
	}

	return ""
}

func reviewDiffSelectedLineSnippet(renderedRows []reviewDiffRenderedRow, selectedLineIndexes []int) string {
	selectedLines := make([]string, 0, len(selectedLineIndexes))
	for _, lineIndex := range selectedLineIndexes {
		lineText, ok := reviewDiffRenderedRowText(renderedRows, lineIndex)
		if !ok {
			continue
		}
		selectedLines = append(selectedLines, lineText)
	}

	return strings.Join(selectedLines, "\n")
}

func reviewDiffSelectedCharacterSnippet(renderedRows []reviewDiffRenderedRow, document detailDocument, start detailPosition, end detailPosition) string {
	if document.comparePositions(start, end) > 0 {
		start, end = end, start
	}
	start = document.clampPosition(start)
	end = document.clampPosition(end)

	selectedLines := make([]string, 0, end.line-start.line+1)
	for lineIndex := start.line; lineIndex <= end.line; lineIndex++ {
		lineText, ok := reviewDiffRenderedRowText(renderedRows, lineIndex)
		if !ok {
			continue
		}

		lineTextRunes := []rune(lineText)
		if len(lineTextRunes) == 0 {
			selectedLines = append(selectedLines, "")
			continue
		}

		documentLine := document.lines[lineIndex]
		contentStartColumn := len(documentLine) - len(lineTextRunes)
		selectedStartColumn := 0
		selectedEndColumn := len(lineTextRunes) - 1
		if lineIndex == start.line {
			selectedStartColumn = maxInt(0, start.column-contentStartColumn)
		}
		if lineIndex == end.line {
			selectedEndColumn = minInt(len(lineTextRunes)-1, end.column-contentStartColumn)
		}
		if selectedStartColumn > selectedEndColumn || selectedEndColumn < 0 || selectedStartColumn >= len(lineTextRunes) {
			continue
		}

		selectedLines = append(selectedLines, string(lineTextRunes[selectedStartColumn:selectedEndColumn+1]))
	}

	return strings.Join(selectedLines, "\n")
}

func reviewDiffRenderedRowText(renderedRows []reviewDiffRenderedRow, lineIndex int) (string, bool) {
	if lineIndex < 0 || lineIndex >= len(renderedRows) {
		return "", false
	}
	anchor := renderedRows[lineIndex].Anchor
	if anchor == nil {
		return "", false
	}
	return anchor.Line.Text, true
}
