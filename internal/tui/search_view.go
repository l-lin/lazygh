package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const bottomPromptPrefix = "/"

func (program *Program) layoutSearchView(gui *gocui.Gui) error {
	view, err := program.layoutBottomPromptView(gui, viewSearchName)
	if err != nil {
		return err
	}

	program.configureSearchView(view)
	program.renderSearchView(view)
	_, err = gui.SetViewOnTop(viewSearchName)
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

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

func (program *Program) configureBottomPromptView(view *gocui.View, editor gocui.Editor, editable bool) {
	view.Title = ""
	view.Frame = false
	view.FrameRunes = nil
	view.FrameColor = gocui.GetColor(theme.ActiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.ActiveTextHex)
	view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	view.BgColor = gocui.ColorDefault
	view.Wrap = false
	view.Highlight = false
	view.Editable = editable
	view.Editor = editor
}

func (program *Program) renderSearchView(view *gocui.View) {
	program.renderBottomPromptView(view, program.currentSearchText(), program.currentSearchCursor())
}

func (program *Program) renderBottomPromptView(view *gocui.View, text string, cursorIndex int) {
	if view == nil {
		return
	}

	view.Clear()
	prompt := bottomPromptPrefix + text
	fmt.Fprint(view, prompt)
	program.setInputCursor(view, prompt, cursorIndex+utf8.RuneCountInString(bottomPromptPrefix))
}

func (program *Program) editSearch(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyEnter || key == gocui.KeyCtrlJ || key == gocui.KeyEsc || key == gocui.KeyCtrlLsqBracket {
		return false
	}
	if program.searchEditor == nil {
		program.searchEditor = newLineEditor(program.model.SearchDraft())
	}
	if !program.searchEditor.HandleKey(key, ch, mod) {
		return false
	}

	program.model.UpdateSearchDraft(program.searchEditor.Text())
	program.configureSearchView(view)
	program.renderSearchView(view)
	return true
}

func (program *Program) userViewTitle() string {
	return "[1]-Connected user" + program.feedbackSuffix(FocusUserView) + program.searchSummarySuffix(program.model.UserSearchQuery(), len(program.model.VisibleUsers()))
}

func (program *Program) detailViewTitle() string {
	suffix := program.feedbackSuffix(FocusDetailView) + program.searchSummarySuffix(program.model.DetailSearchQuery(), countSearchMatches(program.detailViewContent(), program.model.DetailSearchQuery()))
	if program.shouldShowPullRequestDetailTabs() {
		return suffix
	}
	return "[0]-Detail" + suffix
}

func (program *Program) pullRequestsViewTitle() string {
	query := program.model.PullRequestSearchQuery(program.model.ActivePullRequestTab())
	return program.feedbackSuffix(FocusPullRequestsView) + program.searchSummarySuffix(query, len(program.model.VisiblePullRequests()))
}

func (program *Program) searchSummarySuffix(query string, count int) string {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return ""
	}

	return fmt.Sprintf(" / %q (%d %s)", trimmedQuery, count, pluralize(count, "match", "matches"))
}

func (program *Program) setInputCursor(view *gocui.View, value string, cursorIndex int) {
	if view == nil {
		return
	}

	innerWidth := view.InnerWidth()
	if innerWidth < 1 {
		innerWidth = 1
	}

	valueWidth := utf8.RuneCountInString(value)
	if cursorIndex < 0 {
		cursorIndex = 0
	}
	if cursorIndex > valueWidth {
		cursorIndex = valueWidth
	}

	originX := 0
	if cursorIndex >= innerWidth {
		originX = cursorIndex - innerWidth + 1
	}
	cursorX := cursorIndex - originX
	if cursorX < 0 {
		cursorX = 0
	}
	if cursorX >= innerWidth {
		cursorX = innerWidth - 1
	}

	view.SetOrigin(originX, 0)
	view.SetCursor(cursorX, 0)
}

func (program *Program) currentSearchText() string {
	if program.searchEditor != nil {
		return program.searchEditor.Text()
	}

	return program.model.SearchDraft()
}

func (program *Program) currentSearchCursor() int {
	if program.searchEditor != nil {
		return program.searchEditor.Cursor()
	}

	return utf8.RuneCountInString(program.model.SearchDraft())
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}
