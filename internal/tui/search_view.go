package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	searchViewTotalHeight = 3
	searchViewMaxWidth    = 60
	searchViewMinWidth    = 32
)

func (program *Program) layoutSearchView(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	totalWidth := searchViewMaxWidth
	if totalWidth > maxX-4 {
		totalWidth = maxX - 4
	}
	if totalWidth < searchViewMinWidth {
		totalWidth = max(10, maxX)
	}
	if totalWidth > maxX {
		totalWidth = maxX
	}

	x0 := clampCoordinate((maxX-totalWidth)/2, maxX)
	y0 := clampCoordinate(maxY-searchViewTotalHeight, maxY)
	x1 := x0 + totalWidth - 1
	y1 := y0 + searchViewTotalHeight - 1
	if x1 >= maxX {
		x1 = maxX - 1
		x0 = clampCoordinate(x1-totalWidth+1, maxX)
	}
	if y1 >= maxY {
		y1 = maxY - 1
		y0 = clampCoordinate(y1-searchViewTotalHeight+1, maxY)
	}

	view, err := gui.SetView(viewSearchName, x0, y0, x1, y1, 0)
	if err != nil && !isUnknownViewError(err) {
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

func (program *Program) configureSearchView(view *gocui.View) {
	view.Title = program.searchViewTitle()
	view.Frame = true
	view.FrameRunes = roundFrameRunes
	view.FrameColor = gocui.GetColor(theme.ActiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.ActiveTextHex)
	view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	view.BgColor = gocui.ColorDefault
	view.Wrap = false
	view.Highlight = false
	view.Editable = true
	view.Editor = gocui.EditorFunc(program.editSearch)
}

func (program *Program) renderSearchView(view *gocui.View) {
	view.Clear()
	searchText := program.currentSearchText()
	fmt.Fprint(view, searchText)
	program.setInputCursor(view, searchText, program.currentSearchCursor())
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
	return "[1]-Connected user" + program.searchSummarySuffix(program.model.UserSearchQuery(), len(program.model.VisibleUsers()))
}

func (program *Program) detailViewTitle() string {
	suffix := program.searchSummarySuffix(program.model.DetailSearchQuery(), countSearchMatches(program.detailViewContent(), program.model.DetailSearchQuery()))
	if program.shouldShowPullRequestDetailTabs() {
		return suffix
	}
	return "[0]-Detail" + suffix
}

func (program *Program) pullRequestsViewTitle() string {
	query := program.model.PullRequestSearchQuery(program.model.ActivePullRequestTab())
	return program.searchSummarySuffix(query, len(program.model.VisiblePullRequests()))
}

func (program *Program) searchViewTitle() string {
	label := program.searchTargetLabel()
	query := strings.TrimSpace(program.model.SearchDraft())
	if query == "" {
		return fmt.Sprintf("Search %s", label)
	}

	count := program.searchMatchCount()
	return fmt.Sprintf("Search %s (%d %s)", label, count, pluralize(count, "match", "matches"))
}

func (program *Program) searchTargetLabel() string {
	switch program.model.SearchTarget() {
	case FocusPullRequestsView:
		if program.model.SearchTargetPullRequestTab() == RequestedPullRequestsTab {
			return "requested pull requests"
		}
		return "my pull requests"
	case FocusDetailView:
		return "detail"
	default:
		return "connected user"
	}
}

func (program *Program) searchMatchCount() int {
	switch program.model.SearchTarget() {
	case FocusPullRequestsView:
		return len(filterItemsByIndexes(program.model.PullRequests(program.model.SearchTargetPullRequestTab()), matchingItemIndexes(program.model.PullRequests(program.model.SearchTargetPullRequestTab()), program.model.SearchDraft())))
	case FocusDetailView:
		return countSearchMatches(program.detailViewContent(), program.model.SearchDraft())
	default:
		return len(filterItemsByIndexes(program.model.Users(), matchingItemIndexes(program.model.Users(), program.model.SearchDraft())))
	}
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
