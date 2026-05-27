package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestReviewSessionReadModel_GivenStoryReviewChapterSelection_WhenDerivingDetailContent_ThenItRendersTheChapterNarrative(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"# Chapter 1\n\nNarrative body": "Rendered chapter body"}}
	subject := reviewSessionReadModel{
		active:          true,
		mainContentKind: MainContentKindStoryChapter,
		summary:         githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}},
		pendingReviewID: "PRR_pending",
		mode:            reviewSessionModeStory,
		story: reviewStoryData{
			Chapters: []reviewStoryChapter{{ID: "chapter-1", Title: "Chapter 1", Narrative: "Narrative body"}},
			Tree:     reviewDiffTree{Rows: []reviewDiffTreeRow{{ID: "chapter-1", Kind: reviewDiffTreeRowKindChapter, ChapterIndex: 0}}},
		},
		diffResult:       pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{{Path: "internal/tui/render.go"}}}},
		diffResultKnown:  true,
		markdownRenderer: renderer,
		detailWrapWidth:  80,
	}

	actual := subject.detailContent()

	if actual != "Rendered chapter body" {
		t.Fatalf("expected detail content %q, actual %q", "Rendered chapter body", actual)
	}
	if renderer.lastMarkdown != "# Chapter 1\n\nNarrative body" {
		t.Fatalf("expected markdown %q, actual %q", "# Chapter 1\n\nNarrative body", renderer.lastMarkdown)
	}
}

func TestReviewSessionReadModel_GivenSelectedDirectoryRow_WhenSelectingDiffFile_ThenItUsesTheFirstDescendantFile(t *testing.T) {
	subject := reviewSessionReadModel{
		active:              true,
		mainContentKind:     MainContentKindReviewDiff,
		summary:             githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}},
		pendingReviewID:     "PRR_pending",
		selectedFileTreeRow: 0,
		diffResult: pullRequestDiffResult{data: reviewDiffData{
			Files: []reviewDiffFile{{Path: "internal/tui/render.go"}, {Path: "internal/tui/model.go"}},
			FileTree: reviewDiffTree{Rows: []reviewDiffTreeRow{
				{ID: "internal", Depth: 0, Kind: reviewDiffTreeRowKindDirectory, FileIndex: -1, Foldable: true},
				{ID: "internal:render.go", Depth: 1, Kind: reviewDiffTreeRowKindFile, FileIndex: 0},
				{ID: "internal:model.go", Depth: 1, Kind: reviewDiffTreeRowKindFile, FileIndex: 1},
			}},
		}},
		diffResultKnown: true,
	}

	actual, ok := subject.selectedDiffFile()

	if !ok {
		t.Fatalf("expected a selected diff file")
	}
	if actual.Path != "internal/tui/render.go" {
		t.Fatalf("expected selected diff file %q, actual %q", "internal/tui/render.go", actual.Path)
	}
}

func TestReviewSessionReadModelAdapter_GivenOutOfBoundsSelection_WhenReadingFilesAndVisibleLine_ThenItKeepsDurableSelectionUntouched(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.startReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_pending")
	subject.navigationState.reviewSession.selectedFileTreeRow = 99

	expectedSelectableRows, ok := subject.reviewSessionSelectableRows()
	if !ok || len(expectedSelectableRows) == 0 {
		t.Fatalf("expected selectable review rows, actual %v", expectedSelectableRows)
	}
	expectedVisibleLine := adjustVisibleSelection(99, expectedSelectableRows, 0)

	_ = subject.reviewSessionFiles()
	actualVisibleLine := subject.reviewSessionSelectedVisibleLine()

	if actualVisibleLine != expectedVisibleLine {
		t.Fatalf("expected selected visible line %d, actual %d", expectedVisibleLine, actualVisibleLine)
	}
	if actual := subject.navigationState.reviewSession.selectedFileTreeRow; actual != 99 {
		t.Fatalf("expected durable selected file tree row %d to stay untouched by read access, actual %d", 99, actual)
	}
}
