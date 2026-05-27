package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestDetailWorkflowRuntime struct {
	workflowShellRuntime
	pullRequestDetailFromPersistentCache func(githubdomain.PullRequest) (pullRequestDetailResult, bool)
	pullRequestDetailCached              func(githubdomain.PullRequest) bool
	getPullRequestDetail                 func(githubdomain.PullRequest) (githubdomain.PullRequestDetail, error)
	getPendingPullRequestReviewState     func(githubdomain.PullRequest) (pendingPullRequestReviewState, bool)
}

type pullRequestDiffWorkflowRuntime struct {
	workflowShellRuntime
	pullRequestDiffFromPersistentCache  func(githubdomain.PullRequest) (pullRequestDiffResult, bool)
	pullRequestDiffCached               func(githubdomain.PullRequest) bool
	getPullRequestDiff                  func(githubdomain.PullRequest) (githubdomain.PullRequestDiff, error)
	shouldLoadPullRequestDiffTeamOwners bool
	getPullRequestFileTeamOwners        func(string, int, []string) (map[string][]string, error)
}

type hydratePullRequestDetailFromCacheCmd struct {
	summary githubdomain.PullRequest
}

type loadPullRequestDetailCmd struct {
	summary githubdomain.PullRequest
}

type hydratePullRequestDiffFromCacheCmd struct {
	summary githubdomain.PullRequest
}

type loadPullRequestDiffCmd struct {
	summary githubdomain.PullRequest
}

func newPullRequestDetailWorkflowRuntime(program *Program, gui *gocui.Gui) pullRequestDetailWorkflowRuntime {
	runtime := pullRequestDetailWorkflowRuntime{workflowShellRuntime: newWorkflowShellRuntime(program, gui)}
	if program == nil {
		return runtime
	}

	runtime.pullRequestDetailFromPersistentCache = program.pullRequestDetailFromPersistentCache
	runtime.pullRequestDetailCached = func(summary githubdomain.PullRequest) bool {
		return pullRequestDetailCachedInMemory(program.pullRequestDetailCache, summary)
	}
	if program.detailQueries != nil {
		runtime.getPullRequestDetail = func(summary githubdomain.PullRequest) (githubdomain.PullRequestDetail, error) {
			repository := pullRequestRepositoryName(summary.Repository)
			return program.detailQueries.GetPullRequestDetail(repository, summary.Number)
		}
	}
	if program.reviewMutations != nil {
		runtime.getPendingPullRequestReviewState = func(summary githubdomain.PullRequest) (pendingPullRequestReviewState, bool) {
			repository := pullRequestRepositoryName(summary.Repository)
			pendingReviewID, found, err := program.reviewMutations.GetPendingPullRequestReviewID(repository, summary.Number)
			if err != nil {
				return pendingPullRequestReviewState{}, false
			}
			state := pendingPullRequestReviewState{}
			if found {
				state.id = strings.TrimSpace(pendingReviewID)
			}
			return state, true
		}
	}
	return runtime
}

func newPullRequestDiffWorkflowRuntime(program *Program, gui *gocui.Gui) pullRequestDiffWorkflowRuntime {
	runtime := pullRequestDiffWorkflowRuntime{workflowShellRuntime: newWorkflowShellRuntime(program, gui)}
	if program == nil {
		return runtime
	}

	runtime.pullRequestDiffFromPersistentCache = program.pullRequestDiffFromPersistentCache
	runtime.pullRequestDiffCached = func(summary githubdomain.PullRequest) bool {
		return pullRequestDiffCachedInMemory(program.pullRequestDiffCache, summary)
	}
	if program.detailQueries != nil {
		runtime.shouldLoadPullRequestDiffTeamOwners = program.shouldLoadPullRequestDiffTeamOwners()
		runtime.getPullRequestFileTeamOwners = program.detailQueries.GetPullRequestFileTeamOwners
		runtime.getPullRequestDiff = func(summary githubdomain.PullRequest) (githubdomain.PullRequestDiff, error) {
			repository := pullRequestRepositoryName(summary.Repository)
			rawDiff, err := program.detailQueries.GetPullRequestDiff(repository, summary.Number)
			if err == nil {
				rawDiff = loadPullRequestDiffFileTeamOwners(runtime, repository, summary.Number, rawDiff)
			}
			return rawDiff, err
		}
	}
	return runtime
}

