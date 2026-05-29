package tui

import "testing"

func TestStartupStateModel_GivenAppStartAndSpinnerTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := startupStateModel{loadingSpinnerFrameIndex: 2}

	started := subject.withAppStarted()
	advanced := started.withAdvancedLoadingSpinnerFrame(len(loadingSpinnerFrames))
	unchanged := started.withAdvancedLoadingSpinnerFrame(0)

	if !started.appStarted {
		t.Fatal("expected the started state to mark the program as started")
	}
	if actual := advanced.loadingSpinnerFrameIndex; actual != 3 {
		t.Fatalf("expected advanced spinner frame index %d, actual %d", 3, actual)
	}
	if actual := unchanged.loadingSpinnerFrameIndex; actual != 2 {
		t.Fatalf("expected unchanged spinner frame index %d, actual %d", 2, actual)
	}
	if subject.appStarted {
		t.Fatal("expected the original state to stay unstarted")
	}
	if actual := subject.loadingSpinnerFrameIndex; actual != 2 {
		t.Fatalf("expected the original spinner frame index %d, actual %d", 2, actual)
	}
}
