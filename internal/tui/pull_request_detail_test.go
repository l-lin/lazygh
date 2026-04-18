package tui

import (
	"errors"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestRenderPullRequestDetail_GivenRichMetadata_WhenFormatting_ThenItShowsTheStructuredMetadataBlock(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered markdown body"}
	summary := githubcli.PullRequest{
		Title:      "Fallback title",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	}
	detail := githubcli.PullRequestDetail{
		Title:            "Add a real detail pane",
		Number:           42,
		Body:             "## Summary\n\n- render markdown",
		Author:           &githubcli.PullRequestAuthor{Login: "octocat"},
		State:            "OPEN",
		IsDraft:          false,
		CreatedAt:        "2026-04-18T10:00:00Z",
		UpdatedAt:        "2026-04-18T12:30:00Z",
		Labels:           []githubcli.PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
		BaseRefName:      "main",
		HeadRefName:      "feature/detail",
		MergeStateStatus: "CLEAN",
		Mergeable:        "MERGEABLE",
		Additions:        12,
		Deletions:        3,
		ChangedFiles:     5,
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{
			{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Name: "test", Status: "COMPLETED", Conclusion: "FAILURE"},
		},
	}

	actual := renderPullRequestDetail(summary, detail, renderer, 48)

	for _, expected := range []string{
		"Add a real detail pane #42",
		"Status: OPEN main ← feature/detail",
		"Repo: acme/widgets",
		"Author: @octocat",
		"Created: 2026-04-18 10:00 UTC",
		"Updated: 2026-04-18 12:30 UTC",
		"Labels: bug, backend",
		"Merge Status: CLEAN",
		"Checks: 1 passing, 1 failing",
		"Mergeable: yes",
		"Changes: 5 files +12 -3",
		"Rendered markdown body",
		"Comments",
		"No comments yet.",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actual)
		}
	}
	if renderer.lastWidth != 48 {
		t.Fatalf("expected width %d, actual %d", 48, renderer.lastWidth)
	}
	if renderer.lastMarkdown != "## Summary\n\n- render markdown" {
		t.Fatalf("expected markdown %q, actual %q", "## Summary\n\n- render markdown", renderer.lastMarkdown)
	}
}

func TestRenderPullRequestDetail_GivenComments_WhenFormatting_ThenItKeepsUsernamesClearlyVisible(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Body":          "Rendered body",
		"**Ship it**":   "Rendered comment one",
		"Needs changes": "Rendered comment two",
	}}
	summary := githubcli.PullRequest{Number: 7, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{
		Title:    "Comment showcase",
		Number:   7,
		Body:     "Body",
		Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, CreatedAt: "2026-04-18T13:00:00Z", Body: "**Ship it**"}, {Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, CreatedAt: "2026-04-18T14:15:00Z", Body: "Needs changes"}},
	}

	actual := renderPullRequestDetail(summary, detail, renderer, 60)

	for _, expected := range []string{"@reviewer-one", "2026-04-18 13:00 UTC", "Rendered comment one", "@reviewer-two", "2026-04-18 14:15 UTC", "Rendered comment two"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actual)
		}
	}
}

func TestGlamourMarkdownRenderer_GivenMarkdownAndNarrowWidth_WhenRendering_ThenItWrapsAndRemovesRawMarkdownMarkers(t *testing.T) {
	renderer := glamourMarkdownRenderer{}

	actual, actualErr := renderer.Render("# Heading\n\n- one two three four five six seven", 18)

	then_noError(t, actualErr)
	if strings.Contains(actual, "# Heading") {
		t.Fatalf("expected rendered markdown to omit raw heading markers, actual %q", actual)
	}
	if !strings.Contains(actual, "Heading") {
		t.Fatalf("expected rendered markdown to contain heading text, actual %q", actual)
	}
	if !strings.Contains(actual, "\n") {
		t.Fatalf("expected wrapped output to contain a newline, actual %q", actual)
	}
}

func TestRenderPullRequestDetail_GivenMarkdownRendererFailure_WhenFormatting_ThenItFallsBackToRawMarkdown(t *testing.T) {
	renderer := &fakeMarkdownRenderer{err: errors.New("boom")}
	summary := githubcli.PullRequest{Number: 9, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Title: "Fallback body", Number: 9, Body: "## Summary\n\n- keep the source"}

	actual := renderPullRequestDetail(summary, detail, renderer, 40)

	for _, expected := range []string{"Markdown rendering failed", "## Summary", "- keep the source"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actual)
		}
	}
}

type fakeMarkdownRenderer struct {
	output       string
	outputs      map[string]string
	err          error
	lastMarkdown string
	lastWidth    int
}

func (renderer *fakeMarkdownRenderer) Render(markdown string, width int) (string, error) {
	renderer.lastMarkdown = markdown
	renderer.lastWidth = width
	if renderer.err != nil {
		return "", renderer.err
	}
	if renderer.outputs != nil {
		if output, ok := renderer.outputs[markdown]; ok {
			return output, nil
		}
	}
	return renderer.output, nil
}
