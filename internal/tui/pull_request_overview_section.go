package tui

import (
	"fmt"
	"sort"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	pullRequestOverviewSuccessIcon = ""
	pullRequestOverviewFailureIcon = ""
	pullRequestOverviewPendingIcon = ""
	pullRequestOverviewMutedIcon   = "•"
)

type pullRequestOverviewStatus int

const (
	pullRequestOverviewStatusMuted pullRequestOverviewStatus = iota
	pullRequestOverviewStatusPending
	pullRequestOverviewStatusSuccess
	pullRequestOverviewStatusFailure
)

type pullRequestOverviewSection struct {
	Reviewers   pullRequestOverviewBlock
	MergeChecks pullRequestOverviewBlock
	Builds      pullRequestOverviewBlock
}

type pullRequestOverviewBlock struct {
	Title   string
	Summary string
	Entries []pullRequestOverviewEntry
}

type pullRequestOverviewEntry struct {
	Label    string
	Detail   string
	Status   pullRequestOverviewStatus
	ShowIcon bool
}

type pullRequestReviewerOverview struct {
	Entries               []pullRequestOverviewEntry
	ApprovedCount         int
	PendingCount          int
	ChangesRequestedCount int
}

func buildPullRequestOverviewSection(detail githubcli.PullRequestDetail) pullRequestOverviewSection {
	reviewers := buildPullRequestReviewerOverview(detail)
	return pullRequestOverviewSection{
		Reviewers:   buildPullRequestReviewersBlock(reviewers),
		MergeChecks: buildPullRequestMergeChecksBlock(detail, reviewers),
		Builds:      buildPullRequestBuildsBlock(detail),
	}
}

func renderPullRequestOverviewSection(section pullRequestOverviewSection, width int) string {
	blocks := []pullRequestOverviewBlock{section.Reviewers, section.MergeChecks, section.Builds}
	renderedBlocks := make([]string, 0, len(blocks))
	for _, block := range blocks {
		renderedBlock := renderPullRequestOverviewBlock(block, width)
		if strings.TrimSpace(renderedBlock) == "" {
			continue
		}
		renderedBlocks = append(renderedBlocks, renderedBlock)
	}
	return strings.Join(renderedBlocks, "\n\n")
}

func renderPullRequestOverviewBlock(block pullRequestOverviewBlock, width int) string {
	entries := renderPullRequestOverviewEntries(block.Entries)
	if strings.TrimSpace(entries) == "" {
		return ""
	}

	heading := strings.TrimSpace(block.Title)
	if summary := strings.TrimSpace(block.Summary); summary != "" {
		heading += " (" + summary + ")"
	}
	if heading == "" {
		return renderRoundedCommentBox(entries, width)
	}

	return styleText(heading, foregroundColorEscape(theme.InactiveTitleHex)) + "\n" + renderRoundedCommentBox(entries, width)
}

func renderPullRequestOverviewEntries(entries []pullRequestOverviewEntry) string {
	if len(entries) == 0 {
		return ""
	}

	lines := make([]string, 0, len(entries)*2)
	for _, entry := range entries {
		label := strings.TrimSpace(entry.Label)
		if label != "" {
			lines = append(lines, renderPullRequestOverviewEntryLabel(entry))
		}
		if detail := strings.TrimSpace(entry.Detail); detail != "" {
			lines = append(lines, renderPullRequestOverviewEntryDetail(detail))
		}
	}

	return strings.Join(lines, "\n")
}

func renderPullRequestOverviewEntryLabel(entry pullRequestOverviewEntry) string {
	text := strings.TrimSpace(entry.Label)
	if text == "" {
		return ""
	}
	if entry.ShowIcon {
		text = pullRequestOverviewStatusIcon(entry.Status) + " " + text
	}
	return styleText(text, foregroundColorEscape(pullRequestOverviewStatusHex(entry.Status)))
}

func renderPullRequestOverviewEntryDetail(detail string) string {
	return styleText("  "+strings.TrimSpace(detail), foregroundColorEscape(theme.InactiveTitleHex))
}

func buildPullRequestReviewersBlock(reviewers pullRequestReviewerOverview) pullRequestOverviewBlock {
	if len(reviewers.Entries) == 0 {
		return pullRequestOverviewBlock{
			Title: "Reviewers",
			Entries: []pullRequestOverviewEntry{{
				Label:    "No reviewers yet.",
				Status:   pullRequestOverviewStatusMuted,
				ShowIcon: false,
			}},
		}
	}

	return pullRequestOverviewBlock{
		Title:   "Reviewers",
		Summary: fmt.Sprintf("%d/%d", reviewers.ApprovedCount, len(reviewers.Entries)),
		Entries: reviewers.Entries,
	}
}