func pullRequestDetailCachedInMemory(cache map[string]pullRequestDetailResult, summary githubdomain.PullRequest) bool {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return false
	}
	_, ok := cache[key]
	return ok
}

func pullRequestDiffCachedInMemory(cache map[string]pullRequestDiffResult, summary githubdomain.PullRequest) bool {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return false
	}
	_, ok := cache[key]
	return ok
}

func (command hydratePullRequestDetailFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newPullRequestDetailWorkflowRuntime(program, gui)
	if runtime.pullRequestDetailCached == nil || runtime.pullRequestDetailFromPersistentCache == nil || runtime.executeUpdate == nil {
		return
	}
	if runtime.pullRequestDetailCached(command.summary) {
		return
	}
	result, ok := runtime.pullRequestDetailFromPersistentCache(command.summary)
	if !ok {
		return
	}
	runtime.executeUpdate(MsgPullRequestDetailCacheHydrated{Summary: command.summary, Result: result})
}

func (command loadPullRequestDetailCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newPullRequestDetailWorkflowRuntime(program, gui)
	if runtime.getPullRequestDetail == nil || runtime.dispatchAsyncMessage == nil {
		return
	}

	summary := command.summary
	runWorkflowCommandAsync(runtime.runAsync, func() {
		runtime.dispatchAsyncMessage(loadPullRequestDetailResult(runtime, summary))
	})
}

func loadPullRequestDetailResult(runtime pullRequestDetailWorkflowRuntime, summary githubdomain.PullRequest) MsgPullRequestDetailLoaded {
	detail, err := runtime.getPullRequestDetail(summary)
	pendingReviewState := pendingPullRequestReviewState{}
	pendingReviewStateKnown := false
	if runtime.getPendingPullRequestReviewState != nil {
		pendingReviewState, pendingReviewStateKnown = runtime.getPendingPullRequestReviewState(summary)
	}
	return MsgPullRequestDetailLoaded{
		Summary:                 summary,
		Detail:                  detail,
		Err:                     err,
		PendingReviewState:      pendingReviewState,
		PendingReviewStateKnown: pendingReviewStateKnown,
	}
}

func (command hydratePullRequestDiffFromCacheCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newPullRequestDiffWorkflowRuntime(program, gui)
	if runtime.pullRequestDiffCached == nil || runtime.pullRequestDiffFromPersistentCache == nil || runtime.executeUpdate == nil {
		return
	}
	if runtime.pullRequestDiffCached(command.summary) {
		return
	}
	result, ok := runtime.pullRequestDiffFromPersistentCache(command.summary)
	if !ok {
		return
	}
	runtime.executeUpdate(MsgPullRequestDiffCacheHydrated{Summary: command.summary, Result: result})
}

func (command loadPullRequestDiffCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newPullRequestDiffWorkflowRuntime(program, gui)
	if runtime.getPullRequestDiff == nil || runtime.dispatchAsyncMessage == nil {
		return
	}

	summary := command.summary
	runWorkflowCommandAsync(runtime.runAsync, func() {
		runtime.dispatchAsyncMessage(loadPullRequestDiffResult(runtime, summary))
	})
}

func loadPullRequestDiffResult(runtime pullRequestDiffWorkflowRuntime, summary githubdomain.PullRequest) MsgPullRequestDiffLoaded {
	rawDiff, err := runtime.getPullRequestDiff(summary)
	return MsgPullRequestDiffLoaded{Summary: summary, RawDiff: rawDiff, Err: err}
}

func loadPullRequestDiffFileTeamOwners(runtime pullRequestDiffWorkflowRuntime, repository string, number int, rawDiff githubdomain.PullRequestDiff) githubdomain.PullRequestDiff {
	if rawDiff.FileTeamOwnersAttempted || !runtime.shouldLoadPullRequestDiffTeamOwners || runtime.getPullRequestFileTeamOwners == nil {
		return rawDiff
	}

	rawDiff.FileTeamOwnersAttempted = true
	filePaths := pullRequestDiffFilePaths(rawDiff.Files)
	if len(filePaths) == 0 {
		return rawDiff
	}

	teamOwnersByPath, err := runtime.getPullRequestFileTeamOwners(repository, number, filePaths)
	if err != nil || len(teamOwnersByPath) == 0 {
		return rawDiff
	}

	rawDiff.Files = pullRequestDiffFilesWithTeamOwners(rawDiff.Files, teamOwnersByPath)
	return rawDiff
}
