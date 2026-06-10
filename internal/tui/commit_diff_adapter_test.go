package tui

import (
	"strings"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestBuildCommitDiffReviewData_GivenCommitFilesWithPatches_WhenBuilding_ThenItParsesFilesAndHunks(t *testing.T) {
	subject := githubdomain.CommitDiff{Files: []githubdomain.PullRequestDiffFile{
		{
			Path:       "internal/tui/render.go",
			ChangeType: "modified",
			Additions:  1,
			Deletions:  1,
			Patch: strings.Join([]string{
				"@@ -42,2 +42,2 @@",
				" context line",
				"-old line",
				"+new line",
			}, "\n"),
		},
		{
			Path:       "docs/new.md",
			ChangeType: "added",
			Additions:  2,
			Patch: strings.Join([]string{
				"@@ -0,0 +1,2 @@",
				"+# Title",
				"+Body",
			}, "\n"),
		},
	}}

	actual := buildCommitDiffReviewData(subject)

	if actual.Stats.ChangedFiles != 2 || actual.Stats.Additions != 3 || actual.Stats.Deletions != 1 {
		t.Fatalf("expected stats {files:2 additions:3 deletions:1}, actual %+v", actual.Stats)
	}
	if len(actual.Files) != 2 {
		t.Fatalf("expected 2 files, actual %d", len(actual.Files))
	}
	if actual.Files[0].Path != "internal/tui/render.go" || len(actual.Files[0].Hunks) != 1 {
		t.Fatalf("expected the first parsed file to keep its path and hunks, actual %+v", actual.Files[0])
	}
	if actual.Files[0].Hunks[0].Header != "@@ -42,2 +42,2 @@" {
		t.Fatalf("expected the modified file hunk header %q, actual %q", "@@ -42,2 +42,2 @@", actual.Files[0].Hunks[0].Header)
	}
	if actual.Files[0].Hunks[0].Lines[1].Kind != reviewDiffDeletionLine || actual.Files[0].Hunks[0].Lines[2].Kind != reviewDiffAdditionLine {
		t.Fatalf("expected deletion and addition lines in the modified file, actual %+v", actual.Files[0].Hunks[0].Lines)
	}
	if actual.Files[1].Path != "docs/new.md" || actual.Files[1].ChangeType != reviewDiffChangeTypeAdded || len(actual.Files[1].Hunks) != 1 {
		t.Fatalf("expected the added file to keep its path, change type, and hunks, actual %+v", actual.Files[1])
	}
}
