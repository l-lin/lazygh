package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestAppliedSearchFooterText_GivenAnEmptyDetailQueryInReviewMode_WhenComputingTheFooter_ThenItSkipsRenderingTheDiffBody(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"Render thread": "Rendered render thread"}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = renderer
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiffWithComments())}
	subject.startReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_perf")

	actual := subject.footerPresenter().paneFooterStateFor(FocusDetailView).Text()

	if actual != "" {
		t.Fatalf("expected empty footer text, actual %q", actual)
	}
	if renderer.callCount != 0 {
		t.Fatalf("expected no markdown rendering for an empty detail query, actual %d", renderer.callCount)
	}
}

func TestCurrentDetailDocument_GivenTheSameReviewDiff_WhenBuildingItTwice_ThenItReusesTheCachedDiffRendering(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"Render thread": "Rendered render thread"}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = renderer
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiffWithComments())}
	subject.startReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_perf")

	firstDocument := subject.currentDetailDocument(nil)
	secondDocument := subject.currentDetailDocument(nil)

	if string(firstDocument.text) != string(secondDocument.text) {
		t.Fatalf("expected cached review diff text %q, actual %q", string(firstDocument.text), string(secondDocument.text))
	}
	if renderer.callCount != 1 {
		t.Fatalf("expected one markdown render for the cached review diff document, actual %d", renderer.callCount)
	}
}

func TestCurrentDetailDocument_GivenAStoryReviewDiffAlreadyCached_WhenBuildingItRepeatedly_ThenItAvoidsHotPathAllocations(t *testing.T) {
	subject := given_benchmarkStoryReviewProgram()
	storyData := buildReviewStoryData(story.Review{
		Summary: "A calmer way to review the pull request.",
		Chapters: []story.Chapter{
			{ID: "chapter-1", Title: "The Renderer Wakes", Narrative: "## Chapter 1", Files: []string{"internal/tui/render.go"}},
			{ID: "chapter-2", Title: "The Model Answers", Narrative: "## Chapter 2", Files: []string{"internal/tui/model.go"}},
		},
	}, subject.pullRequestDiffCache["acme/widgets#42"].data.Files)
	subject.startStoryReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_story", storyData)
	subject.navigationState.reviewSession.selectedFileTreeRow = 1

	_ = subject.currentDetailDocument(nil)
	actual := testing.AllocsPerRun(20, func() {
		_ = subject.currentDetailDocument(nil)
	})

	if actual > 60 {
		t.Fatalf("expected cached story-review diff lookup to stay below the regression ceiling, actual %.2f allocs/run", actual)
	}
}
