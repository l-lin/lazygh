package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestGlamourMarkdownRenderer_GivenALongParagraphAndDisabledWordWrap_WhenRendering_ThenItDoesNotInsertSoftWrapNewlines(t *testing.T) {
	renderer := glamourMarkdownRenderer{}
	markdown := strings.Repeat("No soft wrap should appear in this paragraph. ", 4)

	actual, actualErr := renderer.Render(markdown, disabledMarkdownWordWrap)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 200)
	if actualDocument.lineCount() != 1 {
		t.Fatalf("expected a single visible paragraph line, actual %d in %q", actualDocument.lineCount(), actual)
	}
}

func TestGlamourMarkdownRenderer_GivenALongParagraphAndPositiveWidth_WhenRendering_ThenItInsertsSoftWrapNewlines(t *testing.T) {
	renderer := glamourMarkdownRenderer{}
	markdown := strings.Repeat("Wrap this paragraph on words. ", 6)

	actual, actualErr := renderer.Render(markdown, 20)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 200)
	if actualDocument.lineCount() < 2 {
		t.Fatalf("expected multiple visible paragraph lines, actual %d in %q", actualDocument.lineCount(), actual)
	}
}

func TestLayout_GivenDescriptionTabWithALongMarkdownParagraph_WhenBuildingViewZeroDocument_ThenItWordWrapsTheDetailBody(t *testing.T) {
	markdown := strings.Repeat("description ", 24)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {Title: "First PR", Number: 42, Body: markdown, BaseRefName: "main", HeadRefName: "feature/wrap", State: "OPEN"},
		},
	})
	gui := given_headlessGuiWithSize(t, 40, 20)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualDescriptionLineCount := 0
	for _, line := range detailView.BufferLines() {
		if strings.Contains(line, "description") {
			actualDescriptionLineCount++
		}
	}
	if actualDescriptionLineCount < 2 {
		t.Fatalf("expected the description to wrap onto multiple visible lines, actual %d in %q", actualDescriptionLineCount, detailView.Buffer())
	}
}

func TestLayout_GivenCommentsTabWithALongMarkdownComment_WhenBuildingViewZeroDocument_ThenItWrapsTheCommentBody(t *testing.T) {
	markdown := strings.Repeat("wrap this comment body ", 18)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				BaseRefName: "main",
				HeadRefName: "feature/wrap-comments",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					CreatedAt: "2026-05-05T10:00:00Z",
					Body:      markdown,
				}},
			},
		},
	})
	subject.activeDetailTab = CommentsDetailTab
	gui := given_headlessGuiWithSize(t, 60, 20)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualWrappedLineCount := 0
	for _, line := range detailView.BufferLines() {
		if strings.Contains(line, "wrap this comment body") || strings.Contains(line, "comment body") {
			actualWrappedLineCount++
		}
	}
	if actualWrappedLineCount < 2 {
		t.Fatalf("expected the comment body to wrap onto multiple visible lines, actual %d in %q", actualWrappedLineCount, detailView.Buffer())
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
