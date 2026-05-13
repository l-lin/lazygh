package tui

type pullRequestDetailDocumentCacheKey struct {
	pullRequestKey string
	tab            DetailTab
	width          int
}

func (program *Program) currentPullRequestDetailDocumentCacheKey(width int) (pullRequestDetailDocumentCacheKey, bool) {
	if program.reviewModeActive() || width < 1 {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestDetailDocumentCacheKey{}, false
	}
	if result, ok := program.pullRequestDetailForSummary(summary); !ok || result.err != nil {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	return pullRequestDetailDocumentCacheKey{
		pullRequestKey: pullRequestDetailKey(summary.Repository, summary.Number),
		tab:            program.activeDetailTab,
		width:          width,
	}, true
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
	if program.pullRequestDetailDocumentCache == nil {
		program.pullRequestDetailDocumentCache = map[pullRequestDetailDocumentCacheKey]detailDocument{}
	}

	program.pullRequestDetailDocumentCache[key] = document
}

func (program *Program) cachePullRequestConversationDocument(key pullRequestDetailDocumentCacheKey, document browserConversationDocument) {
	if program.pullRequestConversationDocumentCache == nil {
		program.pullRequestConversationDocumentCache = map[pullRequestDetailDocumentCacheKey]browserConversationDocument{}
	}

	program.pullRequestConversationDocumentCache[key] = document
}

func (program *Program) invalidatePullRequestDetailDocumentCache() {
	if len(program.pullRequestDetailDocumentCache) == 0 && len(program.pullRequestConversationDocumentCache) == 0 && len(program.pullRequestChangesRenderedRowsCache) == 0 {
		return
	}

	program.pullRequestDetailDocumentCache = map[pullRequestDetailDocumentCacheKey]detailDocument{}
	program.pullRequestConversationDocumentCache = map[pullRequestDetailDocumentCacheKey]browserConversationDocument{}
	program.pullRequestChangesRenderedRowsCache = map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow{}
}
