package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDetailCursorReadModel_GivenPullRequestDescriptionInputs_WhenResolvingContext_ThenItKeepsTheSelectionOnTheSnapshot(t *testing.T) {
	selection := given_detailCursorSelectionForTests("first line\nsecond line", 80)
	subject := detailCursorReadModel{
		selection: selection,
		descriptionSummary: githubdomain.PullRequest{
			Number:     42,
			Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
		},
		descriptionDetail: githubdomain.PullRequestDetail{Body: "Body 42"},
		descriptionKnown:  true,
	}

	actual, ok := subject.pullRequestDescriptionContext()
	if !ok {
		t.Fatal("expected a description cursor context")
	}
	if actual.summary.Number != 42 {
		t.Fatalf("expected description summary number %d, actual %d", 42, actual.summary.Number)
	}
	if actual.detail.Body != "Body 42" {
		t.Fatalf("expected description detail body %q, actual %q", "Body 42", actual.detail.Body)
	}
	if actual.selection.state.cursor.line != selection.state.cursor.line {
		t.Fatalf("expected description cursor line %d, actual %d", selection.state.cursor.line, actual.selection.state.cursor.line)
	}
}

func TestDetailCursorReadModel_GivenBrowserChangesRows_WhenResolvingContext_ThenItKeepsTheRenderedRowsOnTheSnapshot(t *testing.T) {
	selection := given_detailCursorSelectionForTests("diff line", 72)
	subject := detailCursorReadModel{
		selection: selection,
		browserChangesSummary: githubdomain.PullRequest{
			Number:     42,
			Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
		},
		browserChangesRenderedRows: []reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindInlineCommentHeader}},
		browserChangesKnown:        true,
	}

	actual, ok := subject.browserChangesContext()
	if !ok {
		t.Fatal("expected a browser changes cursor context")
	}
	if actual.summary.Number != 42 {
		t.Fatalf("expected browser changes summary number %d, actual %d", 42, actual.summary.Number)
	}
	if len(actual.renderedRows) != 1 {
		t.Fatalf("expected browser changes rendered row count %d, actual %d", 1, len(actual.renderedRows))
	}
	if actual.selection.document.width != selection.document.width {
		t.Fatalf("expected browser changes document width %d, actual %d", selection.document.width, actual.selection.document.width)
	}
}

func TestDetailCursorReadModel_GivenReviewDiffRows_WhenResolvingContext_ThenItKeepsTheRenderedRowsOnTheSnapshot(t *testing.T) {
	selection := given_detailCursorSelectionForTests("review line", 64)
	subject := detailCursorReadModel{
		selection: selection,
		reviewDiffSummary: githubdomain.PullRequest{
			Number:     42,
			Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
		},
		reviewDiffRenderedRows: []reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindInlineCommentHeader}},
		reviewDiffKnown:        true,
	}

	actual, ok := subject.reviewDiffContext()
	if !ok {
		t.Fatal("expected a review diff cursor context")
	}
	if actual.summary.Number != 42 {
		t.Fatalf("expected review diff summary number %d, actual %d", 42, actual.summary.Number)
	}
	if len(actual.renderedRows) != 1 {
		t.Fatalf("expected review diff rendered row count %d, actual %d", 1, len(actual.renderedRows))
	}
	if actual.selection.document.width != selection.document.width {
		t.Fatalf("expected review diff document width %d, actual %d", selection.document.width, actual.selection.document.width)
	}
}

func given_detailCursorSelectionForTests(text string, width int) detailCursorSelection {
	document := newDetailDocumentWithWrap(text, width, false)
	state := newDetailViewState()
	state.sync(document, 1)
	if document.rowCount() > 1 {
		state.cursor.line = 1
	}
	return detailCursorSelection{document: document, state: state}
}
