package tui

import "testing"

func TestListViewportRuntimeState_GivenPendingPlacements_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	original := newListViewportRuntimeState()

	withPlacement := original.withPendingPlacement(viewPullRequestsName, viewportPlacementCenter)
	withoutPlacement := withPlacement.withoutPendingPlacement(viewPullRequestsName)

	if _, ok := original.pendingPlacement(viewPullRequestsName); ok {
		t.Fatal("expected the original state to keep no pending placement")
	}
	if actual, ok := withPlacement.pendingPlacement(viewPullRequestsName); !ok || actual != viewportPlacementCenter {
		t.Fatalf("expected the updated state to store placement %v, actual %v (present=%v)", viewportPlacementCenter, actual, ok)
	}
	if _, ok := withoutPlacement.pendingPlacement(viewPullRequestsName); ok {
		t.Fatal("expected the cleared copy to drop the pending placement")
	}
	if actual, ok := withPlacement.pendingPlacement(viewPullRequestsName); !ok || actual != viewportPlacementCenter {
		t.Fatalf("expected the intermediate state to stay unchanged after clearing the copy, actual %v (present=%v)", actual, ok)
	}
}

func TestKeybindingRuntimeState_GivenRegisteredFingerprint_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	original := keybindingRuntimeState{registeredFingerprint: "old"}

	actual := original.withRegisteredFingerprint("new")

	if original.registeredFingerprintValue() != "old" {
		t.Fatalf("expected the original fingerprint %q, actual %q", "old", original.registeredFingerprintValue())
	}
	if actual.registeredFingerprintValue() != "new" {
		t.Fatalf("expected the updated fingerprint %q, actual %q", "new", actual.registeredFingerprintValue())
	}
}
