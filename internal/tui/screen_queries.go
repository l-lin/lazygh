package tui

func isReviewScreenMode(mode ScreenMode) bool {
	return mode == ScreenModeReview || mode == ScreenModeStoryReview
}

func browserSideFocus(lastSideFocus Focus) Focus {
	switch lastSideFocus {
	case FocusPullRequestsView, FocusNotificationsView:
		return lastSideFocus
	default:
		return FocusUserView
	}
}

func notificationLoadEligible(mode ScreenMode, sideFocus Focus) bool {
	return !isReviewScreenMode(mode) && sideFocus == FocusNotificationsView
}

func pullRequestDiffLoadEligible(mode ScreenMode, showsPullRequestDetailTabs bool, activeDetailTab DetailTab) bool {
	return isReviewScreenMode(mode) || (showsPullRequestDetailTabs && activeDetailTab == ChangesDetailTab)
}

func reviewTreeSearchEligible(mode ScreenMode, searchActive bool, searchTarget Focus) bool {
	return isReviewScreenMode(mode) && searchActive && searchTarget == FocusPullRequestsView
}

func reviewTreeFoldEligible(mode ScreenMode, focus Focus, blocked bool) bool {
	return isReviewScreenMode(mode) && focus == FocusPullRequestsView && !blocked
}

func (program *Program) reviewModeActive() bool {
	if program == nil {
		return false
	}
	return isReviewScreenMode(program.screenState().Mode)
}

func (program *Program) storyReviewModeActive() bool {
	if program == nil {
		return false
	}
	return program.screenState().Mode == ScreenModeStoryReview
}
