package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) applyScheduledPullRequestRefreshDue(message MsgScheduledPullRequestRefreshDue) []Cmd {
	commands := make([]Cmd, 0, len(message.Tabs)+2)
	if message.Generation == program.pullRequestRefreshScheduleGeneration {
		seenTabs := make(map[PullRequestTab]struct{}, len(message.Tabs))
		for _, tab := range message.Tabs {
			if _, seen := seenTabs[tab]; seen {
				continue
			}
			seenTabs[tab] = struct{}{}
			search, configured := program.searchBackedPullRequestSearch(tab)
			if !configured || search.Refresh <= 0 || program.isPastedPullRequestTab(tab) {
				continue
			}
			commands = append(commands, scheduledPullRequestListReloadCmd{tab: tab})
		}
		if message.RefreshPasted {
			pasted := program.scheduledPastedPullRequestRefresh()
			if len(pasted.detailSummaries) > 0 {
				commands = append(commands, pasted)
			}
		}
	}

	// The scheduler stays backpressured until every refresh command has been planned.
	return append(commands, acknowledgePullRequestRefreshDueCmd{generation: message.Generation})
}

func (program *Program) scheduledPastedPullRequestRefresh() scheduledPullRequestRefreshCmd {
	summaries := program.pastedPullRequests.summaries()
	return scheduledPullRequestRefreshCmd{
		detailSummaries: append([]githubdomain.PullRequest(nil), summaries...),
		diffSummaries:   append([]githubdomain.PullRequest(nil), program.unreadPullRequests(summaries)...),
	}
}
