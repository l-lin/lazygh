package tui

type paneFrame struct {
	x0 int
	y0 int
	x1 int
	y1 int
}

type mainPaneLayout struct {
	user                paneFrame
	userVisible         bool
	pullRequests        paneFrame
	pullRequestsVisible bool
	detail              paneFrame
	detailVisible       bool
}

func calculateMainPaneLayout(maxX int, contentMaxY int, layoutSize PaneLayoutSize, fullscreenPane Focus) mainPaneLayout {
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

	userFrame, pullRequestsFrame, detailFrame := calculateSidebarFrames(maxX, contentMaxY, sidebarWidth)
	return mainPaneLayout{
		user:                userFrame,
		userVisible:         true,
		pullRequests:        pullRequestsFrame,
		pullRequestsVisible: true,
		detail:              detailFrame,
		detailVisible:       true,
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

func calculateSidebarFrames(maxX int, contentMaxY int, sidebarWidth int) (paneFrame, paneFrame, paneFrame) {
	sidebarX1 := sidebarWidth - 1
	detailX0 := sidebarX1 + 1
	if detailX0 >= maxX {
		detailX0 = maxX / 2
		sidebarX1 = detailX0 - 1
	}

	userY1, pullRequestsY0 := calculateSidebarSplitY(contentMaxY)
	return paneFrame{x0: 0, y0: 0, x1: sidebarX1, y1: userY1}, paneFrame{x0: 0, y0: pullRequestsY0, x1: sidebarX1, y1: contentMaxY - 1}, paneFrame{x0: detailX0, y0: 0, x1: maxX - 1, y1: contentMaxY - 1}
}

func calculateSidebarSplitY(contentMaxY int) (int, int) {
	userHeight := userViewTotalHeight
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
