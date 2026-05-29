package tui

import "strings"

type reviewDiffRenderCacheKey struct {
	repositoryName     string
	pullRequestNumber  int
	pendingReviewID    string
	filePath           string
	width              int
	collapsedSignature uint64
}

type reviewDiffRenderCacheEntry struct {
	rows     []reviewDiffRenderedRow
	document detailDocument
}

func (program *Program) reviewDiffRenderKey(file reviewDiffFile, width int) reviewDiffRenderCacheKey {
	if width < 1 {
		width = 1
	}

	key := reviewDiffRenderCacheKey{
		filePath:           strings.TrimSpace(file.Path),
		width:              width,
		collapsedSignature: reviewDiffCollapsedStateSignature(file, program.navigationState.reviewSession.collapsedThreadIDs),
	}
	if !program.reviewModeActive() {
		return key
	}

	key.repositoryName = pullRequestRepositoryName(program.navigationState.reviewSession.summary.Repository)
	key.pullRequestNumber = program.navigationState.reviewSession.summary.Number
	key.pendingReviewID = strings.TrimSpace(program.navigationState.reviewSession.pendingReviewID)
	return key
}

func reviewDiffCollapsedStateSignature(file reviewDiffFile, collapsedThreadIDs map[string]bool) uint64 {
	if len(file.Threads) == 0 {
		return 0
	}

	const (
		fnv64Offset = 1469598103934665603
		fnv64Prime  = 1099511628211
	)

	hash := uint64(fnv64Offset)
	for _, thread := range file.Threads {
		threadID := strings.TrimSpace(thread.ID)
		for index := 0; index < len(threadID); index++ {
			hash ^= uint64(threadID[index])
			hash *= fnv64Prime
		}
		if reviewDiffThreadCollapsed(thread, collapsedThreadIDs) {
			hash ^= uint64('1')
		} else {
			hash ^= uint64('0')
		}
		hash *= fnv64Prime
		hash ^= uint64(0xff)
		hash *= fnv64Prime
	}
	return hash
}

func (program *Program) cachedReviewDiffRenderEntry(key reviewDiffRenderCacheKey) (reviewDiffRenderCacheEntry, bool) {
	if program.reviewDiffRenderCache == nil {
		return reviewDiffRenderCacheEntry{}, false
	}

	entry, ok := program.reviewDiffRenderCache[key]
	return entry, ok
}

func (program *Program) storeReviewDiffRenderEntry(key reviewDiffRenderCacheKey, entry reviewDiffRenderCacheEntry) {
	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withReviewDiffRenderEntryCached(key, entry)
	})
}

func (program *Program) invalidateReviewDiffRenderCache() {
	if len(program.reviewDiffRenderCache) == 0 {
		return
	}

	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withReviewDiffRenderCacheReset()
	})
}
