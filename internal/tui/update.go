package tui

type updateResult struct {
	handled  bool
	commands []Cmd
}

type updateRoute func(Msg) updateResult

func handledUpdate(commands []Cmd) updateResult {
	return updateResult{handled: true, commands: commands}
}

func ignoredUpdate() updateResult {
	return updateResult{}
}

func (program *Program) routeUpdateCategory(msg Msg, routes ...updateRoute) updateResult {
	for _, route := range routes {
		if result := route(msg); result.handled {
			return result
		}
	}
	return ignoredUpdate()
}

func (program *Program) routeLifecycleAndEditorMessages(msg Msg) updateResult {
	return program.routeUpdateCategory(msg,
		program.routeBootstrapFocusAndSidePaneSelection,
		program.routeSearchPromptAndDraftUpdate,
		program.routeRuntimeConfigMessages,
		program.routeFeedbackErrorAndModalEditorLifecycle,
		program.routeBuildRunPopupLifecycle,
	)
}

func (program *Program) routeNavigationAndDetailMessages(msg Msg) updateResult {
	return program.routeUpdateCategory(msg,
		program.routeBrowserAndReviewNavigation,
		program.routeDetailMotionAndLiveSync,
	)
}

func (program *Program) routeBrowserClipboardAndLinkMessages(msg Msg) updateResult {
	return program.routeUpdateCategory(msg,
		program.routeBrowserAndClipboardCompletions,
		program.routeURLClipboardBrowserAndLinkFollowUps,
	)
}

func (program *Program) routeNotificationAndSearchMessages(msg Msg) updateResult {
	return program.routeUpdateCategory(msg,
		program.routeNotificationReviewTreeAndSearchNavigation,
		program.routeSearchSubmissionAndPopupSearchEditor,
	)
}

func (program *Program) routeMutationMessages(msg Msg) updateResult {
	return program.routeUpdateCategory(msg,
		program.routePullRequestFeatureRequests,
		program.routeMutationApplyResultsAndOptimisticFollowUp,
		program.routePopupEditorSubmissionAndMutationRequests,
	)
}

func (program *Program) routeWorkflowMessages(msg Msg) updateResult {
	return program.routeUpdateCategory(msg,
		program.routeWorkflowPlanningAndCacheHydration,
		program.routeAsyncLoadResultsAndTimerTicks,
		program.routeAsyncFeatureCompletions,
	)
}

func (program *Program) routeActionsPopupMessages(msg Msg) updateResult {
	return program.routeUpdateCategory(msg,
		program.routeActionsPopupChromeLifecycle,
	)
}

func Update(program *Program, msg Msg) []Cmd {
	if program == nil || msg == nil {
		return nil
	}

	if result := program.routeLifecycleAndEditorMessages(msg); result.handled {
		return result.commands
	}
	if result := program.routeNavigationAndDetailMessages(msg); result.handled {
		return result.commands
	}
	if result := program.routeBrowserClipboardAndLinkMessages(msg); result.handled {
		return result.commands
	}
	if result := program.routeNotificationAndSearchMessages(msg); result.handled {
		return result.commands
	}
	if result := program.routeMutationMessages(msg); result.handled {
		return result.commands
	}
	if result := program.routeWorkflowMessages(msg); result.handled {
		return result.commands
	}
	if result := program.routeActionsPopupMessages(msg); result.handled {
		return result.commands
	}

	return nil
}
