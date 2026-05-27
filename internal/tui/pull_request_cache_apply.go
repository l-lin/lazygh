package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type pullRequestDetailCacheApplyOptions struct {
	clearInFlight        bool
	invalidateDocuments  bool
	invalidatePersistent bool
}

type pullRequestDiffCacheApplyOptions struct {
	clearInFlight          bool
	invalidateReviewRender bool
	invalidateDetailDocs   bool
	invalidatePersistent   bool
}

func (program *Program) applyPullRequestDetailCacheResult(repository string, number int, result pullRequestDetailResult, options pullRequestDetailCacheApplyOptions) bool {
	if program == nil {
		return false
	}

	key := pullRequestMutationCacheKey(repository, number)
	if key == "" {
		return false
	}

	program.pullRequestDetailCache[key] = result
	if options.clearInFlight {
		delete(program.pullRequestDetailLoadInFlight, key)
	}
	if options.invalidateDocuments {
		program.invalidatePullRequestDetailDocumentCache()
	}
	if options.invalidatePersistent {
		program.invalidatePersistentPullRequest(repository, number)
	}
	return true
}

func (program *Program) applyPullRequestDiffCacheResult(repository string, number int, result pullRequestDiffResult, options pullRequestDiffCacheApplyOptions) bool {
	if program == nil {
		return false
	}

	key := pullRequestMutationCacheKey(repository, number)
	if key == "" {
		return false
	}

	program.pullRequestDiffCache[key] = result
	if options.clearInFlight {
		delete(program.pullRequestDiffLoadInFlight, key)
	}
	if options.invalidateReviewRender {
		program.invalidateReviewDiffRenderCache()
	}
	if options.invalidateDetailDocs {
		program.invalidatePullRequestDetailDocumentCache()
	}
	if options.invalidatePersistent {
		program.invalidatePersistentPullRequest(repository, number)
	}
	return true
}

func (program *Program) mutatePullRequestDetailOptimistically(repository string, number int, mutate func(*githubdomain.PullRequestDetail) bool) bool {
	if program == nil || mutate == nil {
		return false
	}

	key := pullRequestMutationCacheKey(repository, number)
	if key == "" {
		return false
	}
	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		return false
	}

	detail := result.detail
	if !mutate(&detail) {
		return false
	}

	result.detail = detail
	result.needsRefresh = true
	return program.applyPullRequestDetailCacheResult(repository, number, result, pullRequestDetailCacheApplyOptions{invalidateDocuments: true, invalidatePersistent: true})
}

func (program *Program) mutatePullRequestDiffOptimistically(repository string, number int, mutate func(*reviewDiffData) bool) bool {
	if program == nil || mutate == nil {
		return false
	}

	key := pullRequestMutationCacheKey(repository, number)
	if key == "" {
		return false
	}
	result, ok := program.pullRequestDiffCache[key]
	if !ok || result.err != nil {
		return false
	}

	data := result.data
	if !mutate(&data) {
		return false
	}

	result.data = data
	result.needsRefresh = true
	return program.applyPullRequestDiffCacheResult(repository, number, result, pullRequestDiffCacheApplyOptions{invalidateReviewRender: true, invalidateDetailDocs: true, invalidatePersistent: true})
}
