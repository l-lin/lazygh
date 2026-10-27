package tui

import "testing"

func TestBoundedHalfWidth_GivenANarrowViewport_WhenSizing_ThenItFallsBackAndClampsInsideTheViewport(t *testing.T) {
	actual := boundedHalfWidth(14, modalEditorMinWidth, modalEditorFallbackWidth)

	if actual != 10 {
		t.Fatalf("expected width %d, actual %d", 10, actual)
	}
}

func TestModalEditorFrame_GivenViewportSizes_WhenLayingOutInput_ThenItIsTopAlignedAndHorizontallyCentered(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		width    int
		height   int
		expected paneFrame
	}{
		{name: "standard", width: 120, height: 40, expected: paneFrame{x0: 30, y0: 0, x1: 89, y1: 6}},
		{name: "narrow", width: 14, height: 20, expected: paneFrame{x0: 2, y0: 0, x1: 11, y1: 6}},
		{name: "short", width: 120, height: 5, expected: paneFrame{x0: 30, y0: 0, x1: 89, y1: 4}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			// GIVEN an open input form in the viewport.
			input := screenLayoutInput{modalEditor: modalEditorLayoutState{visible: true}}

			// WHEN its overlay is laid out.
			actual := overlayFrameForView(input, viewModalEditorName, scenario.width, scenario.height)

			// THEN it stays at the top and fits inside the viewport.
			expected := scenario.expected
			if actual.Frame != expected {
				t.Fatalf("expected frame %+v, actual %+v", expected, actual.Frame)
			}
		})
	}
}

func TestCenteredOverlayFrame_GivenAnOversizedOverlay_WhenCentering_ThenItClampsToTheViewport(t *testing.T) {
	actual := centeredOverlayFrame(12, 5, 20, 8)
	expected := paneFrame{x0: 0, y0: 0, x1: 11, y1: 4}

	if actual != expected {
		t.Fatalf("expected frame %+v, actual %+v", expected, actual)
	}
}
