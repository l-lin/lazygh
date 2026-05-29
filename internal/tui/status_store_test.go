package tui

import "testing"

func TestStatusStore_GivenFeedbackAndGHCommandStatus_WhenSettingAndClearing_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := statusStore{feedbackMessage: "stale", ghCommandLoadingMessage: "Running `old`."}

	feedbackSet := subject.withFeedback(" saved ")
	feedbackCleared := feedbackSet.withoutFeedback()
	loadingStarted := feedbackCleared.withGHCommandLoadingStarted(" gh pr ready ")
	loadingCleared := loadingStarted.withGHCommandLoadingCleared()

	if actual := feedbackSet.feedbackMessage; actual != "saved" {
		t.Fatalf("expected feedback message %q, actual %q", "saved", actual)
	}
	if actual := feedbackSet.ghCommandLoadingMessage; actual != "Running `old`." {
		t.Fatalf("expected gh command loading message %q, actual %q", "Running `old`.", actual)
	}
	if actual := feedbackCleared.feedbackMessage; actual != "" {
		t.Fatalf("expected feedback message %q, actual %q", "", actual)
	}
	if actual := loadingStarted.feedbackMessage; actual != "" {
		t.Fatalf("expected loading start to clear feedback, actual %q", actual)
	}
	if actual := loadingStarted.ghCommandLoadingMessage; actual != "Running `gh pr ready`." {
		t.Fatalf("expected gh command loading message %q, actual %q", "Running `gh pr ready`.", actual)
	}
	if actual := loadingCleared.ghCommandLoadingMessage; actual != "" {
		t.Fatalf("expected gh command loading message %q, actual %q", "", actual)
	}
	if actual := subject.feedbackMessage; actual != "stale" {
		t.Fatalf("expected the original feedback message %q, actual %q", "stale", actual)
	}
	if actual := subject.ghCommandLoadingMessage; actual != "Running `old`." {
		t.Fatalf("expected the original gh command loading message %q, actual %q", "Running `old`.", actual)
	}
}

func TestStatusStore_GivenFeedback_WhenStartingAndFinishingStoryReviewLoading_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := statusStore{feedbackMessage: "stale"}

	started := subject.withStoryReviewLoadingStarted()
	finished := started.withStoryReviewLoadingFinished()

	if !started.storyReviewLoading {
		t.Fatal("expected story review loading to start")
	}
	if actual := started.feedbackMessage; actual != "" {
		t.Fatalf("expected story review loading to clear feedback, actual %q", actual)
	}
	if finished.storyReviewLoading {
		t.Fatal("expected story review loading to finish")
	}
	if actual := finished.feedbackMessage; actual != "" {
		t.Fatalf("expected finished feedback message %q, actual %q", "", actual)
	}
	if actual := subject.feedbackMessage; actual != "stale" {
		t.Fatalf("expected the original feedback message %q, actual %q", "stale", actual)
	}
	if subject.storyReviewLoading {
		t.Fatal("expected the original story review loading flag to stay false")
	}
}
