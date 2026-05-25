package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (store *sessionStore) planLoad(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil || !program.hasSessionQueries() || store.connectedUserLoadStarted {
		return nil
	}

	store.connectedUserLoadStarted = true
	return []Cmd{loadConnectedUserCmd{}}
}

func (store *pullRequestListStore) planLoad(program *Program, gui *gocui.Gui, tab PullRequestTab) []Cmd {
	if store == nil || program == nil || gui == nil || store.pullRequestsLoadStarted(tab) || program.model.ActivePullRequestTab() != tab {
		return nil
	}

	commands := []Cmd{hydratePullRequestsFromCacheCmd{tab: tab}}
	if !program.hasPullRequestListQueries() {
		return commands
	}

	store.setPullRequestsLoadStarted(tab, true)
	store.setPullRequestsLoading(tab, true)
	return append(commands, loadPullRequestsCmd{tab: tab})
}

func (store *pullRequestListStore) planReload(program *Program, gui *gocui.Gui, tab PullRequestTab) []Cmd {
	if store == nil || program == nil || gui == nil {
		return nil
	}

	commands := []Cmd{hydratePullRequestsFromCacheCmd{tab: tab}}
	if !program.hasPullRequestListQueries() {
		return commands
	}

	store.setPullRequestsLoadStarted(tab, true)
	store.setPullRequestsLoading(tab, true)
	return append(commands, loadPullRequestsCmd{tab: tab})
}

func (store *notificationStore) planLoad(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil || program.reviewModeActive() || store.notificationsLoadStarted {
		return nil
	}

	commands := []Cmd{hydrateNotificationsFromCacheCmd{}}
	if !program.hasNotificationQueries() {
		return commands
	}

	store.notificationsLoadStarted = true
	store.notificationsLoading = true
	store.notificationsLoadingDetailMessage = notificationsLoadingDetail
	return append(commands, loadNotificationsCmd{})
}

func (store *notificationStore) planReload(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil || program.reviewModeActive() {
		return nil
	}

	commands := []Cmd{hydrateNotificationsFromCacheCmd{}}
	if !program.hasNotificationQueries() {
		return commands
	}

	store.notificationsLoadStarted = true
	store.notificationsLoading = true
	store.notificationsLoadingDetailMessage = notificationsLoadingDetail
	return append(commands, loadNotificationsCmd{})
}

func (store *detailStore) planSelectedPullRequestDetailLoad(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil {
		return nil
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return nil
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" || store.pullRequestDetailLoadInFlight[key] {
		return nil
	}

	program.hydratePullRequestDetailFromCache(summary)
	cachedResult, cached := program.pullRequestDetailForSummary(summary)
	if !program.pullRequestDetailNeedsRefresh(summary, cachedResult, cached) || !program.hasDetailQueries() {
		return nil
	}

	store.pullRequestDetailLoadInFlight[key] = true
	return []Cmd{loadPullRequestDetailCmd{summary: summary}}
}

func (store *reviewStore) planSelectedPullRequestDiffLoad(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil {
		return nil
	}

	summary, ok := program.selectedPullRequestSummaryForDiff()
	if !ok {
		return nil
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" || store.pullRequestDiffLoadInFlight[key] {
		return nil
	}

	program.hydratePullRequestDiffFromCache(summary)
	cachedResult, cached := program.pullRequestDiffForSummary(summary)
	if !program.pullRequestDiffNeedsRefresh(summary, cachedResult, cached) || !program.hasDetailQueries() {
		return nil
	}

	store.pullRequestDiffLoadInFlight[key] = true
	return []Cmd{loadPullRequestDiffCmd{summary: summary}}
}

func (store *detailStore) planSelectedNotificationDetailLoad(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil || program.reviewModeActive() || program.model.currentSideFocus() != FocusNotificationsView {
		return nil
	}

	notification, ok := program.model.SelectedNotification()
	if !ok {
		return nil
	}
	if _, ok := notification.PullRequestSummary(); ok {
		return nil
	}

	if repository, number, ok := notification.IssueIdentity(); ok {
		key := notificationDetailKey(repository, number)
		if key == "" || store.issueDetailLoadInFlight[key] || program.issueDetailLoaded(key) || !program.hasNotificationQueries() {
			return nil
		}
		store.issueDetailLoadInFlight[key] = true
		return []Cmd{loadIssueDetailCmd{repository: repository, number: number}}
	}

	if repository, id, ok := notification.ReleaseIdentity(); ok {
		key := notificationDetailKey(repository, id)
		if key == "" || store.releaseDetailLoadInFlight[key] || program.releaseDetailLoaded(key) || !program.hasNotificationQueries() {
			return nil
		}
		store.releaseDetailLoadInFlight[key] = true
		return []Cmd{loadReleaseDetailCmd{repository: repository, id: id}}
	}

	return nil
}

func (store *imageLoadCoordinator) planCurrentDetailImageHTMLLoads(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil || !program.hasMarkdownHTMLRenderer() {
		return nil
	}

	commands := make([]Cmd, 0)
	for _, source := range program.currentDetailImageHTMLSources() {
		if !source.canLoadRenderedHTML() {
			continue
		}
		if store.detailImageHTMLLoadInFlight[source.key] || store.detailImageHTMLLoadFailed[source.key] {
			continue
		}

		sourceCopy := source
		store.detailImageHTMLLoadInFlight[source.key] = true
		commands = append(commands, loadCurrentDetailImageHTMLCmd{source: sourceCopy})
	}
	return commands
}

func (store *imageLoadCoordinator) planCurrentDetailImageLoads(program *Program, gui *gocui.Gui) []Cmd {
	if store == nil || program == nil || gui == nil || program.detailImageStore == nil {
		return nil
	}

	commands := make([]Cmd, 0)
	for _, source := range program.currentDetailImageHTMLSources() {
		preparedMarkdown := prepareMarkdownForImageRendering(source.markdown, source.renderedHTML)
		for _, occurrence := range collectMarkdownImageOccurrences(preparedMarkdown) {
			imageURL := strings.TrimSpace(occurrence.imageURL)
			if imageURL == "" {
				continue
			}
			if _, ok := program.detailImageStore.ImageBySource(imageURL); ok {
				continue
			}
			if store.detailImageLoadInFlight[imageURL] || store.detailImageLoadFailed[imageURL] {
				continue
			}

			imageURLCopy := imageURL
			store.detailImageLoadInFlight[imageURLCopy] = true
			commands = append(commands, loadCurrentDetailImageCmd{imageURL: imageURLCopy})
		}
	}
	return commands
}
