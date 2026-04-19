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

	if program.model.SearchActive() {
		if err := program.layoutSearchView(gui); err != nil {
			return err
		}
	} else {
		if err := gui.DeleteView(viewSearchName); err != nil && !isUnknownViewError(err) {
			return err
		}
	}

	program.maybeLoadConnectedUser(gui)
	program.maybeLoadMyPullRequests(gui)
	program.maybeLoadRequestedPullRequests(gui)
	program.maybeLoadSelectedPullRequestDetail(gui)
	return program.syncCurrentView(gui)
}

func (program *Program) configureDetailView(view *gocui.View) {
	program.detailWrapWidth = effectiveMarkdownWidth(view.InnerWidth())
	program.applyViewStyle(view, FocusDetailView, program.detailViewTitle(), false)
	if program.shouldShowPullRequestDetailTabs() {
		view.TitlePrefix = "[0]"
		view.Tabs = program.detailTabLabels()
		view.TabIndex = int(program.activeDetailTab)
		view.SelFgColor = gocui.GetColor(theme.ActiveTextHex) | gocui.AttrBold
	}
	view.Wrap = true
}

func (program *Program) configureUserView(view *gocui.View) {
	program.applyViewStyle(view, FocusUserView, program.userViewTitle(), true)
	view.Wrap = false
}

func (program *Program) configurePullRequestsView(view *gocui.View) {
	program.applyViewStyle(view, FocusPullRequestsView, program.pullRequestsViewTitle(), true)
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

	visibleUsers := program.model.VisibleUsers()
	if len(visibleUsers) == 0 && strings.TrimSpace(program.model.UserSearchQuery()) != "" {
		fmt.Fprintln(view, searchNoMatchesMessage(program.model.UserSearchQuery()))
		return
	}

	for _, item := range visibleUsers {
		fmt.Fprintln(view, item.Title)
	}

	program.selectListLine(view, program.model.SelectedVisibleUserIndex())
}

func (program *Program) renderPullRequestsView(view *gocui.View) {
	view.Clear()

	query := program.model.PullRequestSearchQuery(program.model.ActivePullRequestTab())
	visiblePullRequests := program.model.VisiblePullRequests()
	if len(visiblePullRequests) == 0 && strings.TrimSpace(query) != "" {
		fmt.Fprintln(view, searchNoMatchesMessage(query))
		return
	}

	for _, item := range visiblePullRequests {
		fmt.Fprintln(view, item.Title)
	}

	program.selectListLine(view, program.model.SelectedVisiblePullRequestIndex(program.model.ActivePullRequestTab()))
}

func (program *Program) renderDetailView(view *gocui.View) {
	program.detailWrapWidth = effectiveMarkdownWidth(view.InnerWidth())
	program.resetDetailViewOriginIfNeeded(view)
	view.Clear()

	detailContent := program.detailViewContent()
	highlightedContent, _ := highlightSearchMatches(detailContent, program.model.DetailSearchQuery())
	fmt.Fprint(view, highlightedContent)
}

func (program *Program) detailViewContent() string {
	if program.model.currentSideFocus() == FocusPullRequestsView {
		row, ok := program.model.SelectedPullRequestRow()
		if ok && row.Summary != nil && pullRequestDetailKey(row.Summary.Repository, row.Summary.Number) != "" {
			if result, ok := program.pullRequestDetailForSummary(*row.Summary); ok {
				if result.err != nil {
					return renderPullRequestDetailError(*row.Summary, result.err)
				}

				header := renderPullRequestDetailHeader(*row.Summary, result.detail)
				content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)
				if program.activeDetailTab == CommentsDetailTab {
					content = renderPullRequestCommentsTab(result.detail.Comments, program.markdownRenderer, program.detailWrapWidth)
				}
				return fmt.Sprintf("%s\n\n%s", header, content)
			}
			return renderPullRequestDetailLoading(*row.Summary)
		}
	}

	item, ok := program.model.detailItem()
	if !ok {
		return "No detail available."
	}

	return program.fallbackDetailViewContent(item)
}

func (program *Program) fallbackDetailViewContent(item Item) string {
	header := program.detailHeader(item)
	body := strings.TrimSpace(item.Detail)
	if body == "" {
		body = "No description available. Even the dummy data is disappointed."
	}

	return fmt.Sprintf("%s\n\n%s", header, body)
}

func (program *Program) detailHeader(item Item) string {
	source := "Connected user"
	if program.model.currentSideFocus() == FocusPullRequestsView {
		source = fmt.Sprintf("%s tab", program.model.ActivePullRequestTab().Label())
	}

	return fmt.Sprintf("%s\n%s", source, item.Title)
}

func (program *Program) resetDetailViewOriginIfNeeded(view *gocui.View) {
	if view == nil {
		return
	}

	identity := program.currentDetailIdentity()
	if identity == program.lastDetailIdentity {
		return
	}

	program.lastDetailIdentity = identity
	view.SetOrigin(0, 0)
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
	if view == nil || len(view.BufferLines()) == 0 {
		return
	}

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

func searchNoMatchesMessage(query string) string {
	return fmt.Sprintf("No matches for %q.", strings.TrimSpace(query))
}
