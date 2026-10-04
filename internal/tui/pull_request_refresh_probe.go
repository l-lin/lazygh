package tui

import (
	"maps"

	"github.com/jesseduffield/gocui"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type scheduledPullRequestProbeCmd struct {
	summaries []githubdomain.PullRequest
	batchID   uint64
}

type MsgScheduledPullRequestProbed struct {
	Summary githubdomain.PullRequest
	Err     error
	BatchID uint64
}

func (MsgScheduledPullRequestProbed) isMsg() {}

func (command scheduledPullRequestProbeCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	if !program.hasDetailQueries() || program.scheduledPullRequestRefreshBatchCancelled(command.batchID) {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	runtime := newWorkflowShellRuntime(program, gui)
	if runtime.dispatchAsyncMessage == nil {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	queries := program.detailQueries
	for _, summary := range uniqueScheduledPullRequestSummaries(command.summaries, nil) {
		key := pullRequestDetailKey(summary.Repository, summary.Number)
		if !program.pullRequestProbeRetries[key].ready(program.currentTime()) {
			continue
		}
		program.addScheduledPullRequestRefreshWork(command.batchID, 1)
		runtime.runPullRequestAsync(key, func() {
			remote, err := queries.GetPullRequestSummary(pullRequestRepositoryName(summary.Repository), summary.Number)
			if err != nil {
				remote = summary
			}
			runtime.dispatchAsyncMessage(MsgScheduledPullRequestProbed{Summary: remote, Err: err, BatchID: command.batchID})
		})
	}
	program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
}

func (program *Program) applyScheduledPullRequestProbed(message MsgScheduledPullRequestProbed) []Cmd {
	if program.scheduledPullRequestRefreshBatchCancelled(message.BatchID) {
		return program.completeScheduledPullRequestRefreshWork(message.BatchID, nil)
	}
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	retries := maps.Clone(program.pullRequestProbeRetries)
	if retries == nil {
		retries = map[string]pullRequestRefreshRetry{}
	}
	if message.Err != nil {
		retries[key] = retries[key].afterFailure(program.currentTime())
		program.pullRequestProbeRetries = retries
		return program.completeScheduledPullRequestRefreshWork(message.BatchID, message.Err)
	}
	delete(retries, key)
	program.pullRequestProbeRetries = retries
	program.mutateLoadedPullRequestSummaries(message.Summary, func(summary *githubdomain.PullRequest) {
		summary.Title = message.Summary.Title
		summary.Body = message.Summary.Body
		summary.URL = message.Summary.URL
		summary.State = message.Summary.State
		summary.IsDraft = message.Summary.IsDraft
		summary.UpdatedAt = message.Summary.UpdatedAt
	})
	program.refreshPullRequestUnreadMarkers()
	targets := program.pullRequestsNeedingScheduledRefresh([]githubdomain.PullRequest{message.Summary})
	program.addScheduledPullRequestRefreshWork(message.BatchID, 1)
	commands := []Cmd{scheduledPullRequestRefreshCmd{detailSummaries: targets, diffSummaries: targets, batchID: message.BatchID}}
	return append(commands, program.completeScheduledPullRequestRefreshWork(message.BatchID, nil)...)
}
