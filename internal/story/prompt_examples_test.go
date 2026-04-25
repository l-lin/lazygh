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
