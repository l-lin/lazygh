package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	sidebarWidthPercent = 35
	minimumSidebarWidth = 32
	minimumDetailWidth  = 40
	userViewTotalHeight = 3
)

var roundFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

func (program *Program) layout(gui *gocui.Gui) error {
	program.gui = gui
	maxX, maxY := gui.Size()
	contentMaxY := program.layoutContentHeight(maxY)

	program.maybeLoadConnectedUser(gui)
	program.maybeLoadMyPullRequests(gui)
	program.maybeLoadRequestedPullRequests(gui)
	program.maybeLoadSelectedPullRequestDetail(gui)
	program.maybeLoadSelectedPullRequestDiff(gui)

	mainPaneLayout := calculateMainPaneLayout(maxX, contentMaxY, program.model.PaneLayoutSize(), program.model.FullscreenPane())

	userView, err := setPaneView(gui, viewUserName, mainPaneLayout.userVisible, mainPaneLayout.user)
	if err != nil {
		return err
	}
	if userView != nil {
		program.configureUserView(userView)
		program.renderUserView(userView)
	}

	pullRequestsView, err := setPaneView(gui, viewPullRequestsName, mainPaneLayout.pullRequestsVisible, mainPaneLayout.pullRequests)
	if err != nil {
		return err
	}
	if pullRequestsView != nil {
		program.configurePullRequestsView(pullRequestsView)
		program.renderPullRequestsView(pullRequestsView)
	}

	detailView, err := setPaneView(gui, viewDetailName, mainPaneLayout.detailVisible, mainPaneLayout.detail)
	if err != nil {
		return err
	}
	if detailView != nil {
		program.configureDetailView(detailView)
		program.renderDetailView(detailView)
	}

	if err := program.layoutPaneFooterViews(gui); err != nil {
		return err
	}

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

	if program.modalEditorVisible() {
		if err := program.layoutModalEditorView(gui); err != nil {
			return err
		}
	} else {
		if err := gui.DeleteView(viewModalEditorName); err != nil && !isUnknownViewError(err) {
			return err
		}
	}

	if program.model.ActionsPopupVisible() {
		if err := program.layoutActionsPopupViews(gui); err != nil {
			return err
		}
		if program.model.ActionsPopupSearchActive() {
			if err := program.layoutActionsPopupSearchView(gui); err != nil {
				return err
			}
		} else {
			if err := gui.DeleteView(viewActionsPopupSearchName); err != nil && !isUnknownViewError(err) {
				return err
			}
		}
	} else {
		for _, viewName := range []string{viewActionsPopupSearchName, viewActionsPopupName} {
			if err := gui.DeleteView(viewName); err != nil && !isUnknownViewError(err) {
				return err
			}
		}
	}

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
	view.Wrap = false
	view.Editable = false
	view.Editor = nil
}

func (program *Program) configureUserView(view *gocui.View) {
	if program.reviewSession.active {
		program.applyViewStyle(view, FocusUserView, program.userViewTitle(), false)
		view.Wrap = false
		view.Editable = false
		view.Editor = nil
		return
	}

	program.configureSelectableListView(view, FocusUserView, program.userViewTitle(), program.model.UserSearchQuery())
}

func (program *Program) configurePullRequestsView(view *gocui.View) {
	if program.reviewSession.active {
		program.configureSelectableListView(view, FocusPullRequestsView, program.pullRequestsViewTitle(), "")
		return
	}

	program.configureSelectableListView(view, FocusPullRequestsView, program.pullRequestsViewTitle(), program.model.PullRequestSearchQuery(program.model.ActivePullRequestTab()))
	view.TitlePrefix = "[2]"
	view.Tabs = program.pullRequestsTabLabels()
	view.TabIndex = int(program.model.ActivePullRequestTab())
	view.SelFgColor = gocui.GetColor(theme.ActiveTextHex) | gocui.AttrBold
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
	if program.reviewSession.active {
		renderReadOnlyTextView(view, program.reviewSessionMetadataContent())
		return
	}

	program.renderSelectableListView(view, selectableListViewState{
		focus:               FocusUserView,
		query:               program.model.UserSearchQuery(),
		items:               program.model.VisibleUsers(),
		selectedVisibleLine: program.model.SelectedVisibleUserIndex(),
	})
}

func (program *Program) renderPullRequestsView(view *gocui.View) {
	if program.reviewSession.active {
		program.renderSelectableListView(view, selectableListViewState{
			focus:               FocusPullRequestsView,
			query:               "",
			items:               program.reviewSessionFiles(),
			selectedVisibleLine: program.reviewSessionSelectedVisibleLine(),
		})
		return
	}

	program.renderSelectableListView(view, selectableListViewState{
		focus:               FocusPullRequestsView,
		query:               program.model.PullRequestSearchQuery(program.model.ActivePullRequestTab()),
		items:               program.model.VisiblePullRequests(),
		selectedVisibleLine: program.model.SelectedVisiblePullRequestIndex(program.model.ActivePullRequestTab()),
	})
}

func (program *Program) renderDetailView(view *gocui.View) {
	program.detailWrapWidth = effectiveMarkdownWidth(view.InnerWidth())
	detailDocument := program.currentDetailDocument(view)
	program.syncDetailViewState(detailDocument, view.InnerHeight())
	view.Clear()

	searchMatchRanges := detailSearchMatchRanges(program.detailViewState.searchMatches)
	for rowIndex, row := range detailDocument.rows {
		if rowIndex > 0 {
			fmt.Fprint(view, "\n")
		}
		fmt.Fprint(view, renderDetailRow(detailDocument, row, searchMatchRanges, program.detailViewState))
	}

	cursorRow, cursorColumn := program.detailViewState.screenPosition(detailDocument)
	view.SetOrigin(0, program.detailViewState.originRow)
	view.SetCursor(cursorColumn, cursorRow-program.detailViewState.originRow)
}

func (program *Program) detailViewContent() string {
	if program.reviewSession.active {
		return program.reviewSessionDetailContent()
	}
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
					content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)
				}
				return renderPullRequestDetailContent(header, content)
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

	return renderPullRequestDetailContent(header, body)
}

func (program *Program) detailHeader(item Item) string {
	source := "Connected user"
	if program.model.currentSideFocus() == FocusPullRequestsView {
		source = fmt.Sprintf("%s tab", program.model.ActivePullRequestTab().Label())
	}

	return fmt.Sprintf("%s\n%s", source, item.Title)
}

func (program *Program) currentDetailDocument(view *gocui.View) detailDocument {
	width := program.detailWrapWidth
	if view != nil && view.InnerWidth() > 0 {
		width = view.InnerWidth()
	}
	if width < 1 {
		width = 1
	}

	return newDetailDocument(program.detailViewContent(), width)
}

func (program *Program) syncDetailViewState(detailDocument detailDocument, viewportHeight int) {
	identity := program.currentDetailIdentity()
	if identity != program.lastDetailIdentity {
		program.lastDetailIdentity = identity
		program.detailViewState.reset()
	}

	program.detailViewState.sync(detailDocument, viewportHeight)
	program.detailViewState.syncSearch(detailDocument, program.model.DetailSearchQuery())
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

func (program *Program) layoutContentHeight(maxY int) int {
	if maxY < 1 {
		return 1
	}
	if program.bottomPromptVisible() && maxY > 1 {
		return maxY - 1
	}
	return maxY
}

func (program *Program) bottomPromptVisible() bool {
	return program.model.ActionsPopupSearchActive()
}
