package tui

import (
	"reflect"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
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

func TestBuildReviewDiffFileTree_GivenSingleChildDirectoryChains_WhenProjecting_ThenItCollapsesDirectoriesButKeepsFileRowsToBasenames(t *testing.T) {
	files := []reviewDiffFile{
		{Path: "internal/tui/render.go"},
		{Path: "cmd/lazygh/main.go"},
	}

	actual := buildReviewDiffFileTree(files)

	expected := []reviewDiffTreeRow{
		{VisibleRowIndex: 0, Depth: 0, Label: "internal/tui/", FileIndex: -1},
		{VisibleRowIndex: 1, Depth: 1, Label: "render.go", FileIndex: 0},
		{VisibleRowIndex: 2, Depth: 0, Label: "cmd/lazygh/", FileIndex: -1},
		{VisibleRowIndex: 3, Depth: 1, Label: "main.go", FileIndex: 1},
	}
	if !reflect.DeepEqual(actual.Rows, expected) {
		t.Fatalf("expected rows %+v, actual %+v", expected, actual.Rows)
	}
}

func TestBuildReviewDiffFileTree_GivenMixedSiblingDirectories_WhenProjecting_ThenItKeepsBranchRowsAndFileBasenamesWithStableVisibleIndexes(t *testing.T) {
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
		{VisibleRowIndex: 4, Depth: 1, Label: "githubcli/", FileIndex: -1},
		{VisibleRowIndex: 5, Depth: 2, Label: "client.go", FileIndex: 2},
	}
	if !reflect.DeepEqual(actual.Rows, expected) {
		t.Fatalf("expected rows %+v, actual %+v", expected, actual.Rows)
	}
}

