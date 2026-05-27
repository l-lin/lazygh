package tui

import (
	"fmt"
	"strconv"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

type browserDetailSection struct {
	id                           string
	header                       string
	headerFocusOffset            int
	body                         string
	collapsed                    bool
	overviewBlockTitle           string
	overviewEntries              []pullRequestOverviewEntry
	comment                      *githubdomain.PullRequestComment
	inlineComment                *githubdomain.PullRequestInlineComment
	inlineThread                 *githubdomain.PullRequestReviewThread
	inlineThreadBodyCommentIndex []int
}

type browserDetailSectionCursor struct {
	section                  browserDetailSection
	headerLine               int
	headerFocusLine          int
	bodyLine                 int
	inBody                   bool
	inlineThreadCommentIndex int
}

func renderBrowserDetailSectionHeader(title string, collapsed bool, foregroundHex string) string {
	chevron := browserDetailExpandedChevron
	if collapsed {
		chevron = browserDetailCollapsedChevron
	}
	return styleText(chevron+" "+strings.TrimSpace(title), foregroundColorEscape(foregroundHex))
}

func renderBrowserDetailSections(sections []browserDetailSection, includeBlankLines bool) string {
	if len(sections) == 0 {
		return ""
	}

	lines := make([]string, 0, len(sections)*4)
	for index, section := range sections {
		lines = append(lines, section.header)
		if !section.collapsed {
			trimmedBody := strings.TrimSpace(section.body)
			if trimmedBody != "" {
				lines = append(lines, strings.Split(trimmedBody, "\n")...)
			}
		}
		if includeBlankLines && index < len(sections)-1 {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

func browserDetailSectionAtCursor(sections []browserDetailSection, cursorLine int, includeBlankLines bool) (browserDetailSectionCursor, bool) {
	lineIndex := cursorLine
	currentLine := 0
	for sectionIndex, section := range sections {
		headerLine := currentLine
		headerLineCount := renderedTextLineCount(strings.TrimSpace(section.header))
		headerFocusLine := headerLine + minInt(section.headerFocusOffset, maxInt(headerLineCount-1, 0))
		if lineIndex >= 0 && lineIndex < headerLineCount {
			return browserDetailSectionCursor{section: section, headerLine: headerLine, headerFocusLine: headerFocusLine, bodyLine: -1}, true
		}
		lineIndex -= headerLineCount
		currentLine += headerLineCount

		if !section.collapsed {
			bodyLineCount := 0
			if strings.TrimSpace(section.body) != "" {
				bodyLineCount = renderedTextLineCount(strings.TrimSpace(section.body))
				if lineIndex >= 0 && lineIndex < bodyLineCount {
					return browserDetailSectionCursor{section: section, headerLine: headerLine, headerFocusLine: headerFocusLine, bodyLine: lineIndex, inBody: true}, true
				}
			}
			lineIndex -= bodyLineCount
			currentLine += bodyLineCount
		}

		if includeBlankLines && sectionIndex < len(sections)-1 {
			if lineIndex == 0 {
				return browserDetailSectionCursor{}, false
			}
			lineIndex--
			currentLine++
		}
	}

	return browserDetailSectionCursor{}, false
}

func browserDetailSectionID(pullRequestKey string, prefix string, index int, stableID string) string {
	trimmedStableID := strings.TrimSpace(stableID)
	if trimmedStableID == "" {
		trimmedStableID = strconv.Itoa(index)
	}
	return strings.TrimSpace(pullRequestKey) + ":" + strings.TrimSpace(prefix) + ":" + trimmedStableID
}

func browserDetailSectionCollapsed(collapsedSectionStates map[string]bool, sectionID string, collapsedByDefault bool) bool {
	if collapsedSectionStates != nil {
		if collapsed, ok := collapsedSectionStates[strings.TrimSpace(sectionID)]; ok {
			return collapsed
		}
	}
	return collapsedByDefault
}

func (program *Program) browserDetailSectionCollapsed(sectionID string, collapsedByDefault bool) bool {
	if program == nil {
		return collapsedByDefault
	}
	return browserDetailSectionCollapsed(program.browserCollapsedSectionStates, sectionID, collapsedByDefault)
}

func buildPullRequestOverviewSections(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail, width int, collapsedSectionStates map[string]bool) []browserDetailSection {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	overview := buildPullRequestOverviewSection(detail)
	blocks := []pullRequestOverviewBlock{overview.Reviewers, overview.MergeChecks, overview.Builds}
	sections := make([]browserDetailSection, 0, len(blocks))
	for index, block := range blocks {
		entries := renderPullRequestOverviewEntries(block.Entries)
		if strings.TrimSpace(entries) == "" {
			continue
		}

		sectionID := browserDetailSectionID(pullRequestKey, "overview", index, strings.ToLower(strings.ReplaceAll(strings.TrimSpace(block.Title), " ", "-")))
		collapsedByDefault := pullRequestOverviewBlockCollapsedByDefault(block)
		collapsed := browserDetailSectionCollapsed(collapsedSectionStates, sectionID, collapsedByDefault)
		sections = append(sections, browserDetailSection{
			id:                 sectionID,
			header:             renderBrowserDetailSectionHeader(pullRequestOverviewBlockHeadingText(block), collapsed, pullRequestOverviewStatusHex(block.Status)),
			body:               renderRoundedCommentBox(entries, width),
			collapsed:          collapsed,
			overviewBlockTitle: block.Title,
			overviewEntries:    append([]pullRequestOverviewEntry(nil), block.Entries...),
		})
	}
	return sections
}

func (program *Program) currentPullRequestOverviewSections(summary any, detail any, width int) []browserDetailSection {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return nil
	}
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		return nil
	}
	return buildPullRequestOverviewSections(summaryValue, detailValue, width, program.browserCollapsedSectionStates)
}

func pullRequestOverviewBlockCollapsedByDefault(block pullRequestOverviewBlock) bool {
	if strings.EqualFold(strings.TrimSpace(block.Title), "Reviewers") {
		return block.Status == pullRequestOverviewStatusSuccess
	}
	return block.Status != pullRequestOverviewStatusFailure
}

func (program *Program) renderCurrentPullRequestOverview(summary any, detail any, width int) string {
	return renderBrowserDetailSections(program.currentPullRequestOverviewSections(summary, detail, width), false)
}

func browserDescriptionOverviewStartLine(summary any, detail any) int {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return 0
	}
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		return 0
	}
	header := strings.TrimSpace(renderPullRequestBrowserHeader(summaryValue, detailValue))
	if header == "" {
		return 0
	}
	return renderedTextLineCount(header) + 1
}

