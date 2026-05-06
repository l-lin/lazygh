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
	return effectivePullRequestStatus(firstNonEmpty(detail.State, summary.State), detail.IsDraft || summary.IsDraft)
}

func effectivePullRequestStatus(state string, isDraft bool) string {
	normalizedState := strings.ToUpper(strings.TrimSpace(state))
	if normalizedState == "" {
		normalizedState = "-"
	}
	if isDraft {
		return "DRAFT"
	}
	return normalizedState
}

func renderPullRequestStatusBadge(status string) string {
	label := strings.TrimSpace(detailStatusIcon + " " + status)
	statusStyle, ok := pullRequestStatusStyleFor(status)
	if !ok {
		return label
	}

	return renderRoundedPill(label, statusStyle.foregroundHex, statusStyle.backgroundHex)
}

type pullRequestStatusStyle struct {
	foregroundHex string
	backgroundHex string
}

func pullRequestStatusStyleFor(status string) (pullRequestStatusStyle, bool) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "OPEN":
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusOpenForegroundHex, backgroundHex: theme.PullRequestStatusOpenBackgroundHex}, true
	case "DRAFT":
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusDraftForegroundHex, backgroundHex: theme.PullRequestStatusDraftBackgroundHex}, true
	case "CLOSED":
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusClosedForegroundHex, backgroundHex: theme.PullRequestStatusClosedBackgroundHex}, true
	case "MERGED":
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusMergedForegroundHex, backgroundHex: theme.PullRequestStatusMergedBackgroundHex}, true
	default:
		return pullRequestStatusStyle{}, false
	}
}

func renderPullRequestContextLine(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	parts := []string{fmt.Sprintf("%s %s#%d", detailRepositoryIcon, pullRequestRepositoryName(summary.Repository), firstNonZero(detail.Number, summary.Number))}
	authorLogin := pullRequestAuthorLogin(detail.Author)
	if authorLogin != "-" {
		parts = append(parts, fmt.Sprintf("%s %s", detailAuthorIcon, authorLogin))
	}
	if createdAt := formattedOptionalTimestamp(detail.CreatedAt); createdAt != "" {
		parts = append(parts, "Created: "+createdAt)
	}
	if updatedAt := formattedOptionalTimestamp(firstNonEmpty(detail.UpdatedAt, summary.UpdatedAt)); updatedAt != "" {
		parts = append(parts, "Updated: "+updatedAt)
	}
	return strings.Join(parts, "  ")
}

func renderPullRequestLabelsLine(labels []githubcli.PullRequestLabel) string {
	entries := make([]string, 0, len(labels))
	for _, label := range labels {
		trimmedName := strings.TrimSpace(label.Name)
		if trimmedName == "" {
			continue
		}
		entries = append(entries, detailLabelIcon+" "+trimmedName)
	}
	if len(entries) == 0 {
		return ""
	}
	return strings.Join(entries, "  ")
}

func renderPullRequestAssigneesLine(assignees []githubcli.PullRequestAuthor) string {
	entries := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		login := pullRequestAssigneeLogin(&assignee)
		if login == "-" {
			continue
		}
		entries = append(entries, detailAssigneesIcon+" "+login)
	}
	if len(entries) == 0 {
		return ""
	}
	return strings.Join(entries, "  ")
}

func renderPullRequestReviewRequestsLine(reviewRequests []githubcli.PullRequestReviewRequest) string {
	entries := make([]string, 0, len(reviewRequests))
	for _, reviewRequest := range reviewRequests {
		label := pullRequestReviewRequestLabel(reviewRequest.RequestedReviewer)
		if label == "-" {
			continue
		}
		entries = append(entries, detailReviewRequestsIcon+" "+label)
	}
	if len(entries) == 0 {
		return ""
	}
	return strings.Join(entries, "  ")
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

func pullRequestAssigneeLogin(author *githubcli.PullRequestAuthor) string {
	if author == nil {
		return "-"
	}
	return formatLogin(author.Login)
}

func pullRequestReviewRequestLabel(reviewer githubcli.PullRequestRequestedReviewer) string {
	if login := formatLogin(reviewer.Login); login != "-" {
		return login
	}

	organizationLogin := ""
	if reviewer.Organization != nil {
		organizationLogin = strings.TrimSpace(reviewer.Organization.Login)
	}
	slug := strings.TrimSpace(reviewer.Slug)
	if organizationLogin != "" && slug != "" {
		return "@" + organizationLogin + "/" + slug
	}
	if slug != "" {
		return "@" + slug
	}
	if trimmedName := strings.TrimSpace(reviewer.Name); trimmedName != "" {
		return trimmedName
	}
	return "-"
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

func renderPullRequestChurnParts(detail githubcli.PullRequestDetail) []string {
	if !pullRequestChurnAvailable(detail) {
		return nil
	}

	return []string{
		styleText(fmt.Sprintf("+%d", detail.Additions), foregroundColorEscape(theme.DiffAdditionForegroundHex)),
		styleText(fmt.Sprintf("-%d", detail.Deletions), foregroundColorEscape(theme.DiffDeletionForegroundHex)),
	}
}

func pullRequestChurnAvailable(detail githubcli.PullRequestDetail) bool {
	return detail.ChangedFiles > 0 || detail.Additions > 0 || detail.Deletions > 0
}

func formattedOptionalTimestamp(value string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return ""
	}
	return formatTimestamp(trimmedValue)
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
