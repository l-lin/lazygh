package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) configureModalEditorView(view *gocui.View) {
	configureFramedOverlayView(view, program.modalEditorTitle(), "")
	view.Wrap = program.modalEditorVisible() && !program.overlayState.modalEditor.isLineEditor()
	view.Highlight = false
	view.Editable = true
	view.Editor = gocui.EditorFunc(program.editModalEditor)
}

func (program *Program) renderModalEditorView(view *gocui.View) {
	if view == nil || !program.modalEditorVisible() {
		return
	}

	view.Clear()
	text := program.overlayState.modalEditor.Text()
	fmt.Fprint(view, text)
	if view.Wrap {
		program.setWrappedMultilineInputCursor(view, text, program.overlayState.modalEditor.Cursor())
		return
	}
	column, row := program.overlayState.modalEditor.CursorXY()
	program.setMultilineInputCursor(view, column, row)
}

func (program *Program) editModalEditor(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyAltEnter || key == gocui.KeyEsc {
		return false
	}
	if !program.modalEditorVisible() {
		return false
	}
	if program.overlayState.modalEditor.isLineEditor() {
		intent, ok := lineEditorIntentFromKey(key, ch, mod)
		if !ok {
			return false
		}
		return program.dispatchEditorMessage(MsgModalEditorLineInputRequested{Intent: intent})
	}
	intent, ok := multilineEditorIntentFromKey(key, ch, mod)
	if !ok {
		return false
	}
	return program.dispatchEditorMessage(MsgModalEditorMultilineInputRequested{Intent: intent})
}

func (program *Program) modalEditorTitle() string {
	if !program.modalEditorVisible() {
		return ""
	}

	title := strings.TrimSpace(program.overlayState.modalEditor.title)
	message := strings.TrimSpace(program.overlayState.modalEditor.errorMessage)
	if message == "" {
		return title
	}
	if title == "" {
		return message
	}

	return fmt.Sprintf("%s · %s", title, message)
}

func (program *Program) setMultilineInputCursor(view *gocui.View, column int, row int) {
	if view == nil {
		return
	}

	innerWidth := max(view.InnerWidth(), 1)
	innerHeight := max(view.InnerHeight(), 1)

	if column < 0 {
		column = 0
	}
	if row < 0 {
		row = 0
	}

	originX := 0
	if column >= innerWidth {
		originX = column - innerWidth + 1
	}
	originY := 0
	if row >= innerHeight {
		originY = row - innerHeight + 1
	}

	cursorX := column - originX
	cursorY := row - originY
	if cursorX < 0 {
		cursorX = 0
	}
	if cursorX >= innerWidth {
		cursorX = innerWidth - 1
	}
	if cursorY < 0 {
		cursorY = 0
	}
	if cursorY >= innerHeight {
		cursorY = innerHeight - 1
	}

	view.SetOrigin(originX, originY)
	view.SetCursor(cursorX, cursorY)
}
