package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestPullRequestOverviewStatusIcon_GivenSuccessAndFailureStatuses_WhenFormatting_ThenItUsesTheUpdatedGlyphs(t *testing.T) {
	testCases := []struct {
		name     string
		status   pullRequestOverviewStatus
		expected string
	}{
		{name: "success", status: pullRequestOverviewStatusSuccess, expected: ""},
		{name: "failure", status: pullRequestOverviewStatusFailure, expected: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := pullRequestOverviewStatusIcon(testCase.status)

			if actual != testCase.expected {
				t.Fatalf("expected icon %q, actual %q", testCase.expected, actual)
			}
		})
	}
}

func TestRenderPullRequestBrowserHeader_GivenReviewersAndChecks_WhenFormatting_ThenItKeepsOnlyTheMainOverviewMetadata(t *testing.T) {
	summary := githubcli.PullRequest{Title: "Overview PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{
		Number:            42,
		Author:            &githubcli.PullRequestAuthor{Login: "octocat"},
		State:             "OPEN",
		CreatedAt:         "2026-04-18T10:00:00Z",
		UpdatedAt:         "2026-04-18T12:30:00Z",
		BaseRefName:       "main",
		HeadRefName:       "feature/overview",
		Assignees:         []githubcli.PullRequestAuthor{{Login: "assignee-one"}},
		ReviewRequests:    []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}},
		Reviews:           []githubcli.PullRequestReview{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"}},
	}

	actualDocument := newDetailDocument(renderPullRequestBrowserHeader(summary, detail), 120)
	actualText := string(actualDocument.text)

	for _, expected := range []string{"acme/widgets#42 Overview PR", "Created by", "@octocat", "Assigned to", "@assignee-one", "main ← feature/overview", detailStatusIcon + " OPEN"} {
		if !strings.Contains(actualText, expected) {
			t.Fatalf("expected browser header to contain %q, actual %q", expected, actualText)
		}
	}
	for _, unexpected := range []string{detailChecksIcon, detailReviewRequestsIcon, detailApprovalIcon, detailAssigneesIcon} {
		if strings.Contains(actualText, unexpected) {
			t.Fatalf("expected browser header to omit %q, actual %q", unexpected, actualText)
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
		pullRequestOverviewFailureIcon + " Reviewers (1/4)",
		"@acme/platform",
		"@reviewer-approved",
		"@reviewer-blocked",
		pullRequestOverviewFailureIcon + " Merge Checks",
		"Reviews",
		"1 reviewer has requested changes.",
		"1 passing, 1 failing, 1 pending",
		"No conflicts with base branch",
		"Changes can be cleanly merged.",
		pullRequestOverviewFailureIcon + " Builds",
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

	reviewersHeadingLineIndex, reviewersHeadingLine := given_detailDocumentLineContaining(t, actualDocument, pullRequestOverviewFailureIcon+" Reviewers (1/4)")
	reviewersHeadingIconIndex := given_runeIndexInString(t, reviewersHeadingLine, pullRequestOverviewFailureIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[reviewersHeadingLineIndex][reviewersHeadingIconIndex]; actualStylePrefix != foregroundColorEscape(theme.FailureHex) {
		t.Fatalf("expected reviewers heading prefix %q, actual %q", foregroundColorEscape(theme.FailureHex), actualStylePrefix)
	}

	approvedLineIndex, approvedLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-approved")
	approvedIconIndex := given_runeIndexInString(t, approvedLine, pullRequestOverviewSuccessIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[approvedLineIndex][approvedIconIndex]; actualStylePrefix != foregroundColorEscape(theme.SuccessHex) {
		t.Fatalf("expected approved reviewer prefix %q, actual %q", foregroundColorEscape(theme.SuccessHex), actualStylePrefix)
	}

	blockedLineIndex, blockedLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-blocked")
	blockedIconIndex := given_runeIndexInString(t, blockedLine, pullRequestOverviewFailureIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[blockedLineIndex][blockedIconIndex]; actualStylePrefix != foregroundColorEscape(theme.FailureHex) {
		t.Fatalf("expected blocked reviewer prefix %q, actual %q", foregroundColorEscape(theme.FailureHex), actualStylePrefix)
	}

	mergeChecksHeadingLineIndex, mergeChecksHeadingLine := given_detailDocumentLineContaining(t, actualDocument, pullRequestOverviewFailureIcon+" Merge Checks")
	mergeChecksHeadingIconIndex := given_runeIndexInString(t, mergeChecksHeadingLine, pullRequestOverviewFailureIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[mergeChecksHeadingLineIndex][mergeChecksHeadingIconIndex]; actualStylePrefix != foregroundColorEscape(theme.FailureHex) {
		t.Fatalf("expected merge checks heading prefix %q, actual %q", foregroundColorEscape(theme.FailureHex), actualStylePrefix)
	}

	buildsHeadingLineIndex, buildsHeadingLine := given_detailDocumentLineContaining(t, actualDocument, pullRequestOverviewFailureIcon+" Builds")
	buildsHeadingIconIndex := given_runeIndexInString(t, buildsHeadingLine, pullRequestOverviewFailureIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[buildsHeadingLineIndex][buildsHeadingIconIndex]; actualStylePrefix != foregroundColorEscape(theme.FailureHex) {
		t.Fatalf("expected builds heading prefix %q, actual %q", foregroundColorEscape(theme.FailureHex), actualStylePrefix)
	}
}

func TestRenderPullRequestOverviewSection_GivenMultipleBlocks_WhenFormatting_ThenItDoesNotInsertBlankLinesBetweenOverviewSections(t *testing.T) {
	detail := githubcli.PullRequestDetail{
		ReviewRequests:    []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-one"}}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"}},
	}

	actual := renderPullRequestOverviewSection(buildPullRequestOverviewSection(detail), 80)

	if strings.Contains(actual, "\n\n") {
		t.Fatalf("expected overview sections to stay consecutive without blank spacer lines, actual %q", actual)
	}
}

func TestRenderPullRequestOverviewSection_GivenGenericStatusColorOverrides_WhenFormatting_ThenItUsesTheGenericThemeColors(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.Palette{SuccessHex: "#7FB069", FailureHex: "#E46876", PendingHex: "#727169", MutedHex: "#8A8980"})

	detail := githubcli.PullRequestDetail{
		ReviewRequests:    []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-pending"}}},
		Reviews:           []githubcli.PullRequestReview{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-blocked"}, State: "CHANGES_REQUESTED", SubmittedAt: "2026-04-21T11:00:00Z"}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"}},
	}

	actualDocument := newDetailDocument(renderPullRequestOverviewSection(buildPullRequestOverviewSection(detail), 80), 80)

	pendingLineIndex, pendingLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-pending")
	pendingIconIndex := given_runeIndexInString(t, pendingLine, pullRequestOverviewPendingIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[pendingLineIndex][pendingIconIndex]; actualStylePrefix != foregroundColorEscape(theme.PendingHex) {
		t.Fatalf("expected pending reviewer prefix %q, actual %q", foregroundColorEscape(theme.PendingHex), actualStylePrefix)
	}

	blockedLineIndex, blockedLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-blocked")
	blockedIconIndex := given_runeIndexInString(t, blockedLine, pullRequestOverviewFailureIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[blockedLineIndex][blockedIconIndex]; actualStylePrefix != foregroundColorEscape(theme.FailureHex) {
		t.Fatalf("expected blocked reviewer prefix %q, actual %q", foregroundColorEscape(theme.FailureHex), actualStylePrefix)
	}

	successLineIndex, successLine := given_detailDocumentLineContaining(t, actualDocument, "lint (Successful)")
	successIconIndex := given_runeIndexInString(t, successLine, pullRequestOverviewSuccessIcon)
	if actualStylePrefix := actualDocument.lineStylePrefixes[successLineIndex][successIconIndex]; actualStylePrefix != foregroundColorEscape(theme.SuccessHex) {
		t.Fatalf("expected successful build prefix %q, actual %q", foregroundColorEscape(theme.SuccessHex), actualStylePrefix)
	}
}

func TestRenderPullRequestOverviewSection_GivenMultipleEntriesInTheSameBlock_WhenFormatting_ThenItDoesNotInsertBlankLinesBetweenThem(t *testing.T) {
	detail := githubcli.PullRequestDetail{
		ReviewRequests: []githubcli.PullRequestReviewRequest{
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-one"}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-two"}},
		},
	}

	actual := renderPullRequestOverviewSection(buildPullRequestOverviewSection(detail), 80)

	if strings.Contains(actual, "@reviewer-one\n│                                                                          │\n│ 󰦖 @reviewer-two") {
		t.Fatalf("expected reviewers to stay compact without blank spacer lines, actual %q", actual)
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
