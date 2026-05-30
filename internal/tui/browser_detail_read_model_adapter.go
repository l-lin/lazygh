package tui

func (program *Program) browserDetailSectionCollapsed(sectionID string, collapsedByDefault bool) bool {
	if program == nil {
		return collapsedByDefault
	}
	return browserDetailSectionCollapsed(program.browserCollapsedSectionStates, sectionID, collapsedByDefault)
}

func (program *Program) currentBrowserDetailReadModel(summary any, detail any, width int) (browserDetailReadModel, bool) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return browserDetailReadModel{}, false
	}
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		return browserDetailReadModel{}, false
	}

	model := browserDetailReadModel{summary: summaryValue, detail: detailValue, width: maxInt(width, 1)}
	if program == nil {
		return model, true
	}
	model.markdownRenderer = program.markdownRenderer
	model.wordWrapEnabled = program.detailWordWrapEnabled()
	model.connectedUserLogin = program.currentConnectedUserLogin()
	model.collapsedSectionStates = program.browserCollapsedSectionStates
	return model, true
}

func (program *Program) currentPullRequestOverviewSections(summary any, detail any, width int) []browserDetailSection {
	model, ok := program.currentBrowserDetailReadModel(summary, detail, width)
	if !ok {
		return nil
	}
	return model.overviewSections()
}

func (program *Program) renderCurrentPullRequestOverview(summary any, detail any, width int) string {
	model, ok := program.currentBrowserDetailReadModel(summary, detail, width)
	if !ok {
		return ""
	}
	return model.renderOverview()
}

func (program *Program) browserOverviewSectionAtCursor(summary any, detail any, width int, cursorLine int) (browserDetailSectionCursor, bool) {
	model, ok := program.currentBrowserDetailReadModel(summary, detail, width)
	if !ok {
		return browserDetailSectionCursor{}, false
	}
	return model.overviewSectionAtCursor(cursorLine)
}

func (program *Program) currentPullRequestConversationDocument(summary any, detail any, width int) browserConversationDocument {
	model, ok := program.currentBrowserDetailReadModel(summary, detail, width)
	if !ok {
		return browserConversationDocument{}
	}
	if program == nil {
		return model.conversationDocument()
	}
	if cacheKey, ok := pullRequestConversationDocumentCacheKey(model.summary, model.width); ok {
		if document, ok := program.pullRequestConversationDocumentForKey(cacheKey); ok {
			return document
		}

		document := model.conversationDocument()
		program.cachePullRequestConversationDocument(cacheKey, document)
		return document
	}

	return model.conversationDocument()
}

func (program *Program) currentPullRequestConversationSections(summary any, detail any, width int) []browserDetailSection {
	document := program.currentPullRequestConversationDocument(summary, detail, width)
	return append([]browserDetailSection(nil), document.sections...)
}

func (program *Program) renderCurrentPullRequestConversationsTab(summary any, detail any, width int) string {
	document := program.currentPullRequestConversationDocument(summary, detail, width)
	if len(document.sections) == 0 {
		return "No comments yet."
	}
	return document.text
}

func (program *Program) browserConversationSectionAtCursor(summary any, detail any, width int, cursorLine int) (browserDetailSectionCursor, bool) {
	if cursorLine < 0 {
		return browserDetailSectionCursor{}, false
	}
	return program.currentPullRequestConversationDocument(summary, detail, width).sectionAtCursor(cursorLine)
}
