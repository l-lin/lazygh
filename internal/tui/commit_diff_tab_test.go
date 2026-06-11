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

func TestBrowserMode_GivenACommitDiffTab_WhenPressingEnterAndZA_ThenItTogglesTheContainingFileVisibility(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "+new line")

	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	then_noError(t, toggleHandler(gui, detailView))
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") {
		t.Fatalf("expected enter to collapse the commit diff file, actual %q", detailView.Buffer())
	}
	for _, hidden := range []string{"@@ -42,2 +42,2 @@", "+new line"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected enter to hide %q from the collapsed commit diff file, actual %q", hidden, detailView.Buffer())
		}
	}
	if !strings.Contains(detailView.Buffer(), "internal/tui/model.go") || !strings.Contains(detailView.Buffer(), "+new model") {
		t.Fatalf("expected collapsing one commit diff file to keep the other file visible, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go")

	then_noError(t, toggleHandler(gui, detailView))
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") || !strings.Contains(detailView.Buffer(), "+new line") {
		t.Fatalf("expected enter to expand the commit diff file again, actual %q", detailView.Buffer())
	}
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "+new line")

	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	collapseHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
	then_noError(t, prefixHandler(gui, detailView))
	then_noError(t, collapseHandler(gui, detailView))
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") {
		t.Fatalf("expected za to collapse the commit diff file, actual %q", detailView.Buffer())
	}
	for _, hidden := range []string{"@@ -42,2 +42,2 @@", "+new line"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected za to hide %q from the collapsed commit diff file, actual %q", hidden, detailView.Buffer())
		}
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go")
}

func TestBrowserMode_GivenCommitDiffTabFiles_WhenPressingZMAndZR_ThenItClosesAndOpensEveryFileWhileKeepingTheCursorOnTheSameFileHeader(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/model.go")

	prefixHandler, closeAllHandler, openAllHandler := given_detailBulkFoldHandlers(t, subject)
	then_noError(t, prefixHandler(gui, detailView))
	then_noError(t, closeAllHandler(gui, detailView))
	for _, expected := range []string{" " + reviewDiffHeaderPathIcon + " internal/tui/render.go", " " + reviewDiffHeaderPathIcon + " internal/tui/model.go"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected zM to collapse every commit diff file and keep %q visible, actual %q", expected, detailView.Buffer())
		}
	}
	for _, hidden := range []string{"+new line", "+new model"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected zM to hide %q from collapsed commit diff files, actual %q", hidden, detailView.Buffer())
		}
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go")

	then_noError(t, prefixHandler(gui, detailView))
	then_noError(t, openAllHandler(gui, detailView))
	for _, expected := range []string{" " + reviewDiffHeaderPathIcon + " internal/tui/render.go", " " + reviewDiffHeaderPathIcon + " internal/tui/model.go", "+new line", "+new model"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected zR to reopen every commit diff file and show %q, actual %q", expected, detailView.Buffer())
		}
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go")
}

func TestBrowserMode_GivenACommitDiffTabWithCollapsedFiles_WhenOpeningADifferentCommit_ThenItDoesNotReuseThePreviousCommitFoldState(t *testing.T) {
	loader := given_commitDiffTabLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	detailView := given_commitDiffTabOnCommitHeader(t, gui, subject, "● 2222222 newer commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/model.go")

	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	then_noError(t, toggleHandler(gui, detailView))
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/model.go") {
		t.Fatalf("expected the newer commit diff file to collapse before retargeting, actual %q", detailView.Buffer())
	}

	subject.setDetailActiveTab(CommitsDetailTab)
	then_noError(t, subject.afterStateChange(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "● 1111111 older commit")
	then_noError(t, subject.displayCommitChangesShortcut(gui, detailView))

	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/model.go") {
		t.Fatalf("expected the older commit diff file to start expanded instead of reusing the previous commit fold state, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "+new model") {
		t.Fatalf("expected the older commit diff body to stay visible after retargeting, actual %q", detailView.Buffer())
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
				Files: []githubcli.PullRequestDiffFile{
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
						Path:       "internal/tui/model.go",
						ChangeType: "modified",
						Additions:  1,
						Deletions:  1,
						Patch: strings.Join([]string{
							"@@ -7,1 +7,1 @@",
							"-old model",
							"+new model",
						}, "\n"),
					},
				},
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
