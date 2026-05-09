package tui

import (
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
)

type browserConversationDocument struct {
	text               string
	sections           []browserDetailSection
	sectionHeaderLines []browserConversationSectionHeaderLines
	lineMap            []browserConversationLineMapEntry
}

type browserConversationSectionHeaderLines struct {
	headerLine      int
	headerFocusLine int
}

type browserConversationLineMapEntry struct {
	sectionIndex             int
	bodyLine                 int
	inBody                   bool
	inlineThreadCommentIndex int
}

func buildBrowserConversationDocument(sections []browserDetailSection) browserConversationDocument {
	document := browserConversationDocument{
		text:               renderBrowserDetailSections(sections, false),
		sections:           append([]browserDetailSection(nil), sections...),
		sectionHeaderLines: make([]browserConversationSectionHeaderLines, 0, len(sections)),
		lineMap:            make([]browserConversationLineMapEntry, 0),
	}

	currentLine := 0
	for sectionIndex, section := range sections {
		headerLineCount := renderedTextLineCount(strings.TrimSpace(section.header))
		headerFocusLine := currentLine + minInt(section.headerFocusOffset, maxInt(headerLineCount-1, 0))
		document.sectionHeaderLines = append(document.sectionHeaderLines, browserConversationSectionHeaderLines{
			headerLine:      currentLine,
			headerFocusLine: headerFocusLine,
		})
		for range headerLineCount {
			document.lineMap = append(document.lineMap, browserConversationLineMapEntry{sectionIndex: sectionIndex, bodyLine: -1, inlineThreadCommentIndex: -1})
		}
		currentLine += headerLineCount

		if section.collapsed || strings.TrimSpace(section.body) == "" {
			continue
		}
		bodyLineCount := renderedTextLineCount(strings.TrimSpace(section.body))
		for bodyLine := range bodyLineCount {
			commentIndex := -1
			if bodyLine < len(section.inlineThreadBodyCommentIndex) {
				commentIndex = section.inlineThreadBodyCommentIndex[bodyLine]
			}
			document.lineMap = append(document.lineMap, browserConversationLineMapEntry{sectionIndex: sectionIndex, bodyLine: bodyLine, inBody: true, inlineThreadCommentIndex: commentIndex})
		}
		currentLine += bodyLineCount
	}

	return document
}

func (document browserConversationDocument) sectionAtCursor(cursorLine int) (browserDetailSectionCursor, bool) {
	if cursorLine < 0 || cursorLine >= len(document.lineMap) {
		return browserDetailSectionCursor{}, false
	}

	line := document.lineMap[cursorLine]
	if line.sectionIndex < 0 || line.sectionIndex >= len(document.sections) || line.sectionIndex >= len(document.sectionHeaderLines) {
		return browserDetailSectionCursor{}, false
	}

	section := document.sections[line.sectionIndex]
	headerLines := document.sectionHeaderLines[line.sectionIndex]
	return browserDetailSectionCursor{
		section:                  section,
		headerLine:               headerLines.headerLine,
		headerFocusLine:          headerLines.headerFocusLine,
		bodyLine:                 line.bodyLine,
		inBody:                   line.inBody,
		inlineThreadCommentIndex: line.inlineThreadCommentIndex,
	}, true
}

func (program *Program) currentPullRequestConversationDocument(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, width int) browserConversationDocument {
	if cacheKey, ok := pullRequestConversationDocumentCacheKey(summary, width); ok {
		if document, ok := program.pullRequestConversationDocumentForKey(cacheKey); ok {
			return document
		}

		document := buildBrowserConversationDocument(program.buildPullRequestConversationSections(summary, detail, width))
		program.cachePullRequestConversationDocument(cacheKey, document)
		return document
	}

	return buildBrowserConversationDocument(program.buildPullRequestConversationSections(summary, detail, width))
}

func pullRequestConversationDocumentCacheKey(summary githubcli.PullRequest, width int) (pullRequestDetailDocumentCacheKey, bool) {
	if width < 1 {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	if pullRequestKey == "" {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	return pullRequestDetailDocumentCacheKey{pullRequestKey: pullRequestKey, tab: CommentsDetailTab, width: width}, true
}
