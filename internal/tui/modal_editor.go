package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	viewModalEditorName      = "modal-editor"
	modalEditorTotalHeight   = 7
	modalEditorFallbackWidth = 60
	modalEditorMinWidth      = 40
)

type modalEditorState struct {
	title        string
	editor       *multilineEditor
	errorMessage string
	submit       func(string) error
}

func newModalEditorState(title string, initialText string, submit func(string) error) *modalEditorState {
	if submit == nil {
		submit = func(string) error { return nil }
	}

	return &modalEditorState{
		title:  strings.TrimSpace(title),
		editor: newMultilineEditor(initialText),
		submit: submit,
	}
}

func (program *Program) modalEditorVisible() bool {
	return program != nil && program.modalEditor != nil
}

func (program *Program) openModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error) error {
	program.modalEditor = newModalEditorState(title, initialText, submit)
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) closeModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	program.modalEditor = nil
	if gui == nil {
		return nil
	}

	actualErr := gui.DeleteView(viewModalEditorName)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		return actualErr
	}

	return program.refreshViews(gui)
}

func (program *Program) submitModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	if program.modalEditor == nil {
		return nil
	}

	program.modalEditor.errorMessage = ""
	if err := program.modalEditor.submit(program.modalEditor.editor.Text()); err != nil {
		program.modalEditor.errorMessage = strings.TrimSpace(err.Error())
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}

	return program.closeModalEditor(gui, nil)
}

func (program *Program) layoutModalEditorView(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	totalWidth := maxX / 2
	if totalWidth < modalEditorMinWidth {
		totalWidth = modalEditorFallbackWidth
	}
	if totalWidth > maxX-4 {
		totalWidth = max(10, maxX-4)
	}
	if totalWidth > maxX {
		totalWidth = maxX
	}
	if totalWidth < 1 {
		totalWidth = 1
	}

	totalHeight := modalEditorTotalHeight
	if totalHeight > maxY {
		totalHeight = maxY
	}
	if totalHeight < 1 {
		totalHeight = 1
	}

	x0 := clampCoordinate((maxX-totalWidth)/2, maxX)
	y0 := clampCoordinate((maxY-totalHeight)/2, maxY)
	x1 := x0 + totalWidth - 1
	y1 := y0 + totalHeight - 1
	if x1 >= maxX {
		x1 = maxX - 1
		x0 = clampCoordinate(x1-totalWidth+1, maxX)
	}
	if y1 >= maxY {
		y1 = maxY - 1
		y0 = clampCoordinate(y1-totalHeight+1, maxY)
	}

	view, err := gui.SetView(viewModalEditorName, x0, y0, x1, y1, 0)
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
	view.Title = program.modalEditorTitle()
	view.Frame = true
	view.FrameRunes = roundFrameRunes
	view.FrameColor = gocui.GetColor(theme.ActiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.ActiveTextHex)
	view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	view.BgColor = gocui.ColorDefault
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
	text := program.modalEditor.editor.Text()
	fmt.Fprint(view, text)
	column, row := program.modalEditor.editor.CursorXY()
	program.setMultilineInputCursor(view, column, row)
}

func (program *Program) editModalEditor(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyAltEnter || key == gocui.KeyEsc || key == gocui.KeyCtrlLsqBracket {
		return false
	}
	if program.modalEditor == nil {
		return false
	}
	if !program.modalEditor.editor.HandleKey(key, ch, mod) {
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
