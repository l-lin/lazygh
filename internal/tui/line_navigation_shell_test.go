package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgLineNavigationRequestedWithDetailFocus_WhenApplying_ThenItReturnsATypedDetailLineNavigationCommand(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "user-1", Detail: given_multilineDetail(4)}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	actual := Update(subject, MsgLineNavigationRequested{Delta: 1})

	if len(actual) != 1 {
		t.Fatalf("expected one line-navigation command, actual %d", len(actual))
	}
	command, ok := actual[0].(detailLineNavigationCmd)
	if !ok {
		t.Fatalf("expected a detailLineNavigationCmd, actual %T", actual[0])
	}
	if command.Delta != 1 {
		t.Fatalf("expected detail line-navigation delta %d, actual %d", 1, command.Delta)
	}
}

func TestUpdate_GivenMsgLineNavigationRequestedWithReviewFilesFocus_WhenApplying_ThenItAdjustsTheReviewSelectionWithoutATypedShellCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.startReviewSession(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}, "PRR_nav")
	subject.navigationState.reviewSession.selectedFileTreeRow = 0

	actual := Update(subject, MsgLineNavigationRequested{Delta: 1})

	if len(actual) != 0 {
		t.Fatalf("expected no shell command for review file selection, actual %d", len(actual))
	}
	if subject.navigationState.reviewSession.selectedFileTreeRow != 1 {
		t.Fatalf("expected selected review row %d, actual %d", 1, subject.navigationState.reviewSession.selectedFileTreeRow)
	}
}
