package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type browserDetailReadModel struct {
	summary                githubdomain.PullRequest
	detail                 githubdomain.PullRequestDetail
	width                  int
	markdownRenderer       MarkdownRenderer
	wordWrapEnabled        bool
	connectedUserLogin     string
	collapsedSectionStates map[string]bool
}

func (model browserDetailReadModel) overviewSections() []browserDetailSection {
	return buildPullRequestOverviewSections(model.summary, model.detail, maxInt(model.width, 1), model.collapsedSectionStates)
}

func (model browserDetailReadModel) renderOverview() string {
	return renderBrowserDetailSections(model.overviewSections(), false)
}

func (model browserDetailReadModel) overviewSectionAtCursor(cursorLine int) (browserDetailSectionCursor, bool) {
	overviewStartLine := browserDescriptionOverviewStartLine(model.summary, model.detail)
	relativeCursorLine := cursorLine - overviewStartLine
	if relativeCursorLine < 0 {
		return browserDetailSectionCursor{}, false
	}

	sectionAtCursor, ok := browserDetailSectionAtCursor(model.overviewSections(), relativeCursorLine, false)
	if !ok {
		return browserDetailSectionCursor{}, false
	}
	sectionAtCursor.headerLine += overviewStartLine
	sectionAtCursor.headerFocusLine += overviewStartLine
	return sectionAtCursor, true
}

func (model browserDetailReadModel) conversationSections() []browserDetailSection {
	return buildPullRequestConversationSections(model.summary, model.detail, maxInt(model.width, 1), model.markdownRenderer, model.wordWrapEnabled, model.connectedUserLogin, model.collapsedSectionStates)
}

func (model browserDetailReadModel) conversationDocument() browserConversationDocument {
	return buildBrowserConversationDocument(model.conversationSections())
}

func (model browserDetailReadModel) renderConversationsTab() string {
	document := model.conversationDocument()
	if len(document.sections) == 0 {
		return "No comments yet."
	}
	return document.text
}

func (model browserDetailReadModel) conversationSectionAtCursor(cursorLine int) (browserDetailSectionCursor, bool) {
	if cursorLine < 0 {
		return browserDetailSectionCursor{}, false
	}
	return model.conversationDocument().sectionAtCursor(cursorLine)
}
