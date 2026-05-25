package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgMoveReviewComment_WhenApplying_ThenItSelectsTheTargetFileAndReturnsATypedFocusCommand(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiffWithComments(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, subject))

	actual := Update(subject, MsgMoveReviewComment{Direction: reviewNavigationForward})

	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")
	if len(actual) != 1 {
		t.Fatalf("expected one focus-review-comment command, actual %d", len(actual))
	}
	command, ok := actual[0].(focusReviewCommentCmd)
	if !ok {
		t.Fatalf("expected a focusReviewCommentCmd, actual %T", actual[0])
	}
	if command.RenderedLine < 0 {
		t.Fatalf("expected a non-negative rendered line, actual %d", command.RenderedLine)
	}
}
