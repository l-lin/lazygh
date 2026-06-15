package githubcli

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestGetPullRequestDetail_GivenValidGhResponsesWithInlineComments_WhenFetching_ThenReturnsTheRichPullRequestDetail(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"Add a real detail pane","number":42,"url":"https://github.com/acme/widgets/pull/42","body":"## Summary\n\n- render markdown\n- show comments","author":{"login":"octocat","name":"Octo Cat","is_bot":false},"state":"OPEN","isDraft":false,"createdAt":"2026-04-18T10:00:00Z","updatedAt":"2026-04-18T12:30:00Z","labels":[{"name":"bug"},{"name":"backend"}],"assignees":[{"login":"assignee-one","name":"Assignee One","is_bot":false},{"login":"assignee-two","name":"Assignee Two","is_bot":false}],"reviewRequests":[{"__typename":"User","login":"reviewer-requested","name":"Reviewer Requested"},{"__typename":"Team","name":"Platform","slug":"acme/platform"}],"baseRefName":"main","headRefName":"feature/detail","mergeStateStatus":"CLEAN","mergeable":"MERGEABLE","autoMergeRequest":{"enabledAt":"2026-04-18T12:45:00Z"},"comments":[{"author":{"login":"reviewer"},"body":"Looks good to me","createdAt":"2026-04-18T13:00:00Z","url":"https://github.com/acme/widgets/pull/42#issuecomment-1"}],"commits":[{"oid":"e9a3253762e768badaa1d4a5b3d267416d1e42f4","messageHeadline":"reintroduce interactive gh pr","messageBody":"this commit adds gh pr back","authoredDate":"2019-10-04T15:23:39Z","committedDate":"2019-10-04T15:57:48Z","authors":[{"email":"vilmibm@github.com","login":"vilmibm","name":"nate smith"}]}],"additions":12,"deletions":3,"changedFiles":5,"statusCheckRollup":[{"__typename":"CheckRun","name":"lint","status":"COMPLETED","conclusion":"SUCCESS","workflowName":"CI"}]}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"isMergeQueueEnabled":true,"isInMergeQueue":true,"mergeQueueEntry":{"id":" MQE_1 ","state":" QUEUED ","position":3,"estimatedTimeToMerge":11}}}}}`)},
			{stdout: []byte(`{"status":"diverged","ahead_by":2,"behind_by":3,"total_commits":2}`)},
			{stdout: []byte(`[{"bucket":"pass","completedAt":"2026-04-18T13:05:00Z","description":"lint passed","event":"pull_request","link":"https://github.com/acme/widgets/actions/runs/1","name":"lint","startedAt":"2026-04-18T13:00:00Z","state":"SUCCESS","workflow":"CI"}]`)},
			{stdout: []byte(`[[{"node_id":"PRRC_1","user":{"login":"reviewer-inline"},"body":"Please keep the blank line.","created_at":"2026-04-18T14:00:00Z","html_url":"https://github.com/acme/widgets/pull/42#discussion_r1","path":"internal/tui/render.go","line":252,"original_line":252,"side":"RIGHT","start_side":"RIGHT","subject_type":"LINE","diff_hunk":"@@ -250,3 +250,4 @@\n header := renderPullRequestDetailHeader(*row.Summary, result.detail)\n content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)\n-if program.activeDetailTab == CommentsDetailTab {\n+if program.activeDetailTab == CommentsDetailTab {\n  content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)"}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"thread-1","isResolved":false,"isOutdated":false,"path":"internal/tui/render.go","line":252,"diffSide":"RIGHT","comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"author":{"login":"reviewer-inline"},"body":"Please keep the blank line.","createdAt":"2026-04-18T14:00:00Z","url":"https://github.com/acme/widgets/pull/42#discussion_r1","diffHunk":"@@ -250,3 +250,4 @@\n header := renderPullRequestDetailHeader(*row.Summary, result.detail)\n content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)\n-if program.activeDetailTab == CommentsDetailTab {\n+if program.activeDetailTab == CommentsDetailTab {\n  content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)"}]}}]}}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reactionGroups":[],"comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
			{stdout: []byte(`{"data":{"nodes":[{"id":"PRRC_1","reactionGroups":[]}]}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"pr", "view", "42", "-R", "acme/widgets", "--json", pullRequestDetailJSONFields}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestMergeQueueQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/compare/main...feature/detail"}},
		{name: "gh", args: []string{"pr", "checks", "42", "-R", "acme/widgets", "--json", "bucket,completedAt,description,event,link,name,startedAt,state,workflow"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42/comments?per_page=100", "--paginate", "--slurp"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReactionTargetsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewCommentReactionGroupsQuery, "-F", "ids[]=PRRC_1"}},
	})

	expected := PullRequestDetail{
		Title:               "Add a real detail pane",
		Number:              42,
		URL:                 "https://github.com/acme/widgets/pull/42",
		Body:                "## Summary\n\n- render markdown\n- show comments",
		Author:              &PullRequestAuthor{Login: "octocat", Name: "Octo Cat", IsBot: false},
		State:               "OPEN",
		IsDraft:             false,
		CreatedAt:           "2026-04-18T10:00:00Z",
		UpdatedAt:           "2026-04-18T12:30:00Z",
		Labels:              []PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
		Assignees:           []PullRequestAuthor{{Login: "assignee-one", Name: "Assignee One", IsBot: false}, {Login: "assignee-two", Name: "Assignee Two", IsBot: false}},
		ReviewRequests:      []PullRequestReviewRequest{{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested", Name: "Reviewer Requested"}}, {RequestedReviewer: PullRequestRequestedReviewer{TypeName: "Team", Name: "Platform", Slug: "acme/platform"}}},
		BaseRefName:         "main",
		HeadRefName:         "feature/detail",
		MergeStateStatus:    "CLEAN",
		Mergeable:           "MERGEABLE",
		IsMergeQueueEnabled: true,
		IsInMergeQueue:      true,
		MergeQueueEntry: &PullRequestMergeQueueEntry{
			ID:                   "MQE_1",
			State:                "QUEUED",
			Position:             3,
			EstimatedTimeToMerge: 11,
		},
		AutoMergeRequest:  &PullRequestAutoMergeRequest{EnabledAt: "2026-04-18T12:45:00Z"},
		OutOfDateWithBase: true,
		Comments: []PullRequestComment{{
			Author:    &PullRequestCommentAuthor{Login: "reviewer"},
			Body:      "Looks good to me",
			CreatedAt: "2026-04-18T13:00:00Z",
			URL:       "https://github.com/acme/widgets/pull/42#issuecomment-1",
		}},
		Commits: []PullRequestCommit{{
			OID:             "e9a3253762e768badaa1d4a5b3d267416d1e42f4",
			MessageHeadline: "reintroduce interactive gh pr",
			MessageBody:     "this commit adds gh pr back",
			AuthoredDate:    "2019-10-04T15:23:39Z",
			CommittedDate:   "2019-10-04T15:57:48Z",
			Authors: []PullRequestCommitAuthor{{
				Email: "vilmibm@github.com",
				Login: "vilmibm",
				Name:  "nate smith",
			}},
		}},
		InlineComments: []PullRequestInlineComment{{
			ID:           "PRRC_1",
			Author:       &PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:         "Please keep the blank line.",
			CreatedAt:    "2026-04-18T14:00:00Z",
			URL:          "https://github.com/acme/widgets/pull/42#discussion_r1",
			Path:         "internal/tui/render.go",
			Line:         252,
			OriginalLine: 252,
			Side:         "RIGHT",
			StartSide:    "RIGHT",
			SubjectType:  "LINE",
			DiffHunk:     "@@ -250,3 +250,4 @@\n header := renderPullRequestDetailHeader(*row.Summary, result.detail)\n content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)\n-if program.activeDetailTab == CommentsDetailTab {\n+if program.activeDetailTab == CommentsDetailTab {\n  content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)",
		}},
		InlineCommentThreads: []PullRequestReviewThread{{
			ID:         "thread-1",
			IsResolved: false,
			Path:       "internal/tui/render.go",
			Line:       252,
			DiffSide:   "RIGHT",
			Comments: []PullRequestComment{{
				Author:    &PullRequestCommentAuthor{Login: "reviewer-inline"},
				Body:      "Please keep the blank line.",
				CreatedAt: "2026-04-18T14:00:00Z",
				URL:       "https://github.com/acme/widgets/pull/42#discussion_r1",
				DiffHunk:  "@@ -250,3 +250,4 @@\n header := renderPullRequestDetailHeader(*row.Summary, result.detail)\n content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)\n-if program.activeDetailTab == CommentsDetailTab {\n+if program.activeDetailTab == CommentsDetailTab {\n  content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)",
			}},
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
			Link:         "https://github.com/acme/widgets/actions/runs/1",
		}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected detail %+v, actual %+v", expected, actual)
	}
}

func TestGetPullRequestDetail_GivenMissingOptionalFields_WhenFetching_ThenItNormalizesTheResponse(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"  Ship it  ","number":7,"url":"  https://github.com/acme/widgets/pull/7  ","body":"  body  ","author":null,"state":"  OPEN  ","createdAt":" 2026-04-18T10:00:00Z ","updatedAt":" 2026-04-18T12:30:00Z ","labels":[{"name":"  needs-review  "}],"assignees":[{"login":"  assignee-one  ","name":"  Assignee One  "}],"reviewRequests":[{"__typename":" User ","login":" reviewer-one ","name":" Reviewer One "},{"__typename":" Team ","name":" Platform ","slug":" acme/platform "}],"baseRefName":"  main  ","headRefName":"  branch  ","mergeStateStatus":"  BLOCKED  ","mergeable":"  UNKNOWN  ","comments":[{"author":null,"body":"  first  ","createdAt":" 2026-04-18T13:00:00Z ","url":"  https://example.com/comment  "}],"commits":[{"oid":"  abcdef1234567890  ","messageHeadline":"  Trim me  ","messageBody":"  body  ","authoredDate":" 2026-04-18T14:00:00Z ","committedDate":" 2026-04-18T14:05:00Z ","authors":[{"email":"  dev@example.com  ","login":" reviewer-one ","name":" Reviewer One "}]}],"additions":1,"deletions":2,"changedFiles":3,"statusCheckRollup":[{"__typename":"  CheckRun  ","name":"  lint  ","status":"  COMPLETED  ","conclusion":"  FAILURE  ","workflowName":"  CI  "}]}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"isMergeQueueEnabled":false,"isInMergeQueue":false,"mergeQueueEntry":null}}}}`)},
			{stdout: []byte(`{"status":"behind","ahead_by":0,"behind_by":1,"total_commits":0}`)},
			{stdout: []byte(`[]`)},
			{stdout: []byte(`[[{"node_id":"  PRRC_kw123  ","user":{"login":"  reviewer-inline  "},"body":"  inline body  ","created_at":" 2026-04-18T14:00:00Z ","html_url":"  https://example.com/discussion  ","path":"  internal/tui/pull_request_detail.go  ","line":19,"original_line":21,"side":"  LEFT  ","start_side":"  LEFT  ","subject_type":"  LINE  ","diff_hunk":"  @@ -19,1 +21,1 @@\n-old\n+new  "}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"  thread-1  ","isResolved":true,"isOutdated":false,"path":"  internal/tui/pull_request_detail.go  ","originalLine":21,"diffSide":"  LEFT  ","comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"author":{"login":"  reviewer-inline  "},"body":"  inline body  ","createdAt":" 2026-04-18T14:00:00Z ","url":"  https://example.com/discussion  ","diffHunk":"  @@ -19,1 +21,1 @@\n-old\n+new  "}]}}]}}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reactionGroups":[],"comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
			{stdout: []byte(`{"data":{"nodes":[{"id":"PRRC_kw123","reactionGroups":[]}]}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 7)

	then_noError(t, actualErr)
	if actual.Title != "Ship it" {
		t.Fatalf("expected title %q, actual %q", "Ship it", actual.Title)
	}
	if actual.Author != nil {
		t.Fatalf("expected nil author, actual %+v", actual.Author)
	}
	if len(actual.Labels) != 1 || actual.Labels[0].Name != "needs-review" {
		t.Fatalf("expected normalized labels, actual %+v", actual.Labels)
	}
	if len(actual.Assignees) != 1 || actual.Assignees[0].Login != "assignee-one" || actual.Assignees[0].Name != "Assignee One" {
		t.Fatalf("expected normalized assignees, actual %+v", actual.Assignees)
	}
	if len(actual.ReviewRequests) != 2 || actual.ReviewRequests[0].RequestedReviewer.TypeName != "User" || actual.ReviewRequests[0].RequestedReviewer.Login != "reviewer-one" || actual.ReviewRequests[0].RequestedReviewer.Name != "Reviewer One" || actual.ReviewRequests[1].RequestedReviewer.TypeName != "Team" || actual.ReviewRequests[1].RequestedReviewer.Name != "Platform" || actual.ReviewRequests[1].RequestedReviewer.Slug != "acme/platform" || actual.ReviewRequests[1].RequestedReviewer.Organization != nil {
		t.Fatalf("expected normalized review requests, actual %+v", actual.ReviewRequests)
	}
	if len(actual.Comments) != 1 || actual.Comments[0].Author != nil || actual.Comments[0].Body != "first" {
		t.Fatalf("expected normalized comments, actual %+v", actual.Comments)
	}
	if len(actual.Commits) != 1 {
		t.Fatalf("expected 1 normalized commit, actual %d", len(actual.Commits))
	}
	if actual.Commits[0].OID != "abcdef1234567890" || actual.Commits[0].MessageHeadline != "Trim me" || actual.Commits[0].MessageBody != "body" || actual.Commits[0].AuthoredDate != "2026-04-18T14:00:00Z" || actual.Commits[0].CommittedDate != "2026-04-18T14:05:00Z" || len(actual.Commits[0].Authors) != 1 || actual.Commits[0].Authors[0].Email != "dev@example.com" || actual.Commits[0].Authors[0].Login != "reviewer-one" || actual.Commits[0].Authors[0].Name != "Reviewer One" {
		t.Fatalf("expected normalized commits, actual %+v", actual.Commits)
	}
	if len(actual.InlineComments) != 1 {
		t.Fatalf("expected 1 inline comment, actual %d", len(actual.InlineComments))
	}
	inlineComment := actual.InlineComments[0]
	if inlineComment.Author == nil || inlineComment.Author.Login != "reviewer-inline" || inlineComment.Body != "inline body" || inlineComment.URL != "https://example.com/discussion" || inlineComment.Path != "internal/tui/pull_request_detail.go" || inlineComment.DiffHunk != "@@ -19,1 +21,1 @@\n-old\n+new" || inlineComment.Side != "LEFT" || inlineComment.StartSide != "LEFT" || inlineComment.SubjectType != "LINE" {
		t.Fatalf("expected normalized inline comment, actual %+v", inlineComment)
	}
	if len(actual.InlineCommentThreads) != 1 {
		t.Fatalf("expected 1 inline thread, actual %d", len(actual.InlineCommentThreads))
	}
	inlineThread := actual.InlineCommentThreads[0]
	if inlineThread.ID != "thread-1" || !inlineThread.IsResolved || inlineThread.Path != "internal/tui/pull_request_detail.go" || inlineThread.OriginalLine != 21 || inlineThread.DiffSide != "LEFT" || len(inlineThread.Comments) != 1 || inlineThread.Comments[0].DiffHunk != "@@ -19,1 +21,1 @@\n-old\n+new" {
		t.Fatalf("expected normalized inline thread, actual %+v", inlineThread)
	}
	if !actual.OutOfDateWithBase {
		t.Fatal("expected the detail to be marked out of date with the base branch")
	}
	if len(actual.StatusCheckRollup) != 1 {
		t.Fatalf("expected 1 status check, actual %d", len(actual.StatusCheckRollup))
	}
	check := actual.StatusCheckRollup[0]
	if check.TypeName != "CheckRun" || check.Name != "lint" || check.Status != "COMPLETED" || check.Conclusion != "FAILURE" || check.WorkflowName != "CI" {
		t.Fatalf("expected normalized status check, actual %+v", check)
	}
}

func TestGetPullRequestDetail_GivenApprovalReviews_WhenFetching_ThenItReturnsNormalizedReviewsAndReviewDecision(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"Approvals","number":42,"body":"Body","state":"OPEN","reviewDecision":" APPROVED ","reviews":[{"author":{"login":" reviewer-one "},"state":" APPROVED ","submittedAt":" 2026-04-21T10:00:00Z "},{"author":{"login":" reviewer-two "},"state":" COMMENTED ","submittedAt":" 2026-04-21T11:00:00Z "}]}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"isMergeQueueEnabled":false,"isInMergeQueue":false,"mergeQueueEntry":null}}}}`)},
			{stdout: []byte(`[]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reactionGroups":[],"comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	if actual.ReviewDecision != "APPROVED" {
		t.Fatalf("expected review decision %q, actual %q", "APPROVED", actual.ReviewDecision)
	}
	if len(actual.Reviews) != 2 {
		t.Fatalf("expected 2 normalized reviews, actual %d", len(actual.Reviews))
	}
	if actual.Reviews[0].Author == nil || actual.Reviews[0].Author.Login != "reviewer-one" || actual.Reviews[0].State != "APPROVED" || actual.Reviews[0].SubmittedAt != "2026-04-21T10:00:00Z" {
		t.Fatalf("expected the first normalized review, actual %+v", actual.Reviews[0])
	}
	if actual.Reviews[1].Author == nil || actual.Reviews[1].Author.Login != "reviewer-two" || actual.Reviews[1].State != "COMMENTED" || actual.Reviews[1].SubmittedAt != "2026-04-21T11:00:00Z" {
		t.Fatalf("expected the second normalized review, actual %+v", actual.Reviews[1])
	}
}

