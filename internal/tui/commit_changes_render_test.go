package tui

import (
	"strings"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestRenderPullRequestCommitsTabForSummary_GivenValidCommitChangesURL_WhenFormatting_ThenItMakesOnlyTheCommitHeaderAnUnderlinedHyperlink(t *testing.T) {
	summary := githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}}
	commits := []githubdomain.PullRequestCommit{{
		OID:             "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		MessageHeadline: "newer commit",
		MessageBody:     "Newer body",
		CommittedDate:   "2026-05-20T10:00:00Z",
		Authors:         []githubdomain.PullRequestCommitAuthor{{Name: "Newer Dev"}},
	}}
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}

	actualDocument := newDetailDocument(renderPullRequestCommitsTabForSummary(summary, commits, renderer, 72), 72)
	headerLineIndex, headerLine := given_detailDocumentLineContaining(t, actualDocument, "newer commit")
	authorsLineIndex, authorsLine := given_detailDocumentLineContaining(t, actualDocument, "Authors: Newer Dev")
	bodyLineIndex, bodyLine := given_detailDocumentLineContaining(t, actualDocument, "Rendered newer body")
	headerIndex := given_runeIndexInString(t, headerLine, "2222222")
	authorsIndex := given_runeIndexInString(t, authorsLine, "Authors")
	bodyIndex := given_runeIndexInString(t, bodyLine, "Rendered")
	expectedURL := "https://github.com/acme/widgets/pull/42/changes/2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	if actualPrefix := actualDocument.lineStylePrefixes[headerLineIndex][headerIndex]; !strings.Contains(actualPrefix, underlineEscape) {
		t.Fatalf("expected the commit header to be underlined, actual prefix %q", actualPrefix)
	}
	if actualTarget := actualDocument.lineHyperlinkTargets[headerLineIndex][headerIndex]; actualTarget != expectedURL {
		t.Fatalf("expected the commit header link target %q, actual %q", expectedURL, actualTarget)
	}
	if actualPrefix := actualDocument.lineStylePrefixes[authorsLineIndex][authorsIndex]; strings.Contains(actualPrefix, underlineEscape) {
		t.Fatalf("expected the authors line to stay plain, actual prefix %q", actualPrefix)
	}
	if actualTarget := actualDocument.lineHyperlinkTargets[authorsLineIndex][authorsIndex]; actualTarget != "" {
		t.Fatalf("expected the authors line to have no link target, actual %q", actualTarget)
	}
	if actualPrefix := actualDocument.lineStylePrefixes[bodyLineIndex][bodyIndex]; strings.Contains(actualPrefix, underlineEscape) {
		t.Fatalf("expected the body line to stay plain, actual prefix %q", actualPrefix)
	}
	if actualTarget := actualDocument.lineHyperlinkTargets[bodyLineIndex][bodyIndex]; actualTarget != "" {
		t.Fatalf("expected the body line to have no link target, actual %q", actualTarget)
	}
}

func TestCommitAtCursor_GivenCommitHeaderCursorContext_WhenResolving_ThenItReturnsTheCommitAtTheHeaderLine(t *testing.T) {
	summary := githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}}
	commits := []githubdomain.PullRequestCommit{{
		OID:             "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		MessageHeadline: "newer commit",
		MessageBody:     "Newer body",
		CommittedDate:   "2026-05-20T10:00:00Z",
		Authors:         []githubdomain.PullRequestCommitAuthor{{Name: "Newer Dev"}},
	}}
	document := newDetailDocument(renderPullRequestCommitsTabForSummary(summary, commits, &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}, 72), 72)
	headerLineIndex, headerLine := given_detailDocumentLineContaining(t, document, "newer commit")
	headerIndex := given_runeIndexInString(t, headerLine, "2222222")
	subject := detailCursorReadModel{
		selection: detailCursorSelection{
			document: document,
			state:    detailViewState{cursor: detailPosition{line: headerLineIndex, column: headerIndex}},
		},
		pullRequestCommitsSummary: summary,
		pullRequestCommits:        commits,
		pullRequestCommitsKnown:   true,
	}

	context, ok := subject.pullRequestCommitsContext()
	if !ok {
		t.Fatal("expected a pull request commits cursor context")
	}
	actual, ok := commitAtCursor(context)
	if !ok {
		t.Fatal("expected a commit at the header cursor")
	}
	if actual.OID != commits[0].OID {
		t.Fatalf("expected commit oid %q, actual %q", commits[0].OID, actual.OID)
	}
}

func TestCommitAtCursor_GivenCommitMetadataCursorContext_WhenResolving_ThenItReturnsNoCommit(t *testing.T) {
	summary := githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}}
	commits := []githubdomain.PullRequestCommit{{
		OID:             "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		MessageHeadline: "newer commit",
		MessageBody:     "Newer body",
		CommittedDate:   "2026-05-20T10:00:00Z",
		Authors:         []githubdomain.PullRequestCommitAuthor{{Name: "Newer Dev"}},
	}}
	document := newDetailDocument(renderPullRequestCommitsTabForSummary(summary, commits, &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}, 72), 72)
	metadataLineIndex, metadataLine := given_detailDocumentLineContaining(t, document, "Authors: Newer Dev")
	metadataIndex := given_runeIndexInString(t, metadataLine, "Authors")
	subject := detailCursorReadModel{
		selection: detailCursorSelection{
			document: document,
			state:    detailViewState{cursor: detailPosition{line: metadataLineIndex, column: metadataIndex}},
		},
		pullRequestCommitsSummary: summary,
		pullRequestCommits:        commits,
		pullRequestCommitsKnown:   true,
	}

	context, ok := subject.pullRequestCommitsContext()
	if !ok {
		t.Fatal("expected a pull request commits cursor context")
	}
	if _, ok := commitAtCursor(context); ok {
		t.Fatal("expected no commit on the metadata line")
	}
}
