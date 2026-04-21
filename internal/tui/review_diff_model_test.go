package tui

import (
	"reflect"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestBuildReviewDiffData_GivenSingleFileUnifiedDiff_WhenParsing_ThenItKeepsFileStatsHunksAndLineNumbers(t *testing.T) {
	raw := githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/internal/tui/render.go b/internal/tui/render.go",
			"index 1111111..2222222 100644",
			"--- a/internal/tui/render.go",
			"+++ b/internal/tui/render.go",
			"@@ -10,3 +10,4 @@ func render() {",
			" context line",
			"-old line",
			"+new line",
			"+another line",
			" }",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 2, Deletions: 1}},
	}

	actual := buildReviewDiffData(raw)

	expectedStats := reviewDiffStats{ChangedFiles: 1, Additions: 2, Deletions: 1}
	if actual.Stats != expectedStats {
		t.Fatalf("expected stats %+v, actual %+v", expectedStats, actual.Stats)
	}
	if len(actual.Files) != 1 {
		t.Fatalf("expected 1 parsed file, actual %d", len(actual.Files))
	}

	actualFile := actual.Files[0]
	if actualFile.Path != "internal/tui/render.go" || actualFile.ChangeType != reviewDiffChangeTypeModified || actualFile.Additions != 2 || actualFile.Deletions != 1 {
		t.Fatalf("expected parsed file stats to be preserved, actual %+v", actualFile)
	}
	if len(actualFile.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, actual %d", len(actualFile.Hunks))
	}
	if actualFile.Hunks[0].Header != "@@ -10,3 +10,4 @@ func render() {" {
		t.Fatalf("expected hunk header %q, actual %q", "@@ -10,3 +10,4 @@ func render() {", actualFile.Hunks[0].Header)
	}

	expectedLines := []reviewDiffLine{
		{Kind: reviewDiffContextLine, Text: "context line", LeftLine: 10, RightLine: 10, Side: reviewDiffLineSideBoth},
		{Kind: reviewDiffDeletionLine, Text: "old line", LeftLine: 11, Side: reviewDiffLineSideLeft},
		{Kind: reviewDiffAdditionLine, Text: "new line", RightLine: 11, Side: reviewDiffLineSideRight},
		{Kind: reviewDiffAdditionLine, Text: "another line", RightLine: 12, Side: reviewDiffLineSideRight},
		{Kind: reviewDiffContextLine, Text: "}", LeftLine: 12, RightLine: 13, Side: reviewDiffLineSideBoth},
	}
	if !reflect.DeepEqual(actualFile.Hunks[0].Lines, expectedLines) {
		t.Fatalf("expected parsed lines %+v, actual %+v", expectedLines, actualFile.Hunks[0].Lines)
	}
}

func TestBuildReviewDiffData_GivenQuotedPathsInUnifiedDiff_WhenParsing_ThenItPreservesTheWholePath(t *testing.T) {
	raw := githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			`diff --git "a/docs/my guide.md" "b/docs/my guide.md"`,
			"index 1111111..2222222 100644",
			`--- "a/docs/my guide.md"`,
			`+++ "b/docs/my guide.md"`,
			"@@ -1,1 +1,1 @@",
			"-old title",
			"+new title",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{{Path: "docs/my guide.md", ChangeType: "modified", Additions: 1, Deletions: 1}},
	}

	actual := buildReviewDiffData(raw)

	if len(actual.Files) != 1 {
		t.Fatalf("expected 1 parsed file, actual %d", len(actual.Files))
	}
	if actual.Files[0].Path != "docs/my guide.md" || len(actual.Files[0].Hunks) != 1 {
		t.Fatalf("expected the quoted path and parsed hunk to survive, actual %+v", actual.Files[0])
	}
}

func TestBuildReviewDiffData_GivenMultiFileUnifiedDiff_WhenParsing_ThenItBuildsEachFileInOrder(t *testing.T) {
	raw := githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/cmd/lazygh/main.go b/cmd/lazygh/main.go",
			"index 1111111..2222222 100644",
			"--- a/cmd/lazygh/main.go",
			"+++ b/cmd/lazygh/main.go",
			"@@ -1,1 +1,1 @@",
			"-old main",
			"+new main",
			"diff --git a/internal/tui/render.go b/internal/tui/render.go",
			"index 3333333..4444444 100644",
			"--- a/internal/tui/render.go",
			"+++ b/internal/tui/render.go",
			"@@ -5,1 +5,2 @@",
			" context line",
			"+new render line",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{
			{Path: "cmd/lazygh/main.go", ChangeType: "modified", Additions: 1, Deletions: 1},
			{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 0},
		},
	}

	actual := buildReviewDiffData(raw)

	if len(actual.Files) != 2 {
		t.Fatalf("expected 2 parsed files, actual %d", len(actual.Files))
	}
	if actual.Files[0].Path != "cmd/lazygh/main.go" || actual.Files[1].Path != "internal/tui/render.go" {
		t.Fatalf("expected file order [cmd/lazygh/main.go internal/tui/render.go], actual [%s %s]", actual.Files[0].Path, actual.Files[1].Path)
	}
	if actual.Files[0].Hunks[0].Lines[0].Kind != reviewDiffDeletionLine || actual.Files[1].Hunks[0].Lines[1].Kind != reviewDiffAdditionLine {
		t.Fatalf("expected both files to keep their parsed hunk lines, actual %+v", actual.Files)
	}
}

