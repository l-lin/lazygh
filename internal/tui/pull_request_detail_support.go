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
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusOpenHex, backgroundHex: theme.PullRequestStatusOpenBackgroundHex}, true
	case "DRAFT":
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusDraftHex, backgroundHex: theme.PullRequestStatusDraftBackgroundHex}, true
	case "CLOSED":
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusClosedHex, backgroundHex: theme.PullRequestStatusClosedBackgroundHex}, true
	case "MERGED":
		return pullRequestStatusStyle{foregroundHex: theme.PullRequestStatusMergedHex, backgroundHex: theme.PullRequestStatusMergedBackgroundHex}, true
	default:
		return pullRequestStatusStyle{}, false
	}
}

func renderPullRequestTitleAndReferenceLine(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	reference := pullRequestReference(summary, detail)
	title := pullRequestTitleText(firstNonEmpty(detail.Title, summary.Title))
	if reference == "" {
		return stylePullRequestTitleText(title)
	}
	return stylePullRequestReferenceText(reference) + " " + stylePullRequestTitleText(title)
}

func pullRequestReference(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	number := firstNonZero(detail.Number, summary.Number)
	if number <= 0 {
		return repository
	}
	return fmt.Sprintf("%s#%d", repository, number)
}

func renderPullRequestLifecycleLine(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) string {
	authorBadge := renderPullRequestActorBadge(pullRequestAuthorLogin(detail.Author))
	createdAt := formattedOptionalTimestamp(detail.CreatedAt)
	updatedAt := formattedOptionalTimestamp(firstNonEmpty(detail.UpdatedAt, summary.UpdatedAt))

	var builder strings.Builder
	if authorBadge != "" || createdAt != "" {
		builder.WriteString(stylePullRequestTitleText("Created"))
		if authorBadge != "" {
			builder.WriteString(stylePullRequestTitleText(" by "))
			builder.WriteString(authorBadge)
		}
		if createdAt != "" {
			builder.WriteString(stylePullRequestMutedText(" the " + createdAt))
		}
		if updatedAt != "" {
			builder.WriteString(stylePullRequestMutedText(" (last updated at " + updatedAt + ")"))
		}
		return builder.String()
	}
	if updatedAt != "" {
		return stylePullRequestMutedText("Last updated at " + updatedAt)
	}
	return ""
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
	return stylePullRequestMutedText(strings.Join(entries, "  "))
}

func renderPullRequestAssigneesLine(assignees []githubcli.PullRequestAuthor) string {
	entries := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		badge := renderPullRequestActorBadge(pullRequestAssigneeLogin(&assignee))
		if badge == "" {
			continue
		}
		entries = append(entries, badge)
	}
	if len(entries) == 0 {
		return ""
	}
	return stylePullRequestTitleText("Assigned to ") + strings.Join(entries, " ")
}

func renderPullRequestActorBadge(login string) string {
	trimmedLogin := strings.TrimSpace(login)
	if trimmedLogin == "" || trimmedLogin == "-" {
		return ""
	}
	return renderRoundedPill(trimmedLogin, theme.CommentAuthorBadgeHex, theme.CommentAuthorBadgeBackgroundHex)
}

func stylePullRequestReferenceText(text string) string {
	return styleText(text, foregroundColorEscape(theme.PullRequestReferenceHex))
}

func stylePullRequestTitleText(text string) string {
	return styleText(text, foregroundColorEscape(theme.PullRequestTitleHex))
}

func stylePullRequestMutedText(text string) string {
	return styleText(text, foregroundColorEscape(theme.PullRequestReferenceHex))
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
		approvals = append(approvals, styleText(detailApprovalIcon, foregroundColorEscape(theme.DiffAdditionHex))+" "+formatLogin(login))
	}
	return strings.Join(approvals, "  ")
}

func approvedPullRequestReviewerLogins(reviews []githubcli.PullRequestReview) []string {
	if len(reviews) == 0 {
		return nil
	}

	latestReviewByLogin := latestPullRequestReviews(reviews)
	approverLogins := make([]string, 0, len(latestReviewByLogin))
	for login, review := range latestReviewByLogin {
		if strings.EqualFold(strings.TrimSpace(review.State), "APPROVED") {
			approverLogins = append(approverLogins, login)
		}
	}
	sort.Strings(approverLogins)
	return approverLogins
}

func latestPullRequestReviews(reviews []githubcli.PullRequestReview) map[string]githubcli.PullRequestReview {
	latestByLogin := map[string]githubcli.PullRequestReview{}
	latestIndexes := map[string]int{}
	for index, review := range reviews {
		login := pullRequestReviewAuthorLogin(review.Author)
		if login == "" {
			continue
		}

		latestReview, ok := latestByLogin[login]
		if !ok || pullRequestReviewIsLater(review, index, latestReview, latestIndexes[login]) {
			latestByLogin[login] = review
			latestIndexes[login] = index
		}
	}
	return latestByLogin
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
		classification := classifyPullRequestStatusCheck(check)
		switch classification.SummaryKind {
		case pullRequestStatusCheckSummaryKindPassing:
			passing++
		case pullRequestStatusCheckSummaryKindFailing:
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
		styleText(fmt.Sprintf("+%d", detail.Additions), foregroundColorEscape(theme.DiffAdditionHex)),
		styleText(fmt.Sprintf("-%d", detail.Deletions), foregroundColorEscape(theme.DiffDeletionHex)),
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
