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

func (store reviewStore) withDiffWorkflowStateReset() reviewStore {
	store.pullRequestDiffCache = map[string]pullRequestDiffResult{}
	store.pullRequestDiffLoadInFlight = map[string]bool{}
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
