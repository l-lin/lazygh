package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const shortPullRequestCommitOIDLength = 7

type pullRequestHeaderOptions struct {
	includeStatusChecks bool
	includeReviewers    bool
}

func renderPullRequestDetailHeader(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	return renderPullRequestHeader(summary, detail, pullRequestHeaderOptions{includeStatusChecks: true, includeReviewers: true})
}

func renderPullRequestBrowserHeader(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	return renderPullRequestHeader(summary, detail, pullRequestHeaderOptions{})
}

func renderPullRequestHeader(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, options pullRequestHeaderOptions) string {
	headerLines := filterEmptyStrings([]string{
		renderPullRequestTitleAndReferenceLine(summary, detail),
		renderPullRequestLifecycleLine(summary, detail),
		renderPullRequestAssigneesLine(detail.Assignees),
	})
	metadataLines := filterEmptyStrings([]string{
		renderPullRequestMetaLineWithOptions(summary, detail, options.includeStatusChecks),
		renderPullRequestLabelsLine(detail.Labels),
		renderPullRequestReactionLine(detail.ReactionGroups),
	})
	if options.includeReviewers {
		metadataLines = append(metadataLines, filterEmptyStrings([]string{
			renderPullRequestReviewRequestsLine(detail.ReviewRequests),
			renderPullRequestApprovalsLine(detail.Reviews),
		})...)
	}
	if len(metadataLines) == 0 {
		return strings.Join(headerLines, "\n")
	}
	if len(headerLines) == 0 {
		return strings.Join(metadataLines, "\n")
	}
	return strings.Join(append(append(headerLines, ""), metadataLines...), "\n")
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

func renderPullRequestCommitsTab(commits []githubcli.PullRequestCommit, renderer MarkdownRenderer, width int) string {
	if len(commits) == 0 {
		return "No commits yet."
	}

	sections := make([]string, 0, len(commits))
	for _, commit := range commits {
		sections = append(sections, renderPullRequestCommitSection(commit, renderer, width))
	}
	return strings.Join(sections, "\n\n")
}

func renderPullRequestChangesTab(files []reviewDiffFile, renderer MarkdownRenderer, width int) string {
	return renderPullRequestChangesRows(buildPullRequestChangesRenderedRows(files, renderer, width))
}

func renderPullRequestChangesTabError(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub misplaced the diff again."
	}
	return fmt.Sprintf("Could not load pull request changes.\n\n%s", message)
}

func renderPullRequestCommitSection(commit githubcli.PullRequestCommit, renderer MarkdownRenderer, width int) string {
	sectionParts := []string{renderPullRequestCommitHeader(commit)}
	metadataLines := filterEmptyStrings([]string{
		renderPullRequestCommitAuthorsLine(commit.Authors),
		renderPullRequestCommitTimestampsLine(commit),
	})
	if len(metadataLines) > 0 {
		sectionParts = append(sectionParts, strings.Join(metadataLines, "\n"))
	}
	if body := strings.TrimSpace(renderMarkdownWithFallback(commit.MessageBody, renderer, commentBoxInnerWidth(width), "")); body != "" {
		sectionParts = append(sectionParts, body)
	}
	return renderRoundedCommentBox(strings.Join(sectionParts, "\n\n"), width)
}

func renderPullRequestCommitHeader(commit githubcli.PullRequestCommit) string {
	shortOID := shortPullRequestCommitOID(commit.OID)
	headline := strings.TrimSpace(commit.MessageHeadline)
	switch {
	case shortOID != "" && headline != "":
		return stylePullRequestReferenceText(shortOID) + " " + stylePullRequestTitleText(headline)
	case headline != "":
		return stylePullRequestTitleText(headline)
	case shortOID != "":
		return stylePullRequestReferenceText(shortOID)
	default:
		return stylePullRequestTitleText("Commit")
	}
}

func renderPullRequestCommitAuthorsLine(authors []githubcli.PullRequestCommitAuthor) string {
	labels := pullRequestCommitAuthorLabels(authors)
	if len(labels) == 0 {
		return ""
	}
	return "Authors: " + strings.Join(labels, ", ")
}

func renderPullRequestCommitTimestampsLine(commit githubcli.PullRequestCommit) string {
	parts := filterEmptyStrings([]string{
		renderPullRequestCommitTimestampPart("Authored", commit.AuthoredDate),
		renderPullRequestCommitTimestampPart("Committed", commit.CommittedDate),
	})
	return strings.Join(parts, "  ")
}

func renderPullRequestCommitTimestampPart(label string, value string) string {
	formatted := formattedOptionalTimestamp(value)
	if formatted == "" {
		return ""
	}
	return strings.TrimSpace(label) + " " + formatted
}

func pullRequestCommitAuthorLabels(authors []githubcli.PullRequestCommitAuthor) []string {
	labels := make([]string, 0, len(authors))
	seen := map[string]bool{}
	for _, author := range authors {
		label := pullRequestCommitAuthorLabel(author)
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		labels = append(labels, label)
	}
	return labels
}

func pullRequestCommitAuthorLabel(author githubcli.PullRequestCommitAuthor) string {
	if trimmedName := strings.TrimSpace(author.Name); trimmedName != "" {
		return trimmedName
	}
	if login := formatLogin(author.Login); login != "-" {
		return login
	}
	return strings.TrimSpace(author.Email)
}

func shortPullRequestCommitOID(oid string) string {
	trimmedOID := strings.TrimSpace(oid)
	if runeCountInt(trimmedOID) <= shortPullRequestCommitOIDLength {
		return trimmedOID
	}
	return string([]rune(trimmedOID)[:shortPullRequestCommitOIDLength])
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
	return renderPullRequestMetaLineWithOptions(summary, detail, true)
}

func renderPullRequestMetaLineWithOptions(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, includeStatusChecks bool) string {
	parts := []string{renderPullRequestStatusBadge(detailStatus(detail, summary))}

	baseRefName := strings.TrimSpace(detail.BaseRefName)
	headRefName := strings.TrimSpace(detail.HeadRefName)
	if baseRefName != "" || headRefName != "" {
		parts = append(parts, fmt.Sprintf("%s ← %s", valueOrDash(baseRefName), valueOrDash(headRefName)))
	}

	if includeStatusChecks {
		checkSummary := summarizeStatusChecks(detail.StatusCheckRollup)
		if checkSummary != "-" {
			parts = append(parts, fmt.Sprintf("%s %s", detailChecksIcon, checkSummary))
		}
	}

	parts = append(parts, renderPullRequestChurnParts(detail)...)
	return strings.Join(filterEmptyStrings(parts), "  ")
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

func renderPullRequestBrowserDetailContent(header string, overview string, content string, width int) string {
	trimmedHeader := strings.TrimRight(header, "\n")
	trimmedOverview := strings.TrimSpace(overview)
	trimmedContent := strings.TrimLeft(content, "\n")

	sections := filterEmptyStrings([]string{trimmedHeader, trimmedOverview})
	if len(sections) == 0 {
		return trimmedContent
	}
	if trimmedContent == "" {
		return strings.Join(sections, "\n\n")
	}

	headerAndOverview := strings.Join(sections, "\n\n")
	return strings.Join([]string{headerAndOverview, renderPullRequestDetailSectionSeparator(width), trimmedContent}, "\n")
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
