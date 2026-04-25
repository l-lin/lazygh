package story

import (
	"errors"
	"reflect"
	"testing"
)

func TestBuildCommand_GivenTemplateWithPromptFilePlaceholder_WhenBuilding_ThenItSubstitutesThePlaceholder(t *testing.T) {
	actual := BuildCommand([]string{"pi", "-p", "@{{prompt_file}}"}, "/tmp/story-prompt.txt")

	expected := []string{"pi", "-p", "@/tmp/story-prompt.txt"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected command %v, actual %v", expected, actual)
	}
}

func TestBuildCommand_GivenTemplateWithoutPromptFilePlaceholder_WhenBuilding_ThenItAppendsThePromptFilePath(t *testing.T) {
	actual := BuildCommand([]string{"story-cli", "--prompt-file"}, "/tmp/story-prompt.txt")

	expected := []string{"story-cli", "--prompt-file", "/tmp/story-prompt.txt"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected command %v, actual %v", expected, actual)
	}
}

func TestDecodeReview_GivenJSONWrappedInMarkdownFences_WhenDecoding_ThenItParsesTheStory(t *testing.T) {
	actual, actualErr := DecodeReview("```json\n{\n  \"summary\": \"Overview\",\n  \"chapters\": [{\n    \"id\": \"chapter-1\",\n    \"title\": \"Act I\",\n    \"narrative\": \"A beginning\",\n    \"files\": [\"internal/tui/review.go\"]\n  }]\n}\n```")

	then_noError(t, actualErr)
	expected := Review{
		Summary: "Overview",
		Chapters: []Chapter{{
			ID:        "chapter-1",
			Title:     "Act I",
			Narrative: "A beginning",
			Files:     []string{"internal/tui/review.go"},
		}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected review %+v, actual %+v", expected, actual)
	}
}

func TestNormalizeReview_GivenUnassignedFiles_WhenNormalizing_ThenItAppendsAnUnassignedChapter(t *testing.T) {
	actual, actualErr := NormalizeReview(Review{
		Summary: "Overview",
		Chapters: []Chapter{{
			ID:        "chapter-1",
			Title:     "Act I",
			Narrative: "A beginning",
			Files:     []string{"internal/tui/review.go"},
		}},
	}, []DiffItem{{File: "internal/tui/review.go"}, {File: "internal/tui/model.go"}})

	then_noError(t, actualErr)
	if len(actual.Chapters) != 2 {
		t.Fatalf("expected two chapters, actual %+v", actual.Chapters)
	}
	lastChapter := actual.Chapters[1]
	if lastChapter.ID != UnassignedChapterID {
		t.Fatalf("expected unassigned chapter id %q, actual %q", UnassignedChapterID, lastChapter.ID)
	}
	if !reflect.DeepEqual(lastChapter.Files, []string{"internal/tui/model.go"}) {
		t.Fatalf("expected unassigned files %v, actual %v", []string{"internal/tui/model.go"}, lastChapter.Files)
	}
}

func TestGenerator_GivenMissingAgentCommand_WhenGenerating_ThenItReturnsAnAgentConfigurationError(t *testing.T) {
	subject := NewGenerator(nil)

	_, actualErr := subject.Generate(Config{}, Request{})

	if !errors.Is(actualErr, ErrAgentNotConfigured) {
		t.Fatalf("expected error %v, actual %v", ErrAgentNotConfigured, actualErr)
	}
}

func then_noError(t *testing.T, actualErr error) {
	t.Helper()

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
