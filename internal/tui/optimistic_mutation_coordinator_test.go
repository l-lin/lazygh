package tui

import "testing"

func TestOptimisticMutationCoordinator_GivenSequenceState_WhenGeneratingTheNextMutationID_ThenItReturnsAnUpdatedCopyAndFormattedID(t *testing.T) {
	subject := optimisticMutationCoordinator{optimisticMutationSequence: 7}

	actualCoordinator, actualID := subject.nextOptimisticMutationID(" comment ")

	if actualCoordinator.optimisticMutationSequence != 8 {
		t.Fatalf("expected updated sequence %d, actual %d", 8, actualCoordinator.optimisticMutationSequence)
	}
	if actualID != "optimistic:comment:8" {
		t.Fatalf("expected optimistic mutation id %q, actual %q", "optimistic:comment:8", actualID)
	}
	if subject.optimisticMutationSequence != 7 {
		t.Fatalf("expected original sequence %d, actual %d", 7, subject.optimisticMutationSequence)
	}
}