func TestGetPullRequestDetail_GivenSubmittedReviewBody_WhenFetching_ThenItPreservesTheNormalizedReviewText(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"Review comments","number":42,"body":"Body","state":"OPEN","reviews":[{"id":" PRR_1 ","author":{"login":" reviewer-one "},"body":"  looks good but I think you missed RecommendedContentProvider  ","state":" COMMENTED ","submittedAt":" 2026-06-15T06:54:59Z "}]}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"isMergeQueueEnabled":false,"isInMergeQueue":false,"mergeQueueEntry":null}}}}`)},
			{stdout: []byte(`[]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reactionGroups":[],"comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	if len(actual.Reviews) != 1 {
		t.Fatalf("expected 1 normalized review, actual %d", len(actual.Reviews))
	}
	if actualReviewID := actual.Reviews[0].ID; actualReviewID != "PRR_1" {
		t.Fatalf("expected normalized review id %q, actual %q", "PRR_1", actualReviewID)
	}
	if actualBody := actual.Reviews[0].Body; actualBody != "looks good but I think you missed RecommendedContentProvider" {
		t.Fatalf("expected normalized review body %q, actual %q", "looks good but I think you missed RecommendedContentProvider", actualBody)
	}
}

func TestGetPullRequestDetail_GivenPendingInlineReviewComments_WhenFetching_ThenItPreservesTheThreadCommentState(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"Pending detail","number":42,"body":"Body","state":"OPEN"}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"isMergeQueueEnabled":false,"isInMergeQueue":false,"mergeQueueEntry":null}}}}`)},
			{stdout: []byte(`[]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"thread-1","isResolved":false,"isOutdated":false,"path":"internal/tui/render.go","line":12,"diffSide":"RIGHT","comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"PRRC_1","state":" PENDING ","author":{"login":"reviewer-inline"},"body":"Draft feedback","createdAt":"2026-04-18T14:00:00Z"}]}}]}}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reactionGroups":[],"comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	if len(actual.InlineCommentThreads) != 1 {
		t.Fatalf("expected 1 inline thread, actual %d", len(actual.InlineCommentThreads))
	}
	if len(actual.InlineCommentThreads[0].Comments) != 1 {
		t.Fatalf("expected 1 inline thread comment, actual %+v", actual.InlineCommentThreads[0].Comments)
	}
	if actual.InlineCommentThreads[0].Comments[0].State != "PENDING" {
		t.Fatalf("expected inline thread comment state %q, actual %+v", "PENDING", actual.InlineCommentThreads[0].Comments[0])
	}
}

