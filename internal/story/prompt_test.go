package story

import (
	"strings"
	"testing"
)

func TestDefaultPrompt_GivenNoUserOverride_WhenReading_ThenItAsksForReadableMarkdownNarratives(t *testing.T) {
	actual := DefaultPrompt()

	for _, expected := range []string{
		"Write each chapter narrative as readable markdown",
		"Prefer multiple lines with spacing over one dense sentence",
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
		`"narrative": "## Why\n\n- Explain the behavior shift\n- Call out reviewer checkpoints"`,
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected prompt to contain %q, actual %q", expected, actual)
		}
	}
}
