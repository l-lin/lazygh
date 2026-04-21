package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	defaultDetailWrapWidth      = 80
	minimumMarkdownRenderWidth  = 20
	markdownRenderFailurePrefix = "Markdown rendering failed. Showing source."
	maximumBranchLabelWidth     = 28

	detailDescriptionIcon           = "󰈙"
	detailCommentsIcon              = " "
	detailRepositoryIcon            = ""
	detailBranchIcon                = ""
	detailStatusIcon                = ""
	detailChecksIcon                = "󰄬"
	detailInlineCommentLocationIcon = "󰈔"
)

type DetailTab int

const (
	DescriptionDetailTab DetailTab = iota
	CommentsDetailTab
)

type MarkdownRenderer interface {
	Render(markdown string, width int) (string, error)
}

type glamourMarkdownRenderer struct{}

type pullRequestDetailResult struct {
	detail githubcli.PullRequestDetail
	err    error
}

func (tab DetailTab) Label() string {
	switch tab {
	case CommentsDetailTab:
		return fmt.Sprintf("%s Comments", detailCommentsIcon)
	default:
		return fmt.Sprintf("%s Description", detailDescriptionIcon)
	}
}

func (glamourMarkdownRenderer) Render(markdown string, width int) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(prettyMarkdownStyle()),
		glamour.WithWordWrap(effectiveMarkdownWidth(width)),
	)
	if err != nil {
		return "", err
	}

	rendered, err := renderer.Render(strings.TrimSpace(markdown))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(rendered), nil
}

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

func renderPullRequestDetailLoading(summary githubcli.PullRequest) string {
	return renderPullRequestDetailContent(
		renderPullRequestDetailHeader(summary, githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, State: summary.State, UpdatedAt: summary.UpdatedAt}),
		fmt.Sprintf("Loading pull request detail...\nRunning `gh pr view %d -R %s --json ...`.", summary.Number, pullRequestRepositoryName(summary.Repository)),
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

func detailBody(detail githubcli.PullRequestDetail, summary githubcli.PullRequest) string {
	return firstNonEmpty(detail.Body, summary.Body)
}

func detailStatus(detail githubcli.PullRequestDetail, summary githubcli.PullRequest) string {
	state := strings.ToUpper(strings.TrimSpace(firstNonEmpty(detail.State, summary.State)))
	if state == "" {
		state = "-"
	}
	if detail.IsDraft || summary.IsDraft {
		return "DRAFT"
	}
	return state
}

func pullRequestTitleLine(title string, number int) string {
	trimmedTitle := pullRequestTitleText(title)
	if number <= 0 {
		return trimmedTitle
	}
	return fmt.Sprintf("%s #%d", trimmedTitle, number)
}

func pullRequestTitleText(title string) string {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return "Untitled pull request"
	}
	return trimmedTitle
}

func pullRequestAuthorLogin(author *githubcli.PullRequestAuthor) string {
	if author == nil {
		return "-"
	}
	return formatLogin(author.Login)
}

func pullRequestCommentAuthorLogin(author *githubcli.PullRequestCommentAuthor) string {
	if author == nil {
		return "-"
	}
	return formatLogin(author.Login)
}

func summarizeStatusChecks(checks []githubcli.PullRequestStatusCheck) string {
	if len(checks) == 0 {
		return "-"
	}

	passing := 0
	failing := 0
	pending := 0
	for _, check := range checks {
		status := strings.ToUpper(strings.TrimSpace(check.Status))
		conclusion := strings.ToUpper(strings.TrimSpace(check.Conclusion))
		switch {
		case status != "COMPLETED":
			pending++
		case conclusion == "SUCCESS" || conclusion == "NEUTRAL" || conclusion == "SKIPPED":
			passing++
		case conclusion == "FAILURE" || conclusion == "TIMED_OUT" || conclusion == "CANCELLED" || conclusion == "STARTUP_FAILURE" || conclusion == "ACTION_REQUIRED":
			failing++
		default:
			pending++
		}
	}

	parts := make([]string, 0, 3)
	if passing > 0 {
		parts = append(parts, fmt.Sprintf("%d passing", passing))
	}
	if failing > 0 {
		parts = append(parts, fmt.Sprintf("%d failing", failing))
	}
	if pending > 0 {
		parts = append(parts, fmt.Sprintf("%d pending", pending))
	}
	if len(parts) == 0 {
		return "-"
	}

	return strings.Join(parts, ", ")
}

func mergeableText(mergeable string) string {
	switch strings.ToUpper(strings.TrimSpace(mergeable)) {
	case "MERGEABLE":
		return "yes"
	case "":
		return "-"
	default:
		return "no"
	}
}

func formatCommentCount(count int) string {
	return fmt.Sprintf("%d %s", count, pluralize(count, "comment", "comments"))
}

func formatTimestamp(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "-"
	}

	parsedTime, err := time.Parse(time.RFC3339, trimmedValue)
	if err != nil {
		return trimmedValue
	}

	return parsedTime.UTC().Format("2006-01-02 15:04 UTC")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue != "" {
			return trimmedValue
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func effectiveMarkdownWidth(width int) int {
	if width < minimumMarkdownRenderWidth {
		return defaultDetailWrapWidth
	}
	return width
}

func compactBranchLabel(label string) string {
	runes := []rune(strings.TrimSpace(label))
	if len(runes) <= maximumBranchLabelWidth {
		return string(runes)
	}

	prefixWidth := maximumBranchLabelWidth/2 - 1
	suffixWidth := maximumBranchLabelWidth - prefixWidth - 1
	return string(runes[:prefixWidth]) + "…" + string(runes[len(runes)-suffixWidth:])
}
