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