func TestReviewDiffTreeItems_GivenDirectoriesAndFiles_WhenFormatting_ThenItPrefixesRowsWithIcons(t *testing.T) {
	tree := reviewDiffTree{Rows: []reviewDiffTreeRow{
		{VisibleRowIndex: 0, Depth: 0, Label: "internal/", FileIndex: -1},
		{VisibleRowIndex: 1, Depth: 1, Label: "tui/", FileIndex: -1},
		{VisibleRowIndex: 2, Depth: 2, Label: "notes.txt", FileIndex: 0},
	}}

	actual := reviewDiffTreeItems(tree, nil)

	expected := []Item{
		{Title: "󰝰 internal/"},
		{Title: "  󰝰 tui/"},
		{Title: "     notes.txt"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected items %+v, actual %+v", expected, actual)
	}
}

func TestReviewDiffTreeItems_GivenKnownFileTypes_WhenFormatting_ThenItUsesSpecificFileIcons(t *testing.T) {
	tree := reviewDiffTree{Rows: []reviewDiffTreeRow{
		{VisibleRowIndex: 0, Depth: 0, Label: "main.go", FileIndex: 0},
		{VisibleRowIndex: 1, Depth: 0, Label: "complete_health_reminders_job.rb", FileIndex: 1},
		{VisibleRowIndex: 2, Depth: 0, Label: "patient-context.yaml", FileIndex: 2},
	}}

	actual := reviewDiffTreeItems(tree, nil)

	expected := []Item{
		{Title: " main.go"},
		{Title: " complete_health_reminders_job.rb"},
		{Title: " patient-context.yaml"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected items %+v, actual %+v", expected, actual)
	}
}

func TestRenderReviewDiffFile_GivenChangedFile_WhenRendering_ThenItPrefixesTheHeaderWithTheFileIcon(t *testing.T) {
	file := reviewDiffFile{
		Path:        "engines/preventive_continuous_care/app/jobs/preventive_continuous_care/complete_health_reminders_job.rb:43",
		Additions:   1,
		Deletions:   1,
		ChangeType:  reviewDiffChangeTypeModified,
		Placeholder: "No textual diff is available.",
	}

	actualDocument := newDetailDocument(renderReviewDiffFile(file, nil, 160), 160)
	actualHeader := string(actualDocument.lines[0])

	expected := detailInlineCommentLocationIcon + " engines/preventive_continuous_care/app/jobs/preventive_continuous_care/complete_health_reminders_job.rb:43  +1  -1"
	if actualHeader != expected {
		t.Fatalf("expected review diff header %q, actual %q", expected, actualHeader)
	}
}

func TestRenderReviewDiffFile_GivenPatchlessBinaryRenamedAndDeletedFiles_WhenRendering_ThenItShowsReadablePlaceholders(t *testing.T) {
	files := []reviewDiffFile{
		{Path: "assets/logo.png", ChangeType: reviewDiffChangeTypeAdded, Placeholder: reviewDiffPlaceholder(reviewDiffFile{Path: "assets/logo.png", ChangeType: reviewDiffChangeTypeAdded})},
		{Path: "docs/guide.md", PreviousPath: "docs/old-guide.md", ChangeType: reviewDiffChangeTypeRenamed, Placeholder: reviewDiffPlaceholder(reviewDiffFile{Path: "docs/guide.md", PreviousPath: "docs/old-guide.md", ChangeType: reviewDiffChangeTypeRenamed})},
		{Path: "README.md", ChangeType: reviewDiffChangeTypeRemoved, Placeholder: reviewDiffPlaceholder(reviewDiffFile{Path: "README.md", ChangeType: reviewDiffChangeTypeRemoved})},
	}

	if !strings.Contains(renderReviewDiffFile(files[0], nil, 120), "GitHub did not return a textual patch") {
		t.Fatalf("expected a binary-or-patchless placeholder, actual %q", renderReviewDiffFile(files[0], nil, 120))
	}
	if !strings.Contains(renderReviewDiffFile(files[1], nil, 120), "Renamed from docs/old-guide.md") {
		t.Fatalf("expected a rename placeholder, actual %q", renderReviewDiffFile(files[1], nil, 120))
	}
	if !strings.Contains(renderReviewDiffFile(files[2], nil, 120), "Deleted file README.md") {
		t.Fatalf("expected a deletion placeholder, actual %q", renderReviewDiffFile(files[2], nil, 120))
	}
}

func TestBuildReviewDiffData_GivenReviewThreads_WhenParsing_ThenItKeepsThemGroupedOnTheMatchingFile(t *testing.T) {
	raw := githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/internal/tui/render.go b/internal/tui/render.go",
			"index 1111111..2222222 100644",
			"--- a/internal/tui/render.go",
			"+++ b/internal/tui/render.go",
			"@@ -10,1 +10,2 @@",
			" context line",
			"+new line",
			"diff --git a/internal/tui/model.go b/internal/tui/model.go",
			"index 3333333..4444444 100644",
			"--- a/internal/tui/model.go",
			"+++ b/internal/tui/model.go",
			"@@ -20,1 +20,0 @@",
			"-old model",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{
			{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 0},
			{Path: "internal/tui/model.go", ChangeType: "modified", Additions: 0, Deletions: 1},
		},
		Threads: []githubcli.PullRequestReviewThread{
			{
				ID:       "thread-1",
				Path:     "internal/tui/render.go",
				Line:     11,
				DiffSide: "RIGHT",
				Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Please split this", CreatedAt: "2026-04-20T10:00:00Z"}},
			},
			{
				ID:           "thread-2",
				Path:         "internal/tui/model.go",
				OriginalLine: 20,
				DiffSide:     "LEFT",
				Comments:     []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, Body: "Remove this", CreatedAt: "2026-04-20T10:05:00Z"}},
			},
		},
	}

	actual := buildReviewDiffData(raw)

	if len(actual.Files) != 2 {
		t.Fatalf("expected 2 parsed files, actual %d", len(actual.Files))
	}
	if len(actual.Files[0].Threads) != 1 || len(actual.Files[1].Threads) != 1 {
		t.Fatalf("expected one grouped thread per file, actual %+v", actual.Files)
	}
	if actual.Files[0].Threads[0].Comments[0].Body != "Please split this" || actual.Files[1].Threads[0].Comments[0].Body != "Remove this" {
		t.Fatalf("expected grouped thread bodies to stay with their files, actual %+v", actual.Files)
	}
}

func TestBuildReviewDiffData_GivenAThreadWithoutAMatchingFile_WhenParsing_ThenItCreatesAPlaceholderFileForTheThread(t *testing.T) {
	raw := githubcli.PullRequestDiff{Threads: []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "docs/spec.md",
		Line:     7,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Missing patch context", CreatedAt: "2026-04-20T10:00:00Z"}},
	}}}

	actual := buildReviewDiffData(raw)

	if len(actual.Files) != 1 {
		t.Fatalf("expected 1 file, actual %d", len(actual.Files))
	}
	if actual.Files[0].Path != "docs/spec.md" {
		t.Fatalf("expected placeholder file path %q, actual %q", "docs/spec.md", actual.Files[0].Path)
	}
	if len(actual.Files[0].Threads) != 1 {
		t.Fatalf("expected 1 grouped thread, actual %+v", actual.Files[0].Threads)
	}
	if !strings.Contains(actual.Files[0].Placeholder, "No textual diff is available for docs/spec.md.") {
		t.Fatalf("expected placeholder to mention the thread-only path, actual %q", actual.Files[0].Placeholder)
	}
}

