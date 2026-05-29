package tui

import "strings"

func (store detailStore) withPullRequestDetailLoadPlanned(key string) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.pullRequestDetailLoadInFlight = copyWorkflowStringBoolMap(store.pullRequestDetailLoadInFlight)
	store.pullRequestDetailLoadInFlight[trimmedKey] = true
	return store
}

func (store detailStore) withPullRequestDetailLoadCleared(key string) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" || !store.pullRequestDetailLoadInFlight[trimmedKey] {
		return store
	}
	store.pullRequestDetailLoadInFlight = copyWorkflowStringBoolMap(store.pullRequestDetailLoadInFlight)
	delete(store.pullRequestDetailLoadInFlight, trimmedKey)
	return store
}

func (store detailStore) withPullRequestDetailCached(key string, result pullRequestDetailResult) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.pullRequestDetailCache = copyPullRequestDetailResults(store.pullRequestDetailCache)
	store.pullRequestDetailCache[trimmedKey] = result
	return store
}

func (store detailStore) withoutPullRequestDetail(key string) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" || !detailResultKnown(store.pullRequestDetailCache, trimmedKey) {
		return store
	}
	store.pullRequestDetailCache = copyPullRequestDetailResults(store.pullRequestDetailCache)
	delete(store.pullRequestDetailCache, trimmedKey)
	return store
}

func (store detailStore) withIssueDetailLoadPlanned(key string) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.issueDetailLoadInFlight = copyWorkflowStringBoolMap(store.issueDetailLoadInFlight)
	store.issueDetailLoadInFlight[trimmedKey] = true
	return store
}

func (store detailStore) withIssueDetailLoaded(key string, result issueDetailResult) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.issueDetailCache = copyIssueDetailResults(store.issueDetailCache)
	store.issueDetailCache[trimmedKey] = result
	if store.issueDetailLoadInFlight[trimmedKey] {
		store.issueDetailLoadInFlight = copyWorkflowStringBoolMap(store.issueDetailLoadInFlight)
		delete(store.issueDetailLoadInFlight, trimmedKey)
	}
	return store
}

func (store detailStore) withReleaseDetailLoadPlanned(key string) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.releaseDetailLoadInFlight = copyWorkflowStringBoolMap(store.releaseDetailLoadInFlight)
	store.releaseDetailLoadInFlight[trimmedKey] = true
	return store
}

func (store detailStore) withReleaseDetailLoaded(key string, result releaseDetailResult) detailStore {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return store
	}
	store.releaseDetailCache = copyReleaseDetailResults(store.releaseDetailCache)
	store.releaseDetailCache[trimmedKey] = result
	if store.releaseDetailLoadInFlight[trimmedKey] {
		store.releaseDetailLoadInFlight = copyWorkflowStringBoolMap(store.releaseDetailLoadInFlight)
		delete(store.releaseDetailLoadInFlight, trimmedKey)
	}
	return store
}

func (store detailStore) withWorkflowStateReset() detailStore {
	store.pullRequestDetailCache = map[string]pullRequestDetailResult{}
	store.pullRequestDetailLoadInFlight = map[string]bool{}
	store.issueDetailCache = map[string]issueDetailResult{}
	store.issueDetailLoadInFlight = map[string]bool{}
	store.releaseDetailCache = map[string]releaseDetailResult{}
	store.releaseDetailLoadInFlight = map[string]bool{}
	return store
}

func detailResultKnown(results map[string]pullRequestDetailResult, key string) bool {
	_, ok := results[key]
	return ok
}

func copyPullRequestDetailResults(source map[string]pullRequestDetailResult) map[string]pullRequestDetailResult {
	copied := make(map[string]pullRequestDetailResult, len(source))
	for key, result := range source {
		copied[key] = result
	}
	return copied
}

func copyIssueDetailResults(source map[string]issueDetailResult) map[string]issueDetailResult {
	copied := make(map[string]issueDetailResult, len(source))
	for key, result := range source {
		copied[key] = result
	}
	return copied
}

func copyReleaseDetailResults(source map[string]releaseDetailResult) map[string]releaseDetailResult {
	copied := make(map[string]releaseDetailResult, len(source))
	for key, result := range source {
		copied[key] = result
	}
	return copied
}

func copyWorkflowStringBoolMap(source map[string]bool) map[string]bool {
	copied := make(map[string]bool, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}
