package tui

import "github.com/jesseduffield/gocui"

const bottomPromptPrefix = "/"

func (program *Program) layoutBottomPromptView(gui *gocui.Gui, viewName string) (*gocui.View, error) {
	maxX, maxY := gui.Size()
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}

	x0 := -1
	y0 := maxY - 2
	x1 := maxX
	y1 := maxY
	view, err := gui.SetView(viewName, x0, y0, x1, y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return nil, err
	}

	return view, nil
}

func (program *Program) configureSearchView(view *gocui.View) {
	program.configureBottomPromptView(view, gocui.EditorFunc(program.editSearch), true)
}

func (program *Program) renderSearchView(view *gocui.View) {
	presenter := program.searchViewPresenter()
	program.renderBottomPromptView(view, presenter.promptText(), presenter.promptCursor())
}

func (program *Program) editSearch(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyEnter || key == gocui.KeyCtrlJ || key == gocui.KeyEsc {
		return false
	}
	if !program.searchWidget.hasEditor() {
		program.searchWidget.openEditor(program.model.SearchDraft())
	}
	if !program.searchWidget.editor.HandleKey(key, ch, mod) {
		return false
	}

	return program.dispatchEditorMessage(MsgSearchDraftChanged{Query: program.searchWidget.editor.Text()})
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}
