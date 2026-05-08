package tui

type paneFrame struct {
	x0 int
	y0 int
	x1 int
	y1 int
}

type mainPaneLayout struct {
	user                 paneFrame
	userVisible          bool
	pullRequests         paneFrame
	pullRequestsVisible  bool
	notifications        paneFrame
	notificationsVisible bool
	detail               paneFrame
	detailVisible        bool
}

const browserCollapsedSidebarPaneTotalHeight = 5

func calculateMainPaneLayout(maxX int, contentMaxY int, layoutSize PaneLayoutSize, fullscreenPane Focus) mainPaneLayout {
	return calculateMainPaneLayoutWithSidebarState(maxX, contentMaxY, layoutSize, fullscreenPane, FocusUserView, userViewTotalHeight, true)
}

func calculateMainPaneLayoutWithUserViewHeight(maxX int, contentMaxY int, layoutSize PaneLayoutSize, fullscreenPane Focus, userHeight int) mainPaneLayout {
	return calculateMainPaneLayoutWithSidebarState(maxX, contentMaxY, layoutSize, fullscreenPane, FocusUserView, userHeight, true)
}

func calculateMainPaneLayoutWithSidebarState(maxX int, contentMaxY int, layoutSize PaneLayoutSize, fullscreenPane Focus, focus Focus, userHeight int, showNotifications bool) mainPaneLayout {
	if maxX < 1 {
		maxX = 1
	}
	if contentMaxY < 1 {
		contentMaxY = 1
	}

	if layoutSize == PaneLayoutFullscreen {
		fullscreenFrame := paneFrame{x0: 0, y0: 0, x1: maxX - 1, y1: contentMaxY - 1}
		layout := mainPaneLayout{}
		switch fullscreenPane {
		case FocusPullRequestsView:
			layout.pullRequestsVisible = true
			layout.pullRequests = fullscreenFrame
		case FocusNotificationsView:
			layout.notificationsVisible = true
			layout.notifications = fullscreenFrame
		case FocusDetailView:
			layout.detailVisible = true
			layout.detail = fullscreenFrame
		default:
			layout.userVisible = true
			layout.user = fullscreenFrame
		}
		return layout
	}

	sidebarWidth := defaultSidebarWidth(maxX)
	if layoutSize == PaneLayoutHalfWidth {
		sidebarWidth = maxX / 2
	}
	if sidebarWidth < 1 {
		sidebarWidth = 1
	}

	if !showNotifications {
		userFrame, pullRequestsFrame, detailFrame := calculateSidebarFrames(maxX, contentMaxY, sidebarWidth, userHeight)
		return mainPaneLayout{
			user:                userFrame,
			userVisible:         true,
			pullRequests:        pullRequestsFrame,
			pullRequestsVisible: true,
			detail:              detailFrame,
			detailVisible:       true,
		}
	}

	userFrame, pullRequestsFrame, notificationsFrame, detailFrame := calculateSidebarFramesWithNotifications(maxX, contentMaxY, sidebarWidth, userHeight, focus == FocusNotificationsView)
	return mainPaneLayout{
		user:                 userFrame,
		userVisible:          true,
		pullRequests:         pullRequestsFrame,
		pullRequestsVisible:  true,
		notifications:        notificationsFrame,
		notificationsVisible: true,
		detail:               detailFrame,
		detailVisible:        true,
	}
}

func defaultSidebarWidth(maxX int) int {
	sidebarWidth := maxX * sidebarWidthPercent / 100
	maxSidebarWidth := maxX - minimumDetailWidth
	if maxSidebarWidth < minimumSidebarWidth {
		return maxX / 2
	}
	if sidebarWidth < minimumSidebarWidth {
		sidebarWidth = minimumSidebarWidth
	}
	if sidebarWidth > maxSidebarWidth {
		sidebarWidth = maxSidebarWidth
	}

	return sidebarWidth
}

func calculateSidebarFrames(maxX int, contentMaxY int, sidebarWidth int, userHeight int) (paneFrame, paneFrame, paneFrame) {
	sidebarX1, detailX0 := sidebarColumns(maxX, sidebarWidth)
	userY1, pullRequestsY0 := calculateSidebarSplitY(contentMaxY, userHeight)
	return paneFrame{x0: 0, y0: 0, x1: sidebarX1, y1: userY1}, paneFrame{x0: 0, y0: pullRequestsY0, x1: sidebarX1, y1: contentMaxY - 1}, paneFrame{x0: detailX0, y0: 0, x1: maxX - 1, y1: contentMaxY - 1}
}

func calculateSidebarFramesWithNotifications(maxX int, contentMaxY int, sidebarWidth int, userHeight int, notificationsActive bool) (paneFrame, paneFrame, paneFrame, paneFrame) {
	sidebarX1, detailX0 := sidebarColumns(maxX, sidebarWidth)
	userY1, lowerPaneY0 := calculateSidebarSplitY(contentMaxY, userHeight)
	remainingHeight := contentMaxY - lowerPaneY0
	collapsedHeight := collapsedSidebarPaneHeight(remainingHeight)
	expandedHeight := remainingHeight - collapsedHeight
	if expandedHeight < 2 {
		expandedHeight = 2
		collapsedHeight = maxInt(2, remainingHeight-expandedHeight)
	}

	pullRequestsHeight := expandedHeight
	notificationsHeight := collapsedHeight
	if notificationsActive {
		pullRequestsHeight = collapsedHeight
		notificationsHeight = expandedHeight
	}

	pullRequestsY1 := lowerPaneY0 + pullRequestsHeight - 1
	notificationsY0 := pullRequestsY1 + 1
	notificationsY1 := notificationsY0 + notificationsHeight - 1
	if notificationsY1 >= contentMaxY {
		notificationsY1 = contentMaxY - 1
	}

	return paneFrame{x0: 0, y0: 0, x1: sidebarX1, y1: userY1},
		paneFrame{x0: 0, y0: lowerPaneY0, x1: sidebarX1, y1: pullRequestsY1},
		paneFrame{x0: 0, y0: notificationsY0, x1: sidebarX1, y1: notificationsY1},
		paneFrame{x0: detailX0, y0: 0, x1: maxX - 1, y1: contentMaxY - 1}
}

func sidebarColumns(maxX int, sidebarWidth int) (int, int) {
	sidebarX1 := sidebarWidth - 1
	detailX0 := sidebarX1 + 1
	if detailX0 >= maxX {
		detailX0 = maxX / 2
		sidebarX1 = detailX0 - 1
	}
	return sidebarX1, detailX0
}

func collapsedSidebarPaneHeight(remainingHeight int) int {
	if remainingHeight < 4 {
		return maxInt(2, remainingHeight/2)
	}
	if browserCollapsedSidebarPaneTotalHeight >= remainingHeight {
		return maxInt(2, remainingHeight/2)
	}
	return browserCollapsedSidebarPaneTotalHeight
}

func calculateSidebarSplitY(contentMaxY int, userHeight int) (int, int) {
	if userHeight >= contentMaxY {
		userHeight = contentMaxY / 2
	}
	if userHeight < 2 {
		userHeight = 2
	}

	userY1 := userHeight - 1
	pullRequestsY0 := userY1 + 1
	if pullRequestsY0 >= contentMaxY {
		pullRequestsY0 = contentMaxY / 2
		userY1 = pullRequestsY0 - 1
	}

	return userY1, pullRequestsY0
}
