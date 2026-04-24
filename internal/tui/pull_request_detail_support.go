package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

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

func renderPullRequestStatusBadge(status string) string {
	foregroundHex, backgroundHex := pullRequestStatusBadgeColors(status)
	if foregroundHex == "" || backgroundHex == "" {
		return status
	}

	return styleText(status, foregroundColorEscape(foregroundHex), backgroundColorEscape(backgroundHex))
}

func pullRequestStatusBadgeColors(status string) (string, string) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "OPEN":
		return theme.PullRequestStatusOpenForegroundHex, theme.PullRequestStatusOpenBackgroundHex
	case "DRAFT":
		return theme.PullRequestStatusDraftForegroundHex, theme.PullRequestStatusDraftBackgroundHex
	case "CLOSED":
		return theme.PullRequestStatusClosedForegroundHex, theme.PullRequestStatusClosedBackgroundHex
	case "MERGED":
		return theme.PullRequestStatusMergedForegroundHex, theme.PullRequestStatusMergedBackgroundHex
	default:
		return "", ""
	}
}

func renderPullRequestApprovalsLine(reviews []githubcli.PullRequestReview) string {
	approverLogins := approvedPullRequestReviewerLogins(reviews)
	if len(approverLogins) == 0 {
		return ""
	}

	approvals := make([]string, 0, len(approverLogins))
	for _, login := range approverLogins {
		approvals = append(approvals, styleText(detailApprovalIcon, foregroundColorEscape(theme.DiffAdditionForegroundHex))+" "+formatLogin(login))
	}
	return strings.Join(approvals, "  ")
}

func approvedPullRequestReviewerLogins(reviews []githubcli.PullRequestReview) []string {
	if len(reviews) == 0 {
		return nil
	}

	latestReviewByLogin := map[string]githubcli.PullRequestReview{}
	latestReviewIndexes := map[string]int{}
	for index, review := range reviews {
		login := pullRequestReviewAuthorLogin(review.Author)
		if login == "" {
			continue
		}

		latestReview, ok := latestReviewByLogin[login]
		if !ok || pullRequestReviewIsLater(review, index, latestReview, latestReviewIndexes[login]) {
			latestReviewByLogin[login] = review
			latestReviewIndexes[login] = index
		}
	}

	approverLogins := make([]string, 0, len(latestReviewByLogin))
	for login, review := range latestReviewByLogin {
		if strings.EqualFold(strings.TrimSpace(review.State), "APPROVED") {
			approverLogins = append(approverLogins, login)
		}
	}
	sort.Strings(approverLogins)
	return approverLogins
}

func pullRequestReviewIsLater(candidate githubcli.PullRequestReview, candidateIndex int, current githubcli.PullRequestReview, currentIndex int) bool {
	candidateTime, candidateHasTime := parsePullRequestReviewSubmittedAt(candidate.SubmittedAt)
	currentTime, currentHasTime := parsePullRequestReviewSubmittedAt(current.SubmittedAt)
	if candidateHasTime && currentHasTime {
		if candidateTime.After(currentTime) {
			return true
		}
		if candidateTime.Before(currentTime) {
			return false
		}
	}
	return candidateIndex > currentIndex
}

func parsePullRequestReviewSubmittedAt(value string) (time.Time, bool) {
	submittedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return submittedAt, true
}

func pullRequestReviewAuthorLogin(author *githubcli.PullRequestCommentAuthor) string {
	if author == nil {
		return ""
	}
	return strings.TrimSpace(author.Login)
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
