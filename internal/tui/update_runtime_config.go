package tui

import (
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) routeRuntimeConfigMessages(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgKeymapOverridesApplied:
		program.applyKeymapOverridesApplied(actual)
		return handledUpdate(nil)
	case MsgPullRequestSearchesApplied:
		program.applyPullRequestSearchesApplied(actual)
		return handledUpdate(nil)
	case MsgDisplayConfigApplied:
		program.applyDisplayConfigApplied(actual)
		return handledUpdate(nil)
	case MsgLinksConfigApplied:
		program.applyLinksConfigApplied(actual)
		return handledUpdate(nil)
	case MsgCacheConfigApplied:
		program.applyCacheConfigApplied(actual)
		return handledUpdate(nil)
	case MsgStoryReviewConfigApplied:
		program.applyStoryReviewConfigApplied(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) applyKeymapOverridesApplied(message MsgKeymapOverridesApplied) {
	program.setRuntimeKeymapOverrides(message.Overrides)
}

func (program *Program) applyDisplayConfigApplied(message MsgDisplayConfigApplied) {
	program.setRuntimeDisplayConfig(message.Config)
	program.restylePullRequestRows()
	program.restyleNotificationRows()
}

func (program *Program) applyLinksConfigApplied(message MsgLinksConfigApplied) {
	resolved := appconfig.ResolveLinksConfig(message.Config)
	program.setLinkOpener(configuredSystemLinkOpener(program.linkOpener, resolved.OpenCommand))
}

func (program *Program) applyCacheConfigApplied(message MsgCacheConfigApplied) {
	program.setPersistentPullRequestCache(message.PullRequestCache)
	program.setNotificationDoneStore(message.NotificationDoneStore)
	program.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestsLoaded(append([]githubdomain.PullRequest(nil), message.PastedPullRequests...))
	})
	program.syncPastedPullRequestTab()
}

func (program *Program) applyStoryReviewConfigApplied(message MsgStoryReviewConfigApplied) {
	program.setRuntimeStoryReviewConfig(message.Config)
	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withStoryReviewCacheReset()
	})
}