func TestRenderReviewDiffFile_GivenInlineReviewThreads_WhenRendering_ThenItPlacesTheThreadBoxesAfterTheMatchingDiffLine(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered thread body"}
	file := reviewDiffFile{
		Path:       "internal/tui/render.go",
		Additions:  1,
		Deletions:  0,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,1 +10,2 @@",
			Lines: []reviewDiffLine{
				{Kind: reviewDiffContextLine, Text: "context line", LeftLine: 10, RightLine: 10, Side: reviewDiffLineSideBoth},
				{Kind: reviewDiffAdditionLine, Text: "new line", RightLine: 11, Side: reviewDiffLineSideRight},
				{Kind: reviewDiffContextLine, Text: "tail line", LeftLine: 11, RightLine: 12, Side: reviewDiffLineSideBoth},
			},
		}},
		Threads: []reviewDiffThread{{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     11,
			Side:     reviewDiffLineSideRight,
			Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Thread body", CreatedAt: "2026-04-20T10:00:00Z"}},
		}},
	}

	actualDocument := newDetailDocument(renderReviewDiffFile(file, renderer, 96), 96)
	additionLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "new line")
	metadataLineIndex, metadataLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-one")
	bodyLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "Rendered thread body")
	tailLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "tail line")
	borderLineIndex, borderLine := given_detailDocumentLineContaining(t, actualDocument, "╭")

	if metadataLineIndex <= additionLineIndex || metadataLineIndex >= tailLineIndex {
		t.Fatalf("expected the inline thread to render after the matching diff line and before the next diff line, actual %q", string(actualDocument.text))
	}
	if metadataLineIndex != borderLineIndex+1 {
		t.Fatalf("expected the metadata line to render inside the box immediately after the top border, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, detailCommentsIcon+" @reviewer-one") {
		t.Fatalf("expected the metadata line to contain the author badge, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, "2026-04-20 10:00 UTC") {
		t.Fatalf("expected the metadata line to keep the timestamp on the same line, actual %q", metadataLine)
	}
	if bodyLineIndex != metadataLineIndex+1 {
		t.Fatalf("expected the thread body to render after the metadata line, actual %q", string(actualDocument.text))
	}
	borderIndex := given_runeIndexInString(t, borderLine, "╭")
	if actualStylePrefix := actualDocument.lineStylePrefixes[borderLineIndex][borderIndex]; actualStylePrefix != foregroundColorEscape(theme.InactiveBorderHex) {
		t.Fatalf("expected the thread comment box border prefix %q, actual %q", foregroundColorEscape(theme.InactiveBorderHex), actualStylePrefix)
	}
	if renderer.lastMarkdown != "Thread body" {
		t.Fatalf("expected markdown renderer input %q, actual %q", "Thread body", renderer.lastMarkdown)
	}
}

func TestRenderReviewDiffFile_GivenInlineReviewThreadStatusBadges_WhenRendering_ThenItShowsThemOnTheHeaderLine(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered thread body"}
	file := reviewDiffFile{
		Path:       "internal/tui/render.go",
		Additions:  1,
		Deletions:  0,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,1 +10,2 @@",
			Lines:  []reviewDiffLine{{Kind: reviewDiffAdditionLine, Text: "new line", RightLine: 11, Side: reviewDiffLineSideRight}},
		}},
		Threads: []reviewDiffThread{{
			ID:         "thread-1",
			Path:       "internal/tui/render.go",
			Line:       11,
			Side:       reviewDiffLineSideRight,
			IsOutdated: true,
			Comments:   []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Thread body", CreatedAt: "2026-04-20T10:00:00Z", State: "PENDING"}},
		}},
	}

	actualDocument := newDetailDocument(renderReviewDiffFile(file, renderer, 96), 96)
	headerLineIndex, headerLine := given_detailDocumentLineContaining(t, actualDocument, "internal/tui/render.go:11 R11")

	for _, expected := range []string{"Pending", "Unresolved", "Outdated"} {
		if !strings.Contains(headerLine, expected) {
			t.Fatalf("expected the header line to contain %q, actual %q", expected, headerLine)
		}
	}
	pendingIndex := given_runeIndexInString(t, headerLine, "Pending")
	if actualStylePrefix := actualDocument.lineStylePrefixes[headerLineIndex][pendingIndex]; !strings.Contains(actualStylePrefix, foregroundColorEscape(theme.PendingHex)) {
		t.Fatalf("expected the pending header prefix to contain %q, actual %q", foregroundColorEscape(theme.PendingHex), actualStylePrefix)
	}
}
