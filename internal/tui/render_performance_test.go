package tui

import "testing"

func TestRefreshViews_GivenReviewModeDescription_WhenRefreshingRepeatedly_ThenItStaysBelowTheRegressionCeiling(t *testing.T) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(t)
	defer gui.Close()

	actual := testing.AllocsPerRun(10, func() {
		then_noError(t, subject.refreshViews(gui))
	})

	const expectedMaximumAllocsPerRun = 22000.0
	if actual > expectedMaximumAllocsPerRun {
		t.Fatalf("expected review-description refresh allocations to stay below %.0f allocs/run, actual %.2f", expectedMaximumAllocsPerRun, actual)
	}
}

func TestRefreshViews_GivenActionsPopupOpenOnReviewDescription_WhenRefreshingRepeatedly_ThenItStaysBelowTheRegressionCeiling(t *testing.T) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(t)
	defer gui.Close()
	then_noError(t, subject.openActionsPopup(gui, nil))

	actual := testing.AllocsPerRun(10, func() {
		then_noError(t, subject.refreshViews(gui))
	})

	const expectedMaximumAllocsPerRun = 26000.0
	if actual > expectedMaximumAllocsPerRun {
		t.Fatalf("expected actions-popup refresh allocations to stay below %.0f allocs/run, actual %.2f", expectedMaximumAllocsPerRun, actual)
	}
}
