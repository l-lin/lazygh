package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	sidebarWidthPercent = 46
	minimumSidebarWidth = 32
	minimumDetailWidth  = 40
	userViewTotalHeight = 3
)

var roundFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

func (program *Program) layout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()

	sidebarWidth := maxX * sidebarWidthPercent / 100
	maxSidebarWidth := maxX - minimumDetailWidth
	if maxSidebarWidth < minimumSidebarWidth {
		sidebarWidth = maxX / 2
	} else {
		if sidebarWidth < minimumSidebarWidth {
			sidebarWidth = minimumSidebarWidth
		}
		if sidebarWidth > maxSidebarWidth {
			sidebarWidth = maxSidebarWidth
		}
	}
	if sidebarWidth < 1 {
		sidebarWidth = 1
	}

	sidebarX1 := sidebarWidth - 1
	detailX0 := sidebarX1 + 1
	if detailX0 >= maxX {
		detailX0 = maxX / 2
		sidebarX1 = detailX0 - 1
	}

	userHeight := userViewTotalHeight
	if userHeight >= maxY {
		userHeight = maxY / 2
	}
	if userHeight < 2 {
		userHeight = 2
	}
	userY1 := userHeight - 1
	pullRequestsY0 := userY1 + 1
	if pullRequestsY0 >= maxY {
		pullRequestsY0 = maxY / 2
		userY1 = pullRequestsY0 - 1
	}

	userView, err := gui.SetView(viewUserName, 0, 0, sidebarX1, userY1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureUserView(userView)
	program.renderUserView(userView)

	pullRequestsView, err := gui.SetView(viewPullRequestsName, 0, pullRequestsY0, sidebarX1, maxY-1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configurePullRequestsView(pullRequestsView)
	program.renderPullRequestsView(pullRequestsView)

	detailView, err := gui.SetView(viewDetailName, detailX0, 0, maxX-1, maxY-1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}
	program.configureDetailView(detailView)
	program.renderDetailView(detailView)

	if program.helpVisible {
		if err := program.layoutHelpView(gui); err != nil {
			return err
		}
	} else {
		if err := gui.DeleteView(viewHelpName); err != nil && !isUnknownViewError(err) {
			return err
		}
	}

	program.maybeLoadConnectedUser(gui)
	program.maybeLoadMyPullRequests(gui)
	program.maybeLoadRequestedPullRequests(gui)
	return program.syncCurrentView(gui)
}

func (program *Program) configureDetailView(view *gocui.View) {
	program.applyViewStyle(view, FocusDetailView, "[0]-Detail", false)
	view.Wrap = true
}

func (program *Program) configureUserView(view *gocui.View) {
	program.applyViewStyle(view, FocusUserView, "[1]-Connected user", true)
	view.Wrap = false
}

func (program *Program) configurePullRequestsView(view *gocui.View) {
	program.applyViewStyle(view, FocusPullRequestsView, "", true)
	view.TitlePrefix = "[2]"
	view.Tabs = program.pullRequestsTabLabels()
	view.TabIndex = int(program.model.ActivePullRequestTab())
	view.SelFgColor = gocui.GetColor(theme.ActiveTextHex) | gocui.AttrBold
	view.Wrap = false
}

func (program *Program) applyViewStyle(view *gocui.View, focus Focus, title string, selectable bool) {
	isUnderlyingFocus := program.model.Focus() == focus
	showsSelection := program.shouldHighlightSelection(focus, selectable)

	view.Title = title
	view.TitlePrefix = ""
	view.Tabs = nil
	view.TabIndex = 0
	view.Frame = true
	view.FrameRunes = roundFrameRunes
	view.Highlight = showsSelection
	view.HighlightInactive = showsSelection && !isUnderlyingFocus
	view.FrameColor = gocui.GetColor(theme.InactiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.InactiveTitleHex)
	view.SelBgColor = gocui.GetColor(theme.SelectedLineBackgroundHex)
	view.SelFgColor = gocui.GetColor(theme.ActiveTextHex)
	view.InactiveViewSelBgColor = gocui.GetColor(theme.SelectedLineBackgroundHex)

	if isUnderlyingFocus {
		view.FgColor = gocui.GetColor(theme.ActiveTextHex)
		return
	}

	view.FgColor = gocui.GetColor(theme.InactiveTextHex)
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

	for _, item := range program.model.CurrentPullRequests() {
		fmt.Fprintln(view, item.Title)
	}

	program.selectListLine(view, program.model.SelectedPullRequestIndex(program.model.ActivePullRequestTab()))
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
}

func (program *Program) detailHeader(item Item) string {
	source := "Connected user"
	if program.model.currentSideFocus() == FocusPullRequestsView {
		source = fmt.Sprintf("%s tab", program.model.ActivePullRequestTab().Label())
	}

	return fmt.Sprintf("%s\n%s", source, item.Title)
}

func (program *Program) pullRequestsTabLabels() []string {
	return []string{
		program.pullRequestsTabLabel(MyPullRequestsTab),
		program.pullRequestsTabLabel(RequestedPullRequestsTab),
	}
}

func (program *Program) pullRequestsTabLabel(tab PullRequestTab) string {
	label := tab.Label()
	count, ok := program.pullRequestsCount(tab)
	if !ok {
		return label
	}

	return fmt.Sprintf("%s (%d)", label, count)
}

func (program *Program) pullRequestsCount(tab PullRequestTab) (int, bool) {
	switch tab {
	case RequestedPullRequestsTab:
		return program.requestedPullRequestsCount, program.requestedPullRequestsCountKnown
	default:
		return program.myPullRequestsCount, program.myPullRequestsCountKnown
	}
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

func (program *Program) shouldHighlightSelection(focus Focus, selectable bool) bool {
	if !selectable {
		return false
	}

	if program.model.Focus() == focus {
		return true
	}

	return program.model.Focus() == FocusDetailView && program.model.currentSideFocus() == focus
}
