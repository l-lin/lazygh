package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestBuildReviewDiffData_GivenRawFileTeamOwners_WhenParsing_ThenItKeepsThemOnTheReviewFile(t *testing.T) {
	raw := githubcli.PullRequestDiff{
		Files: []githubcli.PullRequestDiffFile{{
			Path:       "internal/tui/render.go",
			ChangeType: "modified",
			TeamOwners: []string{"P3C"},
		}},
	}

	actual := buildReviewDiffData(raw)

	expected := []string{"P3C"}
	if !reflect.DeepEqual(actual.Files[0].TeamOwners, expected) {
		t.Fatalf("expected team owners %+v, actual %+v", expected, actual.Files[0].TeamOwners)
	}
}

func TestBuildReviewDiffRenderedRows_GivenTeamOwnedFileAndInlineThread_WhenRendering_ThenItPlacesOwnershipBelowTheFileAndThreadPaths(t *testing.T) {
	file := reviewDiffFile{
		Path:       "internal/tui/render.go",
		Additions:  1,
		Deletions:  0,
		ChangeType: reviewDiffChangeTypeModified,
		TeamOwners: []string{"P3C"},
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,1 +10,2 @@",
			Lines:  []reviewDiffLine{{Kind: reviewDiffAdditionLine, Text: "new line", RightLine: 11, Side: reviewDiffLineSideRight}},
		}},
		Threads: []reviewDiffThread{{
			ID:   "thread-1",
			Path: "internal/tui/render.go",
			Line: 11,
			Side: reviewDiffLineSideRight,
			Comments: []githubcli.PullRequestComment{{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:      "Thread body",
				CreatedAt: "2026-04-20T10:00:00Z",
			}},
		}},
	}

	actual := buildReviewDiffRenderedRows(file, &fakeMarkdownRenderer{output: "Rendered thread body"}, 96)

	if len(actual) < 2 || actual[0].Kind != reviewDiffRenderedRowKindFileHeader {
		t.Fatalf("expected a file header followed by ownership metadata, actual %+v", actual)
	}
	if actual[1].Kind != reviewDiffRenderedRowKindTeamOwners {
		t.Fatalf("expected row 1 to be team ownership metadata, actual %+v", actual[1])
	}
	if !strings.Contains(actual[1].Text, reviewDiffTeamOwnershipIcon+" P3C") {
		t.Fatalf("expected file ownership row to mention the team owner, actual %q", actual[1].Text)
	}

	threadHeaderIndex := -1
	for index, row := range actual {
		if row.Kind == reviewDiffRenderedRowKindInlineCommentHeader {
			threadHeaderIndex = index
			break
		}
	}
	if threadHeaderIndex < 0 {
		t.Fatalf("expected an inline comment header row, actual %+v", actual)
	}
	if threadHeaderIndex+1 >= len(actual) || actual[threadHeaderIndex+1].Kind != reviewDiffRenderedRowKindTeamOwners {
		t.Fatalf("expected team ownership row immediately after the inline comment header, actual %+v", actual)
	}
	if !strings.Contains(actual[threadHeaderIndex+1].Text, reviewDiffTeamOwnershipIcon+" P3C") {
		t.Fatalf("expected inline thread ownership row to mention the team owner, actual %q", actual[threadHeaderIndex+1].Text)
	}
}
