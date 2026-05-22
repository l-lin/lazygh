package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

const shortPullRequestCommitOIDLength = 7

type pullRequestHeaderOptions struct {
	includeStatusChecks bool
	includeReviewers    bool
}

func renderPullRequestDetailHeader(summary any, detail any) string {
	return renderPullRequestHeader(summary, detail, pullRequestHeaderOptions{includeStatusChecks: true, includeReviewers: true})
}

func renderPullRequestBrowserHeader(summary any, detail any) string {
	return renderPullRequestHeader(summary, detail, pullRequestHeaderOptions{})
}

func renderPullRequestHeader(summary any, detail any, options pullRequestHeaderOptions) string {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return ""
	}
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		return ""
	}
	headerLines := filterEmptyStrings([]string{
		renderPullRequestTitleAndReferenceLine(summaryValue, detailValue),
		renderPullRequestLifecycleLine(summaryValue, detailValue),
		renderPullRequestAssigneesLine(detailValue.Assignees),
	})
	metadataLines := filterEmptyStrings([]string{
		renderPullRequestMetaLineWithOptions(summaryValue, detailValue, options.includeStatusChecks),
		renderPullRequestAutoMergeLine(detailValue),
		renderPullRequestLabelsLine(detailValue.Labels),
		renderPullRequestReactionLine(detailValue.ReactionGroups),
		renderPullRequestOutOfDateLine(detailValue),
	})
	if options.includeReviewers {
		metadataLines = append(metadataLines, filterEmptyStrings([]string{
			renderPullRequestReviewRequestsLine(detailValue.ReviewRequests),
			renderPullRequestApprovalsLine(detailValue.Reviews),
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

func renderPullRequestDescription(summary any, detail any, renderer MarkdownRenderer, width int) string {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		summaryValue = githubdomain.PullRequest{}
	}
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		detailValue = githubdomain.PullRequestDetail{}
	}
	return renderMarkdownWithFallback(prepareMarkdownForImageRendering(detailBody(detailValue, summaryValue), detailBodyHTML(detailValue)), renderer, width, "No description available.")
}

func renderPullRequestCommentsTab(comments any, inlineThreads any, inlineComments any, renderer MarkdownRenderer, width int) string {
	sections := buildPullRequestCommentsRenderedSections(toDomainPullRequestComments(comments), toDomainPullRequestReviewThreads(inlineThreads), toDomainPullRequestInlineComments(inlineComments), renderer, width)
	if len(sections) == 0 {
		return "No comments yet."
	}

	texts := make([]string, 0, len(sections))
	for _, section := range sections {
		texts = append(texts, section.text)
	}
	return strings.Join(texts, "\n\n")
}

func renderPullRequestCommitsTab(commits any, renderer MarkdownRenderer, width int) string {
	commitValues := sortedPullRequestCommitsDescending(toDomainPullRequestCommits(commits))
	if len(commitValues) == 0 {
		return "No commits yet."
	}

	sections := make([]string, 0, len(commitValues))
	for _, commit := range commitValues {
		sections = append(sections, renderPullRequestCommitSection(commit, renderer, width))
	}
	return strings.Join(sections, "\n"+renderPullRequestCommitTimelineLine(renderPullRequestCommitTimelineRailPrefix(), "")+"\n")
}

func renderPullRequestChangesTabError(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub misplaced the diff again."
	}
	return fmt.Sprintf("Could not load pull request changes.\n\n%s", message)
}

func renderPullRequestCommitSection(commit githubdomain.PullRequestCommit, renderer MarkdownRenderer, width int) string {
	sectionLines := []string{renderPullRequestCommitTimelineLine(renderPullRequestCommitTimelineDotPrefix(), renderPullRequestCommitHeader(commit))}
	for _, metadataLine := range filterEmptyStrings([]string{
		renderPullRequestCommitAuthorsLine(commit.Authors),
		renderPullRequestCommitTimestampsLine(commit),
	}) {
		sectionLines = append(sectionLines, renderPullRequestCommitTimelineLine(renderPullRequestCommitTimelineRailPrefix(), metadataLine))
	}
	if body := strings.TrimSpace(renderMarkdownWithFallback(prepareMarkdownForImageRendering(commit.MessageBody, commit.MessageBodyHTML), renderer, pullRequestCommitTimelineBodyWidth(width), "")); body != "" {
		sectionLines = append(sectionLines, renderPullRequestCommitTimelineLine(renderPullRequestCommitTimelineRailPrefix(), ""))
		sectionLines = append(sectionLines, renderPullRequestCommitTimelineText(renderPullRequestCommitTimelineRailPrefix(), body))
	}
	return strings.Join(sectionLines, "\n")
}

func renderPullRequestCommitHeader(commit githubdomain.PullRequestCommit) string {
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

func renderPullRequestCommitAuthorsLine(authors []githubdomain.PullRequestCommitAuthor) string {
	labels := pullRequestCommitAuthorLabels(authors)
	if len(labels) == 0 {
		return ""
	}
	return "Authors: " + strings.Join(labels, ", ")
}

func renderPullRequestCommitTimestampsLine(commit githubdomain.PullRequestCommit) string {
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

func pullRequestCommitAuthorLabels(authors []githubdomain.PullRequestCommitAuthor) []string {
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

func pullRequestCommitAuthorLabel(author githubdomain.PullRequestCommitAuthor) string {
	if trimmedName := strings.TrimSpace(author.Name); trimmedName != "" {
		return trimmedName
	}
	if login := formatLogin(author.Login); login != "-" {
		return login
	}
	return strings.TrimSpace(author.Email)
}

func sortedPullRequestCommitsDescending(commits []githubdomain.PullRequestCommit) []githubdomain.PullRequestCommit {
	if len(commits) == 0 {
		return nil
	}

	type orderedPullRequestCommit struct {
		commit      githubdomain.PullRequestCommit
		index       int
		sortTime    time.Time
		hasSortTime bool
	}

	orderedCommits := make([]orderedPullRequestCommit, 0, len(commits))
	for index, commit := range commits {
		sortTime, hasSortTime := pullRequestCommitSortTime(commit)
		orderedCommits = append(orderedCommits, orderedPullRequestCommit{commit: commit, index: index, sortTime: sortTime, hasSortTime: hasSortTime})
	}

	sort.SliceStable(orderedCommits, func(i int, j int) bool {
		left := orderedCommits[i]
		right := orderedCommits[j]
		switch {
		case left.hasSortTime && right.hasSortTime:
			if left.sortTime.After(right.sortTime) {
				return true
			}
			if left.sortTime.Before(right.sortTime) {
				return false
			}
		case left.hasSortTime != right.hasSortTime:
			return left.hasSortTime
		}
		return left.index > right.index
	})

	sortedCommits := make([]githubdomain.PullRequestCommit, 0, len(orderedCommits))
	for _, orderedCommit := range orderedCommits {
		sortedCommits = append(sortedCommits, orderedCommit.commit)
	}
	return sortedCommits
}

func pullRequestCommitSortTime(commit githubdomain.PullRequestCommit) (time.Time, bool) {
	if committedAt, ok := parsePullRequestCommitTimestamp(commit.CommittedDate); ok {
		return committedAt, true
	}
	return parsePullRequestCommitTimestamp(commit.AuthoredDate)
}

func parsePullRequestCommitTimestamp(value string) (time.Time, bool) {
	parsedTime, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return parsedTime, true
}

func renderPullRequestCommitTimelineDotPrefix() string {
	return styleText(detailCommitTimelineDot, foregroundColorEscape(theme.PullRequestReferenceHex))
}

func renderPullRequestCommitTimelineRailPrefix() string {
	return styleCommentBorder("│")
}

func renderPullRequestCommitTimelineLine(prefix string, text string) string {
	if strings.TrimSpace(text) == "" {
		return prefix
	}
	return prefix + " " + strings.TrimRight(text, "\n")
}

func renderPullRequestCommitTimelineText(prefix string, text string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	renderedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		renderedLines = append(renderedLines, renderPullRequestCommitTimelineLine(prefix, line))
	}
	return strings.Join(renderedLines, "\n")
}

func pullRequestCommitTimelineBodyWidth(width int) int {
	bodyWidth := effectiveMarkdownWidth(width) - 2
	if bodyWidth < 1 {
		return 1
	}
	return bodyWidth
}

func shortPullRequestCommitOID(oid string) string {
	trimmedOID := strings.TrimSpace(oid)
	if runeCountInt(trimmedOID) <= shortPullRequestCommitOIDLength {
		return trimmedOID
	}
	return string([]rune(trimmedOID)[:shortPullRequestCommitOIDLength])
}

func renderPullRequestDetailLoading(summary githubdomain.PullRequest, spinner string) string {
	return renderPullRequestDetailContent(
		renderPullRequestDetailHeader(summary, githubdomain.PullRequestDetail{Title: summary.Title, Number: summary.Number, State: summary.State, UpdatedAt: summary.UpdatedAt}),
		strings.TrimSpace(spinner),
	)
}

func renderPullRequestDetailError(summary githubdomain.PullRequest, err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub found a new way to be unhelpful."
	}

	fallback := strings.TrimSpace(pullRequestRow(summary).Item.Detail)
	if fallback == "" {
		fallback = "No fallback detail available."
	}

	return renderPullRequestDetailContent(
		renderPullRequestDetailHeader(summary, githubdomain.PullRequestDetail{Title: summary.Title, Number: summary.Number, State: summary.State, UpdatedAt: summary.UpdatedAt}),
		fmt.Sprintf("Could not load rich pull request detail.\n\n%s\n\n%s", message, fallback),
	)
}

func renderPullRequestMetaLineWithOptions(summary any, detail any, includeStatusChecks bool) string {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		summaryValue = githubdomain.PullRequest{}
	}
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		detailValue = githubdomain.PullRequestDetail{}
	}
	parts := []string{renderPullRequestStatusBadge(detailStatus(detailValue, summaryValue))}

	baseRefName := strings.TrimSpace(detailValue.BaseRefName)
	headRefName := strings.TrimSpace(detailValue.HeadRefName)
	if baseRefName != "" || headRefName != "" {
		parts = append(parts, fmt.Sprintf("%s ← %s", valueOrDash(baseRefName), valueOrDash(headRefName)))
	}

	if includeStatusChecks {
		checkSummary := summarizeStatusChecks(detailValue.StatusCheckRollup)
		if checkSummary != "-" {
			parts = append(parts, fmt.Sprintf("%s %s", detailChecksIcon, checkSummary))
		}
	}

	parts = append(parts, renderPullRequestChurnParts(detailValue)...)
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
	inlineThread *githubdomain.PullRequestReviewThread
}