func TestGetPullRequestDetail_GivenReactionTargets_WhenFetching_ThenItLoadsStableIDsAndReactionGroups(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"Reactions","number":42,"body":"Body","state":"OPEN","comments":[{"id":"IC_kwDOA","author":{"login":"reviewer"},"body":"Looks good","createdAt":"2026-04-18T13:00:00Z","url":"https://github.com/acme/widgets/pull/42#issuecomment-1"}]}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"isMergeQueueEnabled":false,"isInMergeQueue":false,"mergeQueueEntry":null}}}}`)},
			{stdout: []byte(`[[{"node_id":"PRRC_kwDOA","user":{"login":"reviewer-inline"},"body":"Nit: keep spacing","created_at":"2026-04-18T14:00:00Z","html_url":"https://github.com/acme/widgets/pull/42#discussion_r1","path":"internal/tui/render.go","line":12,"original_line":12,"side":"RIGHT","start_side":"RIGHT","subject_type":"LINE","diff_hunk":"@@ -12,1 +12,1 @@\n-old\n+new"}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"id":"PR_kwDOA","reactionGroups":[{"content":"THUMBS_UP","viewerHasReacted":true,"users":{"totalCount":2}},{"content":"HOORAY","viewerHasReacted":false,"users":{"totalCount":1}}],"comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"IC_kwDOA","author":{"login":"reviewer"},"body":"Looks good","createdAt":"2026-04-18T13:00:00Z","url":"https://github.com/acme/widgets/pull/42#issuecomment-1","viewerDidAuthor":false,"reactionGroups":[{"content":"EYES","viewerHasReacted":false,"users":{"totalCount":3}}]}]}}}}}`)},
			{stdout: []byte(`{"data":{"nodes":[{"id":"PRRC_kwDOA","reactionGroups":[{"content":"HEART","viewerHasReacted":true,"users":{"totalCount":1}},{"content":"LAUGH","viewerHasReacted":false,"users":{"totalCount":2}}]}]}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"pr", "view", "42", "-R", "acme/widgets", "--json", pullRequestDetailJSONFields}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestMergeQueueQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42/comments?per_page=100", "--paginate", "--slurp"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReactionTargetsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewCommentReactionGroupsQuery, "-F", "ids[]=PRRC_kwDOA"}},
	})
	if actual.ID != "PR_kwDOA" {
		t.Fatalf("expected pull request reaction id %q, actual %q", "PR_kwDOA", actual.ID)
	}
	expectedPullRequestReactions := []ReactionGroup{
		{Content: ReactionContentThumbsUp, TotalCount: 2, ViewerHasReacted: true},
		{Content: ReactionContentHooray, TotalCount: 1},
	}
	if !reflect.DeepEqual(actual.ReactionGroups, expectedPullRequestReactions) {
		t.Fatalf("expected pull request reactions %+v, actual %+v", expectedPullRequestReactions, actual.ReactionGroups)
	}
	if len(actual.Comments) != 1 {
		t.Fatalf("expected 1 pull request comment, actual %+v", actual.Comments)
	}
	expectedCommentReactions := []ReactionGroup{{Content: ReactionContentEyes, TotalCount: 3}}
	if !reflect.DeepEqual(actual.Comments[0].ReactionGroups, expectedCommentReactions) {
		t.Fatalf("expected pull request comment reactions %+v, actual %+v", expectedCommentReactions, actual.Comments[0].ReactionGroups)
	}
	if len(actual.InlineComments) != 1 {
		t.Fatalf("expected 1 inline comment, actual %+v", actual.InlineComments)
	}
	expectedInlineReactions := []ReactionGroup{
		{Content: ReactionContentLaugh, TotalCount: 2},
		{Content: ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true},
	}
	if !reflect.DeepEqual(actual.InlineComments[0].ReactionGroups, expectedInlineReactions) {
		t.Fatalf("expected inline comment reactions %+v, actual %+v", expectedInlineReactions, actual.InlineComments[0].ReactionGroups)
	}
}

func TestGetPullRequestDetail_GivenInvalidJSON_WhenFetching_ThenReturnsAnInvalidResponseError(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"title":`)}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	_, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	if !errors.Is(actualErr, ErrInvalidPullRequestDetailResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestDetailResponse, actualErr)
	}
}

func TestGetPullRequestDetail_GivenCommandFailure_WhenFetching_ThenReturnsTheViewError(t *testing.T) {
	runner := &fakeRunner{
		stderr: []byte("boom"),
		err:    errors.New("exit status 1"),
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	_, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh pr view") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr view", actualErr)
	}
}
