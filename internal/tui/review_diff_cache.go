package tui

import (
	"fmt"
	"sort"
	"strings"
)

type reviewDiffRenderCacheKey struct {
	identity           string
	width              int
	collapsedSignature string
}

type reviewDiffRenderCacheEntry struct {
	rows     []reviewDiffRenderedRow
	document detailDocument
}

func (program *Program) reviewDiffRenderKey(file reviewDiffFile, width int) reviewDiffRenderCacheKey {
	if width < 1 {
		width = 1
	}

	return reviewDiffRenderCacheKey{
		identity:           program.reviewDiffRenderIdentity(file),
		width:              width,
		collapsedSignature: reviewDiffCollapsedStateSignature(file, program.reviewSession.collapsedThreadIDs),
	}
}

func (program *Program) reviewDiffRenderIdentity(file reviewDiffFile) string {
	path := strings.TrimSpace(file.Path)
	if !program.reviewModeActive() {
		return path
	}

	return fmt.Sprintf(
		"%s#%d:%s:%s",
		pullRequestRepositoryName(program.reviewSession.summary.Repository),
		program.reviewSession.summary.Number,
		strings.TrimSpace(program.reviewSession.pendingReviewID),
		path,
	)
}

func reviewDiffCollapsedStateSignature(file reviewDiffFile, collapsedThreadIDs map[string]bool) string {
	if len(file.Threads) == 0 {
		return ""
	}

	parts := make([]string, 0, len(file.Threads))
	for _, thread := range file.Threads {
		threadID := strings.TrimSpace(thread.ID)
		parts = append(parts, fmt.Sprintf("%s=%t", threadID, reviewDiffThreadCollapsed(thread, collapsedThreadIDs)))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (program *Program) cachedReviewDiffRenderEntry(key reviewDiffRenderCacheKey) (reviewDiffRenderCacheEntry, bool) {
	if program.reviewDiffRenderCache == nil {
		return reviewDiffRenderCacheEntry{}, false
	}

	entry, ok := program.reviewDiffRenderCache[key]
	return entry, ok
}

func (program *Program) storeReviewDiffRenderEntry(key reviewDiffRenderCacheKey, entry reviewDiffRenderCacheEntry) {
	if program.reviewDiffRenderCache == nil {
		program.reviewDiffRenderCache = map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry{}
	}

	program.reviewDiffRenderCache[key] = entry
}

func (program *Program) invalidateReviewDiffRenderCache() {
	if len(program.reviewDiffRenderCache) == 0 {
		return
	}

	program.reviewDiffRenderCache = map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry{}
}
