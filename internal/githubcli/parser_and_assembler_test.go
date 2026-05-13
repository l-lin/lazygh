package githubcli

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParsePullRequestListReviewMetadata_GivenFixture_WhenParsing_ThenItHydratesNormalizedMetadata(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "list_review_metadata.json")

	actual, actualErr := parsePullRequestListReviewMetadata(fixture)

	then_noError(t, actualErr)
	expected := map[string]pullRequestListReviewMetadata{
		"PR_kwDOA": {
			ReviewDecision:         "REVIEW_REQUIRED",
			Mergeable:              "MERGEABLE",
			MergeStateStatus:       "CLEAN",
			StatusCheckRollupState: "SUCCESS",
			ReviewRequests: []PullRequestReviewRequest{
				{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-one", Name: "Reviewer One"}},
				{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "Team", Name: "Platform", Slug: "acme/platform", Organization: &PullRequestReviewRequestOrganization{Login: "acme"}}},
			},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected metadata %+v, actual %+v", expected, actual)
	}
}

func TestParsePullRequestDetailResponse_GivenFixture_WhenParsing_ThenItReturnsTheNormalizedDomainDetail(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "detail.json")

	actual, actualErr := parsePullRequestDetailResponse(fixture)

	then_noError(t, actualErr)
	expected := PullRequestDetail{
		ID:               "PR_kwDOA",
		Title:            "Ship it",
		Number:           42,
		URL:              "https://github.com/acme/widgets/pull/42",
		Body:             "## Summary\n\n- parser cleanup",
		Author:           &PullRequestAuthor{Login: "octocat", Name: "Octo Cat"},
		State:            "OPEN",
		CreatedAt:        "2026-04-18T10:00:00Z",
		UpdatedAt:        "2026-04-18T12:30:00Z",
		Labels:           []PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
		Assignees:        []PullRequestAuthor{{Login: "assignee-one", Name: "Assignee One"}},
		ReviewRequests:   []PullRequestReviewRequest{{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested", Name: "Reviewer Requested"}}, {RequestedReviewer: PullRequestRequestedReviewer{TypeName: "Team", Name: "Platform", Slug: "acme/platform", Organization: &PullRequestReviewRequestOrganization{Login: "acme"}}}},
		BaseRefName:      "main",
		HeadRefName:      "feature/parser-cleanup",
		MergeStateStatus: "CLEAN",
		Mergeable:        "MERGEABLE",
		Comments: []PullRequestComment{{
			ID:              "IC_kwDOA",
			Author:          &PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:            "Looks good",
			CreatedAt:       "2026-04-18T13:00:00Z",
			URL:             "https://github.com/acme/widgets/pull/42#issuecomment-1",
			ReactionGroups:  []ReactionGroup{{Content: ReactionContentThumbsUp, TotalCount: 2, ViewerHasReacted: true}},
			ViewerDidAuthor: false,
		}},
		Commits: []PullRequestCommit{{
			OID:             "abc123",
			MessageHeadline: "split parser",
			MessageBody:     "move dto and mapping",
			AuthoredDate:    "2026-04-18T14:00:00Z",
			CommittedDate:   "2026-04-18T14:05:00Z",
			Authors:         []PullRequestCommitAuthor{{Login: "dev-one", Name: "Dev One", Email: "dev@example.com"}},
		}},
		Reviews: []PullRequestReview{{
			Author:      &PullRequestCommentAuthor{Login: "reviewer-two"},
			State:       "APPROVED",
			SubmittedAt: "2026-04-18T15:00:00Z",
		}},
		Additions:    12,
		Deletions:    3,
		ChangedFiles: 5,
		StatusCheckRollup: []PullRequestStatusCheck{{
			TypeName:     "CheckRun",
			Name:         "lint",
			Status:       "COMPLETED",
			Conclusion:   "SUCCESS",
			WorkflowName: "CI",
		}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected detail %+v, actual %+v", expected, actual)
	}
}

func TestParsePullRequestInlineCommentsResponse_GivenFixture_WhenParsing_ThenItReturnsNormalizedInlineComments(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "inline_comments.json")

	actual, actualErr := parsePullRequestInlineCommentsResponse(fixture)

	then_noError(t, actualErr)
	expected := []PullRequestInlineComment{{
		ID:           "PRRC_kw123",
		Author:       &PullRequestCommentAuthor{Login: "reviewer-inline"},
		Body:         "inline body",
		CreatedAt:    "2026-04-18T14:00:00Z",
		URL:          "https://example.com/discussion",
		Path:         "internal/tui/pull_request_detail.go",
		Line:         19,
		OriginalLine: 21,
		Side:         "LEFT",
		StartSide:    "LEFT",
		SubjectType:  "LINE",
		DiffHunk:     "@@ -19,1 +21,1 @@\n-old\n+new",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected inline comments %+v, actual %+v", expected, actual)
	}
}

func TestParsePullRequestReviewThreadsPage_GivenFixture_WhenParsing_ThenItReturnsTheNormalizedPage(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "review_threads_page.json")

	actual, actualErr := parsePullRequestReviewThreadsPage(fixture)

	then_noError(t, actualErr)
	if !actual.HasNextPage || actual.EndCursor != "THREAD_CURSOR_1" {
		t.Fatalf("expected a next page with cursor %q, actual %+v", "THREAD_CURSOR_1", actual)
	}
	expectedThread := pullRequestReviewThreadPageNode{
		Thread: PullRequestReviewThread{
			ID:                 "thread-1",
			IsResolved:         true,
			ViewerCanResolve:   true,
			ViewerCanUnresolve: true,
			Path:               "internal/tui/render.go",
			Line:               11,
			OriginalLine:       11,
			StartLine:          10,
			OriginalStartLine:  10,
			DiffSide:           "RIGHT",
			StartDiffSide:      "LEFT",
			Comments: []PullRequestComment{{
				ID:              "PRRC_1",
				Author:          &PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:            "First reply",
				CreatedAt:       "2026-04-20T10:00:00Z",
				URL:             "https://example.com/thread/1",
				DiffHunk:        "@@ -1 +1 @@\n-old\n+new",
				State:           "PENDING",
				ReactionGroups:  []ReactionGroup{{Content: ReactionContentRocket, TotalCount: 4}},
				ViewerDidAuthor: false,
			}},
		},
		CommentsHasNextPage: false,
		CommentsEndCursor:   "",
	}
	if !reflect.DeepEqual(actual.Threads, []pullRequestReviewThreadPageNode{expectedThread}) {
		t.Fatalf("expected review thread page %+v, actual %+v", []pullRequestReviewThreadPageNode{expectedThread}, actual.Threads)
	}
}

func TestParsePullRequestReviewThreadCommentsPage_GivenFixture_WhenParsing_ThenItReturnsNormalizedThreadComments(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "review_thread_comments_page.json")

	actual, actualErr := parsePullRequestReviewThreadCommentsPage(fixture)

	then_noError(t, actualErr)
	expected := pullRequestReviewThreadCommentsPage{
		Comments: []PullRequestComment{{
			ID:              "PRRC_2",
			Author:          &PullRequestCommentAuthor{Login: "reviewer-two"},
			Body:            "Second reply",
			CreatedAt:       "2026-04-20T10:05:00Z",
			URL:             "https://example.com/thread/2",
			DiffHunk:        "@@ -1 +1 @@\n-old\n+new",
			State:           "SUBMITTED",
			ReactionGroups:  []ReactionGroup{{Content: ReactionContentThumbsDown, TotalCount: 1, ViewerHasReacted: true}},
			ViewerDidAuthor: true,
		}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected thread comments page %+v, actual %+v", expected, actual)
	}
}

func TestParsePullRequestReactionTargetsPage_GivenFixture_WhenParsing_ThenItReturnsNormalizedReactionTargets(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "reaction_targets.json")

	actual, actualErr := parsePullRequestReactionTargetsPage(fixture)

	then_noError(t, actualErr)
	expected := pullRequestReactionTargetsPage{
		PullRequestID:  "PR_kwDOA",
		ReactionGroups: []ReactionGroup{{Content: ReactionContentThumbsUp, TotalCount: 2, ViewerHasReacted: true}, {Content: ReactionContentHooray, TotalCount: 1}},
		Comments: []PullRequestComment{{
			ID:              "IC_kwDOA",
			Author:          &PullRequestCommentAuthor{Login: "reviewer"},
			Body:            "Looks good",
			CreatedAt:       "2026-04-18T13:00:00Z",
			URL:             "https://github.com/acme/widgets/pull/42#issuecomment-1",
			ReactionGroups:  []ReactionGroup{{Content: ReactionContentEyes, TotalCount: 3}},
			ViewerDidAuthor: false,
		}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected reaction targets %+v, actual %+v", expected, actual)
	}
}

func TestParsePullRequestReviewCommentReactionGroups_GivenFixture_WhenParsing_ThenItReturnsNormalizedReactionGroups(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "review_comment_reaction_groups.json")

	actual, actualErr := parsePullRequestReviewCommentReactionGroups(fixture)

	then_noError(t, actualErr)
	expected := map[string][]ReactionGroup{
		"PRRC_kwDOA": {
			{Content: ReactionContentLaugh, TotalCount: 2},
			{Content: ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected review comment reaction groups %+v, actual %+v", expected, actual)
	}
}

func TestBuildAssembler_GivenFixture_WhenParsingBuildInfos_ThenItReturnsNormalizedBuilds(t *testing.T) {
	fixture := given_pullRequestFixtureBytes(t, "build_infos.json")
	assembler := BuildAssembler{}

	actual, actualErr := assembler.ParseBuildInfos(fixture)

	then_noError(t, actualErr)
	expected := []PullRequestBuildInfo{{
		Bucket:      "fail",
		CompletedAt: "2026-04-18T13:04:00Z",
		Description: "widget smoke test timed out",
		Event:       "pull_request",
		Link:        "https://github.com/acme/widgets/actions/runs/42",
		Name:        "test",
		StartedAt:   "2026-04-18T13:00:00Z",
		State:       "FAILURE",
		Workflow:    "CI",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected build infos %+v, actual %+v", expected, actual)
	}
}

func TestNotificationAssembler_GivenFixtures_WhenParsing_ThenItReturnsNormalizedNotificationsAndDetails(t *testing.T) {
	assembler := NotificationAssembler{}

	notifications, notificationsErr := assembler.ParseList(given_notificationFixtureBytes(t, "threads_with_done.json"))
	then_noError(t, notificationsErr)
	if len(notifications) != 1 {
		t.Fatalf("expected notification count %d, actual %d", 1, len(notifications))
	}
	if notifications[0].ID != "1001" || notifications[0].Subject.Title != "feat: ship notifications" || notifications[0].Subject.Type != NotificationSubjectTypePullRequest {
		t.Fatalf("expected the normalized notification fixture, actual %+v", notifications[0])
	}

	issueDetail, issueErr := assembler.ParseIssueDetail(given_notificationFixtureBytes(t, "issue_detail.json"))
	then_noError(t, issueErr)
	if issueDetail.Title != "Support for skills" || issueDetail.State != "open" || issueDetail.Author == nil || issueDetail.Author.Login != "octocat" {
		t.Fatalf("expected normalized issue detail, actual %+v", issueDetail)
	}

	releaseDetail, releaseErr := assembler.ParseReleaseDetail(given_notificationFixtureBytes(t, "release_detail.json"))
	then_noError(t, releaseErr)
	if releaseDetail.Name != "Notifications 3.5.0" || releaseDetail.TagName != "v3.5.0" || releaseDetail.Author == nil || releaseDetail.Author.Login != "release-bot" {
		t.Fatalf("expected normalized release detail, actual %+v", releaseDetail)
	}
}

func TestGraphQLParsers_GivenGraphQLErrorEnvelopes_WhenParsing_ThenTheyReturnEndpointSpecificErrors(t *testing.T) {
	testCases := []struct {
		name        string
		parse       func([]byte) error
		expectedErr error
	}{
		{
			name: "pull request list metadata",
			parse: func(payload []byte) error {
				_, actualErr := parsePullRequestListReviewMetadata(payload)
				return actualErr
			},
			expectedErr: ErrInvalidPullRequestReviewMetadataResponse,
		},
		{
			name: "review threads page",
			parse: func(payload []byte) error {
				_, actualErr := parsePullRequestReviewThreadsPage(payload)
				return actualErr
			},
			expectedErr: ErrInvalidPullRequestReviewThreadsResponse,
		},
		{
			name: "review thread comments page",
			parse: func(payload []byte) error {
				_, actualErr := parsePullRequestReviewThreadCommentsPage(payload)
				return actualErr
			},
			expectedErr: ErrInvalidPullRequestReviewThreadsResponse,
		},
		{
			name: "reaction targets page",
			parse: func(payload []byte) error {
				_, actualErr := parsePullRequestReactionTargetsPage(payload)
				return actualErr
			},
			expectedErr: ErrInvalidPullRequestReactionTargetsResponse,
		},
		{
			name: "review comment reaction groups",
			parse: func(payload []byte) error {
				_, actualErr := parsePullRequestReviewCommentReactionGroups(payload)
				return actualErr
			},
			expectedErr: ErrInvalidPullRequestReviewCommentReactionGroupsPayload,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualErr := testCase.parse([]byte(`{"errors":[{"message":"viewer is forbidden"}]}`))

			if !errors.Is(actualErr, testCase.expectedErr) {
				t.Fatalf("expected error %v, actual %v", testCase.expectedErr, actualErr)
			}
			if actualErr == nil || !strings.Contains(actualErr.Error(), "viewer is forbidden") {
				t.Fatalf("expected GraphQL error message %q, actual %v", "viewer is forbidden", actualErr)
			}
		})
	}
}

func TestPullRequestDetailAssembler_GivenWorkflowLoaders_WhenAssembling_ThenItStitchesTheRichDetail(t *testing.T) {
	assembler := PullRequestDetailAssembler{
		LoadBaseDetail: func(repository string, number int) (PullRequestDetail, error) {
			if repository != "acme/widgets" || number != 42 {
				t.Fatalf("expected base detail lookup for %s#%d", "acme/widgets", 42)
			}
			return PullRequestDetail{
				Title:             "Ship it",
				Number:            42,
				StatusCheckRollup: []PullRequestStatusCheck{{Name: "lint", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "SUCCESS"}},
				Comments:          []PullRequestComment{{ID: "IC_kwDOA", Body: "Base comment"}},
			}, nil
		},
		HydrateBuildLinks: func(repository string, number int, checks []PullRequestStatusCheck) []PullRequestStatusCheck {
			return mergePullRequestStatusCheckLinks(checks, []PullRequestBuildInfo{{Name: "lint", Workflow: "CI", Link: "https://github.com/acme/widgets/actions/runs/1"}})
		},
		ListInlineComments: func(repository string, number int) ([]PullRequestInlineComment, error) {
			return []PullRequestInlineComment{{ID: "PRRC_kwDOA", Body: "Inline comment"}}, nil
		},
		ListReviewThreads: func(repository string, number int) ([]PullRequestReviewThread, error) {
			return []PullRequestReviewThread{{ID: "thread-1", Comments: []PullRequestComment{{ID: "PRRC_1", Body: "Thread comment"}}}}, nil
		},
		ListReactionTargets: func(repository string, number int) (pullRequestReactionTargets, error) {
			return pullRequestReactionTargets{
				PullRequestID:  "PR_kwDOA",
				ReactionGroups: []ReactionGroup{{Content: ReactionContentThumbsUp, TotalCount: 2, ViewerHasReacted: true}},
				Comments:       []PullRequestComment{{ID: "IC_kwDOA", Body: "Enriched comment", ReactionGroups: []ReactionGroup{{Content: ReactionContentEyes, TotalCount: 3}}}},
			}, nil
		},
		ListReviewCommentReactionGroups: func(ids []string) (map[string][]ReactionGroup, error) {
			expectedIDs := []string{"PRRC_kwDOA"}
			if !reflect.DeepEqual(ids, expectedIDs) {
				t.Fatalf("expected inline reaction target ids %v, actual %v", expectedIDs, ids)
			}
			return map[string][]ReactionGroup{"PRRC_kwDOA": {{Content: ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}}}, nil
		},
	}

	actual, actualErr := assembler.Assemble("acme/widgets", 42)

	then_noError(t, actualErr)
	if actual.ID != "PR_kwDOA" {
		t.Fatalf("expected pull request id %q, actual %q", "PR_kwDOA", actual.ID)
	}
	if len(actual.StatusCheckRollup) != 1 || actual.StatusCheckRollup[0].Link != "https://github.com/acme/widgets/actions/runs/1" {
		t.Fatalf("expected hydrated build links, actual %+v", actual.StatusCheckRollup)
	}
	if len(actual.Comments) != 1 || actual.Comments[0].Body != "Enriched comment" || len(actual.Comments[0].ReactionGroups) != 1 {
		t.Fatalf("expected enriched top-level comments, actual %+v", actual.Comments)
	}
	if len(actual.InlineComments) != 1 || len(actual.InlineComments[0].ReactionGroups) != 1 || actual.InlineComments[0].ReactionGroups[0].Content != ReactionContentHeart {
		t.Fatalf("expected enriched inline comments, actual %+v", actual.InlineComments)
	}
	if len(actual.InlineCommentThreads) != 1 || actual.InlineCommentThreads[0].ID != "thread-1" {
		t.Fatalf("expected review threads to be stitched into the detail, actual %+v", actual.InlineCommentThreads)
	}
}

func TestPullRequestDiffAssembler_GivenWorkflowLoaders_WhenAssembling_ThenItReturnsTheUnifiedDiffFilesAndThreads(t *testing.T) {
	assembler := PullRequestDiffAssembler{
		LoadUnifiedDiff: func(repository string, number int) (string, error) {
			if repository != "acme/widgets" || number != 42 {
				t.Fatalf("expected unified diff lookup for %s#%d", "acme/widgets", 42)
			}
			return "diff --git a/internal/tui/render.go b/internal/tui/render.go\n@@ -1 +1 @@\n-old\n+new", nil
		},
		LoadDiffFiles: func(repository string, number int) ([]PullRequestDiffFile, error) {
			return []PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 1}}, nil
		},
		ListReviewThreads: func(repository string, number int) ([]PullRequestReviewThread, error) {
			return []PullRequestReviewThread{{ID: "thread-1", Path: "internal/tui/render.go", Line: 11, DiffSide: "RIGHT", Comments: []PullRequestComment{{ID: "PRRC_1", Body: "Reply"}}}}, nil
		},
	}

	actual, actualErr := assembler.Assemble("acme/widgets", 42)

	then_noError(t, actualErr)
	expected := PullRequestDiff{
		UnifiedDiff: "diff --git a/internal/tui/render.go b/internal/tui/render.go\n@@ -1 +1 @@\n-old\n+new",
		Files:       []PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 1}},
		Threads:     []PullRequestReviewThread{{ID: "thread-1", Path: "internal/tui/render.go", Line: 11, DiffSide: "RIGHT", Comments: []PullRequestComment{{ID: "PRRC_1", Body: "Reply"}}}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected diff %+v, actual %+v", expected, actual)
	}
}

func given_pullRequestFixtureBytes(t *testing.T, name string) []byte {
	t.Helper()

	actual, actualErr := os.ReadFile(filepath.Join("testdata", "pull_requests", name))
	then_noError(t, actualErr)
	return actual
}
