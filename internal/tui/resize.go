package tui

type PaneLayoutSize int

const (
	PaneLayoutDefault PaneLayoutSize = iota
	PaneLayoutHalfWidth
	PaneLayoutFullscreen
)

func (model *Model) PaneLayoutSize() PaneLayoutSize {
	return model.paneLayoutSize
}

func (model *Model) FullscreenPane() Focus {
	return model.fullscreenPane
}

func (model *Model) GrowFocusedPane() {
	focus, ok := model.resizablePaneFocus()
	if !ok {
		return
	}
	if focus == FocusDetailView {
		model.toggleDetailFullscreen()
		return
	}

	switch model.paneLayoutSize {
	case PaneLayoutHalfWidth:
		model.paneLayoutSize = PaneLayoutFullscreen
		model.fullscreenPane = focus
	case PaneLayoutFullscreen:
		model.paneLayoutSize = PaneLayoutDefault
	default:
		model.paneLayoutSize = PaneLayoutHalfWidth
	}
}

func (model *Model) ShrinkFocusedPane() {
	focus, ok := model.resizablePaneFocus()
	if !ok {
		return
	}
	if focus == FocusDetailView {
		model.toggleDetailFullscreen()
		return
	}

	switch model.paneLayoutSize {
	case PaneLayoutFullscreen:
		model.paneLayoutSize = PaneLayoutHalfWidth
	case PaneLayoutHalfWidth:
		model.paneLayoutSize = PaneLayoutDefault
	default:
		model.paneLayoutSize = PaneLayoutFullscreen
		model.fullscreenPane = focus
	}
}

func (model *Model) PaneVisible(focus Focus) bool {
	if model.paneLayoutSize != PaneLayoutFullscreen {
		return true
	}

	return focus == model.fullscreenPane && isMainPaneFocus(focus)
}

func (model *Model) resizablePaneFocus() (Focus, bool) {
	if model.paneLayoutSize == PaneLayoutFullscreen && isMainPaneFocus(model.fullscreenPane) {
		return model.fullscreenPane, true
	}

	switch model.focus {
	case FocusUserView:
		return FocusUserView, true
	case FocusPullRequestsView:
		return FocusPullRequestsView, true
	case FocusNotificationsView:
		return FocusNotificationsView, true
	case FocusDetailView:
		return FocusDetailView, true
	default:
		return FocusUserView, false
	}
}

func (model *Model) toggleDetailFullscreen() {
	if model.paneLayoutSize == PaneLayoutFullscreen && model.fullscreenPane == FocusDetailView {
		model.paneLayoutSize = model.detailFullscreenReturnSize
		return
	}

	model.detailFullscreenReturnSize = model.paneLayoutSize
	model.paneLayoutSize = PaneLayoutFullscreen
	model.fullscreenPane = FocusDetailView
}

func isMainPaneFocus(focus Focus) bool {
	switch focus {
	case FocusUserView, FocusPullRequestsView, FocusNotificationsView, FocusDetailView:
		return true
	default:
		return false
	}
}
