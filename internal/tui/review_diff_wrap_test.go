package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestReviewMode_GivenASelectedFileWithALongDiffLine_WhenBuildingViewZeroDocument_ThenItDoesNotWrapTheFileDiff(t *testing.T) {
	longLine := strings.Repeat("very-long-diff-segment-", 4)
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": {
				UnifiedDiff: strings.Join([]string{
					"diff --git a/internal/tui/render.go b/internal/tui/render.go",
					"index 1111111..2222222 100644",
					"--- a/internal/tui/render.go",
					"+++ b/internal/tui/render.go",
					"@@ -1,1 +1,1 @@",
					"+" + longLine,
				}, "\n"),
				Files: []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 60, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, actualLine := given_detailDocumentLineContaining(t, document, longLine)

	if actual := reviewDiffDocumentRowCountForLine(document, lineIndex); actual != 1 {
		t.Fatalf("expected the long diff line to stay on one rendered row, actual %d for %q", actual, actualLine)
	}
}

func TestReviewMode_GivenAnInlineThreadWithALongMarkdownComment_WhenBuildingViewZeroDocument_ThenItWrapsTheThreadBody(t *testing.T) {
	threadBody := strings.Repeat("wrap this review comment ", 18)
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": {
				UnifiedDiff: strings.Join([]string{
					"diff --git a/internal/tui/render.go b/internal/tui/render.go",
					"index 1111111..2222222 100644",
					"--- a/internal/tui/render.go",
					"+++ b/internal/tui/render.go",
					"@@ -1,1 +1,1 @@",
					"+new line",
				}, "\n"),
				Files: []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1}},
				Threads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     1,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
						Body:      threadBody,
						CreatedAt: "2026-05-05T10:00:00Z",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 60, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualWrappedLineCount := 0
	for _, line := range detailView.BufferLines() {
		if strings.Contains(line, "wrap this review comment") || strings.Contains(line, "review comment") {
			actualWrappedLineCount++
		}
	}
	if actualWrappedLineCount < 2 {
		t.Fatalf("expected the review thread body to wrap onto multiple visible lines, actual %d in %q", actualWrappedLineCount, detailView.Buffer())
	}
}

func reviewDiffDocumentRowCountForLine(document detailDocument, lineIndex int) int {
	count := 0
	for _, row := range document.rows {
		if row.line == lineIndex {
			count++
		}
	}
	return count
}
