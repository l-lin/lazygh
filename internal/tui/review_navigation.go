package tui

import "github.com/jesseduffield/gocui"

const keymapScopeReviewNavigation = "review_navigation"

type reviewNavigationDirection int

const (
	reviewNavigationBackward reviewNavigationDirection = -1
	reviewNavigationForward  reviewNavigationDirection = 1
)

func (program *Program) handleReviewFileMotionPrefix(gui *gocui.Gui, view *gocui.View, direction reviewNavigationDirection) error {
	if !program.reviewSession.active {
		program.clearPendingSelectionPrefix()
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	viewName := program.reviewNavigationViewName(view)
	target := program.reviewNavigationPrefixTarget(viewName, direction)
	return program.armOrHandleSelectionKeySequence(target, func() error {
		return program.moveReviewSessionFile(gui, int(direction))
	})
}

func (program *Program) reviewNavigationViewName(view *gocui.View) string {
	if view != nil {
		return view.Name()
	}

	switch program.model.Focus() {
	case FocusDetailView:
		return viewDetailName
	case FocusPullRequestsView:
		return viewPullRequestsName
	default:
		return ""
	}
}

func (program *Program) reviewNavigationPrefixTarget(viewName string, direction reviewNavigationDirection) keySequenceTarget {
	action := "previous_prefix"
	if direction > 0 {
		action = "next_prefix"
	}
	return keySequenceTargetFor(viewName, keymapScopeReviewNavigation, action)
}

func (program *Program) moveReviewSessionFile(gui *gocui.Gui, change int) error {
	if !program.reviewSession.active {
		return nil
	}

	originalRow := program.reviewSession.selectedFileTreeRow
	program.adjustReviewSessionSelection(change)
	if program.reviewSession.selectedFileTreeRow == originalRow {
		return nil
	}

	return program.refreshViewsIfGUI(gui)
}
