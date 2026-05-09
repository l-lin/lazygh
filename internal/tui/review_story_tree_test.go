package tui

import "testing"

func TestReviewStoryChapterLabel_GivenATitledChapter_WhenFormatting_ThenItPrefixesOnlyTheNumber(t *testing.T) {
	actual := reviewStoryChapterLabel(0, "The Renderer Wakes", 1)
	expected := "1 - The Renderer Wakes (1 file)"

	if actual != expected {
		t.Fatalf("expected chapter label %q, actual %q", expected, actual)
	}
}

func TestReviewStoryChapterLabel_GivenBlankOrChapterPrefixedTitles_WhenFormatting_ThenItNormalizesToTheChapterNumber(t *testing.T) {
	testCases := []struct {
		name      string
		title     string
		fileCount int
		expected  string
	}{
		{name: "blank title", title: "", fileCount: 0, expected: "2"},
		{name: "already prefixed title", title: "Chapter 2 - The Model Answers", fileCount: 2, expected: "2 - The Model Answers (2 files)"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := reviewStoryChapterLabel(1, testCase.title, testCase.fileCount)
			if actual != testCase.expected {
				t.Fatalf("expected chapter label %q, actual %q", testCase.expected, actual)
			}
		})
	}
}
