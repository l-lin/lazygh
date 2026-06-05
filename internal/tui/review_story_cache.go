package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type storyReviewResult struct {
	story           reviewStoryData
	pendingReviewID string
	sourceUpdatedAt string
}

func (program *Program) storyReviewForSummary(summary any) (storyReviewResult, bool) {
	if program == nil {
		return storyReviewResult{}, false
	}

	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return storyReviewResult{}, false
	}

	key := pullRequestDetailKey(summaryValue.Repository, summaryValue.Number)
	if key == "" {
		return storyReviewResult{}, false
	}

	result, ok := program.storyReviewCache[key]
	if !ok {
		return storyReviewResult{}, false
	}
	return cloneStoryReviewResult(result), true
}

func storyReviewNeedsRefresh(summary githubdomain.PullRequest, result storyReviewResult, ok bool) bool {
	if !ok || strings.TrimSpace(result.pendingReviewID) == "" {
		return true
	}

	currentVersion := pullRequestSummaryVersion(summary)
	if currentVersion == "" {
		return false
	}

	return strings.TrimSpace(result.sourceUpdatedAt) != currentVersion
}

func (program *Program) cacheStoryReview(summary githubdomain.PullRequest, pendingReviewID string, story reviewStoryData) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withStoryReviewCached(key, storyReviewResult{
			story:           cloneReviewStoryData(story),
			pendingReviewID: strings.TrimSpace(pendingReviewID),
			sourceUpdatedAt: pullRequestSummaryVersion(summary),
		})
	})
}

func (program *Program) invalidatePullRequestStoryReview(repository string, number int) {
	key := pullRequestKeyFromIdentity(repository, number)
	if key == "" {
		return
	}

	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withoutStoryReview(key)
	})
}

func cloneStoryReviewResult(result storyReviewResult) storyReviewResult {
	result.story = cloneReviewStoryData(result.story)
	result.pendingReviewID = strings.TrimSpace(result.pendingReviewID)
	result.sourceUpdatedAt = strings.TrimSpace(result.sourceUpdatedAt)
	return result
}

func cloneReviewStoryData(data reviewStoryData) reviewStoryData {
	data.Chapters = cloneReviewStoryChapters(data.Chapters)
	data.Tree = cloneReviewDiffTree(data.Tree)
	return data
}

func cloneReviewStoryChapters(source []reviewStoryChapter) []reviewStoryChapter {
	copied := make([]reviewStoryChapter, 0, len(source))
	for _, chapter := range source {
		chapter.Files = append([]string(nil), chapter.Files...)
		chapter.FileIndexes = append([]int(nil), chapter.FileIndexes...)
		copied = append(copied, chapter)
	}
	return copied
}

func cloneReviewDiffTree(tree reviewDiffTree) reviewDiffTree {
	tree.Rows = append([]reviewDiffTreeRow(nil), tree.Rows...)
	return tree
}
