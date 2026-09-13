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
	tab    PullRequestTab
	Source pullRequestLoadSource
}

type scheduledPullRequestListReloadCmd struct {
	tab PullRequestTab
}

type scheduledPullRequestRefreshCmd struct {
	detailSummaries []githubdomain.PullRequest
	diffSummaries   []githubdomain.PullRequest
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
			runtime.listPullRequests = newPullRequestListQueryCommand(program.pullRequestListQueries, program.runtimeConfig.pullRequestSearches)
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
	runtime := newPullRequestListWorkflowRuntime(program, gui)
	if runtime.listPullRequests == nil || runtime.dispatchAsyncMessage == nil {
		return
	}
	generation := program.pullRequestLoadGeneration(command.tab)
	runWorkflowCommandAsync(runtime.runAsync, func() {
		pullRequests, err := runtime.listPullRequests(command.tab)
		runtime.dispatchAsyncMessage(MsgPullRequestsLoaded{Tab: command.tab, PullRequests: pullRequests, Err: err, Generation: generation, Source: command.Source})
	})
}

func (command scheduledPullRequestListReloadCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || program.isPastedPullRequestTab(command.tab) || program.pullRequestsLoading(command.tab) || !program.hasPullRequestListQueries() {
		return
	}
	search, targetConfigured := program.searchBackedPullRequestSearch(command.tab)
	if !targetConfigured || search.Refresh <= 0 {
		return
	}

	runtime := newPullRequestListWorkflowRuntime(program, gui)
	if runtime.executeWorkflowPlan == nil {
		return
	}
	runtime.executeWorkflowPlan(planScheduledPullRequestListReload(command.tab, true, targetConfigured))
}

func (command scheduledPullRequestRefreshCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || !program.hasDetailQueries() {
		return
	}
	runtime := newWorkflowShellRuntime(program, gui)
	if runtime.executeWorkflowPlan == nil {
		return
	}
	runtime.executeWorkflowPlan(planScheduledPullRequestRefresh(scheduledPullRequestRefreshPlanInput{
		detailSummaries:    append([]githubdomain.PullRequest(nil), command.detailSummaries...),
		diffSummaries:      append([]githubdomain.PullRequest(nil), command.diffSummaries...),
		hasDetailQueries:   true,
		detailLoadInFlight: program.pullRequestDetailLoadInFlight,
		diffLoadInFlight:   program.pullRequestDiffLoadInFlight,
	}))
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