func buildPullRequestCommentsRenderedSections(comments []githubdomain.PullRequestComment, inlineThreads []githubdomain.PullRequestReviewThread, inlineComments []githubdomain.PullRequestInlineComment, renderer MarkdownRenderer, width int) []pullRequestCommentsRenderedSection {
	sections := make([]pullRequestCommentsRenderedSection, 0, len(comments)+maxInt(len(inlineThreads), len(inlineComments)))
	commentBodyWidth := commentBoxInnerWidth(width)
	for _, comment := range comments {
		body := renderMarkdownWithFallback(prepareMarkdownForImageRendering(comment.Body, comment.BodyHTML), renderer, commentBodyWidth, "No comment body.")
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
		body := renderInlineCommentBodyForInlineComment(inlineComment, renderer, commentBodyWidth)
		sections = append(sections, pullRequestCommentsRenderedSection{text: renderPullRequestInlineCommentSection(inlineComment, body, width)})
	}
	return sections
}

func pullRequestDetailCommentCount(detail any) int {
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		return 0
	}
	count := len(detailValue.Comments)
	if len(detailValue.InlineCommentThreads) > 0 {
		for _, inlineThread := range detailValue.InlineCommentThreads {
			count += len(inlineThread.Comments)
		}
		return count
	}
	return count + len(detailValue.InlineComments)
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
