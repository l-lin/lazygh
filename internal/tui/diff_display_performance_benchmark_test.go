package tui

import "testing"

// 2026-06-02 cached full-redraw baseline on the same 10k-line fixture:
// 352ms/op, 403MB/op, 5.89M allocs/op.
func BenchmarkRenderDetailDocumentView_GivenLargeCachedDiff_WhenRenderingVisibleRows(b *testing.B) {
	subject, gui := given_largeCachedDiffRefreshProgram(b, largeDiffDisplayFixtureLineCount)
	defer gui.Close()

	detailView, actualErr := gui.View(viewDetailName)
	if actualErr != nil {
		b.Fatalf("expected a detail view, actual %v", actualErr)
	}
	document := subject.currentDetailDocument(detailView)
	state := subject.detailState.viewState

	renderVisibleDetailDocumentView(detailView, document, state)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		renderVisibleDetailDocumentView(detailView, document, state)
	}
}

// 2026-06-02 large row-build baselines on the same 10k-line fixture:
// Go diff: 236ms/op, 212MB/op, 3.13M allocs/op.
// Plain-text control: 47ms/op, 105MB/op, 1.44M allocs/op.
func BenchmarkBuildReviewDiffRenderedRows_GivenLargeGoDiff_WhenRenderingRows(b *testing.B) {
	subject := given_largeGoReviewDiffFile(b, largeDiffDisplayFixtureLineCount)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = buildReviewDiffRenderedRows(subject, nil, largeDiffDisplayFixtureWidth)
	}
}

func BenchmarkBuildReviewDiffRenderedRows_GivenLargePlainTextDiff_WhenRenderingRows(b *testing.B) {
	subject := given_largePlainTextReviewDiffFile(b, largeDiffDisplayFixtureLineCount)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = buildReviewDiffRenderedRows(subject, nil, largeDiffDisplayFixtureWidth)
	}
}
