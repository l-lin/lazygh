package tui

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	githubcli "github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestRefreshViews_GivenAScrolledDetailDocument_WhenRefreshing_ThenTheDetailBufferContainsOnlyVisibleRows(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: given_numberedDetailBody(40)}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGuiWithSize(t, 80, 10)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	actualView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	targetCursorLine := actualView.InnerHeight() + 3
	for range targetCursorLine {
		then_noError(t, subject.moveSelectionDown(gui, actualView))
		actualView, actualErr = gui.View(viewDetailName)
		then_noError(t, actualErr)
	}

	actualOriginRow := subject.detailState.viewState.originRow
	if actualOriginRow == 0 {
		t.Fatal("expected the detail viewport to scroll off the top row")
	}

	actualLines := given_detailBufferLines(actualView)
	expectedMaximumLineCount := actualView.InnerHeight()
	if len(actualLines) > expectedMaximumLineCount {
		t.Fatalf("expected at most %d visible detail rows, actual %d in %q", expectedMaximumLineCount, len(actualLines), actualView.Buffer())
	}

	hiddenLine := given_numberedDetailLine(1)
	if slices.Contains(actualLines, hiddenLine) {
		t.Fatalf("expected the off-screen line %q to stay out of the visible detail buffer, actual %q", hiddenLine, actualView.Buffer())
	}

	document := subject.currentDetailDocument(actualView)
	expectedTopLine := document.rows[actualOriginRow].text
	if len(actualLines) == 0 || actualLines[0] != expectedTopLine {
		t.Fatalf("expected the viewport to start at %q, actual %v", expectedTopLine, actualLines)
	}
}

func TestRenderVisibleDetailDocumentView_GivenAScrolledDocument_WhenRendering_ThenItOnlyWritesViewportRowsAndKeepsTheCursorRelativeToTheViewport(t *testing.T) {
	actualView := given_detailRenderTestView(t)
	document := newDetailDocument(given_numberedDetailBody(12), actualView.InnerWidth())
	subject := newDetailViewState()
	subject.originRow = 3
	subject.cursor = detailPosition{line: 5, column: 0}
	subject.preferredColumn = 0
	subject.manualViewportScroll = true

	renderVisibleDetailDocumentView(actualView, document, subject)

	actualLines := given_detailBufferLines(actualView)
	expected := make([]string, 0, actualView.InnerHeight())
	for lineNumber := subject.originRow + 1; lineNumber <= minInt(document.rowCount(), subject.originRow+actualView.InnerHeight()); lineNumber++ {
		expected = append(expected, given_numberedDetailLine(lineNumber))
	}
	if !reflect.DeepEqual(actualLines, expected) {
		t.Fatalf("expected visible detail lines %v, actual %v", expected, actualLines)
	}

	actualCursorX, actualCursorY := actualView.Cursor()
	expectedCursorY := document.rowIndexForPosition(subject.cursor) - subject.originRow
	if actualCursorX != 0 || actualCursorY != expectedCursorY {
		t.Fatalf("expected detail cursor 0,%d, actual %d,%d", expectedCursorY, actualCursorX, actualCursorY)
	}
}

func TestRenderVisibleDetailDocumentView_GivenVisibleSearchAndYankHighlights_WhenRendering_ThenItKeepsTheHighlightsInsideTheViewport(t *testing.T) {
	gui := given_headlessGuiWithSize(t, 80, 12)
	defer gui.Close()
	actualView, actualErr := gui.SetView(viewDetailName, 0, 0, 39, 6, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}

	document := newDetailDocument(strings.Join([]string{"plain row", "search target", "YANK target", "omega"}, "\n"), actualView.InnerWidth())
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 1, column: 0}
	subject.searchMatches = document.searchMatches("target")
	subject.searchCacheDocumentID = document.id
	subject.searchCacheQuery = "target"
	subject.yankHighlight = detailYankHighlightState{active: true, start: detailPosition{line: 2, column: 0}, end: detailPosition{line: 2, column: 3}}

	renderVisibleDetailDocumentView(actualView, document, subject)

	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewDetailName, 1, "target")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, 2, "YANK", given_searchHighlightColorHex(t), "yank highlight")
}

func TestRenderVisibleDetailDocumentView_GivenAWrappedDiffContinuationRow_WhenRenderingFromTheMiddle_ThenItKeepsTheContinuationGutter(t *testing.T) {
	actualView := given_detailRenderTestView(t)
	longLine := strings.Repeat("wrapped-gutter-segment-", 4)
	file := given_reviewDiffFileWithLongLine(longLine)
	renderedRows := buildReviewDiffRenderedRows(file, nil, 60)
	document := newReviewDiffDetailDocumentWithWordWrap(renderedRows, 32, true)
	lineIndex, _ := given_detailDocumentLineContaining(t, document, longLine)
	continuationRowIndex := document.lineStartRows[lineIndex] + 1
	if continuationRowIndex >= document.rowCount() {
		t.Fatalf("expected a wrapped continuation row for %q", longLine)
	}

	subject := newDetailViewState()
	subject.originRow = continuationRowIndex
	subject.cursor = document.positionForRow(continuationRowIndex, 0)
	subject.manualViewportScroll = true

	renderVisibleDetailDocumentView(actualView, document, subject)

	actualLine, ok := actualView.Line(0)
	if !ok {
		t.Fatal("expected the wrapped continuation row to be visible")
	}
	actualLine = given_withoutANSIEscapeSequences(actualLine)
	if strings.Contains(actualLine, ": 1 │") {
		t.Fatalf("expected the continuation gutter to hide repeated line numbers, actual %q", actualLine)
	}
	if !strings.Contains(actualLine, "│") {
		t.Fatalf("expected the continuation gutter to keep the separator, actual %q", actualLine)
	}
}

