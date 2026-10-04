package tui

import (
	"github.com/jesseduffield/gocui"
	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestListWorkflowRuntime struct {
	workflowShellRuntime
	pullRequestsFromCache         func(PullRequestTab) ([]githubdomain.PullRequest, bool)
	pullRequestFreshnessFromCache func() ([]persistcache.PullRequestFreshness, bool)
	listPullRequests              func(PullRequestTab) ([]githubdomain.PullRequest, error)
}

type loadPullRequestsCmd struct {
	tab                     PullRequestTab
	Source                  pullRequestLoadSource
	ScheduledRefreshBatchID uint64
}

type scheduledPullRequestListReloadCmd struct {
	tab     PullRequestTab
	batchID uint64
}

type scheduledPullRequestRefreshCmd struct {
	detailSummaries []githubdomain.PullRequest
	diffSummaries   []githubdomain.PullRequest
	batchID         uint64
}

type reloadPullRequestsTabCmd struct {
	tab PullRequestTab
}

type hydratePullRequestsFromCacheCmd struct {
	tab PullRequestTab
}

func newPullRequestListWorkflowRuntime(program *Program, gui *gocui.Gui) pullRequestListWorkflowRuntime {
	runtime := pullRequestListWorkflowRuntime{workflowShellRuntime: newWorkflowShellRuntime(program, gui)}
	if program != nil {
		runtime.pullRequestsFromCache = program.pullRequestsFromCache
		runtime.pullRequestFreshnessFromCache = program.pullRequestFreshnessFromCache
		if program.pullRequestListQueries != nil {
			runtime.listPullRequests = newPullRequestListQueryCommand(program.pullRequestListQueries, program.runtimeConfig.pullRequestConfig.Searches)
		}
	}
	return runtime
}

func newPullRequestListQueryCommand(queries PullRequestListQueries, searches []appconfig.PullRequestSearch) func(PullRequestTab) ([]githubdomain.PullRequest, error) {
	if queries == nil {
		return nil
	}
	return func(tab PullRequestTab) ([]githubdomain.PullRequest, error) {
		command, ok := pullRequestSearchCommandForTab(searches, tab)
		if !ok {
			return nil, nil
		}
		return queries.ListPullRequests(command)
	}
}

func (command loadPullRequestsCmd) execute(program *Program, gui *gocui.Gui) {
	if program != nil && command.ScheduledRefreshBatchID != 0 && program.scheduledPullRequestRefreshBatchCancelled(command.ScheduledRefreshBatchID) {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.ScheduledRefreshBatchID, nil))
		return
	}
	runtime := newPullRequestListWorkflowRuntime(program, gui)
	if runtime.listPullRequests == nil || runtime.dispatchAsyncMessage == nil {
		if command.ScheduledRefreshBatchID != 0 {
			program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.ScheduledRefreshBatchID, nil))
		}
		return
	}
	generation := program.pullRequestLoadGeneration(command.tab)
	runWorkflowCommandAsync(runtime.runAsync, func() {
		pullRequests, err := runtime.listPullRequests(command.tab)
		runtime.dispatchAsyncMessage(MsgPullRequestsLoaded{Tab: command.tab, PullRequests: pullRequests, Err: err, Generation: generation, Source: command.Source, ScheduledRefreshBatchID: command.ScheduledRefreshBatchID})
	})
}

func (command scheduledPullRequestListReloadCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	if program.scheduledPullRequestRefreshBatchCancelled(command.batchID) {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	if program.isPastedPullRequestTab(command.tab) || program.pullRequestsLoading(command.tab) || !program.hasPullRequestListQueries() {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	_, targetConfigured := program.searchBackedPullRequestSearch(command.tab)
	if !targetConfigured || program.runtimeConfig.pullRequestConfig.Refresh <= 0 || !program.pullRequestSearchAutoRefreshEnabled(command.tab) {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}

	runtime := newPullRequestListWorkflowRuntime(program, gui)
	if runtime.executeWorkflowPlan == nil {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	plan := planScheduledPullRequestListReload(command.tab, true, targetConfigured, command.batchID)
	if len(plan.commands) == 0 {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	program.addScheduledPullRequestRefreshWork(command.batchID, len(plan.commands))
	runtime.executeWorkflowPlan(plan)
	program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
}

func (command scheduledPullRequestRefreshCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	if !program.hasDetailQueries() || program.scheduledPullRequestRefreshBatchCancelled(command.batchID) {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	runtime := newWorkflowShellRuntime(program, gui)
	if runtime.executeWorkflowPlan == nil {
		program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
		return
	}
	// Scheduled refresh bypasses the normal selected-detail workflow, so hydrate disk cache first. Otherwise a visible PR can show a loading screen instead of its cached content while the forced network refresh runs.
	for _, summary := range uniqueScheduledPullRequestSummaries(command.detailSummaries, program.pullRequestDetailLoadInFlight) {
		(hydratePullRequestDetailFromCacheCmd{summary: summary}).execute(program, gui)
	}
	for _, summary := range uniqueScheduledPullRequestSummaries(command.diffSummaries, program.pullRequestDiffLoadInFlight) {
		(hydratePullRequestDiffFromCacheCmd{summary: summary}).execute(program, gui)
	}

	var detailSummaries, diffSummaries []githubdomain.PullRequest
	for _, summary := range command.detailSummaries {
		result, ok := program.pullRequestDetailForSummary(summary)
		if result.retry.ready(program.currentTime()) && pullRequestDetailNeedsRefresh(summary, result, ok) {
			detailSummaries = append(detailSummaries, summary)
		}
	}
	for _, summary := range command.diffSummaries {
		result, ok := program.pullRequestDiffForSummary(summary)
		if result.retry.ready(program.currentTime()) && pullRequestDiffNeedsRefresh(summary, result, ok, program.shouldLoadPullRequestDiffTeamOwners()) {
			diffSummaries = append(diffSummaries, summary)
		}
	}
	plan := planScheduledPullRequestRefresh(scheduledPullRequestRefreshPlanInput{
		detailSummaries:    detailSummaries,
		diffSummaries:      diffSummaries,
		hasDetailQueries:   true,
		detailLoadInFlight: program.pullRequestDetailLoadInFlight,
		diffLoadInFlight:   program.pullRequestDiffLoadInFlight,
		batchID:            command.batchID,
	})
	program.addScheduledPullRequestRefreshWork(command.batchID, len(plan.commands))
	runtime.executeWorkflowPlan(plan)
	program.executeCmds(gui, program.completeScheduledPullRequestRefreshWork(command.batchID, nil))
}

func (command hydratePullRequestsFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newPullRequestListWorkflowRuntime(program, gui)
	if runtime.pullRequestsFromCache == nil || runtime.executeUpdate == nil {
		return
	}
	pullRequests, ok := runtime.pullRequestsFromCache(command.tab)
	if !ok {
		return
	}
	var freshness []persistcache.PullRequestFreshness
	if runtime.pullRequestFreshnessFromCache != nil {
		freshness, _ = runtime.pullRequestFreshnessFromCache()
	}
	runtime.executeUpdate(MsgPullRequestsCacheHydrated{Tab: command.tab, PullRequests: pullRequests, Freshness: freshness})
}

func (command reloadPullRequestsTabCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newPullRequestListWorkflowRuntime(program, gui)
	if runtime.executeWorkflowPlan == nil || runtime.pullRequestListReloadPlan == nil {
		return
	}
	runtime.executeWorkflowPlan(runtime.pullRequestListReloadPlan(command.tab))
}
