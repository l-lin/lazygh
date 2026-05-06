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
			{stdout: []byte(`{"title":"Add a real detail pane","number":42,"url":"https://github.com/acme/widgets/pull/42","body":"## Summary\n\n- render markdown\n- show comments","author":{"login":"octocat","name":"Octo Cat","is_bot":false},"state":"OPEN","isDraft":false,"createdAt":"2026-04-18T10:00:00Z","updatedAt":"2026-04-18T12:30:00Z","labels":[{"name":"bug"},{"name":"backend"}],"assignees":[{"login":"assignee-one","name":"Assignee One","is_bot":false},{"login":"assignee-two","name":"Assignee Two","is_bot":false}],"reviewRequests":[{"requestedReviewer":{"__typename":"User","login":"reviewer-requested"}},{"requestedReviewer":{"__typename":"Team","name":"Platform","slug":"platform","organization":{"login":"acme"}}}],"baseRefName":"main","headRefName":"feature/detail","mergeStateStatus":"CLEAN","mergeable":"MERGEABLE","comments":[{"author":{"login":"reviewer"},"body":"Looks good to me","createdAt":"2026-04-18T13:00:00Z","url":"https://github.com/acme/widgets/pull/42#issuecomment-1"}],"additions":12,"deletions":3,"changedFiles":5,"statusCheckRollup":[{"__typename":"CheckRun","name":"lint","status":"COMPLETED","conclusion":"SUCCESS","workflowName":"CI"}]}`)},
			{stdout: []byte(`[[{"user":{"login":"reviewer-inline"},"body":"Please keep the blank line.","created_at":"2026-04-18T14:00:00Z","html_url":"https://github.com/acme/widgets/pull/42#discussion_r1","path":"internal/tui/render.go","line":252,"original_line":252,"side":"RIGHT","start_side":"RIGHT","subject_type":"LINE","diff_hunk":"@@ -250,3 +250,4 @@\n header := renderPullRequestDetailHeader(*row.Summary, result.detail)\n content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)\n-if program.activeDetailTab == CommentsDetailTab {\n+if program.activeDetailTab == CommentsDetailTab {\n  content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)"}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"thread-1","isResolved":false,"isOutdated":false,"path":"internal/tui/render.go","line":252,"diffSide":"RIGHT","comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"author":{"login":"reviewer-inline"},"body":"Please keep the blank line.","createdAt":"2026-04-18T14:00:00Z","url":"https://github.com/acme/widgets/pull/42#discussion_r1","diffHunk":"@@ -250,3 +250,4 @@\n header := renderPullRequestDetailHeader(*row.Summary, result.detail)\n content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)\n-if program.activeDetailTab == CommentsDetailTab {\n+if program.activeDetailTab == CommentsDetailTab {\n  content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)"}]}}]}}}}}`)},
		},
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"pr", "view", "42", "-R", "acme/widgets", "--json", "title,number,url,body,author,state,isDraft,createdAt,updatedAt,labels,assignees,reviewRequests,baseRefName,headRefName,mergeStateStatus,mergeable,comments,reviews,additions,deletions,changedFiles,statusCheckRollup"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42/comments?per_page=100", "--paginate", "--slurp"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
	})

	expected := PullRequestDetail{
		Title:            "Add a real detail pane",
		Number:           42,
		URL:              "https://github.com/acme/widgets/pull/42",
		Body:             "## Summary\n\n- render markdown\n- show comments",
		Author:           &PullRequestAuthor{Login: "octocat", Name: "Octo Cat", IsBot: false},
		State:            "OPEN",
		IsDraft:          false,
		CreatedAt:        "2026-04-18T10:00:00Z",
		UpdatedAt:        "2026-04-18T12:30:00Z",
		Labels:           []PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
		Assignees:        []PullRequestAuthor{{Login: "assignee-one", Name: "Assignee One", IsBot: false}, {Login: "assignee-two", Name: "Assignee Two", IsBot: false}},
		ReviewRequests:   []PullRequestReviewRequest{{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}, {RequestedReviewer: PullRequestRequestedReviewer{TypeName: "Team", Name: "Platform", Slug: "platform", Organization: &PullRequestReviewRequestOrganization{Login: "acme"}}}},
		BaseRefName:      "main",
		HeadRefName:      "feature/detail",
		MergeStateStatus: "CLEAN",
		Mergeable:        "MERGEABLE",
		Comments: []PullRequestComment{{
			Author:    &PullRequestCommentAuthor{Login: "reviewer"},
			Body:      "Looks good to me",
			CreatedAt: "2026-04-18T13:00:00Z",
			URL:       "https://github.com/acme/widgets/pull/42#issuecomment-1",
		}},
		InlineComments: []PullRequestInlineComment{{
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
		}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected detail %+v, actual %+v", expected, actual)
	}
}

