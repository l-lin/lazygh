package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func renderPullRequestDetailHeader(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	headerLines := []string{
		renderPullRequestContextLine(summary, detail),
		pullRequestTitleText(firstNonEmpty(detail.Title, summary.Title)),
		renderPullRequestMetaLine(summary, detail),
	}
	for _, line := range []string{
		renderPullRequestLabelsLine(detail.Labels),
		renderPullRequestAssigneesLine(detail.Assignees),
		renderPullRequestReviewRequestsLine(detail.ReviewRequests),
		renderPullRequestApprovalsLine(detail.Reviews),
	} {
		if strings.TrimSpace(line) == "" {
			continue
		}
		headerLines = append(headerLines, line)
	}

	return strings.Join(headerLines, "\n")
}

func renderPullRequestDescription(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, renderer MarkdownRenderer, width int) string {
	return renderMarkdownWithFallback(detailBody(detail, summary), renderer, width, "No description available.")
}

func renderPullRequestCommentsTab(comments []githubcli.PullRequestComment, inlineThreads []githubcli.PullRequestReviewThread, inlineComments []githubcli.PullRequestInlineComment, renderer MarkdownRenderer, width int) string {
	sections := buildPullRequestCommentsRenderedSections(comments, inlineThreads, inlineComments, renderer, width)
	if len(sections) == 0 {
		return "No comments yet."
	}

	texts := make([]string, 0, len(sections))
	for _, section := range sections {
		texts = append(texts, section.text)
	}
	return strings.Join(texts, "\n\n")
}

func renderPullRequestDetailLoading(summary githubcli.PullRequest, spinner string) string {
	return renderPullRequestDetailContent(
		renderPullRequestDetailHeader(summary, githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, State: summary.State, UpdatedAt: summary.UpdatedAt}),
		strings.TrimSpace(spinner),
	)
}

func renderPullRequestDetailError(summary githubcli.PullRequest, err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub found a new way to be unhelpful."
	}

	fallback := strings.TrimSpace(pullRequestRow(summary).Item.Detail)
	if fallback == "" {
		fallback = "No fallback detail available."
	}

	return renderPullRequestDetailContent(
		renderPullRequestDetailHeader(summary, githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, State: summary.State, UpdatedAt: summary.UpdatedAt}),
		fmt.Sprintf("Could not load rich pull request detail.\n\n%s\n\n%s", message, fallback),
	)
}

func renderPullRequestMetaLine(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	parts := make([]string, 0, 4)

	baseRefName := strings.TrimSpace(detail.BaseRefName)
	headRefName := strings.TrimSpace(detail.HeadRefName)
	if baseRefName != "" || headRefName != "" {
		parts = append(parts, fmt.Sprintf("%s %s ← %s", detailBranchIcon, compactBranchLabel(valueOrDash(baseRefName)), compactBranchLabel(valueOrDash(headRefName))))
	}

	parts = append(parts, renderPullRequestStatusBadge(detailStatus(detail, summary)))

	checkSummary := summarizeStatusChecks(detail.StatusCheckRollup)
	if checkSummary != "-" {
		parts = append(parts, fmt.Sprintf("%s %s", detailChecksIcon, checkSummary))
	}

	commentCount := pullRequestDetailCommentCount(detail)
	parts = append(parts, fmt.Sprintf("%s %s", detailCommentsIcon, formatCommentCount(commentCount)))

	return strings.Join(parts, "  ")
}

func renderPullRequestDetailContent(header string, content string) string {
	trimmedHeader := strings.TrimRight(header, "\n")
	trimmedContent := strings.TrimLeft(content, "\n")
	if trimmedHeader == "" {
		return trimmedContent
	}
	if trimmedContent == "" {
		return trimmedHeader
	}
	return strings.Join([]string{trimmedHeader, "", trimmedContent}, "\n")
}

func renderPullRequestDetailContentWithSeparator(header string, content string, width int) string {
	trimmedHeader := strings.TrimRight(header, "\n")
	trimmedContent := strings.TrimLeft(content, "\n")
	if trimmedHeader == "" {
		return trimmedContent
	}
	if trimmedContent == "" {
		return trimmedHeader
	}

	return strings.Join([]string{trimmedHeader, renderPullRequestDetailSectionSeparator(width), trimmedContent}, "\n")
}

func renderPullRequestDetailSectionSeparator(width int) string {
	if width < 1 {
		width = defaultDetailWrapWidth
	}

	return styleCommentBorder(strings.Repeat("─", width))
}

type pullRequestCommentsRenderedSection struct {
	text         string
	inlineThread *githubcli.PullRequestReviewThread
}

func buildPullRequestCommentsRenderedSections(comments []githubcli.PullRequestComment, inlineThreads []githubcli.PullRequestReviewThread, inlineComments []githubcli.PullRequestInlineComment, renderer MarkdownRenderer, width int) []pullRequestCommentsRenderedSection {
	sections := make([]pullRequestCommentsRenderedSection, 0, len(comments)+maxInt(len(inlineThreads), len(inlineComments)))
	commentBodyWidth := commentBoxInnerWidth(width)
	for _, comment := range comments {
		body := renderMarkdownWithFallback(comment.Body, renderer, commentBodyWidth, "No comment body.")
		sections = append(sections, pullRequestCommentsRenderedSection{text: renderPullRequestCommentSection(comment, body, width)})
	}
	if len(inlineThreads) > 0 {
		for _, inlineThread := range inlineThreads {
			thread := inlineThread
			sections = append(sections, pullRequestCommentsRenderedSection{text: renderPullRequestInlineCommentThreadSection(thread, renderer, width), inlineThread: &thread})
		}
		return sections
	}
	for _, inlineComment := range inlineComments {
		body := renderInlineCommentBody(inlineComment.Body, renderer, commentBodyWidth)
		sections = append(sections, pullRequestCommentsRenderedSection{text: renderPullRequestInlineCommentSection(inlineComment, body, width)})
	}
	return sections
}

func pullRequestDetailCommentCount(detail githubcli.PullRequestDetail) int {
	count := len(detail.Comments)
	if len(detail.InlineCommentThreads) > 0 {
		for _, inlineThread := range detail.InlineCommentThreads {
			count += len(inlineThread.Comments)
		}
		return count
	}
	return count + len(detail.InlineComments)
}

func renderMarkdownWithFallback(markdown string, renderer MarkdownRenderer, width int, emptyMessage string) string {
	trimmedMarkdown := strings.TrimSpace(markdown)
	if trimmedMarkdown == "" {
		return emptyMessage
	}
	if renderer == nil {
		renderer = glamourMarkdownRenderer{}
	}

	rendered, err := renderer.Render(trimmedMarkdown, width)
	if err != nil {
		return fmt.Sprintf("%s\n\n%s", markdownRenderFailurePrefix, trimmedMarkdown)
	}

	rendered = strings.TrimSpace(rendered)
	if rendered == "" {
		return emptyMessage
	}

	return rendered
}
