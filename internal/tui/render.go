package tui

import (
	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/theme"
)

const (
	sidebarWidthPercent          = 35
	minimumSidebarWidth          = 32
	minimumDetailWidth           = 40
	userViewTotalHeight          = 3
	reviewModeMetadataViewHeight = 3
)

var roundFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

func (program *Program) layout(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	program.gui = gui
	if !program.appStarted {
		return program.dispatch(gui, MsgAppStarted{})
	}

	maxX, maxY := gui.Size()
	if actualErr := program.reloadRegisteredKeybindings(gui); actualErr != nil {
		return actualErr
	}

	return program.applyScreenComposition(gui, program.screenCompositionForSize(maxX, maxY))
}

func (program *Program) sidebarTopPaneHeight() int {
	if program.reviewModeActive() {
		return reviewModeMetadataViewHeight
	}
	return userViewTotalHeight
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
	if program.reviewModeActive() {
		program.applyViewStyle(view, FocusUserView, program.userViewTitle(), false)
		view.Wrap = false
		view.Editable = false
		view.Editor = nil
		return
	}

	program.configureSelectableListView(view, FocusUserView, program.userViewTitle(), program.model.UserSearchQuery())
}

func (program *Program) configurePullRequestsView(view *gocui.View) {
	if program.reviewModeActive() {
		program.configureSelectableListView(view, FocusPullRequestsView, program.pullRequestsViewTitle(), program.reviewFileTreeSearchQuery())
		return
	}

	program.configureSelectableListView(view, FocusPullRequestsView, program.pullRequestsViewTitle(), program.model.PullRequestSearchQuery(program.model.ActivePullRequestTab()))
	view.TitlePrefix = "[2]"
	view.Tabs = program.pullRequestsTabLabels()
	view.TabIndex = int(program.model.ActivePullRequestTab())
	view.SelFgColor = gocui.GetColor(theme.ActiveTextHex) | gocui.AttrBold
}

func (program *Program) configureNotificationsView(view *gocui.View) {
	program.configureSelectableListView(view, FocusNotificationsView, program.notificationsViewTitle(), program.model.NotificationSearchQuery())
	view.TitlePrefix = "[3]"
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
	view.BgColor = gocuiColorOrDefault(theme.BackgroundHex)
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
	if program.reviewModeActive() {
		renderReadOnlyTextView(view, program.reviewSessionMetadataContent())
		return
	}

	program.renderSelectableListView(view, selectableListViewState{
		focus:               FocusUserView,
		query:               program.model.UserSearchQuery(),
		items:               program.model.Users(),
		selectedVisibleLine: program.model.SelectedUserIndex(),
	})
}

func (program *Program) renderPullRequestsView(view *gocui.View) {
	if program.reviewModeActive() {
		query := program.reviewFileTreeSearchQuery()
		if tree, files, ok := program.reviewSessionCurrentTree(); ok && len(tree.Rows) > 0 {
			program.renderReviewDiffTreeView(view, tree, files, query, program.reviewSessionSelectedVisibleLine())
			return
		}
		program.renderSelectableListView(view, selectableListViewState{
			focus:               FocusPullRequestsView,
			query:               query,
			items:               program.reviewSessionFiles(),
			selectedVisibleLine: program.reviewSessionSelectedVisibleLine(),
			renderSelectedLine:  true,
		})
		return
	}

	program.renderSelectableListView(view, selectableListViewState{
		focus:               FocusPullRequestsView,
		query:               program.model.PullRequestSearchQuery(program.model.ActivePullRequestTab()),
		items:               program.model.CurrentPullRequests(),
		selectedVisibleLine: program.model.SelectedPullRequestIndex(program.model.ActivePullRequestTab()),
		renderSelectedLine:  true,
	})
}

func (program *Program) renderNotificationsView(view *gocui.View) {
	program.renderSelectableListView(view, selectableListViewState{
		focus:               FocusNotificationsView,
		query:               program.model.NotificationSearchQuery(),
		items:               program.model.Notifications(),
		selectedVisibleLine: program.model.SelectedNotificationIndex(),
		renderSelectedLine:  true,
	})
}

func (program *Program) renderDetailView(view *gocui.View) {
	program.detailWrapWidth = effectiveMarkdownWidth(view.InnerWidth())
	detailDocument := program.currentDetailDocument(view)
	program.syncDetailViewState(detailDocument, view.InnerHeight())
	renderDetailDocumentView(view, detailDocument, program.detailViewState)
	if program.detailImageManager != nil {
		program.detailImageManager.Sync(detailDocument.images)
	}
}
