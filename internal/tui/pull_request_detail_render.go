package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func renderPullRequestDetailHeader(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	headerLines := []string{
		fmt.Sprintf("%s %s#%d", detailRepositoryIcon, pullRequestRepositoryName(summary.Repository), firstNonZero(detail.Number, summary.Number)),
		pullRequestTitleText(firstNonEmpty(detail.Title, summary.Title)),
		renderPullRequestMetaLine(summary, detail),
	}

	return strings.Join(headerLines, "\n")
}

func renderPullRequestDescription(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, renderer MarkdownRenderer, width int) string {
	return renderMarkdownWithFallback(detailBody(detail, summary), renderer, width, "No description available.")
}

func renderPullRequestCommentsTab(comments []githubcli.PullRequestComment, inlineComments []githubcli.PullRequestInlineComment, renderer MarkdownRenderer, width int) string {
	if len(comments) == 0 && len(inlineComments) == 0 {
		return "No comments yet."
	}

	sections := make([]string, 0, len(comments)+len(inlineComments))
	commentBodyWidth := commentBoxInnerWidth(width)
	for _, comment := range comments {
		body := renderMarkdownWithFallback(comment.Body, renderer, commentBodyWidth, "No comment body.")
		sections = append(sections, renderPullRequestCommentSection(comment, body, width))
	}
	for _, inlineComment := range inlineComments {
		body := renderMarkdownWithFallback(inlineComment.Body, renderer, commentBodyWidth, "No comment body.")
		sections = append(sections, renderPullRequestInlineCommentSection(inlineComment, body, width))
	}
	return strings.Join(sections, "\n\n")
}

func renderPullRequestDetailLoading(summary githubcli.PullRequest, spinner string) string {
	return renderPullRequestDetailContent(
		renderPullRequestDetailHeader(summary, githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, State: summary.State, UpdatedAt: summary.UpdatedAt}),
		renderLoadingBody(spinner, fmt.Sprintf("Running `gh pr view %d -R %s --json ...`.", summary.Number, pullRequestRepositoryName(summary.Repository))),
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

	parts = append(parts, fmt.Sprintf("%s %s", detailStatusIcon, detailStatus(detail, summary)))

	checkSummary := summarizeStatusChecks(detail.StatusCheckRollup)
	if checkSummary != "-" {
		parts = append(parts, fmt.Sprintf("%s %s", detailChecksIcon, checkSummary))
	}

	commentCount := len(detail.Comments) + len(detail.InlineComments)
	parts = append(parts, fmt.Sprintf("%s %s", detailCommentsIcon, formatCommentCount(commentCount)))

	return strings.Join(parts, "  ·  ")
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

	return strings.Repeat("-", width)
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
