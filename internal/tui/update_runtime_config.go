package tui

func (program *Program) routeRuntimeConfigMessages(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgKeymapOverridesApplied:
		program.applyKeymapOverridesApplied(actual)
		return handledUpdate(nil)
	case MsgPullRequestSearchesApplied:
		program.applyPullRequestSearchesApplied(actual)
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

func (program *Program) applyStoryReviewConfigApplied(message MsgStoryReviewConfigApplied) {
	program.setRuntimeStoryReviewConfig(message.Config)
}