func buildPullRequestReviewerOverview(detail githubcli.PullRequestDetail) pullRequestReviewerOverview {
	latestReviews := latestPullRequestReviewsByLogin(detail.Reviews)
	entriesByLabel := map[string]pullRequestOverviewEntry{}
	for login, state := range latestReviews {
		label := formatLogin(login)
		if label == "-" {
			continue
		}
		entriesByLabel[label] = pullRequestOverviewEntry{Label: label, Status: pullRequestOverviewStatusForReviewState(state), ShowIcon: true}
	}

	for _, reviewRequest := range detail.ReviewRequests {
		label := pullRequestReviewRequestLabel(reviewRequest.RequestedReviewer)
		if label == "-" {
			continue
		}
		if _, ok := entriesByLabel[label]; ok {
			continue
		}
		entriesByLabel[label] = pullRequestOverviewEntry{Label: label, Status: pullRequestOverviewStatusPending, ShowIcon: true}
	}

	entries := make([]pullRequestOverviewEntry, 0, len(entriesByLabel))
	reviewers := pullRequestReviewerOverview{}
	for _, entry := range entriesByLabel {
		switch entry.Status {
		case pullRequestOverviewStatusSuccess:
			reviewers.ApprovedCount++
		case pullRequestOverviewStatusFailure:
			reviewers.ChangesRequestedCount++
		default:
			reviewers.PendingCount++
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(left int, right int) bool {
		leftPriority := pullRequestOverviewStatusPriority(entries[left].Status)
		rightPriority := pullRequestOverviewStatusPriority(entries[right].Status)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return entries[left].Label < entries[right].Label
	})

	reviewers.Entries = entries
	return reviewers
}

func latestPullRequestReviewsByLogin(reviews []githubcli.PullRequestReview) map[string]string {
	latestByLogin := latestPullRequestReviews(reviews)
	statesByLogin := make(map[string]string, len(latestByLogin))
	for login, review := range latestByLogin {
		statesByLogin[login] = strings.ToUpper(strings.TrimSpace(review.State))
	}
	return statesByLogin
}

func buildPullRequestMergeChecksBlock(detail githubcli.PullRequestDetail, reviewers pullRequestReviewerOverview) pullRequestOverviewBlock {
	return pullRequestOverviewBlock{
		Title: "Merge Checks",
		Entries: []pullRequestOverviewEntry{
			buildPullRequestReviewSummaryEntry(reviewers),
			buildPullRequestBuildSummaryEntry(detail.StatusCheckRollup),
			buildPullRequestMergeabilityEntry(detail),
		},
	}
}

func buildPullRequestReviewSummaryEntry(reviewers pullRequestReviewerOverview) pullRequestOverviewEntry {
	switch {
	case reviewers.ChangesRequestedCount > 0:
		return pullRequestOverviewEntry{
			Label:    "Reviews",
			Detail:   fmt.Sprintf("%d %s requested changes.", reviewers.ChangesRequestedCount, pluralize(reviewers.ChangesRequestedCount, "reviewer has", "reviewers have")),
			Status:   pullRequestOverviewStatusFailure,
			ShowIcon: true,
		}
	case reviewers.PendingCount > 0:
		return pullRequestOverviewEntry{
			Label:    "Reviews",
			Detail:   fmt.Sprintf("%d %s not approved yet.", reviewers.PendingCount, pluralize(reviewers.PendingCount, "reviewer has", "reviewers have")),
			Status:   pullRequestOverviewStatusPending,
			ShowIcon: true,
		}
	case reviewers.ApprovedCount > 0:
		return pullRequestOverviewEntry{
			Label:    "Reviews",
			Detail:   fmt.Sprintf("%d %s approved.", reviewers.ApprovedCount, pluralize(reviewers.ApprovedCount, "reviewer has", "reviewers have")),
			Status:   pullRequestOverviewStatusSuccess,
			ShowIcon: true,
		}
	default:
		return pullRequestOverviewEntry{
			Label:    "Reviews",
			Detail:   "No reviewer activity yet.",
			Status:   pullRequestOverviewStatusMuted,
			ShowIcon: true,
		}
	}
}

func buildPullRequestBuildSummaryEntry(checks []githubcli.PullRequestStatusCheck) pullRequestOverviewEntry {
	summary := summarizeStatusChecks(checks)
	if summary == "-" {
		return pullRequestOverviewEntry{
			Label:    "Builds",
			Detail:   "No status checks reported.",
			Status:   pullRequestOverviewStatusMuted,
			ShowIcon: true,
		}
	}

	return pullRequestOverviewEntry{
		Label:    "Builds",
		Detail:   summary,
		Status:   pullRequestOverviewStatusForChecks(checks),
		ShowIcon: true,
	}
}

func buildPullRequestMergeabilityEntry(detail githubcli.PullRequestDetail) pullRequestOverviewEntry {
	mergeable := strings.ToUpper(strings.TrimSpace(detail.Mergeable))
	mergeState := strings.ToUpper(strings.TrimSpace(detail.MergeStateStatus))
	switch {
	case mergeable == "MERGEABLE" || mergeState == "CLEAN":
		return pullRequestOverviewEntry{
			Label:    "No conflicts with base branch",
			Detail:   "Changes can be cleanly merged.",
			Status:   pullRequestOverviewStatusSuccess,
			ShowIcon: true,
		}
	case mergeable == "CONFLICTING" || mergeState == "DIRTY":
		return pullRequestOverviewEntry{
			Label:    "Conflicts with base branch",
			Detail:   "Conflicting files must be resolved before merging.",
			Status:   pullRequestOverviewStatusFailure,
			ShowIcon: true,
		}
	case mergeState != "" && mergeState != "UNKNOWN":
		return pullRequestOverviewEntry{
			Label:    "Mergeability",
			Detail:   "GitHub reports " + strings.ToLower(strings.ReplaceAll(mergeState, "_", " ")) + ".",
			Status:   pullRequestOverviewStatusPending,
			ShowIcon: true,
		}
	default:
		return pullRequestOverviewEntry{
			Label:    "Mergeability",
			Detail:   "Mergeability has not been reported yet.",
			Status:   pullRequestOverviewStatusMuted,
			ShowIcon: true,
		}
	}
}

func buildPullRequestBuildsBlock(detail githubcli.PullRequestDetail) pullRequestOverviewBlock {
	entries := make([]pullRequestOverviewEntry, 0, len(detail.StatusCheckRollup))
	for _, check := range detail.StatusCheckRollup {
		entries = append(entries, buildPullRequestBuildEntry(check))
	}
	if len(entries) == 0 {
		entries = append(entries, pullRequestOverviewEntry{Label: "No builds reported.", Status: pullRequestOverviewStatusMuted, ShowIcon: false})
	}

	return pullRequestOverviewBlock{Title: "Builds", Entries: entries}
}

func buildPullRequestBuildEntry(check githubcli.PullRequestStatusCheck) pullRequestOverviewEntry {
	return pullRequestOverviewEntry{
		Label:    fmt.Sprintf("%s (%s)", pullRequestOverviewCheckDisplayName(check), pullRequestOverviewCheckStateLabel(check)),
		Status:   pullRequestOverviewStatusForCheck(check),
		ShowIcon: true,
	}
}

func pullRequestOverviewCheckDisplayName(check githubcli.PullRequestStatusCheck) string {
	workflowName := strings.TrimSpace(check.WorkflowName)
	checkName := strings.TrimSpace(check.Name)
	switch {
	case workflowName != "" && checkName != "" && !strings.EqualFold(workflowName, checkName):
		return workflowName + " / " + checkName
	case workflowName != "":
		return workflowName
	case checkName != "":
		return checkName
	default:
		return "Build"
	}
}

func pullRequestOverviewCheckStateLabel(check githubcli.PullRequestStatusCheck) string {
	return classifyPullRequestStatusCheck(check).StateLabel
}

func pullRequestOverviewStatusForReviewState(state string) pullRequestOverviewStatus {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "APPROVED":
		return pullRequestOverviewStatusSuccess
	case "CHANGES_REQUESTED":
		return pullRequestOverviewStatusFailure
	default:
		return pullRequestOverviewStatusPending
	}
}