func TestRenderVisibleDetailDocumentView_GivenScrolledBrowserDiff_WhenRendering_ThenItPinsTheOwningFileHeader(t *testing.T) {
	actualView := given_detailRenderTestView(t)
	document := given_browserChangesStickyTestDocument(t, 48)
	bodyRowIndex := given_browserChangesStickyTestRowIndexContaining(t, document, "+new line")

	subject := newDetailViewState()
	subject.originRow = bodyRowIndex
	subject.cursor = document.positionForRow(bodyRowIndex, 0)
	subject.manualViewportScroll = true

	renderVisibleDetailDocumentView(actualView, document, subject)

	actualLines := given_detailBufferLines(actualView)
	if len(actualLines) < 2 {
		t.Fatalf("expected pinned header plus body rows, actual %q", actualView.Buffer())
	}
	if !strings.Contains(actualLines[0], "internal/tui/render.go") {
		t.Fatalf("expected the sticky header on the first visible line, actual %q", actualLines[0])
	}
	if !strings.Contains(actualLines[1], "+new line") {
		t.Fatalf("expected the diff body to stay below the sticky header, actual %q", actualLines[1])
	}
}

func TestRenderVisibleDetailDocumentView_GivenWrappedFileHeaderStillVisible_WhenRendering_ThenItShowsNoStickyHeader(t *testing.T) {
	actualView := given_detailRenderTestView(t)
	document := given_browserChangesStickyLongHeaderDocument(t, 24)
	headerLineIndex, _ := given_detailDocumentLineContaining(t, document, "sticky-file-header-context")
	headerContinuationRowIndex := document.lineStartRows[headerLineIndex] + 1
	if headerContinuationRowIndex >= document.rowCount() {
		t.Fatalf("expected a wrapped header continuation row for %q", document.text)
	}

	subject := newDetailViewState()
	subject.originRow = headerContinuationRowIndex
	subject.cursor = document.positionForRow(headerContinuationRowIndex, 0)
	subject.manualViewportScroll = true

	renderVisibleDetailDocumentView(actualView, document, subject)

	actualLines := given_detailBufferLines(actualView)
	if len(actualLines) == 0 {
		t.Fatal("expected visible rows")
	}
	if strings.Contains(actualLines[0], "") {
		t.Fatalf("expected the visible wrapped header continuation instead of a duplicated sticky header, actual %q", actualLines[0])
	}
}

func TestRenderVisibleDetailDocumentView_GivenAnInterFileSpacerAtTheTop_WhenRendering_ThenItShowsNoStickyHeader(t *testing.T) {
	actualView := given_detailRenderTestView(t)
	document := given_browserChangesStickyTestDocument(t, 48)
	spacerRowIndex := given_browserChangesStickyTestSpacerRowIndex(t, document)

	subject := newDetailViewState()
	subject.originRow = spacerRowIndex
	subject.cursor = document.positionForRow(spacerRowIndex, 0)
	subject.manualViewportScroll = true

	renderVisibleDetailDocumentView(actualView, document, subject)

	actualLines := given_detailBufferLines(actualView)
	if len(actualLines) == 0 {
		t.Fatal("expected visible rows")
	}
	if strings.Contains(actualLines[0], "internal/tui/render.go") {
		t.Fatalf("expected no sticky header on the inter-file spacer, actual %q", actualLines[0])
	}
}

func TestRenderVisibleDetailDocumentView_GivenScrolledInlineThread_WhenRendering_ThenItPinsTheOwningFileHeader(t *testing.T) {
	actualView := given_detailRenderTestView(t)
	document := given_browserChangesStickyThreadDocument(t, 48)
	threadRowIndex := given_browserChangesStickyTestRowIndexContaining(t, document, "Needs follow-up")

	subject := newDetailViewState()
	subject.originRow = threadRowIndex
	subject.cursor = document.positionForRow(threadRowIndex, 0)
	subject.manualViewportScroll = true

	renderVisibleDetailDocumentView(actualView, document, subject)

	actualLines := given_detailBufferLines(actualView)
	if len(actualLines) == 0 {
		t.Fatal("expected visible rows")
	}
	if !strings.Contains(actualLines[0], "widget.go") {
		t.Fatalf("expected the sticky header to stay pinned above inline thread rows, actual %q", actualLines[0])
	}
}

