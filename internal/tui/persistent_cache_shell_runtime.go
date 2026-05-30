package tui

import (
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type persistentCacheShellAction interface {
	sync(persistentPullRequestCache)
}

type persistentCacheRuntimeState struct {
	pending []persistentCacheShellAction
}

type savePullRequestsPersistentCacheAction struct {
	search       appconfig.PullRequestSearch
	pullRequests []githubdomain.PullRequest
}

type saveNotificationsPersistentCacheAction struct {
	notifications []githubdomain.Notification
}

type savePullRequestDetailPersistentCacheAction struct {
	summary githubdomain.PullRequest
	detail  githubdomain.PullRequestDetail
}

type savePullRequestDiffPersistentCacheAction struct {
	summary githubdomain.PullRequest
	diff    githubdomain.PullRequestDiff
}

type invalidatePullRequestPersistentCacheAction struct {
	repository string
	number     int
}

func (state persistentCacheRuntimeState) withQueued(action persistentCacheShellAction) persistentCacheRuntimeState {
	if action == nil {
		return state
	}
	updated := append([]persistentCacheShellAction(nil), state.pending...)
	updated = append(updated, action)
	state.pending = updated
	return state
}

func (state persistentCacheRuntimeState) drain() ([]persistentCacheShellAction, persistentCacheRuntimeState) {
	if len(state.pending) == 0 {
		return nil, state
	}
	pending := append([]persistentCacheShellAction(nil), state.pending...)
	state.pending = nil
	return pending, state
}

func (action savePullRequestsPersistentCacheAction) sync(cache persistentPullRequestCache) {
	if cache == nil {
		return
	}
	_ = cache.SavePullRequests(action.search, action.pullRequests)
}

func (action saveNotificationsPersistentCacheAction) sync(cache persistentPullRequestCache) {
	if cache == nil {
		return
	}
	_ = cache.SaveNotifications(action.notifications)
}

func (action savePullRequestDetailPersistentCacheAction) sync(cache persistentPullRequestCache) {
	if cache == nil {
		return
	}
	_ = cache.SavePullRequestDetail(action.summary, action.detail)
}

func (action savePullRequestDiffPersistentCacheAction) sync(cache persistentPullRequestCache) {
	if cache == nil {
		return
	}
	_ = cache.SavePullRequestDiff(action.summary, action.diff)
}

func (action invalidatePullRequestPersistentCacheAction) sync(cache persistentPullRequestCache) {
	if cache == nil {
		return
	}
	_ = cache.InvalidatePullRequest(strings.TrimSpace(action.repository), action.number)
}

func (program *Program) updatePersistentCacheRuntime(update func(persistentCacheRuntimeState) persistentCacheRuntimeState) {
	if program == nil || update == nil {
		return
	}
	program.persistentCacheRuntime = update(program.persistentCacheRuntime)
}

func (program *Program) queuePersistentCacheShellAction(action persistentCacheShellAction) {
	if program == nil || action == nil {
		return
	}
	program.updatePersistentCacheRuntime(func(state persistentCacheRuntimeState) persistentCacheRuntimeState {
		return state.withQueued(action)
	})
}

func (program *Program) syncPersistentCacheShellState() {
	if program == nil {
		return
	}

	var pending []persistentCacheShellAction
	program.updatePersistentCacheRuntime(func(state persistentCacheRuntimeState) persistentCacheRuntimeState {
		pending, state = state.drain()
		return state
	})
	for _, action := range pending {
		if action == nil {
			continue
		}
		action.sync(program.pullRequestCache)
	}
}

func clonePersistentCachePullRequests(pullRequests []githubdomain.PullRequest) []githubdomain.PullRequest {
	if len(pullRequests) == 0 {
		return nil
	}
	return append([]githubdomain.PullRequest(nil), pullRequests...)
}

func clonePersistentCacheNotifications(notifications []githubdomain.Notification) []githubdomain.Notification {
	if len(notifications) == 0 {
		return nil
	}
	return append([]githubdomain.Notification(nil), notifications...)
}

func clonePersistentCachePullRequestDiff(diff githubdomain.PullRequestDiff) githubdomain.PullRequestDiff {
	diff.Files = clonePersistentCachePullRequestDiffFiles(diff.Files)
	diff.Threads = clonePersistentCacheReviewThreads(diff.Threads)
	return diff
}

func clonePersistentCachePullRequestDiffFiles(files []githubdomain.PullRequestDiffFile) []githubdomain.PullRequestDiffFile {
	if len(files) == 0 {
		return nil
	}
	cloned := make([]githubdomain.PullRequestDiffFile, 0, len(files))
	for _, file := range files {
		file.TeamOwners = append([]string(nil), file.TeamOwners...)
		cloned = append(cloned, file)
	}
	return cloned
}

func clonePersistentCacheReviewThreads(threads []githubdomain.ReviewThread) []githubdomain.ReviewThread {
	if len(threads) == 0 {
		return nil
	}
	cloned := make([]githubdomain.ReviewThread, 0, len(threads))
	for _, thread := range threads {
		thread.Comments = clonePullRequestComments(thread.Comments)
		cloned = append(cloned, thread)
	}
	return cloned
}
