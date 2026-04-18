package tui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	defaultDetailWrapWidth      = 80
	minimumMarkdownRenderWidth  = 20
	markdownRenderFailurePrefix = "Markdown rendering failed. Showing source."
)

type MarkdownRenderer interface {
	Render(markdown string, width int) (string, error)
}

type glamourMarkdownRenderer struct{}

type pullRequestDetailResult struct {
	detail githubcli.PullRequestDetail
	err    error
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func (glamourMarkdownRenderer) Render(markdown string, width int) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(effectiveMarkdownWidth(width)),
	)
	if err != nil {
		return "", err
	}

	rendered, err := renderer.Render(strings.TrimSpace(markdown))
	if err != nil {
		return "", err
	}

	return normalizeRenderedMarkdown(rendered), nil
}

func renderPullRequestDetail(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, renderer MarkdownRenderer, width int) string {
	if renderer == nil {
		renderer = glamourMarkdownRenderer{}
	}

	body := renderMarkdownWithFallback(detailBody(detail, summary), renderer, width, "No description available.")
	sections := []string{
		renderPullRequestMetadata(summary, detail),
		body,
		renderPullRequestComments(detail.Comments, renderer, width),
	}

	return strings.Join(sections, "\n\n")
}

func renderPullRequestDetailLoading(summary githubcli.PullRequest) string {
	repository := pullRequestRepositoryName(summary.Repository)
	if strings.TrimSpace(repository) == "" {
		repository = "-"
	}

	return fmt.Sprintf("%s\n\nLoading pull request detail...\nRunning `gh pr view %d -R %s --json ...`.", pullRequestTitleLine(summary.Title, summary.Number), summary.Number, repository)
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

	return fmt.Sprintf("%s\n\nCould not load rich pull request detail.\n\n%s\n\n%s", pullRequestTitleLine(summary.Title, summary.Number), message, fallback)
}

func renderPullRequestMetadata(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	lines := []string{
		pullRequestTitleLine(firstNonEmpty(detail.Title, summary.Title), firstNonZero(detail.Number, summary.Number)),
		fmt.Sprintf("Status: %s %s ← %s", detailStatus(detail, summary), valueOrDash(firstNonEmpty(detail.BaseRefName, "-")), valueOrDash(firstNonEmpty(detail.HeadRefName, "-"))),
		fmt.Sprintf("Repo: %s", pullRequestRepositoryName(summary.Repository)),
		fmt.Sprintf("Author: %s", pullRequestAuthorLogin(detail.Author)),
		fmt.Sprintf("Created: %s", formatTimestamp(detail.CreatedAt)),
		fmt.Sprintf("Updated: %s", formatTimestamp(firstNonEmpty(detail.UpdatedAt, summary.UpdatedAt))),
		fmt.Sprintf("Labels: %s", formatPullRequestLabels(detail.Labels)),
		fmt.Sprintf("Merge Status: %s", valueOrDash(detail.MergeStateStatus)),
		fmt.Sprintf("Checks: %s", summarizeStatusChecks(detail.StatusCheckRollup)),
		fmt.Sprintf("Mergeable: %s", mergeableText(detail.Mergeable)),
		fmt.Sprintf("Changes: %s", formatChanges(detail.ChangedFiles, detail.Additions, detail.Deletions)),
	}

	return strings.Join(lines, "\n")
}

func renderPullRequestComments(comments []githubcli.PullRequestComment, renderer MarkdownRenderer, width int) string {
	if len(comments) == 0 {
		return "Comments\nNo comments yet."
	}

	sections := []string{"Comments"}
	for _, comment := range comments {
		sections = append(sections, fmt.Sprintf("%s · %s\n%s", pullRequestCommentAuthorLogin(comment.Author), formatTimestamp(comment.CreatedAt), renderMarkdownWithFallback(comment.Body, renderer, width, "No comment body.")))
	}
	return strings.Join(sections, "\n\n")
}

func renderMarkdownWithFallback(markdown string, renderer MarkdownRenderer, width int, emptyMessage string) string {
	trimmedMarkdown := strings.TrimSpace(markdown)
	if trimmedMarkdown == "" {
		return emptyMessage
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
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		trimmedTitle = "Untitled pull request"
	}
	if number <= 0 {
		return trimmedTitle
	}
	return fmt.Sprintf("%s #%d", trimmedTitle, number)
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

func formatPullRequestLabels(labels []githubcli.PullRequestLabel) string {
	if len(labels) == 0 {
		return "-"
	}

	formattedLabels := make([]string, 0, len(labels))
	for _, label := range labels {
		formattedLabels = append(formattedLabels, valueOrDash(label.Name))
	}
	return strings.Join(formattedLabels, ", ")
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

func formatChanges(changedFiles int, additions int, deletions int) string {
	fileLabel := pluralize(changedFiles, "file", "files")
	return fmt.Sprintf("%d %s +%d -%d", changedFiles, fileLabel, additions, deletions)
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

func normalizeRenderedMarkdown(rendered string) string {
	withoutANSI := ansiPattern.ReplaceAllString(rendered, "")
	lines := strings.Split(withoutANSI, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " ")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
