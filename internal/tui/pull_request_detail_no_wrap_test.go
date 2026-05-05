package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestGlamourMarkdownRenderer_GivenALongParagraph_WhenRendering_ThenItDoesNotInsertSoftWrapNewlines(t *testing.T) {
	renderer := glamourMarkdownRenderer{}
	markdown := strings.Repeat("No soft wrap should appear in this paragraph. ", 4)

	actual, actualErr := renderer.Render(markdown, 20)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 200)
	if actualDocument.lineCount() != 1 {
		t.Fatalf("expected a single visible paragraph line, actual %d in %q", actualDocument.lineCount(), actual)
	}
}

func TestLayout_GivenDescriptionTabWithALongRenderedLine_WhenBuildingViewZeroDocument_ThenItDoesNotWrapTheDetailBody(t *testing.T) {
	longLine := strings.Repeat("very-long-description-segment-", 4)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body 42", BaseRefName: "main", HeadRefName: "feature/no-wrap", State: "OPEN"},
		},
	})
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 42": longLine}}
	gui := given_headlessGuiWithSize(t, 60, 20)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, actualLine := given_detailDocumentLineContaining(t, document, longLine)

	if actual := detailDocumentRowCountForLine(document, lineIndex); actual != 1 {
		t.Fatalf("expected the description line to stay on one rendered row, actual %d for %q", actual, actualLine)
	}
}

func TestLayout_GivenCommentsTabWithALongRenderedCommentLine_WhenBuildingViewZeroDocument_ThenItDoesNotWrapTheCommentBody(t *testing.T) {
	longLine := strings.Repeat("very-long-comment-segment-", 4)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				BaseRefName: "main",
				HeadRefName: "feature/no-wrap-comments",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					CreatedAt: "2026-05-05T10:00:00Z",
					Body:      "Comment 42",
				}},
			},
		},
	})
	subject.activeDetailTab = CommentsDetailTab
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Comment 42": longLine}}
	gui := given_headlessGuiWithSize(t, 60, 20)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, actualLine := given_detailDocumentLineContaining(t, document, longLine)

	if actual := detailDocumentRowCountForLine(document, lineIndex); actual != 1 {
		t.Fatalf("expected the comment line to stay on one rendered row, actual %d for %q", actual, actualLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenALongRenderedCommentLine_WhenFormatting_ThenTheRoundedBoxExpandsToFitTheVisibleBody(t *testing.T) {
	longLine := strings.Repeat("very-long-comment-segment-", 4)
	renderer := &fakeMarkdownRenderer{output: longLine}
	comments := []githubcli.PullRequestComment{{
		Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
		CreatedAt: "2026-05-05T10:00:00Z",
		Body:      "Comment 42",
	}}

	actual := renderPullRequestCommentsTab(comments, nil, nil, renderer, 40)
	actualDocument := newDetailDocument(actual, 240)
	topBorderLineIndex, topBorderLine := given_detailDocumentLineContaining(t, actualDocument, "╭")
	bodyLineIndex, bodyLine := given_detailDocumentLineContaining(t, actualDocument, longLine)

	if bodyLineIndex != topBorderLineIndex+2 {
		t.Fatalf("expected the body line to stay inside the first comment box, actual %q", bodyLine)
	}
	if len([]rune(topBorderLine)) != len([]rune(bodyLine)) {
		t.Fatalf("expected the comment box width %d to match the visible body width %d", len([]rune(topBorderLine)), len([]rune(bodyLine)))
	}
}

func detailDocumentRowCountForLine(document detailDocument, lineIndex int) int {
	count := 0
	for _, row := range document.rows {
		if row.line == lineIndex {
			count++
		}
	}
	return count
}
