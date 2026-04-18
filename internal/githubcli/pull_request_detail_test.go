package githubcli

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestGetPullRequestDetail_GivenValidGhResponse_WhenFetching_ThenReturnsTheRichPullRequestDetail(t *testing.T) {
	runner := &fakeRunner{
		stdout: []byte(`{"title":"Add a real detail pane","number":42,"url":"https://github.com/acme/widgets/pull/42","body":"## Summary\n\n- render markdown\n- show comments","author":{"login":"octocat","name":"Octo Cat","is_bot":false},"state":"OPEN","isDraft":false,"createdAt":"2026-04-18T10:00:00Z","updatedAt":"2026-04-18T12:30:00Z","labels":[{"name":"bug"},{"name":"backend"}],"baseRefName":"main","headRefName":"feature/detail","mergeStateStatus":"CLEAN","mergeable":"MERGEABLE","comments":[{"author":{"login":"reviewer"},"body":"Looks good to me","createdAt":"2026-04-18T13:00:00Z","url":"https://github.com/acme/widgets/pull/42#issuecomment-1"}],"additions":12,"deletions":3,"changedFiles":5,"statusCheckRollup":[{"__typename":"CheckRun","name":"lint","status":"COMPLETED","conclusion":"SUCCESS","workflowName":"CI"}]}`),
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "view", "42", "-R", "acme/widgets", "--json", "title,number,url,body,author,state,isDraft,createdAt,updatedAt,labels,baseRefName,headRefName,mergeStateStatus,mergeable,comments,additions,deletions,changedFiles,statusCheckRollup"})

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
		stdout: []byte(`{"title":"  Ship it  ","number":7,"url":"  https://github.com/acme/widgets/pull/7  ","body":"  body  ","author":null,"state":"  OPEN  ","createdAt":" 2026-04-18T10:00:00Z ","updatedAt":" 2026-04-18T12:30:00Z ","labels":[{"name":"  needs-review  "}],"baseRefName":"  main  ","headRefName":"  branch  ","mergeStateStatus":"  BLOCKED  ","mergeable":"  UNKNOWN  ","comments":[{"author":null,"body":"  first  ","createdAt":" 2026-04-18T13:00:00Z ","url":"  https://example.com/comment  "}],"additions":1,"deletions":2,"changedFiles":3,"statusCheckRollup":[{"__typename":"  CheckRun  ","name":"  lint  ","status":"  COMPLETED  ","conclusion":"  FAILURE  ","workflowName":"  CI  "}]}`),
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
	if len(actual.Comments) != 1 || actual.Comments[0].Author != nil || actual.Comments[0].Body != "first" {
		t.Fatalf("expected normalized comments, actual %+v", actual.Comments)
	}
	if len(actual.StatusCheckRollup) != 1 {
		t.Fatalf("expected 1 status check, actual %d", len(actual.StatusCheckRollup))
	}
	check := actual.StatusCheckRollup[0]
	if check.TypeName != "CheckRun" || check.Name != "lint" || check.Status != "COMPLETED" || check.Conclusion != "FAILURE" || check.WorkflowName != "CI" {
		t.Fatalf("expected normalized status check, actual %+v", check)
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