func TestReviewDiffThreadTargetForLines_GivenARightSideMultiLineSelection_WhenBuilding_ThenItKeepsStartAndEndMetadata(t *testing.T) {
	selectedLines := []reviewDiffLine{
		{Kind: reviewDiffAdditionLine, Text: "new line", RightLine: 11, Side: reviewDiffLineSideRight},
		{Kind: reviewDiffContextLine, Text: "kept line", LeftLine: 12, RightLine: 12, Side: reviewDiffLineSideBoth},
		{Kind: reviewDiffAdditionLine, Text: "newer line", RightLine: 13, Side: reviewDiffLineSideRight},
	}

	actual, ok := reviewDiffThreadTargetForLines("internal/tui/render.go", selectedLines)

	if !ok {
		t.Fatal("expected a valid review thread target")
	}
	expected := reviewDiffThreadTarget{
		Path:        "internal/tui/render.go",
		Line:        13,
		Side:        "RIGHT",
		StartLine:   11,
		StartSide:   "RIGHT",
		SubjectType: "LINE",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected target %+v, actual %+v", expected, actual)
	}
}

func TestBuildReviewDiffFileTree_GivenSingleChildDirectoryChains_WhenProjecting_ThenItCollapsesThemIntoCompactRows(t *testing.T) {
	files := []reviewDiffFile{
		{Path: "internal/tui/render.go"},
		{Path: "cmd/lazygh/main.go"},
	}

	actual := buildReviewDiffFileTree(files)

	expected := []reviewDiffTreeRow{
		{VisibleRowIndex: 0, Depth: 0, Label: "internal/tui/render.go", FileIndex: 0},
		{VisibleRowIndex: 1, Depth: 0, Label: "cmd/lazygh/main.go", FileIndex: 1},
	}
	if !reflect.DeepEqual(actual.Rows, expected) {
		t.Fatalf("expected rows %+v, actual %+v", expected, actual.Rows)
	}
}

func TestBuildReviewDiffFileTree_GivenMixedSiblingDirectories_WhenProjecting_ThenItKeepsBranchRowsAndStableVisibleIndexes(t *testing.T) {
	files := []reviewDiffFile{
		{Path: "internal/tui/render.go"},
		{Path: "internal/tui/model.go"},
		{Path: "internal/githubcli/client.go"},
	}

	actual := buildReviewDiffFileTree(files)

	expected := []reviewDiffTreeRow{
		{VisibleRowIndex: 0, Depth: 0, Label: "internal/", FileIndex: -1},
		{VisibleRowIndex: 1, Depth: 1, Label: "tui/", FileIndex: -1},
		{VisibleRowIndex: 2, Depth: 2, Label: "render.go", FileIndex: 0},
		{VisibleRowIndex: 3, Depth: 2, Label: "model.go", FileIndex: 1},
		{VisibleRowIndex: 4, Depth: 1, Label: "githubcli/client.go", FileIndex: 2},
	}
	if !reflect.DeepEqual(actual.Rows, expected) {
		t.Fatalf("expected rows %+v, actual %+v", expected, actual.Rows)
	}
}

func TestRenderReviewDiffFile_GivenPatchlessBinaryRenamedAndDeletedFiles_WhenRendering_ThenItShowsReadablePlaceholders(t *testing.T) {
	files := []reviewDiffFile{
		{Path: "assets/logo.png", ChangeType: reviewDiffChangeTypeAdded, Placeholder: reviewDiffPlaceholder(reviewDiffFile{Path: "assets/logo.png", ChangeType: reviewDiffChangeTypeAdded})},
		{Path: "docs/guide.md", PreviousPath: "docs/old-guide.md", ChangeType: reviewDiffChangeTypeRenamed, Placeholder: reviewDiffPlaceholder(reviewDiffFile{Path: "docs/guide.md", PreviousPath: "docs/old-guide.md", ChangeType: reviewDiffChangeTypeRenamed})},
		{Path: "README.md", ChangeType: reviewDiffChangeTypeRemoved, Placeholder: reviewDiffPlaceholder(reviewDiffFile{Path: "README.md", ChangeType: reviewDiffChangeTypeRemoved})},
	}

	if !strings.Contains(renderReviewDiffFile(files[0]), "GitHub did not return a textual patch") {
		t.Fatalf("expected a binary-or-patchless placeholder, actual %q", renderReviewDiffFile(files[0]))
	}
	if !strings.Contains(renderReviewDiffFile(files[1]), "Renamed from docs/old-guide.md") {
		t.Fatalf("expected a rename placeholder, actual %q", renderReviewDiffFile(files[1]))
	}
	if !strings.Contains(renderReviewDiffFile(files[2]), "Deleted file README.md") {
		t.Fatalf("expected a deletion placeholder, actual %q", renderReviewDiffFile(files[2]))
	}
}