func TestGetPullRequestDetail_GivenMissingOptionalFields_WhenFetching_ThenItNormalizesTheResponse(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"  Ship it  ","number":7,"url":"  https://github.com/acme/widgets/pull/7  ","body":"  body  ","author":null,"state":"  OPEN  ","createdAt":" 2026-04-18T10:00:00Z ","updatedAt":" 2026-04-18T12:30:00Z ","labels":[{"name":"  needs-review  "}],"assignees":[{"login":"  assignee-one  ","name":"  Assignee One  "}],"reviewRequests":[{"requestedReviewer":{"__typename":" User ","login":" reviewer-one "}},{"requestedReviewer":{"__typename":" Team ","name":" Platform ","slug":" platform ","organization":{"login":" acme "}}}],"baseRefName":"  main  ","headRefName":"  branch  ","mergeStateStatus":"  BLOCKED  ","mergeable":"  UNKNOWN  ","comments":[{"author":null,"body":"  first  ","createdAt":" 2026-04-18T13:00:00Z ","url":"  https://example.com/comment  "}],"additions":1,"deletions":2,"changedFiles":3,"statusCheckRollup":[{"__typename":"  CheckRun  ","name":"  lint  ","status":"  COMPLETED  ","conclusion":"  FAILURE  ","workflowName":"  CI  "}]}`)},
			{stdout: []byte(`[[{"user":{"login":"  reviewer-inline  "},"body":"  inline body  ","created_at":" 2026-04-18T14:00:00Z ","html_url":"  https://example.com/discussion  ","path":"  internal/tui/pull_request_detail.go  ","line":19,"original_line":21,"side":"  LEFT  ","start_side":"  LEFT  ","subject_type":"  LINE  ","diff_hunk":"  @@ -19,1 +21,1 @@\n-old\n+new  "}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"  thread-1  ","isResolved":true,"isOutdated":false,"path":"  internal/tui/pull_request_detail.go  ","originalLine":21,"diffSide":"  LEFT  ","comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"author":{"login":"  reviewer-inline  "},"body":"  inline body  ","createdAt":" 2026-04-18T14:00:00Z ","url":"  https://example.com/discussion  ","diffHunk":"  @@ -19,1 +21,1 @@\n-old\n+new  "}]}}]}}}}}`)},
		},
	}
	subject := NewClientWithRunner(runner)

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
	if len(actual.ReviewRequests) != 2 || actual.ReviewRequests[0].RequestedReviewer.TypeName != "User" || actual.ReviewRequests[0].RequestedReviewer.Login != "reviewer-one" || actual.ReviewRequests[1].RequestedReviewer.TypeName != "Team" || actual.ReviewRequests[1].RequestedReviewer.Slug != "platform" || actual.ReviewRequests[1].RequestedReviewer.Organization == nil || actual.ReviewRequests[1].RequestedReviewer.Organization.Login != "acme" {
		t.Fatalf("expected normalized review requests, actual %+v", actual.ReviewRequests)
	}
	if len(actual.Comments) != 1 || actual.Comments[0].Author != nil || actual.Comments[0].Body != "first" {
		t.Fatalf("expected normalized comments, actual %+v", actual.Comments)
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
	if len(actual.StatusCheckRollup) != 1 {
		t.Fatalf("expected 1 status check, actual %d", len(actual.StatusCheckRollup))
	}
	check := actual.StatusCheckRollup[0]
	if check.TypeName != "CheckRun" || check.Name != "lint" || check.Status != "COMPLETED" || check.Conclusion != "FAILURE" || check.WorkflowName != "CI" {
		t.Fatalf("expected normalized status check, actual %+v", check)
	}
}

func TestGetPullRequestDetail_GivenApprovalReviews_WhenFetching_ThenItReturnsNormalizedReviews(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"title":"Approvals","number":42,"body":"Body","state":"OPEN","reviews":[{"author":{"login":" reviewer-one "},"state":" APPROVED ","submittedAt":" 2026-04-21T10:00:00Z "},{"author":{"login":" reviewer-two "},"state":" COMMENTED ","submittedAt":" 2026-04-21T11:00:00Z "}]}`)},
			{stdout: []byte(`[]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
		},
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
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

func TestGetPullRequestDetail_GivenInvalidJSON_WhenFetching_ThenReturnsAnInvalidResponseError(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"title":`)}
	subject := NewClientWithRunner(runner)

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
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh pr view") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr view", actualErr)
	}
}
