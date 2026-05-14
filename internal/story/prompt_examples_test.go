package story

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBundledPromptExamples_GivenTheStoryReviewPromptDirectory_WhenReadingTheExamples_ThenEachFileIsPresentAndNonEmpty(t *testing.T) {
	expectedExampleFileNames := []string{
		"caveman.md",
		"default.md",
		"emoji.md",
		"sanderson.md",
	}
	exampleDirectory := filepath.Join("..", "..", "prompts", "story-review")

	for _, fileName := range expectedExampleFileNames {
		contents, actualErr := os.ReadFile(filepath.Join(exampleDirectory, fileName))

		then_noError(t, actualErr)
		if strings.TrimSpace(string(contents)) == "" {
			t.Fatalf("expected %q to be non-empty", fileName)
		}
	}
}

func TestBundledPromptExamples_GivenTheDefaultStoryReviewPrompt_WhenReadingTheBundledFile_ThenItMatchesTheCodeDefault(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "prompts", "story-review", "default.md"))

	then_noError(t, actualErr)
	if actual := strings.TrimSpace(string(contents)); actual != DefaultPrompt() {
		t.Fatalf("expected bundled default prompt %q, actual %q", DefaultPrompt(), actual)
	}
}
