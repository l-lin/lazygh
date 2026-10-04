package tui

func (program *Program) applyPullRequestDetailProgress(message MsgPullRequestDetailProgress) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}

	result, cached := program.pullRequestDetailCache[key]
	detail := clonePullRequestDetail(message.Detail)
	if cached && result.err == nil {
		// Retain cached enrichment until the complete result replaces it.
		retainedDetail := clonePullRequestDetail(result.detail)
		retainedDetail.Body = detail.Body
		retainedDetail.BodyHTML = detail.BodyHTML
		retainedDetail.Title = detail.Title
		retainedDetail.UpdatedAt = detail.UpdatedAt
		detail = retainedDetail
	}
	result.detail = detail
	result.err = nil
	result.needsRefresh = true
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withPullRequestDetailCached(key, result)
	})
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyPullRequestDiffProgress(message MsgPullRequestDiffProgress) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}

	result := program.pullRequestDiffCache[key]
	data := buildReviewDiffData(message.RawDiff)
	cachedFiles := make(map[string]reviewDiffFile, len(result.data.Files))
	for _, file := range result.data.Files {
		cachedFiles[file.Path] = file
	}
	for fileIndex := range data.Files {
		cachedFile := cachedFiles[data.Files[fileIndex].Path]
		if message.RawDiff.Threads == nil {
			data.Files[fileIndex].Threads = cachedFile.Threads
		}
		if !message.RawDiff.FileTeamOwnersAttempted {
			data.Files[fileIndex].TeamOwners = cachedFile.TeamOwners
		}
	}
	result.data = data
	result.err = nil
	result.needsRefresh = true
	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withPullRequestDiffCached(key, result)
	})
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}
