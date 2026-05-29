package tui

import "strings"

func (store reviewStore) withPullRequestDiffLoadPlanned(key string) reviewStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.pullRequestDiffLoadInFlight = copyWorkflowStringBoolMap(store.pullRequestDiffLoadInFlight)
	store.pullRequestDiffLoadInFlight[trimmedKey] = true
	return store
}

func (store reviewStore) withPullRequestDiffLoadCleared(key string) reviewStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" || !store.pullRequestDiffLoadInFlight[trimmedKey] {
		return store
	}
	store.pullRequestDiffLoadInFlight = copyWorkflowStringBoolMap(store.pullRequestDiffLoadInFlight)
	delete(store.pullRequestDiffLoadInFlight, trimmedKey)
	return store
}

func (store reviewStore) withPullRequestDiffCached(key string, result pullRequestDiffResult) reviewStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.pullRequestDiffCache = copyPullRequestDiffResults(store.pullRequestDiffCache)
	store.pullRequestDiffCache[trimmedKey] = result
	return store
}

func (store reviewStore) withoutPullRequestDiff(key string) reviewStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" || !diffResultKnown(store.pullRequestDiffCache, trimmedKey) {
		return store
	}
	store.pullRequestDiffCache = copyPullRequestDiffResults(store.pullRequestDiffCache)
	delete(store.pullRequestDiffCache, trimmedKey)
	return store
}

func (store reviewStore) withPendingPullRequestReviewCached(key string, state pendingPullRequestReviewState) reviewStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.pendingPullRequestReviewCache = copyPendingPullRequestReviewStates(store.pendingPullRequestReviewCache)
	store.pendingPullRequestReviewCache[trimmedKey] = state
	return store
}

func (store reviewStore) withReviewDiffRenderEntryCached(key reviewDiffRenderCacheKey, entry reviewDiffRenderCacheEntry) reviewStore {
	store.reviewDiffRenderCache = copyReviewDiffRenderEntries(store.reviewDiffRenderCache)
	store.reviewDiffRenderCache[key] = cloneReviewDiffRenderCacheEntry(entry)
	return store
}

func (store reviewStore) withoutPendingPullRequestReview(key string) reviewStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	if _, ok := store.pendingPullRequestReviewCache[trimmedKey]; !ok {
		return store
	}
	store.pendingPullRequestReviewCache = copyPendingPullRequestReviewStates(store.pendingPullRequestReviewCache)
	delete(store.pendingPullRequestReviewCache, trimmedKey)
	return store
}

func (store reviewStore) withDiffWorkflowStateReset() reviewStore {
	store.pullRequestDiffCache = map[string]pullRequestDiffResult{}
	store.pullRequestDiffLoadInFlight = map[string]bool{}
	return store
}

func (store reviewStore) withReviewDiffRenderCacheReset() reviewStore {
	store.reviewDiffRenderCache = map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry{}
	return store
}

func (store reviewStore) withPendingReviewCacheReset() reviewStore {
	store.pendingPullRequestReviewCache = map[string]pendingPullRequestReviewState{}
	return store
}

func diffResultKnown(results map[string]pullRequestDiffResult, key string) bool {
	_, ok := results[key]
	return ok
}

func copyPullRequestDiffResults(source map[string]pullRequestDiffResult) map[string]pullRequestDiffResult {
	copied := make(map[string]pullRequestDiffResult, len(source))
	for key, result := range source {
		copied[key] = result
	}
	return copied
}

func copyPendingPullRequestReviewStates(source map[string]pendingPullRequestReviewState) map[string]pendingPullRequestReviewState {
	copied := make(map[string]pendingPullRequestReviewState, len(source))
	for key, state := range source {
		copied[key] = state
	}
	return copied
}

func copyReviewDiffRenderEntries(source map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry) map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry {
	copied := make(map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry, len(source))
	for key, entry := range source {
		copied[key] = cloneReviewDiffRenderCacheEntry(entry)
	}
	return copied
}

func cloneReviewDiffRenderCacheEntry(entry reviewDiffRenderCacheEntry) reviewDiffRenderCacheEntry {
	entry.rows = copyReviewDiffRenderedRows(entry.rows)
	return entry
}
