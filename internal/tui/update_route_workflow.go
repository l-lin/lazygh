package tui

func (program *Program) routeWorkflowPlanningAndCacheHydration(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgConnectedUserLoadPlanned:
		program.applyConnectedUserLoadPlanned()
		return handledUpdate(nil)
	case MsgPullRequestsLoadPlanned:
		program.applyPullRequestsLoadPlanned(actual)
		return handledUpdate(nil)
	case MsgNotificationsLoadPlanned:
		program.applyNotificationsLoadPlanned()
		return handledUpdate(nil)
	case MsgPullRequestDetailLoadPlanned:
		program.applyPullRequestDetailLoadPlanned(actual)
		return handledUpdate(nil)
	case MsgPullRequestDiffLoadPlanned:
		program.applyPullRequestDiffLoadPlanned(actual)
		return handledUpdate(nil)
	case MsgIssueDetailLoadPlanned:
		program.applyIssueDetailLoadPlanned(actual)
		return handledUpdate(nil)
	case MsgReleaseDetailLoadPlanned:
		program.applyReleaseDetailLoadPlanned(actual)
		return handledUpdate(nil)
	case MsgCurrentDetailImageHTMLLoadPlanned:
		program.applyCurrentDetailImageHTMLLoadPlanned(actual)
		return handledUpdate(nil)
	case MsgCurrentDetailImageLoadPlanned:
		program.applyCurrentDetailImageLoadPlanned(actual)
		return handledUpdate(nil)
	case MsgPullRequestsCacheHydrated:
		program.applyPullRequestsCacheHydrated(actual)
		return handledUpdate(nil)
	case MsgNotificationsCacheHydrated:
		program.applyNotificationsCacheHydrated(actual)
		return handledUpdate(nil)
	case MsgPullRequestDetailCacheHydrated:
		program.applyPullRequestDetailCacheHydrated(actual)
		return handledUpdate(nil)
	case MsgPullRequestDiffCacheHydrated:
		program.applyPullRequestDiffCacheHydrated(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeAsyncLoadResultsAndTimerTicks(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgConnectedUserLoaded:
		program.applyConnectedUserLoaded(actual)
		return handledUpdate(nil)
	case MsgPullRequestsLoaded:
		return handledUpdate(program.applyPullRequestsLoaded(actual))
	case MsgNotificationsLoaded:
		return handledUpdate(program.applyNotificationsLoaded(actual))
	case MsgPullRequestDetailLoaded:
		return handledUpdate(program.applyPullRequestDetailLoaded(actual))
	case MsgPullRequestDiffLoaded:
		return handledUpdate(program.applyPullRequestDiffLoaded(actual))
	case MsgCommitDiffLoaded:
		program.applyCommitDiffLoaded(actual)
		return handledUpdate(nil)
	case MsgIssueDetailLoaded:
		program.applyIssueDetailLoaded(actual)
		return handledUpdate(nil)
	case MsgReleaseDetailLoaded:
		program.applyReleaseDetailLoaded(actual)
		return handledUpdate(nil)
	case MsgCurrentDetailImageHTMLLoaded:
		program.applyCurrentDetailImageHTMLLoaded(actual)
		return handledUpdate(nil)
	case MsgCurrentDetailImageLoaded:
		program.applyCurrentDetailImageLoaded(actual)
		return handledUpdate(nil)
	case MsgLoadingSpinnerTick:
		program.applyLoadingSpinnerTick()
		return handledUpdate(nil)
	case MsgTransientErrorPopupExpired:
		program.applyTransientErrorPopupExpired(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeAsyncFeatureCompletions(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgActionsPopupAsyncGHCommandFinished:
		return handledUpdate(program.applyActionsPopupAsyncGHCommandFinished(actual))
	case MsgNotificationMutationStarted:
		program.applyNotificationMutationStarted(actual)
		return handledUpdate(nil)
	case MsgNotificationMutationFinished:
		return handledUpdate(program.applyNotificationMutationFinished(actual))
	case MsgStoryReviewPrepared:
		return handledUpdate(program.applyStoryReviewPrepared(actual))
	case MsgAssigneePickerSearchLoadingStarted:
		program.applyAssigneePickerSearchLoadingStarted(actual)
		return handledUpdate(nil)
	case MsgAssigneePickerSearchLoaded:
		return handledUpdate(program.applyAssigneePickerSearchLoaded(actual))
	case MsgPullRequestBuildRunLoaded:
		return handledUpdate(program.applyPullRequestBuildRunLoaded(actual))
	case MsgPullRequestBuildRunJobLogLoaded:
		return handledUpdate(program.applyPullRequestBuildRunJobLogLoaded(actual))
	default:
		return ignoredUpdate()
	}
}
