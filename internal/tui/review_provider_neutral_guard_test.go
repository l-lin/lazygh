package tui

import (
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"testing"
)

func TestRefactorGuard_GivenReviewAndReactionFiles_WhenScanning_ThenTheyDoNotImportGithubcli(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGoFiles(t, ".", []string{"github.com/l-lin/lazygh/internal/githubcli"})
	guardedMatches := map[string]bool{}
	for _, match := range []string{
		"pending_pull_request_review.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_build.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"reaction_picker.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"reaction_remove.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"reaction_target.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_diff_loader.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_diff_model_support.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_diff_model.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_diff_render.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_diff_target.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_diff_team_owners.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_inline_comment.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_inline_conversation.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_session.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_story_url.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_story.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_submit.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"review_url.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
	} {
		guardedMatches[match] = true
	}

	remainingMatches := make([]string, 0)
	for _, match := range actualMatches {
		if guardedMatches[match] {
			remainingMatches = append(remainingMatches, match)
		}
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected review, story-review, build, and reaction files to stop importing githubcli, actual %v", remainingMatches)
	}
}

func TestBuildReviewDiffData_GivenDomainDiff_WhenBuilding_ThenItKeepsTheProviderNeutralThreads(t *testing.T) {
	diff := githubdomain.PullRequestDiff{
		UnifiedDiff: "diff --git a/main.go b/main.go\n@@ -1 +1,2 @@\n line\n+added",
		Files:       []githubdomain.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1, Patch: "@@ -1 +1,2 @@\n line\n+added"}},
		Threads:     []githubdomain.PullRequestReviewThread{{ID: "thread-42", Path: "main.go", Line: 2, DiffSide: "RIGHT", Comments: []githubdomain.PullRequestComment{{ID: "comment-42", Body: "Ship it"}}}},
	}

	actual := buildReviewDiffData(diff)

	if len(actual.Files) != 1 || len(actual.Files[0].Threads) != 1 {
		t.Fatalf("expected one provider-neutral diff thread, actual %+v", actual.Files)
	}
}

func TestRenderReactionGroup_GivenDomainReactionGroup_WhenRendering_ThenItShowsTheEmojiAndCount(t *testing.T) {
	actual := renderReactionGroup(githubdomain.ReactionGroup{Content: githubdomain.ReactionContentHeart, TotalCount: 2})

	if actual == "" {
		t.Fatal("expected a rendered reaction pill")
	}
}
