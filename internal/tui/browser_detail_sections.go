package tui

import (
	"fmt"
	"strconv"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	browserDetailExpandedChevron  = ""
	browserDetailCollapsedChevron = ""
)

type browserDetailSection struct {
	id           string
	header       string
	body         string
	collapsed    bool
	inlineThread *githubcli.PullRequestReviewThread
}

type browserDetailSectionCursor struct {
	section    browserDetailSection
	headerLine int
	bodyLine   int
	inBody     bool
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
		if lineIndex == 0 {
			return browserDetailSectionCursor{section: section, headerLine: headerLine, bodyLine: -1}, true
		}
		lineIndex--
		currentLine++

		if !section.collapsed {
			bodyLineCount := 0
			if strings.TrimSpace(section.body) != "" {
				bodyLineCount = renderedTextLineCount(strings.TrimSpace(section.body))
				if lineIndex >= 0 && lineIndex < bodyLineCount {
					return browserDetailSectionCursor{section: section, headerLine: headerLine, bodyLine: lineIndex, inBody: true}, true
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

func (program *Program) browserDetailSectionCollapsed(sectionID string, collapsedByDefault bool) bool {
	if program.browserCollapsedSectionStates != nil {
		if collapsed, ok := program.browserCollapsedSectionStates[strings.TrimSpace(sectionID)]; ok {
			return collapsed
		}
	}
	return collapsedByDefault
}

func (program *Program) setBrowserDetailSectionCollapsed(sectionID string, collapsed bool) {
	trimmedSectionID := strings.TrimSpace(sectionID)
	if trimmedSectionID == "" {
		return
	}
	if program.browserCollapsedSectionStates == nil {
		program.browserCollapsedSectionStates = map[string]bool{}
	}
	program.browserCollapsedSectionStates[trimmedSectionID] = collapsed
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) currentPullRequestOverviewSections(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, width int) []browserDetailSection {
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
		collapsed := program.browserDetailSectionCollapsed(sectionID, true)
		sections = append(sections, browserDetailSection{
			id:        sectionID,
			header:    renderBrowserDetailSectionHeader(pullRequestOverviewBlockHeadingText(block), collapsed, pullRequestOverviewStatusHex(block.Status)),
			body:      renderRoundedCommentBox(entries, width),
			collapsed: collapsed,
		})
	}
	return sections
}

func (program *Program) renderCurrentPullRequestOverview(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, width int) string {
	return renderBrowserDetailSections(program.currentPullRequestOverviewSections(summary, detail, width), false)
}

func browserDescriptionOverviewStartLine(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) int {
	header := strings.TrimSpace(renderPullRequestBrowserHeader(summary, detail))
	if header == "" {
		return 0
	}
	return renderedTextLineCount(header) + 1
}

func (program *Program) browserOverviewSectionAtCursor(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, width int, cursorLine int) (browserDetailSectionCursor, bool) {
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
	return sectionAtCursor, true
}

func (program *Program) currentPullRequestConversationSections(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, width int) []browserDetailSection {
	pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
	sections := make([]browserDetailSection, 0, len(detail.Comments)+maxInt(len(detail.InlineCommentThreads), len(detail.InlineComments)))
	commentBodyWidth := commentBoxInnerWidth(width)

	for index, comment := range detail.Comments {
		body := renderMarkdownWithFallback(comment.Body, program.markdownRenderer, commentBodyWidth, "No comment body.")
		sectionID := browserDetailSectionID(pullRequestKey, "comment", index, comment.ID)
		collapsed := program.browserDetailSectionCollapsed(sectionID, false)
		sections = append(sections, browserDetailSection{
			id:        sectionID,
			header:    renderBrowserDetailSectionHeader(renderPullRequestCommentConversationTitle(comment), collapsed, theme.InactiveTitleHex),
			body:      renderPullRequestCommentSection(comment, body, width),
			collapsed: collapsed,
		})
	}

	if len(detail.InlineCommentThreads) > 0 {
		for index, rawThread := range detail.InlineCommentThreads {
			thread := rawThread
			sectionID := browserDetailSectionID(pullRequestKey, "thread", index, thread.ID)
			collapsed := program.browserDetailSectionCollapsed(sectionID, thread.IsResolved)
			sections = append(sections, browserDetailSection{
				id:           sectionID,
				header:       renderBrowserDetailSectionHeader(renderPullRequestInlineThreadConversationTitle(thread), collapsed, theme.InactiveTitleHex),
				body:         renderPullRequestInlineCommentThreadBody(thread, program.markdownRenderer, width),
				collapsed:    collapsed,
				inlineThread: &thread,
			})
		}
		return sections
	}

	for index, comment := range detail.InlineComments {
		body := renderInlineCommentBody(comment.Body, program.markdownRenderer, commentBodyWidth)
		sectionID := browserDetailSectionID(pullRequestKey, "inline-comment", index, comment.ID)
		collapsed := program.browserDetailSectionCollapsed(sectionID, false)
		sections = append(sections, browserDetailSection{
			id:        sectionID,
			header:    renderBrowserDetailSectionHeader(renderPullRequestInlineCommentConversationTitle(comment), collapsed, theme.InactiveTitleHex),
			body:      renderPullRequestInlineCommentSection(comment, body, width),
			collapsed: collapsed,
		})
	}

	return sections
}

func (program *Program) renderCurrentPullRequestConversationsTab(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, width int) string {
	sections := program.currentPullRequestConversationSections(summary, detail, width)
	if len(sections) == 0 {
		return "No comments yet."
	}
	return renderBrowserDetailSections(sections, false)
}

func (program *Program) browserConversationSectionAtCursor(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, width int, cursorLine int) (browserDetailSectionCursor, bool) {
	if cursorLine < 0 {
		return browserDetailSectionCursor{}, false
	}
	return browserDetailSectionAtCursor(program.currentPullRequestConversationSections(summary, detail, width), cursorLine, false)
}

func renderPullRequestCommentConversationTitle(_ githubcli.PullRequestComment) string {
	return "Comment"
}

func renderPullRequestInlineCommentConversationTitle(comment githubcli.PullRequestInlineComment) string {
	line := pullRequestInlineCommentDisplayLine(comment)
	if line <= 0 {
		return "Inline comment"
	}
	return fmt.Sprintf("Comment on line %s%d", pullRequestInlineCommentSideLabel(comment), line)
}

func renderPullRequestInlineThreadConversationTitle(thread githubcli.PullRequestReviewThread) string {
	comment := pullRequestInlineCommentFromThread(thread)
	title := renderPullRequestInlineCommentConversationTitle(comment)
	if thread.IsResolved {
		title += " · resolved"
	}
	return title
}

func renderPullRequestInlineCommentThreadBody(thread githubcli.PullRequestReviewThread, renderer MarkdownRenderer, width int) string {
	return renderPullRequestInlineCommentThreadContent(thread, renderer, width)
}

func pullRequestInlineCommentSideLabel(comment githubcli.PullRequestInlineComment) string {
	_, _, side := pullRequestInlineCommentTargetRange(comment)
	switch side {
	case "LEFT":
		return "L"
	case "RIGHT":
		return "R"
	default:
		return "?"
	}
}

func pullRequestInlineCommentDisplayLine(comment githubcli.PullRequestInlineComment) int {
	_, endLine, _ := pullRequestInlineCommentTargetRange(comment)
	return endLine
}
