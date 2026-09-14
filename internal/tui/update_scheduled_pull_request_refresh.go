package tui

import (
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyScheduledPullRequestRefreshDue(message MsgScheduledPullRequestRefreshDue) []Cmd {
	if message.Generation != program.pullRequestRefreshScheduleGeneration || program.runtimeConfig.pullRequestConfig.Refresh <= 0 {
		return []Cmd{acknowledgePullRequestRefreshDueCmd{generation: message.Generation}}
	}

	triggeredAt := message.TriggeredAt
	if triggeredAt.IsZero() {
		triggeredAt = time.Now()
	}
	batchID := program.beginScheduledPullRequestRefreshBatch(message.Generation, triggeredAt)
	if batchID == 0 {
		return []Cmd{acknowledgePullRequestRefreshDueCmd{generation: message.Generation}}
	}

	commands := make([]Cmd, 0, len(message.Tabs)+2)
	seenTabs := make(map[PullRequestTab]struct{}, len(message.Tabs))
	for _, tab := range message.Tabs {
		if _, seen := seenTabs[tab]; seen {
			continue
		}
		seenTabs[tab] = struct{}{}
		if program.isPastedPullRequestTab(tab) || !program.pullRequestSearchAutoRefreshEnabled(tab) {
			continue
		}
		commands = append(commands, scheduledPullRequestListReloadCmd{tab: tab, batchID: batchID})
	}
	if message.RefreshPasted {
		pasted := program.scheduledPastedPullRequestRefresh()
		commands = append(commands, scheduledPullRequestRefreshCmd{
			detailSummaries: pasted.detailSummaries,
			diffSummaries:   pasted.diffSummaries,
			batchID:         batchID,
		})
	}
	commands = append(commands, finishScheduledPullRequestRefreshPlanningCmd{batchID: batchID})
	return commands
}

func (program *Program) scheduledPastedPullRequestRefresh() scheduledPullRequestRefreshCmd {
	summaries := program.pastedPullRequests.summaries()
	return scheduledPullRequestRefreshCmd{
		detailSummaries: append([]githubdomain.PullRequest(nil), summaries...),
		diffSummaries:   append([]githubdomain.PullRequest(nil), program.unreadPullRequests(summaries)...),
	}
}