func pullRequestOverviewStatusForChecks(checks []githubcli.PullRequestStatusCheck) pullRequestOverviewStatus {
	overall := pullRequestOverviewStatusMuted
	for _, check := range checks {
		switch pullRequestOverviewStatusForCheck(check) {
		case pullRequestOverviewStatusFailure:
			return pullRequestOverviewStatusFailure
		case pullRequestOverviewStatusPending:
			overall = pullRequestOverviewStatusPending
		case pullRequestOverviewStatusSuccess:
			if overall == pullRequestOverviewStatusMuted {
				overall = pullRequestOverviewStatusSuccess
			}
		}
	}
	return overall
}

func pullRequestOverviewStatusForCheck(check githubcli.PullRequestStatusCheck) pullRequestOverviewStatus {
	return classifyPullRequestStatusCheck(check).OverviewStatus
}

func pullRequestOverviewStatusPriority(status pullRequestOverviewStatus) int {
	switch status {
	case pullRequestOverviewStatusFailure:
		return 0
	case pullRequestOverviewStatusPending:
		return 1
	case pullRequestOverviewStatusSuccess:
		return 2
	default:
		return 3
	}
}

func pullRequestOverviewStatusIcon(status pullRequestOverviewStatus) string {
	switch status {
	case pullRequestOverviewStatusSuccess:
		return pullRequestOverviewSuccessIcon
	case pullRequestOverviewStatusFailure:
		return pullRequestOverviewFailureIcon
	case pullRequestOverviewStatusPending:
		return pullRequestOverviewPendingIcon
	default:
		return pullRequestOverviewMutedIcon
	}
}

func pullRequestOverviewStatusHex(status pullRequestOverviewStatus) string {
	switch status {
	case pullRequestOverviewStatusSuccess:
		return theme.SuccessHex
	case pullRequestOverviewStatusFailure:
		return theme.FailureHex
	case pullRequestOverviewStatusPending:
		return theme.PendingHex
	default:
		return theme.MutedHex
	}
}
