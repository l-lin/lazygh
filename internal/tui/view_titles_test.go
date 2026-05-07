package tui

import "testing"

func TestDetailTabLabel_GivenEachBrowserTab_WhenFormatting_ThenItPrefixesTheLabelWithAnIcon(t *testing.T) {
	testCases := []struct {
		name     string
		tab      DetailTab
		expected string
	}{
		{name: "description", tab: DescriptionDetailTab, expected: iconDescription + " Description"},
		{name: "comments", tab: CommentsDetailTab, expected: iconComment + " Comments"},
		{name: "commits", tab: CommitsDetailTab, expected: iconCommit + " Commits"},
		{name: "changes", tab: ChangesDetailTab, expected: iconChanges + " Changes"},
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

	expected := "[1]-" + iconUser + " Connected user"
	if actual != expected {
		t.Fatalf("expected title %q, actual %q", expected, actual)
	}
}

func TestReviewModeTitles_GivenTheRepurposedViews_WhenFormatting_ThenEachViewGetsAnIcon(t *testing.T) {
	expected := map[string]string{
		"metadata":    "[1]-" + iconMetadata + " Metadata",
		"files":       "[2]-" + iconDirectory + " Files",
		"chapters":    "[2]-" + iconChapter + " Chapters",
		"description": "[0]-" + iconDescription + " Description",
		"diff":        "[0]-" + iconChanges + " Diff",
		"chapter":     "[0]-" + iconChapter + " Chapter",
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
