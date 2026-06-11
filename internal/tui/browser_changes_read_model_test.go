package tui

import (
	"strings"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestBrowserChangesReadModel_GivenCursorOnAFileHeader_WhenResolvingTheFilePathAtCursor_ThenItUsesTheSnapshotSelection(t *testing.T) {
	subject, document := given_browserChangesReadModelForTests(t)
	lineIndex, _ := given_detailDocumentLineContaining(t, document, "widget.go")
	subject.selection.state.cursor = detailPosition{line: lineIndex, column: 0}

	actual, ok := subject.filePathAtCursor()

	if !ok {
		t.Fatal("expected a file path at the cursor")
	}
	if actual != "widget.go" {
		t.Fatalf("expected file path %q, actual %q", "widget.go", actual)
	}
}

func TestBrowserChangesReadModel_GivenCursorOnAnInlineThreadHeader_WhenResolvingTheThreadAtCursor_ThenItUsesTheSnapshotSelection(t *testing.T) {
	subject, document := given_browserChangesReadModelForTests(t)
	if len(subject.files) == 0 || len(subject.files[0].Threads) == 0 {
		t.Fatal("expected a browser changes thread")
	}
	thread := subject.files[0].Threads[0]
	headerLineIndex := reviewDiffThreadHeaderLineIndex(subject.renderedRows, thread.ID)
	if headerLineIndex < 0 {
		t.Fatalf("expected a thread header line for %q", thread.ID)
	}
	subject.selection.state.cursor = detailPosition{line: headerLineIndex, column: 0}
	subject.selection.document = document

	actual, actualOK := subject.threadAtCursor()

	if !actualOK {
		t.Fatal("expected a thread at the cursor")
	}
	if actual.ID != thread.ID {
		t.Fatalf("expected thread id %q, actual %q", thread.ID, actual.ID)
	}
}

func TestBrowserChangesReadModel_GivenAFileVisibilityToggle_WhenPlanningTheSync_ThenItRebuildsTheDocumentAndKeepsTheHeaderFocused(t *testing.T) {
	subject, _ := given_browserChangesReadModelForTests(t)

	actual, ok := subject.fileVisibilityPlan("widget.go")

	if !ok {
		t.Fatal("expected a file visibility plan")
	}
	if !actual.focusLineKnown {
		t.Fatal("expected the file visibility plan to keep a focused line")
	}
	if actual.focusLine != 0 {
		t.Fatalf("expected focused line %d, actual %d", 0, actual.focusLine)
	}
	if actualText := string(actual.document.text); actualText == "" || strings.Contains(actualText, "+added line") {
		t.Fatalf("expected the collapsed file plan to hide the diff body, actual %q", actualText)
	}
}

func TestBrowserChangesReadModel_GivenAThreadVisibilityToggle_WhenPlanningTheSync_ThenItRebuildsTheDocumentAndKeepsTheThreadHeaderFocused(t *testing.T) {
	subject, _ := given_browserChangesReadModelForTests(t)
	thread := subject.files[0].Threads[0]

	actual, ok := subject.threadVisibilityPlan(thread)

	if !ok {
		t.Fatal("expected a thread visibility plan")
	}
	if !actual.focusLineKnown {
		t.Fatal("expected the thread visibility plan to keep a focused line")
	}
	if actual.focusLine < 0 || actual.focusLine >= len(actual.document.lines) {
		t.Fatalf("expected focused line within %d lines, actual %d", len(actual.document.lines), actual.focusLine)
	}
	if actualLine := string(actual.document.lines[actual.focusLine]); !strings.Contains(actualLine, "Unresolved") {
		t.Fatalf("expected focused line to stay on the thread header, actual %q", actualLine)
	}
	if actualText := string(actual.document.text); actualText == "" || strings.Contains(actualText, "Needs follow-up") {
		t.Fatalf("expected the collapsed thread plan to hide the thread body, actual %q", actualText)
	}
}

func given_browserChangesReadModelForTests(t *testing.T) (browserChangesReadModel, detailDocument) {
	t.Helper()

	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"Needs follow-up": "Needs follow-up"}}
	diff := githubcli.PullRequestDiff{
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
	files := buildReviewDiffData(diff).Files
	renderedRows := buildPullRequestChangesRenderedRowsForViewerWithWordWrap(files, renderer, 72, true, nil, nil, "")
	document := newReviewDiffDetailDocumentWithWordWrap(renderedRows, 72, true)
	state := newDetailViewState()
	state.sync(document, 8)

	return browserChangesReadModel{
		sectionScopeKey: pullRequestDetailKey(githubdomain.Repository{NameWithOwner: "acme/widgets"}, 42),
		files:           files,
		selection: detailCursorSelection{
			document: document,
			state:    state,
		},
		renderedRows:       renderedRows,
		renderer:           renderer,
		wordWrapEnabled:    true,
		connectedUserLogin: "",
	}, document
}
