package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestRenderPullRequestBrowserHeader_GivenReviewersAndChecks_WhenFormatting_ThenItKeepsOnlyTheMainMetadata(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{
		Number:         42,
		State:          "OPEN",
		BaseRefName:    "main",
		HeadRefName:    "feature/overview",
		Assignees:      []githubcli.PullRequestAuthor{{Login: "assignee-one"}},
		ReviewRequests: []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}},
		Reviews:        []githubcli.PullRequestReview{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"}},
	}

	actual := renderPullRequestBrowserHeader(summary, detail)

	for _, expected := range []string{detailBranchIcon + " main ← feature/overview", detailStatusIcon + " OPEN", detailAssigneesIcon + " @assignee-one"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected browser header to contain %q, actual %q", expected, actual)
		}
	}
	for _, unexpected := range []string{detailChecksIcon, detailReviewRequestsIcon, detailApprovalIcon} {
		if strings.Contains(actual, unexpected) {
			t.Fatalf("expected browser header to omit %q, actual %q", unexpected, actual)
		}
	}
}

func TestRenderPullRequestOverviewSection_GivenPopulatedMetadata_WhenFormatting_ThenItShowsReviewersMergeChecksAndBuildsInSeparateBoxes(t *testing.T) {
	detail := githubcli.PullRequestDetail{
		ReviewRequests: []githubcli.PullRequestReviewRequest{
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Slug: "platform", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
		},
		Reviews: []githubcli.PullRequestReview{
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-blocked"}, State: "CHANGES_REQUESTED", SubmittedAt: "2026-04-21T11:00:00Z"},
		},
		Mergeable:        "MERGEABLE",
		MergeStateStatus: "CLEAN",
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{
			{Name: "lint", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE"},
			{Name: "deploy", Status: "IN_PROGRESS"},
		},
	}

	actualDocument := newDetailDocument(renderPullRequestOverviewSection(buildPullRequestOverviewSection(detail), 80), 80)
	actualText := string(actualDocument.text)

	for _, expected := range []string{
		"Reviewers (1/4)",
		"@acme/platform",
		"@reviewer-approved",
		"@reviewer-blocked",
		"Merge Checks",
		"Reviews",
		"1 reviewer has requested changes.",
		"1 passing, 1 failing, 1 pending",
		"No conflicts with base branch",
		"Changes can be cleanly merged.",
		"Builds",
		"CI / lint (Successful)",
		"CI / test (Failed)",
		"deploy (Pending)",
	} {
		if !strings.Contains(actualText, expected) {
			t.Fatalf("expected overview section to contain %q, actual %q", expected, actualText)
		}
	}

	borderLineIndex, borderLine := given_detailDocumentLineContaining(t, actualDocument, "╭")
	if borderLine == "" {
		t.Fatal("expected the overview section to render boxed blocks")
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[borderLineIndex][0]; actualStylePrefix != foregroundColorEscape(theme.InactiveBorderHex) {
		t.Fatalf("expected overview box border prefix %q, actual %q", foregroundColorEscape(theme.InactiveBorderHex), actualStylePrefix)
	}

	approvedLineIndex, approvedLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-approved")
	approvedIconIndex := given_runeIndexInString(t, approvedLine, pullRequestOverviewSuccessIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[approvedLineIndex][approvedIconIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffAdditionForegroundHex) {
		t.Fatalf("expected approved reviewer prefix %q, actual %q", foregroundColorEscape(theme.DiffAdditionForegroundHex), actualStylePrefix)
	}

	blockedLineIndex, blockedLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-blocked")
	blockedIconIndex := given_runeIndexInString(t, blockedLine, pullRequestOverviewFailureIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[blockedLineIndex][blockedIconIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffDeletionForegroundHex) {
		t.Fatalf("expected blocked reviewer prefix %q, actual %q", foregroundColorEscape(theme.DiffDeletionForegroundHex), actualStylePrefix)
	}
}

func TestRenderPullRequestOverviewSection_GivenPartialMetadata_WhenFormatting_ThenItShowsKnownValuesAndKeepsPlaceholdersForMissingBlocks(t *testing.T) {
	detail := githubcli.PullRequestDetail{
		ReviewRequests:   []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}},
		MergeStateStatus: "BLOCKED",
	}

	actual := renderPullRequestOverviewSection(buildPullRequestOverviewSection(detail), 80)

	for _, expected := range []string{"Reviewers (0/1)", "@reviewer-requested", "1 reviewer has not approved yet.", "No status checks reported.", "GitHub reports blocked.", "No builds reported."} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected overview section to contain %q, actual %q", expected, actual)
		}
	}
}

func TestRenderPullRequestOverviewSection_GivenEmptyMetadata_WhenFormatting_ThenItShowsPlaceholdersForEachBlock(t *testing.T) {
	actual := renderPullRequestOverviewSection(buildPullRequestOverviewSection(githubcli.PullRequestDetail{}), 80)

	for _, expected := range []string{"Reviewers", "No reviewers yet.", "Merge Checks", "No reviewer activity yet.", "No status checks reported.", "Mergeability has not been reported yet.", "Builds", "No builds reported."} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected overview section to contain %q, actual %q", expected, actual)
		}
	}
}
