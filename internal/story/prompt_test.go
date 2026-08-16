package story

import (
	"strings"
	"testing"
)

func TestDefaultPrompt_GivenNoUserOverride_WhenReading_ThenItDescribesTheNarrativeAndVisualGuidance(t *testing.T) {
	actual := DefaultPrompt()

	for _, expected := range []string{
		"Group the changes into a logical, reviewer-friendly story.",
		"Prefer chapters that reflect one cohesive behavior change, refactor step, or debugging thread. For each chapter, explain what changed, why it changed.",
		"When a visual makes the point clearer, include lightweight cues. Use your judgment and don't overwhelm the reader.",
		"Show logic or an algorithm as pseudocode:",
		"Show component interaction, control flow, or data flow with ASCII diagrams:",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected default prompt to contain %q, actual %q", expected, actual)
		}
	}
}

func TestBuildPrompt_GivenStoryReviewRequest_WhenBuilding_ThenItExplainsHowToEncodeMarkdownNarrativesInJSON(t *testing.T) {
	actual := BuildPrompt(Request{
		Metadata:  Metadata{Number: 42, Title: "Render better stories"},
		DiffItems: []DiffItem{{File: "internal/tui/review_story.go"}},
		DiffText:  "diff --git a/internal/tui/review_story.go b/internal/tui/review_story.go",
	}, "")

	for _, expected := range []string{
		"Markdown inside the `narrative` field is allowed and encouraged.",
		`Use escaped newlines (\n) inside the JSON string`,
		`Escape double quotes inside JSON strings with \" so the response stays valid JSON.`,
		`"narrative": "<markdown chapter narrative>"`,
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected prompt to contain %q, actual %q", expected, actual)
		}
	}
}

func TestBuildPrompt_GivenStoryReviewRequest_WhenBuilding_ThenItIncludesVisualGuidanceForReviewNarratives(t *testing.T) {
	actual := BuildPrompt(Request{
		Metadata:  Metadata{Number: 42, Title: "Render better stories"},
		DiffItems: []DiffItem{{File: "internal/tui/review_story.go"}},
		DiffText:  "diff --git a/internal/tui/review_story.go b/internal/tui/review_story.go",
	}, "")

	for _, expected := range []string{
		"When a visual makes the point clearer, include lightweight cues.",
		"Show runtime control flow as a call tree:",
		"When a visual makes the point clearer, include lightweight cues such as pseudocode, call trees, shallow file trees, or ASCII diagrams.",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected prompt to contain %q, actual %q", expected, actual)
		}
	}
}

func TestBuildPrompt_GivenStoryReviewRequestWithDescriptionAndChangedFiles_WhenBuilding_ThenItRendersTheTemplateSections(t *testing.T) {
	actual := BuildPrompt(Request{
		Metadata: Metadata{
			Number:       42,
			Title:        "Render better stories",
			Body:         "Explains the review framing.",
			Author:       "octocat",
			Base:         "main",
			Head:         "story-prompts",
			URL:          "https://example.test/pr/42",
			Additions:    10,
			Deletions:    2,
			ChangedFiles: 2,
		},
		DiffItems: []DiffItem{{File: "internal/story/prompt.go"}, {File: "prompts/story-review/default.md"}},
		DiffText:  "diff --git a/internal/story/prompt.go b/internal/story/prompt.go",
	}, "")

	for _, expected := range []string{
		"## PR Description\nExplains the review framing.",
		"## Changed Files\n- internal/story/prompt.go\n- prompts/story-review/default.md",
		"# Pull Request #42: Render better stories",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected prompt to contain %q, actual %q", expected, actual)
		}
	}
}
