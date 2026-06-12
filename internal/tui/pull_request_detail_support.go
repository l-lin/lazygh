package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

func detailBody(detail githubdomain.PullRequestDetail, summary githubdomain.PullRequest) string {
	return firstNonEmpty(detail.Body, summary.Body)
}

func detailBodyHTML(detail githubdomain.PullRequestDetail) string {
	return strings.TrimSpace(detail.BodyHTML)
}

func detailStatus(detail any, summary any) string {
	detailValue, ok := toDomainPullRequestDetail(detail)
	if !ok {
		detailValue = githubdomain.PullRequestDetail{}
	}
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		summaryValue = githubdomain.PullRequest{}
	}
	return effectivePullRequestStatus(firstNonEmpty(detailValue.State, summaryValue.State), detailValue.IsDraft || summaryValue.IsDraft)
}

func effectivePullRequestStatus(state string, isDraft bool) string {
	normalizedState := strings.ToUpper(strings.TrimSpace(state))
	if normalizedState == "" {
		if isDraft {
			return "DRAFT"
		}
		return "-"
	}
	if normalizedState == "OPEN" && isDraft {
		return "DRAFT"
	}
	return normalizedState
}

func renderPullRequestStatusBadge(status string) string {
	label := strings.TrimSpace(pullRequestStatusIcon(status) + " " + status)
	statusStyle, ok := pullRequestStatusStyleFor(status)
	if !ok {
		return label
	}

	return renderRoundedPill(label, statusStyle.foregroundHex, statusStyle.backgroundHex)
}

func pullRequestStatusIcon(status string) string {
	if strings.EqualFold(strings.TrimSpace(status), "DRAFT") {
		return draftPullRequestIcon
	}
	return pullRequestIcon
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

func renderPullRequestTitleAndReferenceLine(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) string {
	reference := pullRequestReference(summary, detail)
	title := pullRequestTitleText(firstNonEmpty(detail.Title, summary.Title))
	if reference == "" {
		return stylePullRequestTitleText(title)
	}
	return stylePullRequestReferenceText(reference) + " " + stylePullRequestTitleText(title)
}

func pullRequestReference(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) string {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	number := firstNonZero(detail.Number, summary.Number)
	if number <= 0 {
		return repository
	}
	return fmt.Sprintf("%s#%d", repository, number)
}

func renderPullRequestLifecycleLine(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) string {
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

func renderPullRequestLabelsLine(labels []githubdomain.PullRequestLabel) string {
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

func renderPullRequestOutOfDateLine(detail githubdomain.PullRequestDetail) string {
	if !pullRequestOutOfDateWithBase(detail) {
		return ""
	}

	message := "Out of date with base branch"
	if baseRefName := strings.TrimSpace(detail.BaseRefName); baseRefName != "" {
		message += " " + baseRefName
	}
	return styleText(iconWarning+" "+message, foregroundColorEscape(theme.WarningHex))
}

func pullRequestOutOfDateWithBase(detail githubdomain.PullRequestDetail) bool {
	return detail.OutOfDateWithBase || strings.EqualFold(strings.TrimSpace(detail.MergeStateStatus), "BEHIND")
}

func renderPullRequestAssigneesLine(assignees []githubdomain.PullRequestAuthor) string {
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

func renderPullRequestMergeStateLine(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail) string {
	switch {
	case effectivePullRequestInMergeQueue(summary, detail):
		return renderRoundedPill("Queued to merge", theme.PendingHex, theme.PendingBackgroundHex)
	case detail.AutoMergeRequest != nil || summary.AutoMergeRequest != nil:
		return renderRoundedPill("Auto-merge enabled", theme.PendingHex, theme.PendingBackgroundHex)
	default:
		return ""
	}
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

func renderPullRequestReviewRequestsLine(reviewRequests []githubdomain.PullRequestReviewRequest) string {
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

func renderPullRequestApprovalsLine(reviews []githubdomain.PullRequestReview) string {
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

func approvedPullRequestReviewerLogins(reviews any) []string {
	reviewValues := toDomainPullRequestReviews(reviews)
	if len(reviewValues) == 0 {
		return nil
	}

	latestReviewByLogin := latestPullRequestReviews(reviewValues)
	approverLogins := make([]string, 0, len(latestReviewByLogin))
	for login, review := range latestReviewByLogin {
		if strings.EqualFold(strings.TrimSpace(review.State), "APPROVED") {
			approverLogins = append(approverLogins, login)
		}
	}
	sort.Strings(approverLogins)
	return approverLogins
}

func latestPullRequestReviews(reviews any) map[string]githubdomain.PullRequestReview {
	latestByLogin := map[string]githubdomain.PullRequestReview{}
	latestIndexes := map[string]int{}
	for index, review := range toDomainPullRequestReviews(reviews) {
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

func pullRequestReviewIsLater(candidate githubdomain.PullRequestReview, candidateIndex int, current githubdomain.PullRequestReview, currentIndex int) bool {
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

func pullRequestReviewAuthorLogin(author *githubdomain.PullRequestCommentAuthor) string {
	if author == nil {
		return ""
	}
	return strings.TrimSpace(author.Login)
}

func pullRequestTitleText(title string) string {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return "Untitled pull request"
	}
	return trimmedTitle
}

func pullRequestAuthorLogin(author *githubdomain.PullRequestAuthor) string {
	if author == nil {
		return "-"
	}
	return formatLogin(author.Login)
}

func pullRequestAssigneeLogin(author *githubdomain.PullRequestAuthor) string {
	if author == nil {
		return "-"
	}
	return formatLogin(author.Login)
}

func pullRequestReviewRequestLabel(reviewer githubdomain.PullRequestRequestedReviewer) string {
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

func pullRequestCommentAuthorLogin(author *githubdomain.PullRequestCommentAuthor) string {
	if author == nil {
		return "-"
	}
	return formatLogin(author.Login)
}

func summarizeStatusChecks(checks []githubdomain.PullRequestStatusCheck) string {
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

func renderPullRequestChurnParts(detail githubdomain.PullRequestDetail) []string {
	if !pullRequestChurnAvailable(detail) {
		return nil
	}

	return []string{
		styleText(fmt.Sprintf("+%d", detail.Additions), foregroundColorEscape(theme.DiffAdditionHex)),
		styleText(fmt.Sprintf("-%d", detail.Deletions), foregroundColorEscape(theme.DiffDeletionHex)),
	}
}

func pullRequestChurnAvailable(detail githubdomain.PullRequestDetail) bool {
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
