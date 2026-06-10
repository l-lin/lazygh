package tui

import (
	"reflect"
	"strings"
	"testing"

	githubcli "github.com/l-lin/lazygh/internal/githubcli"
)

func TestOpenLinkUnderCursor_GivenGXOnACommitHeader_WhenOpening_ThenItUsesTheConfiguredLinkOpener(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestDetailWithCommitsForCommitChangesTests(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}
	opener := &fakeLinkOpener{}
	subject.linkOpener = opener
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, "2222222 newer commit")
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	goHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'g')
	xHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'x')
	actualErr = goHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = xHandler(gui, detailView)
	then_noError(t, actualErr)

	expected := []string{"https://github.com/acme/widgets/pull/42/changes/2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	if !reflect.DeepEqual(opener.urls, expected) {
		t.Fatalf("expected opened links %v, actual %v", expected, opener.urls)
	}
}

func TestActionsPopup_GivenBrowserCommitsTabCursorOnCommitHeader_WhenOpening_ThenItShowsDisplayCommitChanges(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestDetailWithCommitsForCommitChangesTests(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "2222222 newer commit")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Display commit changes") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "Display commit changes", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommitsTabCursorOnCommitMetadata_WhenOpening_ThenItHidesDisplayCommitChanges(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestDetailWithCommitsForCommitChangesTests(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Authors: Newer Dev")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "Display commit changes") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "Display commit changes", popupView.Buffer())
	}
}

func TestHelpPopup_GivenBrowserCommitHeader_WhenTogglingHelp_ThenItShowsgfForDisplayCommitChanges(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestDetailWithCommitsForCommitChangesTests(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "2222222 newer commit")

	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Display commit changes", "gf")
}

func TestDisplayCommitChangesShortcut_GivenBrowserCommitsTabCommitHeader_WhenPressingGF_ThenItDispatchesWithoutError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestDetailWithCommitsForCommitChangesTests(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Newer body": "Rendered newer body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "2222222 newer commit")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	gHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'g')
	fHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'f')
	actualErr = gHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = fHandler(gui, detailView)
	then_noError(t, actualErr)
}

func given_pullRequestDetailWithCommitsForCommitChangesTests() githubcli.PullRequestDetail {
	return githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      42,
		Body:        "Body 42",
		BaseRefName: "main",
		HeadRefName: "feature/commits",
		State:       "OPEN",
		Commits: []githubcli.PullRequestCommit{{
			OID:             "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			MessageHeadline: "newer commit",
			MessageBody:     "Newer body",
			CommittedDate:   "2026-05-20T10:00:00Z",
			Authors:         []githubcli.PullRequestCommitAuthor{{Name: "Newer Dev"}},
		}},
	}
}
