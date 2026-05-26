package tui

import "strings"

type detailImageHTMLApplyKind int

const (
	detailImageHTMLApplyKindNone detailImageHTMLApplyKind = iota
	detailImageHTMLApplyKindPullRequestDescription
	detailImageHTMLApplyKindPullRequestComment
	detailImageHTMLApplyKindPullRequestInlineThreadComment
	detailImageHTMLApplyKindPullRequestInlineComment
	detailImageHTMLApplyKindPullRequestCommit
	detailImageHTMLApplyKindPullRequestDiffThreadComment
	detailImageHTMLApplyKindIssue
	detailImageHTMLApplyKindRelease
)

type detailImageHTMLApplyTarget struct {
	kind             detailImageHTMLApplyKind
	cacheKey         string
	markdownRevision string
	fallbackMarkdown string
	itemIndex        int
	fileIndex        int
	threadIndex      int
	commentIndex     int
}

func (target detailImageHTMLApplyTarget) canApply() bool {
	return target.kind != detailImageHTMLApplyKindNone && strings.TrimSpace(target.cacheKey) != "" && strings.TrimSpace(target.markdownRevision) != ""
}

func (target detailImageHTMLApplyTarget) matchesMarkdown(markdown string) bool {
	return target.canApply() && target.markdownRevision == detailImageMarkdownRevision(markdown)
}

func (program *Program) applyDetailImageHTMLRendered(target detailImageHTMLApplyTarget, renderedHTML string) bool {
	if program == nil || !target.canApply() {
		return false
	}

	trimmedRenderedHTML := strings.TrimSpace(renderedHTML)
	if trimmedRenderedHTML == "" {
		return false
	}

	switch target.kind {
	case detailImageHTMLApplyKindPullRequestDescription:
		cachedResult, ok := program.pullRequestDetailCache[target.cacheKey]
		if !ok || !target.matchesMarkdown(firstNonEmpty(cachedResult.detail.Body, target.fallbackMarkdown)) {
			return false
		}
		cachedResult.detail.BodyHTML = trimmedRenderedHTML
		program.pullRequestDetailCache[target.cacheKey] = cachedResult
		return true
	case detailImageHTMLApplyKindPullRequestComment:
		cachedResult, ok := program.pullRequestDetailCache[target.cacheKey]
		if !ok || target.itemIndex < 0 || target.itemIndex >= len(cachedResult.detail.Comments) || !target.matchesMarkdown(cachedResult.detail.Comments[target.itemIndex].Body) {
			return false
		}
		cachedResult.detail.Comments[target.itemIndex].BodyHTML = trimmedRenderedHTML
		program.pullRequestDetailCache[target.cacheKey] = cachedResult
		return true
	case detailImageHTMLApplyKindPullRequestInlineThreadComment:
		cachedResult, ok := program.pullRequestDetailCache[target.cacheKey]
		if !ok || target.threadIndex < 0 || target.threadIndex >= len(cachedResult.detail.InlineCommentThreads) || target.commentIndex < 0 || target.commentIndex >= len(cachedResult.detail.InlineCommentThreads[target.threadIndex].Comments) || !target.matchesMarkdown(cachedResult.detail.InlineCommentThreads[target.threadIndex].Comments[target.commentIndex].Body) {
			return false
		}
		cachedResult.detail.InlineCommentThreads[target.threadIndex].Comments[target.commentIndex].BodyHTML = trimmedRenderedHTML
		program.pullRequestDetailCache[target.cacheKey] = cachedResult
		return true
	case detailImageHTMLApplyKindPullRequestInlineComment:
		cachedResult, ok := program.pullRequestDetailCache[target.cacheKey]
		if !ok || target.itemIndex < 0 || target.itemIndex >= len(cachedResult.detail.InlineComments) || !target.matchesMarkdown(cachedResult.detail.InlineComments[target.itemIndex].Body) {
			return false
		}
		cachedResult.detail.InlineComments[target.itemIndex].BodyHTML = trimmedRenderedHTML
		program.pullRequestDetailCache[target.cacheKey] = cachedResult
		return true
	case detailImageHTMLApplyKindPullRequestCommit:
		cachedResult, ok := program.pullRequestDetailCache[target.cacheKey]
		if !ok || target.itemIndex < 0 || target.itemIndex >= len(cachedResult.detail.Commits) || !target.matchesMarkdown(cachedResult.detail.Commits[target.itemIndex].MessageBody) {
			return false
		}
		cachedResult.detail.Commits[target.itemIndex].MessageBodyHTML = trimmedRenderedHTML
		program.pullRequestDetailCache[target.cacheKey] = cachedResult
		return true
	case detailImageHTMLApplyKindPullRequestDiffThreadComment:
		cachedResult, ok := program.pullRequestDiffCache[target.cacheKey]
		if !ok || target.fileIndex < 0 || target.fileIndex >= len(cachedResult.data.Files) || target.threadIndex < 0 || target.threadIndex >= len(cachedResult.data.Files[target.fileIndex].Threads) || target.commentIndex < 0 || target.commentIndex >= len(cachedResult.data.Files[target.fileIndex].Threads[target.threadIndex].Comments) || !target.matchesMarkdown(cachedResult.data.Files[target.fileIndex].Threads[target.threadIndex].Comments[target.commentIndex].Body) {
			return false
		}
		cachedResult.data.Files[target.fileIndex].Threads[target.threadIndex].Comments[target.commentIndex].BodyHTML = trimmedRenderedHTML
		program.pullRequestDiffCache[target.cacheKey] = cachedResult
		return true
	case detailImageHTMLApplyKindIssue:
		cachedResult, ok := program.issueDetailCache[target.cacheKey]
		if !ok || !target.matchesMarkdown(cachedResult.detail.Body) {
			return false
		}
		cachedResult.detail.BodyHTML = trimmedRenderedHTML
		program.issueDetailCache[target.cacheKey] = cachedResult
		return true
	case detailImageHTMLApplyKindRelease:
		cachedResult, ok := program.releaseDetailCache[target.cacheKey]
		if !ok || !target.matchesMarkdown(cachedResult.detail.Body) {
			return false
		}
		cachedResult.detail.BodyHTML = trimmedRenderedHTML
		program.releaseDetailCache[target.cacheKey] = cachedResult
		return true
	default:
		return false
	}
}
