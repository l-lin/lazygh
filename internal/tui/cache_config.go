package tui

import (
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) ApplyCacheConfig(config appconfig.CacheConfig) error {
	if program == nil {
		return nil
	}

	applied, actualErr := executeApplyCacheConfigRuntime(newCacheConfigRuntime(), program.pullRequestCache, config)
	if actualErr != nil {
		return actualErr
	}
	return program.dispatchRuntimeMessage(applied.message())
}

type cacheConfigAppliedState struct {
	pullRequestCache      persistentPullRequestCache
	notificationDoneStore notificationDoneStore
	pastedPullRequests    []githubdomain.PullRequest
}

func (state cacheConfigAppliedState) message() MsgCacheConfigApplied {
	return MsgCacheConfigApplied{PullRequestCache: state.pullRequestCache, NotificationDoneStore: state.notificationDoneStore, PastedPullRequests: append([]githubdomain.PullRequest(nil), state.pastedPullRequests...)}
}
