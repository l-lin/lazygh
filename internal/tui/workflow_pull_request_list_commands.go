package tui

import (
	"github.com/jesseduffield/gocui"
	appconfig "github.com/l-lin/lazygh/internal/config"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestListWorkflowRuntime struct {
	workflowShellRuntime
	pullRequestsFromCache func(PullRequestTab) ([]githubdomain.PullRequest, bool)
	listPullRequests      func(PullRequestTab) ([]githubdomain.PullRequest, error)
}

type loadPullRequestsCmd struct {
	tab PullRequestTab
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
	runWorkflowCommandAsync(runtime.runAsync, func() {
		pullRequests, err := runtime.listPullRequests(command.tab)
		runtime.dispatchAsyncMessage(MsgPullRequestsLoaded{Tab: command.tab, PullRequests: pullRequests, Err: err})
	})
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
	runtime.executeUpdate(MsgPullRequestsCacheHydrated{Tab: command.tab, PullRequests: pullRequests})
}

func (command reloadPullRequestsTabCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newPullRequestListWorkflowRuntime(program, gui)
	if runtime.executeWorkflowPlan == nil || runtime.pullRequestListReloadPlan == nil {
		return
	}
	runtime.executeWorkflowPlan(runtime.pullRequestListReloadPlan(command.tab))
}
