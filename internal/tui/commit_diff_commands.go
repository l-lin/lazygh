package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type commitDiffWorkflowRuntime struct {
	workflowShellRuntime
	getCommitDiff func(repository string, commitOID string) (githubdomain.CommitDiff, error)
}

type loadCommitDiffCmd struct {
	PullRequestKey string
	Repository     string
	CommitOID      string
}

func newCommitDiffWorkflowRuntime(program *Program, gui *gocui.Gui) commitDiffWorkflowRuntime {
	runtime := commitDiffWorkflowRuntime{workflowShellRuntime: newWorkflowShellRuntime(program, gui)}
	if program == nil || program.detailQueries == nil {
		return runtime
	}

	runtime.getCommitDiff = program.detailQueries.GetCommitDiff
	return runtime
}

func (command loadCommitDiffCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newCommitDiffWorkflowRuntime(program, gui)
	if runtime.getCommitDiff == nil || runtime.dispatchAsyncMessage == nil {
		return
	}

	pullRequestKey := strings.TrimSpace(command.PullRequestKey)
	repository := strings.TrimSpace(command.Repository)
	commitOID := strings.TrimSpace(command.CommitOID)
	if pullRequestKey == "" || repository == "" || commitOID == "" {
		return
	}

	runWorkflowCommandAsync(runtime.runAsync, func() {
		diff, err := runtime.getCommitDiff(repository, commitOID)
		runtime.dispatchAsyncMessage(MsgCommitDiffLoaded{PullRequestKey: pullRequestKey, CommitOID: commitOID, Diff: diff, Err: err})
	})
}
