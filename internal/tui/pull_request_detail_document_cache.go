package tui

type pullRequestDetailDocumentCacheKey struct {
	pullRequestKey string
	tab            DetailTab
	width          int
}

func (program *Program) pullRequestDetailDocumentForKey(key pullRequestDetailDocumentCacheKey) (detailDocument, bool) {
	if len(program.pullRequestDetailDocumentCache) == 0 {
		return detailDocument{}, false
	}

	document, ok := program.pullRequestDetailDocumentCache[key]
	return document, ok
}

func (program *Program) pullRequestConversationDocumentForKey(key pullRequestDetailDocumentCacheKey) (browserConversationDocument, bool) {
	if len(program.pullRequestConversationDocumentCache) == 0 {
		return browserConversationDocument{}, false
	}

	document, ok := program.pullRequestConversationDocumentCache[key]
	return document, ok
}

func (program *Program) cachePullRequestDetailDocument(key pullRequestDetailDocumentCacheKey, document detailDocument) {
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withPullRequestDetailDocumentCached(key, document)
	})
}

func (program *Program) cachePullRequestConversationDocument(key pullRequestDetailDocumentCacheKey, document browserConversationDocument) {
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withPullRequestConversationDocumentCached(key, document)
	})
}

func (program *Program) invalidatePullRequestDetailDocumentCache() {
	if len(program.pullRequestDetailDocumentCache) == 0 && len(program.pullRequestConversationDocumentCache) == 0 && len(program.pullRequestChangesRenderedRowsCache) == 0 {
		return
	}

	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withDocumentRenderCachesReset()
	})
}
