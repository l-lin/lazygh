package tui

import appconfig "github.com/l-lin/lazygh/internal/config"

func (program *Program) routeRuntimeConfigMessages(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgKeymapOverridesApplied:
		program.applyKeymapOverridesApplied(actual)
		return handledUpdate(nil)
	case MsgPullRequestSearchesApplied:
		program.applyPullRequestSearchesApplied(actual)
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

func (program *Program) applyLinksConfigApplied(message MsgLinksConfigApplied) {
	resolved := appconfig.ResolveLinksConfig(message.Config)
	program.setLinkOpener(configuredSystemLinkOpener(program.linkOpener, resolved.OpenCommand))
}

func (program *Program) applyCacheConfigApplied(message MsgCacheConfigApplied) {
	program.setPersistentPullRequestCache(message.PullRequestCache)
	program.setNotificationDoneStore(message.NotificationDoneStore)
}

func (program *Program) applyStoryReviewConfigApplied(message MsgStoryReviewConfigApplied) {
	program.setRuntimeStoryReviewConfig(message.Config)
	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withStoryReviewCacheReset()
	})
}
