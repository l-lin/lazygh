package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func BenchmarkCurrentDetailDocument_GivenReviewModeDiff_WhenRenderingRepeatedly(b *testing.B) {
	subject := given_benchmarkStoryReviewProgram()
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	subject.startReviewSession(summary, "PRR_story")
	subject.clampReviewSessionSelection()
	if len(subject.navigationState.reviewSession.story.Tree.Rows) != 0 {
		b.Fatalf("expected plain review mode to keep no story tree, actual %+v", subject.navigationState.reviewSession.story.Tree.Rows)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = subject.currentDetailDocument(nil)
	}
}

func BenchmarkCurrentDetailDocument_GivenStoryReviewDiff_WhenRenderingRepeatedly(b *testing.B) {
	subject := given_benchmarkStoryReviewProgram()
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	storyData := buildReviewStoryData(story.Review{
		Summary: "A calmer way to review the pull request.",
		Chapters: []story.Chapter{
			{ID: "chapter-1", Title: "The Renderer Wakes", Narrative: "## Chapter 1", Files: []string{"internal/tui/render.go"}},
			{ID: "chapter-2", Title: "The Model Answers", Narrative: "## Chapter 2", Files: []string{"internal/tui/model.go"}},
		},
	}, subject.pullRequestDiffCache["acme/widgets#42"].data.Files)
	subject.startStoryReviewSession(summary, "PRR_story", storyData)
	subject.navigationState.reviewSession.selectedFileTreeRow = 1
	if actualKind := subject.mainViewResolver().ContentKind; actualKind != MainContentKindReviewDiff {
		b.Fatalf("expected story review diff content kind %v, actual %v", MainContentKindReviewDiff, actualKind)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = subject.currentDetailDocument(nil)
	}
}

func given_benchmarkStoryReviewProgram() *Program {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	subject := given_pullRequestCommentProgram(given_panelViewContractBrowserModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = renderer
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Body 42", BaseRefName: "main", HeadRefName: "feature/story", State: "OPEN"})}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiffWithComments())}
	return subject
}
