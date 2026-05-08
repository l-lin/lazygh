package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) layoutModalEditorView(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	totalWidth := boundedHalfWidth(maxX, modalEditorMinWidth, modalEditorFallbackWidth)
	totalHeight := modalEditorTotalHeight
	if program.modalEditor != nil {
		totalHeight = program.modalEditor.Height()
	}
	frame := centeredOverlayFrame(maxX, maxY, totalWidth, totalHeight)

	view, err := gui.SetView(viewModalEditorName, frame.x0, frame.y0, frame.x1, frame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}

	program.configureModalEditorView(view)
	program.renderModalEditorView(view)
	_, err = gui.SetViewOnTop(viewModalEditorName)
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

func (program *Program) configureModalEditorView(view *gocui.View) {
	configureFramedOverlayView(view, program.modalEditorTitle(), "")
	view.Wrap = false
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
	column, row := program.modalEditor.CursorXY()
	program.setMultilineInputCursor(view, column, row)
}

func (program *Program) editModalEditor(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyAltEnter || key == gocui.KeyEsc || key == gocui.KeyCtrlLsqBracket {
		return false
	}
	if program.modalEditor == nil {
		return false
	}
	if program.modalEditor.handleKey != nil && program.modalEditor.handleKey(program, view, key, ch, mod) {
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

	innerWidth := view.InnerWidth()
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := view.InnerHeight()
	if innerHeight < 1 {
		innerHeight = 1
	}

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
