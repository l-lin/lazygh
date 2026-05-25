package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenBrowserChangesTabCursorOnAnInlineThread_WhenOpening_ThenItShowsTheResolveInlineCommentAction(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	subject.detailState.activeTab = ChangesDetailTab
	then_noError(t, subject.afterStateChange(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	then_noError(t, subject.openActionsPopup(gui, nil))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Mark inline comment as resolved") {
		t.Fatalf("expected the popup to contain the inline-thread resolve action, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserChangesTabCursorOnAResolvedInlineThread_WhenOpening_ThenItShowsTheUnresolveInlineCommentAction(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithResolvedInlineThreadForChangesResolutionTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithResolvedInlineThreadForChangesResolutionTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	subject.detailState.activeTab = ChangesDetailTab
	then_noError(t, subject.afterStateChange(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go:43")

	then_noError(t, subject.openActionsPopup(gui, nil))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Mark inline comment as unresolved") {
		t.Fatalf("expected the popup to contain the inline-thread unresolve action, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserChangesTabResolveInlineCommentAction_WhenExecuting_ThenItRefreshesTheThreadStateAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	subject.detailState.activeTab = ChangesDetailTab
	then_noError(t, subject.afterStateChange(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	then_noError(t, subject.openActionsPopup(gui, nil))
	subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
	then_noError(t, subject.afterStateChange(gui))
	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))

	if !reflect.DeepEqual(loader.resolveReviewThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected resolved thread ids %v, actual %v", []string{"thread-1"}, loader.resolveReviewThreadIDs)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, inlineCommentResolvedSuccessMessage)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Resolved") {
		t.Fatalf("expected the detail buffer to refresh with the resolved state, actual %q", detailView.Buffer())
	}
}

func given_pullRequestDetailWithResolvedInlineThreadForChangesResolutionTests() githubcli.PullRequestDetail {
	detail := given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()
	detail.InlineCommentThreads[0].IsResolved = true
	return detail
}

func given_pullRequestDiffWithResolvedInlineThreadForChangesResolutionTests() githubcli.PullRequestDiff {
	diff := given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()
	diff.Threads[0].IsResolved = true
	return diff
}
