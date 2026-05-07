package tui

import "testing"

func TestIconFileForPath_GivenKnownAndUnknownPaths_WhenResolving_ThenItUsesTheCentralCatalog(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		expected string
	}{
		{name: "go by filename", filePath: "go.mod", expected: iconFileGo},
		{name: "go by extension", filePath: "internal/tui/render.go", expected: iconFileGo},
		{name: "markdown by extension", filePath: "README.md", expected: iconFileMarkdown},
		{name: "unknown file", filePath: "notes.txt", expected: iconFile},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := iconFileForPath(testCase.filePath)

			if actual != testCase.expected {
				t.Fatalf("expected icon %q, actual %q", testCase.expected, actual)
			}
		})
	}
}

func TestNotificationIcons_GivenTheCentralCatalog_WhenReadingTheKnownKinds_ThenTheyRemainDiscoverable(t *testing.T) {
	if iconNotificationPullRequest != iconPullRequest {
		t.Fatalf("expected pull request notification icon %q, actual %q", iconPullRequest, iconNotificationPullRequest)
	}
	for name, icon := range map[string]string{
		"issue":   iconNotificationIssue,
		"release": iconNotificationRelease,
	} {
		if icon == "" {
			t.Fatalf("expected notification icon for %s to be non-empty", name)
		}
	}
}
