package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestBrowserMode_GivenCommitChangesRequest_WhenOpening_ThenItShowsTheCommitDiffTabWithTheShortSHALabel(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")

	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))

	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (2)", ChangesDetailTab.Label(), detailChangesIcon + " 2222222"}, 4)
	for _, expected := range []string{"internal/tui/render.go", "@@ -42,2 +42,2 @@", "42 : 42 │  context line", "43 :    │ -old line", "   : 43 │ +new line"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the commit diff tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	if !reflect.DeepEqual(loader.commitDiffCalls, []string{"acme/widgets@2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}) {
		t.Fatalf("expected commit diff calls %v, actual %v", []string{"acme/widgets@2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}, loader.commitDiffCalls)
	}
}

func TestBrowserMode_GivenAnExistingCommitDiffTabForTheSameCommit_WhenOpeningAgain_ThenItFocusesTheExistingTabWithoutReloading(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))

	subject.setDetailActiveTab(CommitsDetailTab)
	then_noError(t, subject.afterStateChange(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "● 2222222 newer commit")

	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))

	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (2)", ChangesDetailTab.Label(), detailChangesIcon + " 2222222"}, 4)
	if !reflect.DeepEqual(loader.commitDiffCalls, []string{"acme/widgets@2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}) {
		t.Fatalf("expected the same commit diff to stay cached after reopening, actual %v", loader.commitDiffCalls)
	}
}

func TestBrowserMode_GivenAnExistingCommitDiffTabForADifferentCommit_WhenOpeningAgain_ThenItReusesTheTabAndReplacesItsLabelAndContent(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))

	subject.setDetailActiveTab(CommitsDetailTab)
	then_noError(t, subject.afterStateChange(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "● 1111111 older commit")

	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))

	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (2)", ChangesDetailTab.Label(), detailChangesIcon + " 1111111"}, 4)
	for _, expected := range []string{"internal/tui/model.go", "@@ -7,1 +7,1 @@", "-old model", "+new model"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the retargeted commit diff tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	if reflect.DeepEqual(loader.commitDiffCalls, []string{"acme/widgets@2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}) {
		t.Fatalf("expected a second commit diff load when switching commits, actual %v", loader.commitDiffCalls)
	}
}

func TestBrowserMode_GivenACommitDiffTab_WhenClosingDetailAndReopening_ThenItClearsTheOptionalTab(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))
	then_noError(t, subject.closeDetail(gui, nil))
	then_noError(t, subject.openDetail(gui, nil))
	then_noError(t, subject.afterStateChange(gui))

	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (2)", ChangesDetailTab.Label()}, 0)
}

func TestActionsPopup_GivenACommitDiffTab_WhenListingActions_ThenItStaysReadOnly(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))

	if given_hasActionTitle(subject.currentActionsPopupActions(), "Add inline comment") {
		t.Fatalf("expected commit diff actions %v to stay read-only and hide %q", given_actionTitles(subject.currentActionsPopupActions()), "Add inline comment")
	}
	if given_hasActionTitle(subject.currentActionsPopupActions(), reactionPickerTitle) {
		t.Fatalf("expected commit diff actions %v to stay read-only and hide %q", given_actionTitles(subject.currentActionsPopupActions()), reactionPickerTitle)
	}
}

func given_commitDiffTabOnCommitHeader(t *testing.T, gui *gocui.Gui, subject *Program, header string) *gocui.View {
	t.Helper()

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	then_noError(t, subject.nextDetailTab(gui, nil))
	then_noError(t, subject.nextDetailTab(gui, nil))

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, header)
	return detailView
}

func given_commitDiffTabLoader() *fakePullRequestDetailLoader {
	return &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/commit-diff",
				State:       "OPEN",
				Commits: []githubcli.PullRequestCommit{
					{
						OID:             "1111111aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						MessageHeadline: "older commit",
						CommittedDate:   "2026-05-19T10:00:00Z",
						Authors:         []githubcli.PullRequestCommitAuthor{{Name: "Older Dev"}},
					},
					{
						OID:             "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
						MessageHeadline: "newer commit",
						CommittedDate:   "2026-05-20T10:00:00Z",
						Authors:         []githubcli.PullRequestCommitAuthor{{Name: "Newer Dev"}},
					},
				},
			},
		},
		commitDiffs: map[string]githubcli.CommitDiff{
			"acme/widgets@2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": {
				Files: []githubcli.PullRequestDiffFile{{
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
				}},
			},
			"acme/widgets@1111111aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {
				Files: []githubcli.PullRequestDiffFile{{
					Path:       "internal/tui/model.go",
					ChangeType: "modified",
					Additions:  1,
					Deletions:  1,
					Patch: strings.Join([]string{
						"@@ -7,1 +7,1 @@",
						"-old model",
						"+new model",
					}, "\n"),
				}},
			},
		},
	}
}