func (program *Program) browserOverviewSectionAtCursor(summary any, detail any, width int, cursorLine int) (browserDetailSectionCursor, bool) {
	overviewStartLine := browserDescriptionOverviewStartLine(summary, detail)
	relativeCursorLine := cursorLine - overviewStartLine
	if relativeCursorLine < 0 {
		return browserDetailSectionCursor{}, false
	}
	sectionAtCursor, ok := browserDetailSectionAtCursor(program.currentPullRequestOverviewSections(summary, detail, width), relativeCursorLine, false)
	if !ok {
		return browserDetailSectionCursor{}, false
	}
	sectionAtCursor.headerLine += overviewStartLine
	sectionAtCursor.headerFocusLine += overviewStartLine
	return sectionAtCursor, true
}

func (program *Program) buildPullRequestConversationSections(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail, width int) []browserDetailSection {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	sections := make([]browserDetailSection, 0, len(detail.Comments)+maxInt(len(detail.InlineCommentThreads), len(detail.InlineComments)))
	commentBodyWidth := commentBoxInnerWidth(width)
	connectedUserLogin := program.currentConnectedUserLogin()

	for index, rawComment := range detail.Comments {
		comment := rawComment
		body := renderMarkdownWithFallback(prepareMarkdownForImageRendering(comment.Body, comment.BodyHTML), program.markdownRenderer, commentBodyWidth, "No comment body.")
		sectionID := browserDetailSectionID(pullRequestKey, "comment", index, comment.ID)
		collapsed := program.browserDetailSectionCollapsed(sectionID, false)
		sections = append(sections, browserDetailSection{
			id:        sectionID,
			header:    renderBrowserDetailSectionHeader(renderPullRequestCommentConversationTitle(comment), collapsed, theme.InactiveTitleHex),
			body:      renderPullRequestCommentSectionForViewer(comment, body, width, connectedUserLogin),
			collapsed: collapsed,
			comment:   &comment,
		})
	}

	if len(detail.InlineCommentThreads) > 0 {
		for index, rawThread := range detail.InlineCommentThreads {
			thread := rawThread
			sectionID := browserDetailSectionID(pullRequestKey, "thread", index, thread.ID)
			collapsed := program.browserDetailSectionCollapsed(sectionID, thread.IsResolved)
			sections = append(sections, browserDetailSection{
				id:                           sectionID,
				header:                       renderPullRequestInlineCommentThreadHeader(thread, collapsed, width),
				headerFocusOffset:            1,
				body:                         renderPullRequestInlineCommentThreadBodyForViewer(thread, program.markdownRenderer, width, connectedUserLogin),
				collapsed:                    collapsed,
				inlineThread:                 &thread,
				inlineThreadBodyCommentIndex: inlineThreadBodyCommentIndexesForViewer(thread, program.markdownRenderer, width, connectedUserLogin),
			})
		}
		return sections
	}

	for index, rawComment := range detail.InlineComments {
		comment := rawComment
		body := renderInlineCommentBodyForInlineComment(comment, program.markdownRenderer, commentBodyWidth)
		sectionID := browserDetailSectionID(pullRequestKey, "inline-comment", index, comment.ID)
		collapsed := program.browserDetailSectionCollapsed(sectionID, false)
		sections = append(sections, browserDetailSection{
			id:            sectionID,
			header:        renderBrowserDetailSectionHeader(renderPullRequestInlineCommentConversationTitle(comment), collapsed, theme.InactiveTitleHex),
			body:          renderPullRequestInlineCommentSectionForViewer(comment, body, width, connectedUserLogin),
			collapsed:     collapsed,
			inlineComment: &comment,
		})
	}

	return sections
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

func browserConversationInlineThreadCommentAtCursor(sectionAtCursor browserDetailSectionCursor) (githubdomain.PullRequestComment, bool) {
	if sectionAtCursor.section.inlineThread == nil || !sectionAtCursor.inBody {
		return githubdomain.PullRequestComment{}, false
	}
	thread := *sectionAtCursor.section.inlineThread
	commentIndex := sectionAtCursor.inlineThreadCommentIndex
	if commentIndex < 0 || commentIndex >= len(thread.Comments) {
		return githubdomain.PullRequestComment{}, false
	}
	return thread.Comments[commentIndex], true
}

func renderPullRequestCommentConversationTitle(_ githubdomain.PullRequestComment) string {
	return "Comment"
}

func renderPullRequestInlineCommentConversationTitle(comment githubdomain.PullRequestInlineComment) string {
	line := pullRequestInlineCommentDisplayLine(comment)
	if line <= 0 {
		return "Inline comment"
	}
	return fmt.Sprintf("Comment on line %s%d", pullRequestInlineCommentSideLabel(comment), line)
}
