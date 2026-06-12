package tui

func (program *Program) syncPastedPullRequestTab() (PullRequestTab, bool) {
	if program == nil || program.model == nil {
		return 0, false
	}

	seed, hasPastedTab := program.pastedPullRequests.tabSeedWithRepositoryStyle(program.runtimeConfig.displayConfig.RepositoryStyle)
	preservedRows := make(map[PullRequestTab][]PullRequestRow, len(program.model.PullRequestTabs()))
	seeds := make([]PullRequestTabSeed, 0, len(program.model.PullRequestTabs())+1)
	for _, tab := range program.model.PullRequestTabs() {
		if program.isPastedPullRequestTab(tab) {
			continue
		}
		rows := program.model.PullRequestRows(tab)
		preservedRows[tab] = rows
		seeds = append(seeds, PullRequestTabSeed{Label: program.model.PullRequestTabLabel(tab), PullRequests: pullRequestItems(rows)})
	}
	if hasPastedTab {
		seeds = append(seeds, seed)
	}

	program.model.SetPullRequestTabs(seeds)
	for tab, rows := range preservedRows {
		program.model.SetPullRequestRows(tab, rows)
	}
	if !hasPastedTab {
		return 0, false
	}

	pastedTab := PullRequestTab(len(seeds) - 1)
	program.model.SetPullRequestRows(pastedTab, program.pastedPullRequests.rowsWithRepositoryStyle(program.runtimeConfig.displayConfig.RepositoryStyle))
	return pastedTab, true
}
