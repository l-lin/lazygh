package tui

import (
	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
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
	program.gui = gui
	maxX, maxY := gui.Size()
	contentMaxY := program.layoutContentHeight(maxY)

	program.maybeLoadConnectedUser(gui)
	program.maybeLoadActivePullRequests(gui)
	if !program.reviewSession.active {
		program.maybeLoadNotifications(gui)
		program.maybeLoadSelectedNotificationDetail(gui)
	}
	program.maybeLoadSelectedPullRequestDetail(gui)
	program.maybeLoadSelectedPullRequestDiff(gui)

	mainPaneLayout := calculateMainPaneLayoutWithSidebarState(maxX, contentMaxY, program.model.PaneLayoutSize(), program.model.FullscreenPane(), program.model.currentSideFocus(), program.sidebarTopPaneHeight(), !program.reviewSession.active)

	if err := program.layoutMainPane(gui, viewUserName, mainPaneLayout.userVisible, mainPaneLayout.user, program.configureUserView, program.renderUserView); err != nil {
		return err
	}
	if err := program.layoutMainPane(gui, viewPullRequestsName, mainPaneLayout.pullRequestsVisible, mainPaneLayout.pullRequests, program.configurePullRequestsView, program.renderPullRequestsView); err != nil {
		return err
	}
	if err := program.layoutMainPane(gui, viewNotificationsName, mainPaneLayout.notificationsVisible, mainPaneLayout.notifications, program.configureNotificationsView, program.renderNotificationsView); err != nil {
		return err
	}
	if err := program.layoutMainPane(gui, viewDetailName, mainPaneLayout.detailVisible, mainPaneLayout.detail, program.configureDetailView, program.renderDetailView); err != nil {
		return err
	}

	if err := program.layoutPaneFooterViews(gui); err != nil {
		return err
	}
	if err := program.layoutStatusLineView(gui); err != nil {
		return err
	}

	if err := syncOverlayLayout(gui, program.helpVisible, program.layoutHelpView, viewHelpName); err != nil {
		return err
	}
	if err := syncOverlayLayout(gui, program.searchPromptVisible(), program.layoutSearchView, viewSearchName); err != nil {
		return err
	}
	if err := syncOverlayLayout(gui, program.modalEditorVisible(), program.layoutModalEditorView, viewModalEditorName); err != nil {
		return err
	}
	if err := syncOverlayLayout(gui, program.pullRequestBuildRunPopupVisible(), program.layoutPullRequestBuildRunPopupView, viewPullRequestBuildInfoName); err != nil {
		return err
	}
	if err := syncOverlayLayout(gui, program.model.ActionsPopupVisible(), program.layoutActionsPopupViews, viewActionsPopupSearchName, viewActionsPopupName); err != nil {
		return err
	}
	if program.model.ActionsPopupVisible() {
		if err := syncOverlayLayout(gui, program.model.ActionsPopupSearchActive(), program.layoutActionsPopupSearchView, viewActionsPopupSearchName); err != nil {
			return err
		}
	}
	if err := program.layoutStatusLineKeyHintsView(gui); err != nil {
		return err
	}

	return program.syncCurrentView(gui)
}

func (program *Program) layoutMainPane(gui *gocui.Gui, viewName string, visible bool, frame paneFrame, configure viewConfigurator, render viewRenderer) error {
	view, err := setPaneView(gui, viewName, visible, frame)
	if err != nil {
		return err
	}
	if view == nil {
		return nil
	}

	configure(view)
	render(view)
	return nil
}

type guiLayouter func(*gocui.Gui) error

func syncOverlayLayout(gui *gocui.Gui, visible bool, layout guiLayouter, viewNames ...string) error {
	if visible {
		return layout(gui)
	}

	return deleteViewsIfPresent(gui, viewNames...)
}

func (program *Program) sidebarTopPaneHeight() int {
	if program.reviewSession.active {
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
		items:               program.model.VisiblePullRequests(),
		selectedVisibleLine: program.model.SelectedVisiblePullRequestIndex(program.model.ActivePullRequestTab()),
		renderSelectedLine:  true,
	})
}

func (program *Program) renderNotificationsView(view *gocui.View) {
	program.renderSelectableListView(view, selectableListViewState{
		focus:               FocusNotificationsView,
		query:               program.model.NotificationSearchQuery(),
		items:               program.model.VisibleNotifications(),
		selectedVisibleLine: program.model.SelectedVisibleNotificationIndex(),
		renderSelectedLine:  true,
	})
}

func (program *Program) renderDetailView(view *gocui.View) {
	program.detailWrapWidth = effectiveMarkdownWidth(view.InnerWidth())
	detailDocument := program.currentDetailDocument(view)
	program.syncDetailViewState(detailDocument, view.InnerHeight())
	renderDetailDocumentView(view, detailDocument, program.detailViewState)
}
