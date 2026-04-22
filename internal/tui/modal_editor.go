package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	viewModalEditorName        = "modal-editor"
	modalEditorTotalHeight     = 7
	lineModalEditorTotalHeight = 3
	modalEditorFallbackWidth   = 60
	modalEditorMinWidth        = 40
)

type modalEditorKeyHandler func(*Program, *gocui.View, gocui.Key, rune, gocui.Modifier) bool

type modalEditorState struct {
	title        string
	editor       *multilineEditor
	lineEditor   *lineEditor
	errorMessage string
	submit       func(string) error
	afterSubmit  func(*gocui.Gui)
	totalHeight  int
	handleKey    modalEditorKeyHandler
}

func newModalEditorState(title string, initialText string, submit func(string) error) *modalEditorState {
	return newModalEditorStateWithOptions(title, initialText, submit, modalEditorTotalHeight, false, nil)
}

func newModalEditorStateWithKeyHandler(title string, initialText string, submit func(string) error, handleKey modalEditorKeyHandler) *modalEditorState {
	return newModalEditorStateWithOptions(title, initialText, submit, modalEditorTotalHeight, false, handleKey)
}

func newLineModalEditorState(title string, initialText string, submit func(string) error) *modalEditorState {
	return newModalEditorStateWithOptions(title, initialText, submit, lineModalEditorTotalHeight, true, nil)
}

func newModalEditorStateWithOptions(title string, initialText string, submit func(string) error, totalHeight int, singleLine bool, handleKey modalEditorKeyHandler) *modalEditorState {
	if submit == nil {
		submit = func(string) error { return nil }
	}

	state := &modalEditorState{
		title:       strings.TrimSpace(title),
		submit:      submit,
		totalHeight: totalHeight,
		handleKey:   handleKey,
	}
	if singleLine {
		state.lineEditor = newLineEditor(initialText)
		return state
	}

	state.editor = newMultilineEditor(initialText)
	return state
}

func (state *modalEditorState) Text() string {
	if state == nil {
		return ""
	}
	if state.lineEditor != nil {
		return state.lineEditor.Text()
	}
	if state.editor != nil {
		return state.editor.Text()
	}
	return ""
}

func (state *modalEditorState) CursorXY() (int, int) {
	if state == nil {
		return 0, 0
	}
	if state.lineEditor != nil {
		return state.lineEditor.Cursor(), 0
	}
	if state.editor != nil {
		return state.editor.CursorXY()
	}
	return 0, 0
}

func (state *modalEditorState) HandleKey(key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if state == nil {
		return false
	}
	if state.lineEditor != nil {
		return state.lineEditor.HandleKey(key, ch, mod)
	}
	if state.editor != nil {
		return state.editor.HandleKey(key, ch, mod)
	}
	return false
}

func (state *modalEditorState) Height() int {
	if state == nil || state.totalHeight < 1 {
		return modalEditorTotalHeight
	}
	return state.totalHeight
}

func (program *Program) modalEditorVisible() bool {
	return program != nil && program.modalEditor != nil
}

func (program *Program) openModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error) error {
	return program.openMultilineModalEditor(gui, title, initialText, submit, modalEditorTotalHeight, nil)
}

func (program *Program) openModalEditorWithKeyHandler(gui *gocui.Gui, title string, initialText string, submit func(string) error, handleKey modalEditorKeyHandler) error {
	return program.openMultilineModalEditor(gui, title, initialText, submit, modalEditorTotalHeight, handleKey)
}

func (program *Program) openMultilineModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error, totalHeight int, handleKey modalEditorKeyHandler) error {
	program.modalEditor = newModalEditorStateWithOptions(title, initialText, submit, totalHeight, false, handleKey)
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) openLineModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error) error {
	program.modalEditor = newLineModalEditorState(title, initialText, submit)
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) closeModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	program.modalEditor = nil
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) submitModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	if program.modalEditor == nil {
		return nil
	}

	program.modalEditor.errorMessage = ""
	if err := program.modalEditor.submit(program.modalEditor.Text()); err != nil {
		program.modalEditor.errorMessage = strings.TrimSpace(err.Error())
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}
	if program.modalEditor.afterSubmit != nil {
		program.modalEditor.afterSubmit(gui)
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
	if program.modalEditor != nil {
		totalHeight = program.modalEditor.Height()
	}
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
