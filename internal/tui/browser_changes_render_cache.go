package tui

func pullRequestChangesRenderedRowsCacheKey(summary any, width int) (pullRequestDetailDocumentCacheKey, bool) {
	if width < 1 {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	pullRequestKey := pullRequestDetailKey(summaryValue.Repository, summaryValue.Number)
	if pullRequestKey == "" {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	return pullRequestDetailDocumentCacheKey{pullRequestKey: pullRequestKey, tab: ChangesDetailTab, width: width}, true
}

func (program *Program) pullRequestChangesRenderedRowsForKey(key pullRequestDetailDocumentCacheKey) ([]reviewDiffRenderedRow, bool) {
	if len(program.pullRequestChangesRenderedRowsCache) == 0 {
		return nil, false
	}

	rows, ok := program.pullRequestChangesRenderedRowsCache[key]
	return rows, ok
}

func (program *Program) cachePullRequestChangesRenderedRows(key pullRequestDetailDocumentCacheKey, rows []reviewDiffRenderedRow) {
	if program.pullRequestChangesRenderedRowsCache == nil {
		program.pullRequestChangesRenderedRowsCache = map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow{}
	}

	program.pullRequestChangesRenderedRowsCache[key] = rows
}
