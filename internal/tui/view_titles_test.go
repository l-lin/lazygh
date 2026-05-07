package tui

import "testing"

func TestDetailTabLabel_GivenEachBrowserTab_WhenFormatting_ThenItPrefixesTheLabelWithAnIcon(t *testing.T) {
	testCases := []struct {
		name     string
		tab      DetailTab
		expected string
	}{
		{name: "description", tab: DescriptionDetailTab, expected: " Description"},
		{name: "comments", tab: CommentsDetailTab, expected: " Comments"},
		{name: "commits", tab: CommitsDetailTab, expected: " Commits"},
		{name: "changes", tab: ChangesDetailTab, expected: " Changes"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := testCase.tab.Label()

			if actual != testCase.expected {
				t.Fatalf("expected label %q, actual %q", testCase.expected, actual)
			}
		})
	}
}

func TestUserViewTitle_GivenBrowserMode_WhenFormatting_ThenItShowsAnIcon(t *testing.T) {
	subject := NewProgram()

	actual := subject.userViewTitle()

	if actual != "[1]- Connected user" {
		t.Fatalf("expected title %q, actual %q", "[1]- Connected user", actual)
	}
}

func TestReviewModeTitles_GivenTheRepurposedViews_WhenFormatting_ThenEachViewGetsAnIcon(t *testing.T) {
	expected := map[string]string{
		"metadata":    "[1]-󰋼 Metadata",
		"files":       "[2]- Files",
		"chapters":    "[2]- Chapters",
		"description": "[0]- Description",
		"diff":        "[0]- Diff",
		"chapter":     "[0]- Chapter",
	}

	actual := map[string]string{
		"metadata":    reviewModeMetadataTitle,
		"files":       reviewModeFilesTitle,
		"chapters":    reviewModeChaptersTitle,
		"description": reviewModeDescriptionTitle,
		"diff":        reviewModeDiffTitle,
		"chapter":     reviewModeChapterTitle,
	}

	for key, expectedValue := range expected {
		if actual[key] != expectedValue {
			t.Fatalf("expected %s title %q, actual %q", key, expectedValue, actual[key])
		}
	}
}
