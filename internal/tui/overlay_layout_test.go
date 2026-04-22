package tui

import "testing"

func TestBoundedHalfWidth_GivenANarrowViewport_WhenSizing_ThenItFallsBackAndClampsInsideTheViewport(t *testing.T) {
	actual := boundedHalfWidth(14, modalEditorMinWidth, modalEditorFallbackWidth)

	if actual != 10 {
		t.Fatalf("expected width %d, actual %d", 10, actual)
	}
}

func TestCenteredOverlayFrame_GivenAnOversizedOverlay_WhenCentering_ThenItClampsToTheViewport(t *testing.T) {
	actual := centeredOverlayFrame(12, 5, 20, 8)
	expected := paneFrame{x0: 0, y0: 0, x1: 11, y1: 4}

	if actual != expected {
		t.Fatalf("expected frame %+v, actual %+v", expected, actual)
	}
}
