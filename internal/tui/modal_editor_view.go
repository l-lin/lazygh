package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) configureModalEditorView(view *gocui.View) {
	configureFramedOverlayView(view, program.modalEditorTitle(), "")
	view.Wrap = program.modalEditor != nil && program.modalEditor.lineEditor == nil
	view.Highlight = false
	view.Editable = true
	view.Editor = gocui.EditorFunc(program.editModalEditor)
}

func (program *Program) renderModalEditorView(view *gocui.View) {
	if view == nil || program.modalEditor == nil {
		return
	}

	view.Clear()
	text := program.modalEditor.Text()
	fmt.Fprint(view, text)
	if view.Wrap {
		program.setWrappedMultilineInputCursor(view, text, program.modalEditor.Cursor())
		return
	}
	column, row := program.modalEditor.CursorXY()
	program.setMultilineInputCursor(view, column, row)
}

func (program *Program) editModalEditor(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyAltEnter || key == gocui.KeyEsc {
		return false
	}
	if program.modalEditor == nil {
		return false
	}
	if key == gocui.KeyEnter && mod == gocui.ModNone && program.modalEditor.submitOnEnter {
		_ = program.submitModalEditor(program.gui, nil)
		return true
	}
	if !program.modalEditor.HandleKey(key, ch, mod) {
		return false
	}

	program.modalEditor.errorMessage = ""
	program.configureModalEditorView(view)
	program.renderModalEditorView(view)
	return true
}

func (program *Program) modalEditorTitle() string {
	if program.modalEditor == nil {
		return ""
	}

	title := strings.TrimSpace(program.modalEditor.title)
	message := strings.TrimSpace(program.modalEditor.errorMessage)
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
