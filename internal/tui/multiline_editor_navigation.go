package tui

import "unicode"

func (editor *multilineEditor) MoveCursorLeft() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	editor.cursor--
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorRight() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	editor.cursor++
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorUp() {
	if editor == nil {
		return
	}

	currentLineStart := editor.lineStart(editor.cursor)
	if currentLineStart == 0 {
		editor.cursor = 0
		return
	}

	desiredColumn := editor.verticalDesiredColumn()
	previousLineEnd := currentLineStart - 1
	previousLineStart := editor.lineStart(previousLineEnd)
	previousLineLength := previousLineEnd - previousLineStart
	editor.cursor = previousLineStart + min(desiredColumn, previousLineLength)
}

func (editor *multilineEditor) MoveCursorDown() {
	if editor == nil {
		return
	}

	currentLineEnd := editor.lineEnd(editor.cursor)
	if currentLineEnd >= len(editor.text) {
		editor.cursor = len(editor.text)
		return
	}

	desiredColumn := editor.verticalDesiredColumn()
	nextLineStart := currentLineEnd + 1
	nextLineEnd := editor.lineEnd(nextLineStart)
	nextLineLength := nextLineEnd - nextLineStart
	editor.cursor = nextLineStart + min(desiredColumn, nextLineLength)
}

func (editor *multilineEditor) MoveCursorToLineStart() {
	if editor == nil {
		return
	}

	editor.cursor = editor.lineStart(editor.cursor)
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorToLineEnd() {
	if editor == nil {
		return
	}

	editor.cursor = editor.lineEnd(editor.cursor)
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorWordLeft() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	cursor := editor.cursor
	for cursor > 0 && unicode.IsSpace(editor.text[cursor-1]) {
		cursor--
	}
	for cursor > 0 && !unicode.IsSpace(editor.text[cursor-1]) {
		cursor--
	}
	editor.cursor = cursor
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorWordRight() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	cursor := editor.cursor
	for cursor < len(editor.text) && unicode.IsSpace(editor.text[cursor]) {
		cursor++
	}
	for cursor < len(editor.text) && !unicode.IsSpace(editor.text[cursor]) {
		cursor++
	}
	editor.cursor = cursor
	editor.preferredColumn = -1
}

func (editor *multilineEditor) verticalDesiredColumn() int {
	if editor.preferredColumn >= 0 {
		return editor.preferredColumn
	}

	column, _ := editor.CursorXY()
	editor.preferredColumn = column
	return editor.preferredColumn
}

func (editor *multilineEditor) lineStart(index int) int {
	if editor == nil {
		return 0
	}

	if index < 0 {
		index = 0
	}
	if index > len(editor.text) {
		index = len(editor.text)
	}

	start := index
	for start > 0 && editor.text[start-1] != '\n' {
		start--
	}

	return start
}

func (editor *multilineEditor) lineEnd(index int) int {
	if editor == nil {
		return 0
	}

	if index < 0 {
		index = 0
	}
	if index > len(editor.text) {
		index = len(editor.text)
	}

	end := index
	for end < len(editor.text) && editor.text[end] != '\n' {
		end++
	}

	return end
}

func (editor *multilineEditor) normalizeCursor() {
	if editor.cursor < 0 {
		editor.cursor = 0
	}
	if editor.cursor > len(editor.text) {
		editor.cursor = len(editor.text)
	}
}
