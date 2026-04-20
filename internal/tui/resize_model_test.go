package tui

import "testing"

func TestGrowFocusedPane_GivenUserFocus_WhenCyclingThroughTheResizeStates_ThenItMovesFromDefaultToHalfWidthToFullscreenAndBack(t *testing.T) {
	subject := given_model()

	if subject.PaneLayoutSize() != PaneLayoutDefault {
		t.Fatalf("expected initial layout size %v, actual %v", PaneLayoutDefault, subject.PaneLayoutSize())
	}

	subject.GrowFocusedPane()
	if subject.PaneLayoutSize() != PaneLayoutHalfWidth {
		t.Fatalf("expected layout size %v after the first grow, actual %v", PaneLayoutHalfWidth, subject.PaneLayoutSize())
	}

	subject.GrowFocusedPane()
	if subject.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected layout size %v after the second grow, actual %v", PaneLayoutFullscreen, subject.PaneLayoutSize())
	}
	if subject.FullscreenPane() != FocusUserView {
		t.Fatalf("expected fullscreen pane %v, actual %v", FocusUserView, subject.FullscreenPane())
	}

	subject.GrowFocusedPane()
	if subject.PaneLayoutSize() != PaneLayoutDefault {
		t.Fatalf("expected layout size %v after the third grow, actual %v", PaneLayoutDefault, subject.PaneLayoutSize())
	}
	if subject.Focus() != FocusUserView {
		t.Fatalf("expected focus %v after restoring the default layout, actual %v", FocusUserView, subject.Focus())
	}
}

func TestShrinkFocusedPane_GivenPullRequestsFocus_WhenCyclingThroughTheResizeStates_ThenItMovesFromDefaultToFullscreenToHalfWidthAndBack(t *testing.T) {
	subject := given_model()
	subject.FocusPullRequestsView()

	subject.ShrinkFocusedPane()
	if subject.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected layout size %v after the first shrink, actual %v", PaneLayoutFullscreen, subject.PaneLayoutSize())
	}
	if subject.FullscreenPane() != FocusPullRequestsView {
		t.Fatalf("expected fullscreen pane %v, actual %v", FocusPullRequestsView, subject.FullscreenPane())
	}

	subject.ShrinkFocusedPane()
	if subject.PaneLayoutSize() != PaneLayoutHalfWidth {
		t.Fatalf("expected layout size %v after the second shrink, actual %v", PaneLayoutHalfWidth, subject.PaneLayoutSize())
	}

	subject.ShrinkFocusedPane()
	if subject.PaneLayoutSize() != PaneLayoutDefault {
		t.Fatalf("expected layout size %v after the third shrink, actual %v", PaneLayoutDefault, subject.PaneLayoutSize())
	}
	if subject.Focus() != FocusPullRequestsView {
		t.Fatalf("expected focus %v after restoring the default layout, actual %v", FocusPullRequestsView, subject.Focus())
	}
}

func TestResizeFocusedPane_GivenDetailFocus_WhenGrowingOrShrinking_ThenTheLayoutStateDoesNotChange(t *testing.T) {
	subject := given_model()
	subject.OpenDetail()

	subject.GrowFocusedPane()
	subject.ShrinkFocusedPane()

	if subject.PaneLayoutSize() != PaneLayoutDefault {
		t.Fatalf("expected layout size %v, actual %v", PaneLayoutDefault, subject.PaneLayoutSize())
	}
	if subject.Focus() != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, subject.Focus())
	}
}
