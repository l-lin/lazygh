package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	detailWidthRatio      = 2
	sideWidthRatio        = 1
	minimumSidebarWidth   = 30
	minimumUserViewHeight = 7
	detailKeyHint         = "tab/l next panel · shift+tab/h previous panel · j/k move · enter detail · esc back · ctrl+c quit"
)

func (program *Program) layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()

	detailWidth := maxX * detailWidthRatio / (detailWidthRatio + sideWidthRatio)
	maxDetailWidth := maxX - minimumSidebarWidth
	if maxDetailWidth > 0 && detailWidth > maxDetailWidth {
		detailWidth = maxDetailWidth
	}
	if detailWidth < minimumSidebarWidth {
		detailWidth = maxX / 2
	}
	if detailWidth < 1 {
		detailWidth = 1
	}

	detailX1 := detailWidth - 1
	sideX0 := detailX1 + 1
	if sideX0 >= maxX {
		sideX0 = maxX / 2
		detailX1 = sideX0 - 1
	}

	userHeight := maxY / 4
	if userHeight < minimumUserViewHeight {
		userHeight = minimumUserViewHeight
	}
	if userHeight >= maxY {
		userHeight = maxY / 2
	}
	userY1 := userHeight - 1
	if userY1 < 1 {
		userY1 = 1
	}
	pullRequestsY0 := userY1 + 1
	if pullRequestsY0 >= maxY {
		pullRequestsY0 = maxY / 2
		userY1 = pullRequestsY0 - 1
	}

	detailView, err := gui.SetView(viewDetailName, 0, 0, detailX1, maxY-1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureDetailView(detailView)
	program.renderDetailView(detailView)

	userView, err := gui.SetView(viewUserName, sideX0, 0, maxX-1, userY1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureListView(userView, "1 · Connected user")
	program.renderUserView(userView)

	pullRequestsView, err := gui.SetView(viewPullRequestsName, sideX0, pullRequestsY0, maxX-1, maxY-1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureListView(pullRequestsView, "2 · Pull requests")
	pullRequestsView.Tabs = nil
	pullRequestsView.TabIndex = 0
	program.renderPullRequestsView(pullRequestsView)

	return program.syncCurrentView(gui)
}

func (program *Program) configureDetailView(view *gocui.View) {
	view.Title = "0 · Detail"
	view.Wrap = true
	view.Frame = true
	view.Highlight = false
	view.FrameColor = gocui.GetColor(theme.InactiveBorderHex)
}

func (program *Program) configureListView(view *gocui.View, title string) {
	view.Title = title
	view.Wrap = false
	view.Frame = true
	view.Highlight = true
	view.HighlightInactive = true
	view.FrameColor = gocui.GetColor(theme.InactiveBorderHex)
	view.InactiveViewSelBgColor = gocui.GetColor(theme.InactiveSelectionBackgroundHex)
}

func (program *Program) renderUserView(view *gocui.View) {
	view.Clear()

	for _, item := range program.model.Users() {
		fmt.Fprintln(view, item.Title)
	}

	program.selectListLine(view, program.model.SelectedUserIndex())
}

func (program *Program) renderPullRequestsView(view *gocui.View) {
	view.Clear()
	fmt.Fprintln(view, program.pullRequestTabsLine())

	for _, item := range program.model.CurrentPullRequests() {
		fmt.Fprintln(view, item.Title)
	}

	selectedIndex := program.model.SelectedPullRequestIndex(program.model.ActivePullRequestTab()) + 1
	program.selectListLine(view, selectedIndex)
}

func (program *Program) renderDetailView(view *gocui.View) {
	view.Clear()

	item, ok := program.model.detailItem()
	if !ok {
		fmt.Fprintln(view, "No detail available.")
		return
	}

	header := program.detailHeader(item)
	body := strings.TrimSpace(item.Detail)
	if body == "" {
		body = "No description available. Even the dummy data is disappointed."
	}

	fmt.Fprintln(view, header)
	fmt.Fprintln(view)
	fmt.Fprintln(view, body)
	fmt.Fprintln(view)
	fmt.Fprintln(view, detailKeyHint)
}

func (program *Program) detailHeader(item Item) string {
	source := "Connected user"
	if program.model.currentSideFocus() == FocusPullRequestsView {
		source = fmt.Sprintf("%s tab", program.model.ActivePullRequestTab().Label())
	}

	return fmt.Sprintf("%s\n%s", source, item.Title)
}

func (program *Program) pullRequestTabsLine() string {
	myPullRequestsLabel := MyPullRequestsTab.Label()
	requestedLabel := RequestedPullRequestsTab.Label()

	if program.model.ActivePullRequestTab() == MyPullRequestsTab {
		return fmt.Sprintf("[%s]  %s", myPullRequestsLabel, requestedLabel)
	}

	return fmt.Sprintf("%s  [%s]", myPullRequestsLabel, requestedLabel)
}

func (program *Program) selectListLine(view *gocui.View, selectedIndex int) {
	visibleHeight := view.InnerHeight()
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	originY := 0
	cursorY := selectedIndex
	if cursorY >= visibleHeight {
		originY = cursorY - visibleHeight + 1
		cursorY = visibleHeight - 1
	}
	if cursorY < 0 {
		cursorY = 0
	}

	view.SetOrigin(0, originY)
	view.SetCursor(0, cursorY)
}
