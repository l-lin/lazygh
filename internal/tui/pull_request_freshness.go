package tui

import (
	"sort"
	"strings"
	"time"

	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestFreshnessState struct {
	seen          bool
	seenUpdatedAt string
}

func (state pullRequestFreshnessState) isUnread(updatedAt string) bool {
	if !state.seen {
		return true
	}
	return comparePullRequestUpdatedAt(strings.TrimSpace(updatedAt), state.seenUpdatedAt) > 0
}

func prependPullRequestUnreadMarker(item Item) Item {
	marker := iconPullRequestUnread + " "
	markerSegment := ItemTitleSegment{Text: marker}
	for _, segment := range item.TitleSegments {
		if backgroundHex := strings.TrimSpace(segment.BackgroundHex); backgroundHex != "" {
			markerSegment.BackgroundHex = backgroundHex
			break
		}
	}
	item.Title = marker + item.Title
	item.TitleSegments = append([]ItemTitleSegment{markerSegment}, item.TitleSegments...)
	return item
}

func sortPullRequests(pullRequests []githubdomain.PullRequest) []githubdomain.PullRequest {
	sorted := append([]githubdomain.PullRequest(nil), pullRequests...)
	sort.SliceStable(sorted, func(left int, right int) bool {
		updatedComparison := comparePullRequestUpdatedAt(sorted[left].UpdatedAt, sorted[right].UpdatedAt)
		if updatedComparison != 0 {
			return updatedComparison > 0
		}

		leftRepository := strings.ToLower(strings.TrimSpace(sorted[left].Repository.NameWithOwner))
		rightRepository := strings.ToLower(strings.TrimSpace(sorted[right].Repository.NameWithOwner))
		if leftRepository != rightRepository {
			return leftRepository < rightRepository
		}
		if sorted[left].Number != sorted[right].Number {
			return sorted[left].Number < sorted[right].Number
		}
		return false
	})
	return sorted
}

func comparePullRequestUpdatedAt(left string, right string) int {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == right {
		return 0
	}

	leftTime, leftErr := time.Parse(time.RFC3339Nano, left)
	rightTime, rightErr := time.Parse(time.RFC3339Nano, right)
	if leftErr == nil && rightErr == nil {
		switch {
		case leftTime.Before(rightTime):
			return -1
		case leftTime.After(rightTime):
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(left, right)
}

func (store pullRequestListStore) withPullRequestFreshnessSnapshot(snapshot map[string]pullRequestFreshnessState) pullRequestListStore {
	if !store.pullRequestFreshnessTracking {
		store.pullRequestFreshness = clonePullRequestFreshnessStates(snapshot)
		store.pullRequestFreshnessTracking = true
		return store
	}

	merged := clonePullRequestFreshnessStates(store.pullRequestFreshness)
	for key, snapshotState := range snapshot {
		merged[key] = mergePullRequestFreshnessStates(merged[key], snapshotState)
	}
	store.pullRequestFreshness = merged
	return store
}

func (store pullRequestListStore) withPullRequestFreshnessTrackingEnabled() pullRequestListStore {
	store.pullRequestFreshnessTracking = true
	return store
}

func (store pullRequestListStore) withPullRequestSeen(key string, updatedAt string) pullRequestListStore {
	key = strings.TrimSpace(key)
	if key == "" {
		return store
	}
	store.pullRequestFreshness = clonePullRequestFreshnessStates(store.pullRequestFreshness)
	current := store.pullRequestFreshness[key]
	current.seen = true
	if comparePullRequestUpdatedAt(strings.TrimSpace(updatedAt), current.seenUpdatedAt) > 0 {
		current.seenUpdatedAt = strings.TrimSpace(updatedAt)
	}
	store.pullRequestFreshness[key] = current
	return store
}

func (store pullRequestListStore) withPullRequestTabMembership(tab PullRequestTab, pullRequests []githubdomain.PullRequest) pullRequestListStore {
	memberships := make(map[string]struct{}, len(pullRequests))
	for _, pullRequest := range pullRequests {
		if key := pullRequestDetailKey(pullRequest.Repository, pullRequest.Number); key != "" {
			memberships[key] = struct{}{}
		}
	}

	store.pullRequestTabMembership = clonePullRequestTabMemberships(store.pullRequestTabMembership)
	store.pullRequestTabMembership[tab] = memberships
	store.pullRequestTabMembershipKnown = clonePullRequestTabMembershipKnown(store.pullRequestTabMembershipKnown)
	store.pullRequestTabMembershipKnown[tab] = true
	return store
}

func (store pullRequestListStore) withoutPullRequestFreshnessAbsentFromTabs(tabs []PullRequestTab) pullRequestListStore {
	if len(tabs) == 0 {
		return store
	}
	for _, tab := range tabs {
		if !store.pullRequestTabMembershipKnown[tab] {
			return store
		}
	}

	present := map[string]struct{}{}
	for _, tab := range tabs {
		for key := range store.pullRequestTabMembership[tab] {
			present[key] = struct{}{}
		}
	}

	freshness := clonePullRequestFreshnessStates(store.pullRequestFreshness)
	for key := range freshness {
		if _, ok := present[key]; !ok {
			delete(freshness, key)
		}
	}
	store.pullRequestFreshness = freshness
	return store
}

func (store pullRequestListStore) withoutPullRequestTabMembership(tab PullRequestTab) pullRequestListStore {
	store.pullRequestTabMembership = clonePullRequestTabMemberships(store.pullRequestTabMembership)
	delete(store.pullRequestTabMembership, tab)
	store.pullRequestTabMembershipKnown = clonePullRequestTabMembershipKnown(store.pullRequestTabMembershipKnown)
	delete(store.pullRequestTabMembershipKnown, tab)
	return store
}

func (store pullRequestListStore) withoutPullRequestMemberships() pullRequestListStore {
	store.pullRequestTabMembership = map[PullRequestTab]map[string]struct{}{}
	store.pullRequestTabMembershipKnown = map[PullRequestTab]bool{}
	return store
}

func (store pullRequestListStore) withoutPullRequestFreshness() pullRequestListStore {
	store.pullRequestFreshness = map[string]pullRequestFreshnessState{}
	store.pullRequestFreshnessTracking = false
	return store.withoutPullRequestMemberships()
}

func (store pullRequestListStore) pullRequestFreshnessFor(key string) pullRequestFreshnessState {
	return store.pullRequestFreshness[strings.TrimSpace(key)]
}

func (store pullRequestListStore) pullRequestLoadGeneration(tab PullRequestTab) uint64 {
	return store.pullRequestLoadGenerations[tab]
}

func (store pullRequestListStore) withNextPullRequestLoadGeneration(tab PullRequestTab) (pullRequestListStore, uint64) {
	generations := clonePullRequestLoadGenerations(store.pullRequestLoadGenerations)
	generations[tab]++
	store.pullRequestLoadGenerations = generations
	return store, generations[tab]
}

func mergePullRequestFreshnessStates(left pullRequestFreshnessState, right pullRequestFreshnessState) pullRequestFreshnessState {
	if !left.seen && right.seen {
		return right
	}
	if !right.seen {
		return left
	}
	if comparePullRequestUpdatedAt(right.seenUpdatedAt, left.seenUpdatedAt) > 0 {
		left.seenUpdatedAt = right.seenUpdatedAt
	}
	left.seen = true
	return left
}

func clonePullRequestFreshnessStates(source map[string]pullRequestFreshnessState) map[string]pullRequestFreshnessState {
	cloned := make(map[string]pullRequestFreshnessState, len(source))
	for key, state := range source {
		cloned[key] = state
	}
	return cloned
}

func clonePullRequestTabMemberships(source map[PullRequestTab]map[string]struct{}) map[PullRequestTab]map[string]struct{} {
	cloned := make(map[PullRequestTab]map[string]struct{}, len(source))
	for tab, memberships := range source {
		cloned[tab] = make(map[string]struct{}, len(memberships))
		for key := range memberships {
			cloned[tab][key] = struct{}{}
		}
	}
	return cloned
}

func clonePullRequestTabMembershipKnown(source map[PullRequestTab]bool) map[PullRequestTab]bool {
	cloned := make(map[PullRequestTab]bool, len(source))
	for tab, known := range source {
		cloned[tab] = known
	}
	return cloned
}

func clonePullRequestLoadGenerations(source map[PullRequestTab]uint64) map[PullRequestTab]uint64 {
	cloned := make(map[PullRequestTab]uint64, len(source))
	for tab, generation := range source {
		cloned[tab] = generation
	}
	return cloned
}

func pullRequestFreshnessStatesFromCache(entries []persistcache.PullRequestFreshness) map[string]pullRequestFreshnessState {
	states := make(map[string]pullRequestFreshnessState, len(entries))
	for _, entry := range entries {
		key := pullRequestDetailKey(githubdomain.Repository{NameWithOwner: entry.Repository}, entry.Number)
		if key == "" {
			continue
		}
		states[key] = pullRequestFreshnessState{seen: entry.Seen, seenUpdatedAt: strings.TrimSpace(entry.SeenUpdatedAt)}
	}
	return states
}

func configuredPullRequestTabs(model *Model) []PullRequestTab {
	if model == nil {
		return nil
	}
	return model.PullRequestTabs()
}

func (program *Program) decoratePullRequestRows(rows []PullRequestRow) []PullRequestRow {
	if program == nil || program.model == nil {
		return rows
	}

	style := program.runtimeConfig.displayConfig.RepositoryStyle
	if program.pullRequestListStore == nil || !program.pullRequestListStore.pullRequestFreshnessTracking {
		return restyledPullRequestRowsWithRepositoryStyle(style, rows)
	}
	decorated := make([]PullRequestRow, 0, len(rows))
	for _, row := range rows {
		if row.Summary == nil {
			decorated = append(decorated, row)
			continue
		}
		key := pullRequestDetailKey(row.Summary.Repository, row.Summary.Number)
		state := program.pullRequestFreshnessFor(key)
		decorated = append(decorated, pullRequestRowWithUnread(style, *row.Summary, state.isUnread(row.Summary.UpdatedAt)))
	}
	return decorated
}

func (program *Program) pullRequestFreshnessFor(key string) pullRequestFreshnessState {
	if program == nil || program.pullRequestListStore == nil {
		return pullRequestFreshnessState{}
	}
	return program.pullRequestListStore.pullRequestFreshnessFor(key)
}

func (program *Program) markCurrentPullRequestSeen() {
	if program == nil || program.model == nil || program.model.Focus() != FocusDetailView {
		return
	}
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok && program.model.currentSideFocus() == FocusPullRequestsView {
		summary, ok = program.model.SelectedPullRequestSummary()
	}
	if !ok {
		return
	}
	program.markPullRequestSeen(summary)
}

func (program *Program) markPullRequestSeen(summary githubdomain.PullRequest) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}
	state := program.pullRequestFreshnessFor(key)
	if !state.isUnread(summary.UpdatedAt) {
		return
	}
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestSeen(key, summary.UpdatedAt)
	})
	program.refreshPullRequestUnreadMarkers()
	program.queuePersistentCacheShellAction(markPullRequestSeenPersistentCacheAction{
		repository: pullRequestRepositoryName(summary.Repository),
		number:     summary.Number,
		updatedAt:  summary.UpdatedAt,
	})
}

func (program *Program) applyPullRequestFreshnessSnapshot(entries []persistcache.PullRequestFreshness) {
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestFreshnessSnapshot(pullRequestFreshnessStatesFromCache(entries))
	})
	program.refreshPullRequestUnreadMarkers()
}

func (program *Program) queuePullRequestSearchMembershipReconciliation() {
	if program == nil {
		return
	}
	searches := append([]appconfig.PullRequestSearch(nil), program.runtimeConfig.pullRequestSearches...)
	searches = append(searches, pastedPullRequestsPersistentSearch())
	program.queuePersistentCacheShellAction(reconcilePullRequestSearchesPersistentCacheAction{searches: searches})
}
