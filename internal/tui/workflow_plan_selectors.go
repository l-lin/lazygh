package tui

import "strings"

func (program *Program) sessionLoadPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	return planSessionLoad(sessionLoadPlanInput{hasSessionQueries: program.hasSessionQueries(), loadStarted: program.connectedUserLoadStarted})
}

func (program *Program) pullRequestListLoadPlan(tab PullRequestTab) workflowPlan {
	if program == nil || program.isPastedPullRequestTab(tab) {
		return workflowPlan{}
	}
	return planPullRequestListLoad(pullRequestListLoadPlanInput{
		activeTab:             program.model.ActivePullRequestTab(),
		targetTab:             tab,
		loadStarted:           program.pullRequestsLoadStarted(tab),
		hasPullRequestQueries: program.hasPullRequestListQueries(),
	})
}

func (program *Program) pullRequestListReloadPlan(tab PullRequestTab) workflowPlan {
	if program == nil || program.isPastedPullRequestTab(tab) {
		return workflowPlan{}
	}
	return planPullRequestListLoad(pullRequestListLoadPlanInput{targetTab: tab, hasPullRequestQueries: program.hasPullRequestListQueries(), forceReload: true})
}

func (program *Program) notificationLoadPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	return planNotificationLoad(notificationLoadPlanInput{reviewModeActive: program.reviewModeActive(), loadStarted: program.notificationsLoadStarted, hasNotificationQueries: program.hasNotificationQueries()})
}

func (program *Program) notificationReloadPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	return planNotificationLoad(notificationLoadPlanInput{reviewModeActive: program.reviewModeActive(), hasNotificationQueries: program.hasNotificationQueries(), forceReload: true})
}

func (program *Program) selectedPullRequestDetailLoadPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	return planPullRequestDetailLoad(program.pullRequestDetailLoadPlanInput())
}

func (program *Program) pullRequestDetailLoadPlanInput() pullRequestDetailLoadPlanInput {
	input := pullRequestDetailLoadPlanInput{hasPullRequestDetailPort: program.hasDetailQueries()}
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return input
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	input.summary = summary
	input.hasSelection = key != ""
	input.key = key
	input.loadInFlight = program.pullRequestDetailLoadInFlight[key]
	if key == "" {
		return input
	}
	if actual, ok := program.pullRequestDetailForSummary(summary); ok {
		input.visibleResult = actual
		input.visibleResultLoaded = true
		return input
	}
	if actual, ok := program.pullRequestDetailFromPersistentCache(summary); ok {
		input.visibleResult = actual
		input.visibleResultLoaded = true
		input.hydrateVisibleResult = true
	}
	return input
}

func (program *Program) selectedPullRequestDiffLoadPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	return planPullRequestDiffLoad(program.pullRequestDiffLoadPlanInput())
}

func (program *Program) pullRequestDiffLoadPlanInput() pullRequestDiffLoadPlanInput {
	input := pullRequestDiffLoadPlanInput{hasPullRequestDiffPort: program.hasDetailQueries(), shouldLoadTeamOwners: program.shouldLoadPullRequestDiffTeamOwners()}
	summary, ok := program.selectedPullRequestSummaryForDiff()
	if !ok {
		return input
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	input.summary = summary
	input.hasSelection = key != ""
	input.key = key
	input.loadInFlight = program.pullRequestDiffLoadInFlight[key]
	if key == "" {
		return input
	}
	if actual, ok := program.pullRequestDiffForSummary(summary); ok {
		input.visibleResult = actual
		input.visibleResultLoaded = true
		return input
	}
	if actual, ok := program.pullRequestDiffFromPersistentCache(summary); ok {
		input.visibleResult = actual
		input.visibleResultLoaded = true
		input.hydrateVisibleResult = true
	}
	return input
}

func (program *Program) selectedNotificationDetailLoadPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	return planNotificationDetailLoad(program.notificationDetailLoadPlanInput())
}

func (program *Program) notificationDetailLoadPlanInput() notificationDetailLoadPlanInput {
	input := notificationDetailLoadPlanInput{hasNotificationQueries: program.hasNotificationQueries()}
	if program.reviewModeActive() || program.model.currentSideFocus() != FocusNotificationsView {
		return input
	}

	notification, ok := program.model.SelectedNotification()
	if !ok {
		return input
	}
	if _, ok := notification.PullRequestSummary(); ok {
		return input
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		input.kind = notificationDetailLoadKindIssue
		input.repository = repository
		input.number = number
		input.key = notificationDetailKey(repository, number)
		input.loaded = program.issueDetailLoaded(input.key)
		input.loadInFlight = program.issueDetailLoadInFlight[input.key]
		return input
	}
	if repository, id, ok := notification.ReleaseIdentity(); ok {
		input.kind = notificationDetailLoadKindRelease
		input.repository = repository
		input.number = id
		input.key = notificationDetailKey(repository, id)
		input.loaded = program.releaseDetailLoaded(input.key)
		input.loadInFlight = program.releaseDetailLoadInFlight[input.key]
	}
	return input
}

func (program *Program) currentDetailImageHTMLLoadsPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	return planCurrentDetailImageHTMLLoads(detailImageHTMLLoadPlanInput{
		hasMarkdownHTMLRenderer: program.hasMarkdownHTMLRenderer(),
		sources:                 program.currentDetailImageHTMLSources(),
		loadInFlightByKey:       program.detailImageHTMLLoadInFlight,
		loadFailedByKey:         program.detailImageHTMLLoadFailed,
	})
}

func (program *Program) currentDetailImageLoadsPlan() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}
	sources := program.currentDetailImageHTMLSources()
	return planCurrentDetailImageLoads(detailImageLoadPlanInput{
		detailImageStoreAvailable: program.detailImageStore != nil,
		sources:                   sources,
		imageAlreadyLoadedByURL:   program.loadedCurrentDetailImageURLs(sources),
		loadInFlightByURL:         program.detailImageLoadInFlight,
		loadFailedByURL:           program.detailImageLoadFailed,
	})
}

func (program *Program) loadedCurrentDetailImageURLs(sources []detailImageHTMLSource) map[string]bool {
	actual := map[string]bool{}
	if program == nil || program.detailImageStore == nil {
		return actual
	}

	for _, source := range sources {
		preparedMarkdown := prepareMarkdownForImageRendering(source.markdown, source.renderedHTML)
		for _, occurrence := range collectMarkdownImageOccurrences(preparedMarkdown) {
			imageURL := strings.TrimSpace(occurrence.imageURL)
			if imageURL == "" {
				continue
			}
			if _, ok := program.detailImageStore.ImageBySource(imageURL); ok {
				actual[imageURL] = true
			}
		}
	}
	return actual
}
