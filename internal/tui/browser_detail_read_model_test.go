package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestBrowserDetailReadModel_GivenOverviewSections_WhenResolvingACursor_ThenItKeepsTheSectionBodyContextInsideTheSnapshot(t *testing.T) {
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	})
	detail := githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Number:         42,
		Mergeable:      "MERGEABLE",
		ReviewDecision: "REVIEW_REQUIRED",
		ReviewRequests: []githubcli.PullRequestReviewRequest{{
			RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"},
		}},
	})
	subject := browserDetailReadModel{summary: summary, detail: detail, width: 80}

	cursorLine := browserDescriptionOverviewStartLine(summary, detail) + given_textLineContaining(t, subject.renderOverview(), "@reviewer-requested")
	actual, ok := subject.overviewSectionAtCursor(cursorLine)
	if !ok {
		t.Fatal("expected an overview section at the reviewer line")
	}
	if !actual.inBody {
		t.Fatalf("expected the reviewer line to stay in the overview body, actual %+v", actual)
	}
	if actual.section.overviewBlockTitle != "Reviewers" {
		t.Fatalf("expected the reviewer line to resolve the reviewers block, actual %q", actual.section.overviewBlockTitle)
	}
}

func TestBrowserDetailReadModel_GivenAnInlineThreadConversation_WhenResolvingACursor_ThenItKeepsTheThreadCommentContextInsideTheSnapshot(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"Inline body": "Rendered inline body"}}
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	})
	detail := githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Number: 42,
		InlineCommentThreads: []githubcli.PullRequestReviewThread{{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     43,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				ID:        "PRRC_1",
				Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				Body:      "Inline body",
				CreatedAt: "2026-04-18T10:30:00Z",
				DiffHunk:  "@@ -42,1 +42,1 @@\n-old line\n+new line",
			}},
		}},
	})
	subject := browserDetailReadModel{
		summary:                summary,
		detail:                 detail,
		width:                  80,
		markdownRenderer:       renderer,
		wordWrapEnabled:        true,
		connectedUserLogin:     "octocat",
		collapsedSectionStates: map[string]bool{},
	}

	document := subject.conversationDocument()
	cursorLine := given_textLineContaining(t, document.text, "Rendered inline body")
	actual, ok := subject.conversationSectionAtCursor(cursorLine)
	if !ok {
		t.Fatal("expected a conversation section at the inline thread line")
	}
	if actual.section.inlineThread == nil || actual.section.inlineThread.ID != "thread-1" {
		t.Fatalf("expected inline thread %q, actual %+v", "thread-1", actual.section.inlineThread)
	}
	if !actual.inBody {
		t.Fatalf("expected the inline thread line to stay in the section body, actual %+v", actual)
	}
	if actual.inlineThreadCommentIndex != 0 {
		t.Fatalf("expected inline thread comment index %d, actual %d", 0, actual.inlineThreadCommentIndex)
	}
}

func TestBrowserDetailReadModel_GivenSubmittedReviewBody_WhenRenderingConversations_ThenItIncludesTheReviewSectionInTheCommentsTab(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"looks good but I think you missed RecommendedContentProvider": "Rendered review body"}}
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	})
	detail := githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Number: 42,
		Reviews: []githubcli.PullRequestReview{{
			ID:          "PRR_1",
			Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:        "looks good but I think you missed RecommendedContentProvider",
			State:       "COMMENTED",
			SubmittedAt: "2026-06-15T06:54:59Z",
		}},
	})
	subject := browserDetailReadModel{
		summary:                summary,
		detail:                 detail,
		width:                  80,
		markdownRenderer:       renderer,
		wordWrapEnabled:        true,
		connectedUserLogin:     "octocat",
		collapsedSectionStates: map[string]bool{},
	}

	actual := subject.renderConversationsTab()

	for _, expected := range []string{" Commented review", "Rendered review body", "2026-06-15 06:54 UTC"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected the comments tab to contain %q, actual %q", expected, actual)
		}
	}
}

func given_textLineContaining(t *testing.T, text string, expectedSegment string) int {
	t.Helper()

	for index, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if strings.Contains(line, expectedSegment) {
			return index
		}
	}
	t.Fatalf("expected text to contain %q, actual %q", expectedSegment, text)
	return -1
}
