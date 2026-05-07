package tui

import (
	"reflect"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestReviewDiffThreadTargetForSelection_GivenTheCursorOnADiffLine_WhenBuilding_ThenItReturnsTheSingleLineGitHubTarget(t *testing.T) {
	file := reviewDiffFile{
		Path:       "internal/tui/render.go",
		Additions:  2,
		Deletions:  0,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,1 +10,3 @@",
			Lines: []reviewDiffLine{
				{Kind: reviewDiffContextLine, Text: "context line", LeftLine: 10, RightLine: 10, Side: reviewDiffLineSideBoth},
				{Kind: reviewDiffAdditionLine, Text: "new line", RightLine: 11, Side: reviewDiffLineSideRight},
				{Kind: reviewDiffAdditionLine, Text: "newer line", RightLine: 12, Side: reviewDiffLineSideRight},
			},
		}},
	}
	renderedRows := buildReviewDiffRenderedRows(file, nil, 96)
	document := newDetailDocument(renderReviewDiffFile(file, nil, 96), 96)
	lineIndex, _ := given_detailDocumentLineContaining(t, document, "new line")
	state := newDetailViewState()
	state.cursor = detailPosition{line: lineIndex, column: 0}
	state.sync(document, 20)

	actual, actualErr := reviewDiffThreadTargetForSelection(renderedRows, document, state)

	then_noError(t, actualErr)
	expected := githubcli.PullRequestReviewThreadTarget{
		Path:        "internal/tui/render.go",
		Line:        11,
		Side:        "RIGHT",
		SubjectType: "LINE",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected target %+v, actual %+v", expected, actual)
	}
}

func TestReviewDiffThreadTargetForSelection_GivenALinewiseSelectionCoveringHeadersAndDiffLines_WhenBuilding_ThenItUsesTheSmallestValidRange(t *testing.T) {
	file := reviewDiffFile{
		Path:       "internal/tui/render.go",
		Additions:  2,
		Deletions:  0,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,0 +11,3 @@",
			Lines: []reviewDiffLine{
				{Kind: reviewDiffAdditionLine, Text: "new line", RightLine: 11, Side: reviewDiffLineSideRight},
				{Kind: reviewDiffContextLine, Text: "kept line", LeftLine: 10, RightLine: 12, Side: reviewDiffLineSideBoth},
				{Kind: reviewDiffAdditionLine, Text: "newer line", RightLine: 13, Side: reviewDiffLineSideRight},
			},
		}},
	}
	renderedRows := buildReviewDiffRenderedRows(file, nil, 96)
	document := newDetailDocument(renderReviewDiffFile(file, nil, 96), 96)
	hunkHeaderLineIndex, _ := given_detailDocumentLineContaining(t, document, "@@ -10,0 +11,3 @@")
	lastDiffLineIndex, _ := given_detailDocumentLineContaining(t, document, "newer line")
	state := newDetailViewState()
	state.mode = detailLineVisualMode
	state.visualAnchor = detailPosition{line: hunkHeaderLineIndex, column: 0}
	state.cursor = detailPosition{line: lastDiffLineIndex, column: 0}
	state.sync(document, 20)

	actual, actualErr := reviewDiffThreadTargetForSelection(renderedRows, document, state)

	then_noError(t, actualErr)
	expected := githubcli.PullRequestReviewThreadTarget{
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

func TestReviewDiffThreadTargetForSelection_GivenOnlyInlineCommentDecorationRows_WhenBuilding_ThenItRejectsTheTarget(t *testing.T) {
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
			},
		}},
		Threads: []reviewDiffThread{{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     11,
			Side:     reviewDiffLineSideRight,
			Comments: []githubcli.PullRequestComment{{Body: "Thread body", CreatedAt: "2026-04-20T10:00:00Z"}},
		}},
	}
	renderedRows := buildReviewDiffRenderedRows(file, nil, 96)
	document := newDetailDocument(renderReviewDiffFile(file, nil, 96), 96)
	commentLineIndex, _ := given_detailDocumentLineContaining(t, document, "internal/tui/render.go:11")
	state := newDetailViewState()
	state.cursor = detailPosition{line: commentLineIndex, column: 0}
	state.sync(document, 20)

	_, actualErr := reviewDiffThreadTargetForSelection(renderedRows, document, state)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != reviewThreadTargetUnavailableMessage {
		t.Fatalf("expected error %q, actual %v", reviewThreadTargetUnavailableMessage, actualErr)
	}
}
