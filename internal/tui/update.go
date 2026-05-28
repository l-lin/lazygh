package tui

type updateResult struct {
	handled  bool
	commands []Cmd
}

func handledUpdate(commands []Cmd) updateResult {
	return updateResult{handled: true, commands: commands}
}

func ignoredUpdate() updateResult {
	return updateResult{}
}

func Update(program *Program, msg Msg) []Cmd {
	if program == nil || msg == nil {
		return nil
	}
	defer program.resyncVisibleActionsPopupSearchInUpdate()

	if result := program.routeBootstrapFocusAndSidePaneSelection(msg); result.handled {
		return result.commands
	}
	if result := program.routeSearchPromptAndDraftUpdate(msg); result.handled {
		return result.commands
	}
	if result := program.routeFeedbackErrorAndModalEditorLifecycle(msg); result.handled {
		return result.commands
	}
	if result := program.routeBuildRunPopupLifecycle(msg); result.handled {
		return result.commands
	}
	if result := program.routeBrowserAndReviewNavigation(msg); result.handled {
		return result.commands
	}
	if result := program.routeDetailMotionAndLiveSync(msg); result.handled {
		return result.commands
	}
	if result := program.routeBrowserAndClipboardCompletions(msg); result.handled {
		return result.commands
	}
	if result := program.routeURLClipboardBrowserAndLinkFollowUps(msg); result.handled {
		return result.commands
	}
	if result := program.routeNotificationReviewTreeAndSearchNavigation(msg); result.handled {
		return result.commands
	}
	if result := program.routeSearchSubmissionAndPopupSearchEditor(msg); result.handled {
		return result.commands
	}
	if result := program.routePullRequestFeatureRequests(msg); result.handled {
		return result.commands
	}
	if result := program.routeMutationApplyResultsAndOptimisticFollowUp(msg); result.handled {
		return result.commands
	}
	if result := program.routePopupEditorSubmissionAndMutationRequests(msg); result.handled {
		return result.commands
	}
	if result := program.routeWorkflowPlanningAndCacheHydration(msg); result.handled {
		return result.commands
	}
	if result := program.routeAsyncLoadResultsAndTimerTicks(msg); result.handled {
		return result.commands
	}
	if result := program.routeAsyncFeatureCompletions(msg); result.handled {
		return result.commands
	}
	if result := program.routeActionsPopupChromeLifecycle(msg); result.handled {
		return result.commands
	}

	return nil
}
