package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestReviewDiffDetailDocument_GivenRenderedDiffRows_WhenBuildingTheDocument_ThenItStoresTheLineNumberGutterAsANonSelectablePrefix(t *testing.T) {
	file := buildReviewDiffData(given_reviewSessionPullRequestDiff()).Files[0]
	renderedRows := buildReviewDiffRenderedRows(file, nil, 96)

	document := newReviewDiffDetailDocument(renderedRows, 96)
	lineIndex, actualLine := given_detailDocumentLineContaining(t, document, "-old line")

	if actual := actualLine; actual != "-old line" {
		t.Fatalf("expected diff body line %q, actual %q", "-old line", actual)
	}
	actualPrefix := string(document.prefixLines[lineIndex].runes)
	if actualPrefix == "" {
		t.Fatal("expected a diff gutter prefix")
	}
	if actualPrefix == "-old line" {
		t.Fatalf("expected the gutter prefix to stay outside the selectable text, actual %q", actualPrefix)
	}
	if !strings.HasSuffix(actualPrefix, "│ ") {
		t.Fatalf("expected diff gutter prefix to end with %q, actual %q", "│ ", actualPrefix)
	}
	rowIndex := document.lineStartRows[lineIndex]
	if actual := document.rowSelectionText(rowIndex, rowIndex); actual != "-old line" {
		t.Fatalf("expected selectable row text %q, actual %q", "-old line", actual)
	}
}

func TestReviewDiffDetailDocument_GivenRenderedDiffRows_WhenSearching_ThenItIgnoresTheLineNumberGutter(t *testing.T) {
	file := buildReviewDiffData(given_reviewSessionPullRequestDiff()).Files[0]
	renderedRows := buildReviewDiffRenderedRows(file, nil, 96)

	document := newReviewDiffDetailDocument(renderedRows, 96)

	if actual := len(document.searchMatches("2 :")); actual != 0 {
		t.Fatalf("expected no gutter search matches, actual %d", actual)
	}
	if actual := len(document.searchMatches("old line")); actual != 1 {
		t.Fatalf("expected one body search match, actual %d", actual)
	}
}

func TestBrowserChangesCursor_GivenDiffLineAtLogicalColumnZero_WhenRendering_ThenTheVisibleCursorStartsAfterTheGutterAndCannotMoveIntoIt(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailCursorOnLineContaining(t, gui, subject, "+new line")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, _ := given_detailDocumentLineContaining(t, document, "+new line")
	expectedCursorX := len(document.prefixLines[lineIndex].runes)

	actualCursorX, _ := detailView.Cursor()
	if actualCursorX != expectedCursorX {
		t.Fatalf("expected browser changes cursor x %d, actual %d", expectedCursorX, actualCursorX)
	}
	if subject.detailViewState.cursor.column != 0 {
		t.Fatalf("expected logical diff cursor column %d, actual %d", 0, subject.detailViewState.cursor.column)
	}

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '0')
	actualErr = handler(gui, detailView)
	then_noError(t, actualErr)

	actualCursorX, _ = detailView.Cursor()
	if actualCursorX != expectedCursorX {
		t.Fatalf("expected browser changes cursor x %d after moving to row start, actual %d", expectedCursorX, actualCursorX)
	}
	if subject.detailViewState.cursor.column != 0 {
		t.Fatalf("expected logical diff cursor column %d after moving to row start, actual %d", 0, subject.detailViewState.cursor.column)
	}
}

func TestReviewDiffCursor_GivenDiffLineAtLogicalColumnZero_WhenRendering_ThenTheVisibleCursorStartsAfterTheGutterAndCannotMoveIntoIt(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "+new line")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, _ := given_detailDocumentLineContaining(t, document, "+new line")
	expectedCursorX := len(document.prefixLines[lineIndex].runes)

	actualCursorX, _ := detailView.Cursor()
	if actualCursorX != expectedCursorX {
		t.Fatalf("expected review diff cursor x %d, actual %d", expectedCursorX, actualCursorX)
	}
	if subject.detailViewState.cursor.column != 0 {
		t.Fatalf("expected logical diff cursor column %d, actual %d", 0, subject.detailViewState.cursor.column)
	}

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '0')
	actualErr = handler(gui, detailView)
	then_noError(t, actualErr)

	actualCursorX, _ = detailView.Cursor()
	if actualCursorX != expectedCursorX {
		t.Fatalf("expected review diff cursor x %d after moving to row start, actual %d", expectedCursorX, actualCursorX)
	}
	if subject.detailViewState.cursor.column != 0 {
		t.Fatalf("expected logical diff cursor column %d after moving to row start, actual %d", 0, subject.detailViewState.cursor.column)
	}
}

func given_browserChangesDetailCursorOnLineContaining(t *testing.T, gui *gocui.Gui, subject *Program, segment string) {
	t.Helper()

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	subject.syncDetailViewState(document, detailView.InnerHeight())
	lineIndex, _ := given_detailDocumentLineContaining(t, document, segment)
	subject.detailViewState.cursor = detailPosition{line: lineIndex, column: 0}
	subject.detailViewState.preferredColumn = 0
	subject.detailViewState.sync(document, detailView.InnerHeight())
	actualErr = subject.refreshDetailView(gui)
	then_noError(t, actualErr)
}

func TestGutterCursorBindings_GivenProgram_WhenListingTheDetailKeymaps_ThenTheLeftMotionStillUsesTheNormalBinding(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	then_bindingExists(t, subject.keybindingSpecs(), keybindingSpec{viewName: viewDetailName, key: 'h', handler: subject.moveDetailCursorLeft})
	then_bindingExists(t, subject.keybindingSpecs(), keybindingSpec{viewName: viewDetailName, key: gocui.KeyArrowLeft, handler: subject.moveDetailCursorLeft})
}
