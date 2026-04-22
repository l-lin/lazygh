package tui

const codeFenceSnippet = "```\n```"

type multilineEditor struct {
	text            []rune
	cursor          int
	preferredColumn int
}

func newMultilineEditor(text string) *multilineEditor {
	editor := &multilineEditor{preferredColumn: -1}
	editor.SetText(text)
	return editor
}

func (editor *multilineEditor) Text() string {
	if editor == nil {
		return ""
	}

	return string(editor.text)
}

func (editor *multilineEditor) Cursor() int {
	if editor == nil {
		return 0
	}

	return editor.cursor
}

func (editor *multilineEditor) CursorXY() (int, int) {
	if editor == nil {
		return 0, 0
	}

	column := 0
	row := 0
	for index := 0; index < editor.cursor && index < len(editor.text); index++ {
		if editor.text[index] == '\n' {
			row++
			column = 0
			continue
		}

		column++
	}

	return column, row
}

func (editor *multilineEditor) SetText(text string) {
	if editor == nil {
		return
	}

	editor.text = []rune(text)
	editor.cursor = len(editor.text)
	editor.preferredColumn = -1
}
