package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestDetailImageSourceReadModel_GivenAReviewStoryChapter_WhenResolvingSources_ThenItBuildsTheStoryMarkdownFromTheSnapshot(t *testing.T) {
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	})
	subject := detailImageSourceReadModel{
		reviewModeActive:        true,
		reviewShowsStoryChapter: true,
		summary:                 summary,
		summaryKnown:            true,
		reviewStoryChapter: reviewStoryChapter{
			ID:        "chapter-1",
			Title:     "Architecture",
			Narrative: "![Diagram](./docs/diagram.png)",
		},
		reviewStoryChapterKnown: true,
	}

	actual := subject.sources()
	if len(actual) != 1 {
		t.Fatalf("expected one story image source, actual %d", len(actual))
	}
	if !strings.Contains(actual[0].key, "story:acme/widgets#42:chapter-1") {
		t.Fatalf("expected story source key to keep the chapter identity, actual %q", actual[0].key)
	}
	if actual[0].repository != "acme/widgets" {
		t.Fatalf("expected story source repository %q, actual %q", "acme/widgets", actual[0].repository)
	}
	if actual[0].markdown != "# Architecture\n\n![Diagram](./docs/diagram.png)" {
		t.Fatalf("expected story markdown %q, actual %q", "# Architecture\n\n![Diagram](./docs/diagram.png)", actual[0].markdown)
	}
}

func TestDetailImageSourceReadModel_GivenASelectedReviewDiffFile_WhenResolvingSources_ThenItKeepsTheResolvedFileIndexOnApplyTargets(t *testing.T) {
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	})
	file := buildReviewDiffData(given_reviewSessionPullRequestDiffWithComments()).Files[0]
	subject := detailImageSourceReadModel{
		reviewModeActive:    true,
		summary:             summary,
		summaryKnown:        true,
		reviewDiffFile:      file,
		reviewDiffFileKnown: true,
		reviewDiffFileIndex: 0,
	}

	actual := subject.sources()
	if len(actual) != 1 {
		t.Fatalf("expected one review-diff image source, actual %d", len(actual))
	}
	if actual[0].applyTarget.kind != detailImageHTMLApplyKindPullRequestDiffThreadComment {
		t.Fatalf("expected a review-diff thread apply target, actual %+v", actual[0].applyTarget)
	}
	if actual[0].applyTarget.fileIndex != 0 {
		t.Fatalf("expected review-diff file index %d, actual %d", 0, actual[0].applyTarget.fileIndex)
	}
	if actual[0].markdown != "Render thread" {
		t.Fatalf("expected review-diff markdown %q, actual %q", "Render thread", actual[0].markdown)
	}
}

func TestDetailImageSourceReadModel_GivenANotificationIssue_WhenResolvingSources_ThenItBuildsTheIssueSourceFromTheSnapshot(t *testing.T) {
	subject := detailImageSourceReadModel{
		issueRepository:  "acme/widgets",
		issueNumber:      77,
		issueKnown:       true,
		issueDetail:      githubcli.ToDomainIssueDetail(githubcli.IssueDetail{Body: "![Diagram](https://example.com/diagram.png)", BodyHTML: ""}),
		issueDetailKnown: true,
	}

	actual := subject.sources()
	if len(actual) != 1 {
		t.Fatalf("expected one issue image source, actual %d", len(actual))
	}
	if actual[0].applyTarget.kind != detailImageHTMLApplyKindIssue {
		t.Fatalf("expected an issue apply target, actual %+v", actual[0].applyTarget)
	}
	if actual[0].applyTarget.cacheKey != "acme/widgets#77" {
		t.Fatalf("expected issue cache key %q, actual %q", "acme/widgets#77", actual[0].applyTarget.cacheKey)
	}
	if actual[0].markdown != "![Diagram](https://example.com/diagram.png)" {
		t.Fatalf("expected issue markdown %q, actual %q", "![Diagram](https://example.com/diagram.png)", actual[0].markdown)
	}
}
