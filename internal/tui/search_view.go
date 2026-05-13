package tui

import (
	"fmt"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/theme"
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

func (program *Program) layoutPaneBottomOverlayView(gui *gocui.Gui, viewName string, parentViewName string) (*gocui.View, error) {
	x0, _, x1, y1, err := gui.ViewPosition(parentViewName)
	if err != nil {
		return nil, err
	}

	view, err := gui.SetView(viewName, x0, y1-1, x1, y1+1, 0)
	if err != nil && !isUnknownViewError(err) {
		return nil, err
	}

	return view, nil
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
	view.BgColor = gocuiColorOrDefault(theme.BackgroundHex)
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
	if key == gocui.KeyEnter || key == gocui.KeyCtrlJ || key == gocui.KeyEsc {
		return false
	}
	if program.searchEditor == nil {
		program.searchEditor = newLineEditor(program.model.SearchDraft())
	}
	if !program.searchEditor.HandleKey(key, ch, mod) {
		return false
	}

	program.updateActiveSearchDraft(program.searchEditor.Text())
	program.configureSearchView(view)
	program.renderSearchView(view)
	return true
}

func (program *Program) userViewTitle() string {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return reviewModeMetadataTitle
	}
	return "[1]-" + detailAuthorIcon + " Connected user"
}

func (program *Program) detailViewTitle() string {
	switch program.mainViewResolver().ContentKind {
	case MainContentKindReviewDescription:
		return reviewModeDescriptionTitle
	case MainContentKindStoryChapter:
		return reviewModeChapterTitle
	case MainContentKindReviewDiff:
		return reviewModeDiffTitle
	default:
		if program.shouldShowPullRequestDetailTabs() {
			return ""
		}
		return "[0]-Detail"
	}
}

func (program *Program) notificationsViewTitle() string {
	count, ok := program.notificationsCount()
	if !ok {
		return "Notifications"
	}
	return fmt.Sprintf("Notifications (%d)", count)
}

func (program *Program) notificationsCount() (int, bool) {
	rows := program.model.NotificationRows()
	if len(rows) == 0 {
		return 0, false
	}
	if len(rows) == 1 && rows[0].Notification == nil {
		item := rows[0].Item
		if program.isNotificationLoadingItem(item) || program.isNotificationErrorItem(item) {
			return 0, false
		}
		if item.Title == notificationsEmptyTitle && item.Detail == notificationsEmptyDetail {
			return 0, true
		}
	}

	count := 0
	for _, row := range rows {
		if row.Notification == nil {
			return 0, false
		}
		count++
	}
	return count, true
}

func (program *Program) pullRequestsViewTitle() string {
	switch program.modeDescriptor().Mode() {
	case ScreenModeStoryReview:
		return reviewModeChaptersTitle
	case ScreenModeReview:
		return reviewModeFilesTitle
	default:
		return ""
	}
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
