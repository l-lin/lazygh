package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestCurrentDetailDocument_GivenTheCommentsTabForTheSamePullRequest_WhenBuildingItTwice_ThenItReusesTheCachedDocument(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Comment body": "Rendered comment body",
		"Inline body":  "Rendered inline body",
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = renderer
	subject.activeDetailTab = CommentsDetailTab
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{
		Title:  "First PR",
		Number: 42,
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Comment body",
			CreatedAt: "2026-04-18T10:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{
			Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:         "Inline body",
			CreatedAt:    "2026-04-18T10:30:00Z",
			Path:         "internal/tui/render.go",
			Line:         43,
			OriginalLine: 43,
			Side:         "RIGHT",
			DiffHunk:     "@@ -42,1 +42,1 @@\n-old line\n+new line",
		}},
	}}

	firstDocument := subject.currentDetailDocument(nil)
	secondDocument := subject.currentDetailDocument(nil)

	if string(firstDocument.text) != string(secondDocument.text) {
		t.Fatalf("expected cached detail document %q, actual %q", string(firstDocument.text), string(secondDocument.text))
	}
	if renderer.callCount != 2 {
		t.Fatalf("expected one markdown render per unique comment body, actual %d", renderer.callCount)
	}
}
