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

	return focus == model.fullscreenPane && (focus == FocusUserView || focus == FocusPullRequestsView)
}

func (model *Model) resizablePaneFocus() (Focus, bool) {
	if model.paneLayoutSize == PaneLayoutFullscreen && (model.fullscreenPane == FocusUserView || model.fullscreenPane == FocusPullRequestsView) {
		return model.fullscreenPane, true
	}

	switch model.focus {
	case FocusUserView:
		return FocusUserView, true
	case FocusPullRequestsView:
		return FocusPullRequestsView, true
	default:
		return FocusUserView, false
	}
}
