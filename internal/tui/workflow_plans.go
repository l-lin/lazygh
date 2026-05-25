package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type workflowPlan struct {
	messages []Msg
	commands []Cmd
}

func (plan *workflowPlan) append(other workflowPlan) {
	if len(other.messages) != 0 {
		plan.messages = append(plan.messages, other.messages...)
	}
	if len(other.commands) != 0 {
		plan.commands = append(plan.commands, other.commands...)
	}
}

func (plan *workflowPlan) addMessage(message Msg) {
	if message == nil {
		return
	}
	plan.messages = append(plan.messages, message)
}

func (plan *workflowPlan) addCommand(command Cmd) {
	if command == nil {
		return
	}
	plan.commands = append(plan.commands, command)
}

type sessionLoadPlanInput struct {
	hasSessionQueries bool
	loadStarted       bool
}

func planSessionLoad(input sessionLoadPlanInput) workflowPlan {
	if !input.hasSessionQueries || input.loadStarted {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	actual.addMessage(MsgConnectedUserLoadPlanned{})
	actual.addCommand(loadConnectedUserCmd{})
	return actual
}

type pullRequestListLoadPlanInput struct {
	activeTab             PullRequestTab
	targetTab             PullRequestTab
	loadStarted           bool
	hasPullRequestQueries bool
	forceReload           bool
}

func planPullRequestListLoad(input pullRequestListLoadPlanInput) workflowPlan {
	if !input.forceReload && (input.activeTab != input.targetTab || input.loadStarted) {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	actual.addCommand(hydratePullRequestsFromCacheCmd{tab: input.targetTab})
	if !input.hasPullRequestQueries {
		return actual
	}

	actual.addMessage(MsgPullRequestsLoadPlanned{Tab: input.targetTab})
	actual.addCommand(loadPullRequestsCmd{tab: input.targetTab})
	return actual
}

type notificationLoadPlanInput struct {
	reviewModeActive       bool
	loadStarted            bool
	hasNotificationQueries bool
	forceReload            bool
}

func planNotificationLoad(input notificationLoadPlanInput) workflowPlan {
	if input.reviewModeActive || (!input.forceReload && input.loadStarted) {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	actual.addCommand(hydrateNotificationsFromCacheCmd{})
	if !input.hasNotificationQueries {
		return actual
	}

	actual.addMessage(MsgNotificationsLoadPlanned{})
	actual.addCommand(loadNotificationsCmd{})
	return actual
}

type pullRequestDetailLoadPlanInput struct {
	summary                  githubdomain.PullRequest
	hasSelection             bool
	key                      string
	loadInFlight             bool
	visibleResult            pullRequestDetailResult
	visibleResultLoaded      bool
	hydrateVisibleResult     bool
	hasPullRequestDetailPort bool
}

func planPullRequestDetailLoad(input pullRequestDetailLoadPlanInput) workflowPlan {
	if !input.hasSelection || input.key == "" || input.loadInFlight {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	if input.hydrateVisibleResult {
		actual.addCommand(hydratePullRequestDetailFromCacheCmd{summary: input.summary})
	}
	if !pullRequestDetailNeedsRefresh(input.summary, input.visibleResult, input.visibleResultLoaded) || !input.hasPullRequestDetailPort {
		return actual
	}

	actual.addMessage(MsgPullRequestDetailLoadPlanned{Key: input.key})
	actual.addCommand(loadPullRequestDetailCmd{summary: input.summary})
	return actual
}

type pullRequestDiffLoadPlanInput struct {
	summary                githubdomain.PullRequest
	hasSelection           bool
	key                    string
	loadInFlight           bool
	visibleResult          pullRequestDiffResult
	visibleResultLoaded    bool
	hydrateVisibleResult   bool
	shouldLoadTeamOwners   bool
	hasPullRequestDiffPort bool
}

func planPullRequestDiffLoad(input pullRequestDiffLoadPlanInput) workflowPlan {
	if !input.hasSelection || input.key == "" || input.loadInFlight {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	if input.hydrateVisibleResult {
		actual.addCommand(hydratePullRequestDiffFromCacheCmd{summary: input.summary})
	}
	if !pullRequestDiffNeedsRefresh(input.summary, input.visibleResult, input.visibleResultLoaded, input.shouldLoadTeamOwners) || !input.hasPullRequestDiffPort {
		return actual
	}

	actual.addMessage(MsgPullRequestDiffLoadPlanned{Key: input.key})
	actual.addCommand(loadPullRequestDiffCmd{summary: input.summary})
	return actual
}

type notificationDetailLoadKind int

const (
	notificationDetailLoadKindNone notificationDetailLoadKind = iota
	notificationDetailLoadKindIssue
	notificationDetailLoadKindRelease
)

type notificationDetailLoadPlanInput struct {
	kind                   notificationDetailLoadKind
	repository             string
	number                 int
	key                    string
	loaded                 bool
	loadInFlight           bool
	hasNotificationQueries bool
}

func planNotificationDetailLoad(input notificationDetailLoadPlanInput) workflowPlan {
	if input.kind == notificationDetailLoadKindNone || input.key == "" || input.loaded || input.loadInFlight || !input.hasNotificationQueries {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	switch input.kind {
	case notificationDetailLoadKindIssue:
		actual.addMessage(MsgIssueDetailLoadPlanned{Repository: input.repository, Number: input.number})
		actual.addCommand(loadIssueDetailCmd{repository: input.repository, number: input.number})
	case notificationDetailLoadKindRelease:
		actual.addMessage(MsgReleaseDetailLoadPlanned{Repository: input.repository, ID: input.number})
		actual.addCommand(loadReleaseDetailCmd{repository: input.repository, id: input.number})
	}
	return actual
}

type detailImageHTMLLoadPlanInput struct {
	hasMarkdownHTMLRenderer bool
	sources                 []detailImageHTMLSource
	loadInFlightByKey       map[string]bool
	loadFailedByKey         map[string]bool
}

func planCurrentDetailImageHTMLLoads(input detailImageHTMLLoadPlanInput) workflowPlan {
	if !input.hasMarkdownHTMLRenderer {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	plannedKeys := map[string]bool{}
	for _, source := range input.sources {
		if !source.canLoadRenderedHTML() {
			continue
		}
		if input.loadInFlightByKey[source.key] || input.loadFailedByKey[source.key] || plannedKeys[source.key] {
			continue
		}

		plannedKeys[source.key] = true
		sourceCopy := source
		actual.addMessage(MsgCurrentDetailImageHTMLLoadPlanned{SourceKey: source.key})
		actual.addCommand(loadCurrentDetailImageHTMLCmd{source: sourceCopy})
	}
	return actual
}

type detailImageLoadPlanInput struct {
	detailImageStoreAvailable bool
	sources                   []detailImageHTMLSource
	imageAlreadyLoadedByURL   map[string]bool
	loadInFlightByURL         map[string]bool
	loadFailedByURL           map[string]bool
}

func planCurrentDetailImageLoads(input detailImageLoadPlanInput) workflowPlan {
	if !input.detailImageStoreAvailable {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	plannedURLs := map[string]bool{}
	for _, source := range input.sources {
		preparedMarkdown := prepareMarkdownForImageRendering(source.markdown, source.renderedHTML)
		for _, occurrence := range collectMarkdownImageOccurrences(preparedMarkdown) {
			imageURL := strings.TrimSpace(occurrence.imageURL)
			if imageURL == "" || input.imageAlreadyLoadedByURL[imageURL] || input.loadInFlightByURL[imageURL] || input.loadFailedByURL[imageURL] || plannedURLs[imageURL] {
				continue
			}

			plannedURLs[imageURL] = true
			actual.addMessage(MsgCurrentDetailImageLoadPlanned{ImageURL: imageURL})
			actual.addCommand(loadCurrentDetailImageCmd{imageURL: imageURL})
		}
	}
	return actual
}