func TestRenderVisibleDetailDocumentView_GivenScrolledBrowserDiff_WhenRendering_ThenItFillsThePinnedHeaderLineBackground(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.Palette{StickyFileHeaderBackgroundHex: "#223249"})

	gui := given_headlessGuiWithSize(t, 80, 12)
	defer gui.Close()
	actualView, actualErr := gui.SetView(viewDetailName, 0, 0, 39, 6, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	document := given_browserChangesStickyTestDocument(t, actualView.InnerWidth())
	bodyRowIndex := given_browserChangesStickyTestRowIndexContaining(t, document, "+new line")

	subject := newDetailViewState()
	subject.originRow = bodyRowIndex
	subject.cursor = document.positionForRow(bodyRowIndex, 0)
	subject.manualViewportScroll = true

	renderVisibleDetailDocumentView(actualView, document, subject)

	then_viewLineHasBackgroundColor(t, gui, viewDetailName, 0, given_themeColorHex(t, theme.StickyFileHeaderBackgroundHex), "sticky file header background")
}

func given_numberedDetailBody(lineCount int) string {
	lines := make([]string, 0, lineCount)
	for lineNumber := range lineCount {
		lines = append(lines, given_numberedDetailLine(lineNumber+1))
	}
	return strings.Join(lines, "\n")
}

func given_numberedDetailLine(lineNumber int) string {
	return fmt.Sprintf("detail line %02d", lineNumber)
}

func given_detailBufferLines(view *gocui.View) []string {
	buffer := strings.TrimRight(view.Buffer(), "\n")
	if buffer == "" {
		return nil
	}
	return strings.Split(buffer, "\n")
}

func given_detailRenderTestView(t *testing.T) *gocui.View {
	t.Helper()

	gui := given_headlessGuiWithSize(t, 80, 12)
	t.Cleanup(func() { gui.Close() })
	actualView, actualErr := gui.SetView(viewDetailName, 0, 0, 39, 6, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	return actualView
}

func given_withoutANSIEscapeSequences(text string) string {
	var builder strings.Builder
	insideEscape := false
	for _, character := range text {
		switch {
		case character == '\u001b':
			insideEscape = true
		case insideEscape && character == 'm':
			insideEscape = false
		case !insideEscape:
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

func given_browserChangesStickyTestDocument(t *testing.T, width int) detailDocument {
	t.Helper()

	renderedRows := buildPullRequestChangesRenderedRowsForViewerWithWordWrap(buildReviewDiffData(given_reviewSessionPullRequestDiff()).Files, nil, width, true, nil, nil, "")
	return newBrowserChangesDetailDocumentWithWordWrap(renderedRows, width, true)
}

func given_browserChangesStickyLongHeaderDocument(t *testing.T, width int) detailDocument {
	t.Helper()

	given_file := reviewDiffFile{
		Path:       "internal/tui/sticky-file-header-context/very/long/path/that-wraps.go",
		Additions:  1,
		Deletions:  0,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -1,0 +1,1 @@",
			Lines:  []reviewDiffLine{{Kind: reviewDiffAdditionLine, Text: "wrapped body", RightLine: 1, Side: reviewDiffLineSideRight}},
		}},
	}
	renderedRows := buildPullRequestChangesRenderedRowsForViewerWithWordWrap([]reviewDiffFile{given_file}, nil, width, true, nil, nil, "")
	return newBrowserChangesDetailDocumentWithWordWrap(renderedRows, width, true)
}

func given_browserChangesStickyThreadDocument(t *testing.T, width int) detailDocument {
	t.Helper()

	given_diff := githubcli.PullRequestDiff{
		UnifiedDiff: "diff --git a/widget.go b/widget.go\nindex 0000000..1111111 100644\n--- a/widget.go\n+++ b/widget.go\n@@ -0,0 +1 @@\n+added line\n",
		Files:       []githubcli.PullRequestDiffFile{{Path: "widget.go", ChangeType: "added", Additions: 1}},
		Threads: []githubcli.PullRequestReviewThread{{
			ID:       "thread-1",
			Path:     "widget.go",
			Line:     1,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:      "Needs follow-up",
				CreatedAt: "2026-05-31T09:00:00Z",
			}},
		}},
	}
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"Needs follow-up": "Needs follow-up"}}
	renderedRows := buildPullRequestChangesRenderedRowsForViewerWithWordWrap(buildReviewDiffData(given_diff).Files, renderer, width, true, nil, nil, "")
	return newBrowserChangesDetailDocumentWithWordWrap(renderedRows, width, true)
}

func given_browserChangesStickyTestRowIndexContaining(t *testing.T, document detailDocument, segment string) int {
	t.Helper()

	lineIndex, _ := given_detailDocumentLineContaining(t, document, segment)
	return document.lineStartRows[lineIndex]
}

func given_browserChangesStickyTestSpacerRowIndex(t *testing.T, document detailDocument) int {
	t.Helper()

	for rowIndex, row := range document.rows {
		if row.owningHeaderLine >= 0 || !row.empty {
			continue
		}
		return rowIndex
	}

	t.Fatalf("expected an inter-file spacer row in %q", document.text)
	return -1
}
