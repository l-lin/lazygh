package githubcli

import (
	"errors"
	"testing"
)

func TestAddPullRequestReviewThread_GivenMultiLineTarget_WhenSubmitting_ThenItRunsGhApiGraphQLWithTheThreadInput(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"addPullRequestReviewThread":{"thread":{"id":"PRRT_1"}}}}`)}
	subject := NewReviewServiceWithRunner(runner)
	target := PullRequestReviewThreadTarget{
		Path:        "internal/tui/render.go",
		Line:        13,
		Side:        "RIGHT",
		StartLine:   11,
		StartSide:   "RIGHT",
		SubjectType: "LINE",
	}

	actualErr := subject.AddPullRequestReviewThread("PRR_pending", "Line one\nLine two", target)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + addPullRequestReviewThreadMutation,
		"-f",
		"pullRequestReviewId=PRR_pending",
		"-f",
		"path=internal/tui/render.go",
		"-F",
		"line=13",
		"-f",
		"side=RIGHT",
		"-f",
		"body=Line one\nLine two",
		"-F",
		"startLine=11",
		"-f",
		"startSide=RIGHT",
		"-f",
		"subjectType=LINE",
	})
}

func TestAddPullRequestReviewThread_GivenSingleLineTarget_WhenSubmitting_ThenItOmitsTheRangeFields(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"addPullRequestReviewThread":{"thread":{"id":"PRRT_1"}}}}`)}
	subject := NewReviewServiceWithRunner(runner)
	target := PullRequestReviewThreadTarget{
		Path:        "internal/tui/render.go",
		Line:        13,
		Side:        "RIGHT",
		SubjectType: "LINE",
	}

	actualErr := subject.AddPullRequestReviewThread("PRR_pending", "Please add context", target)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + addPullRequestReviewThreadMutation,
		"-f",
		"pullRequestReviewId=PRR_pending",
		"-f",
		"path=internal/tui/render.go",
		"-F",
		"line=13",
		"-f",
		"side=RIGHT",
		"-f",
		"body=Please add context",
		"-f",
		"subjectType=LINE",
	})
}

func TestAddPullRequestReviewThread_GivenGraphQLErrorPayload_WhenSubmitting_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"Pull request review thread line must be part of the diff"}],"data":{"addPullRequestReviewThread":null}}`)}
	subject := NewReviewServiceWithRunner(runner)
	target := PullRequestReviewThreadTarget{Path: "internal/tui/render.go", Line: 13, Side: "RIGHT", SubjectType: "LINE"}

	actualErr := subject.AddPullRequestReviewThread("PRR_pending", "Please add context", target)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "Pull request review thread line must be part of the diff" {
		t.Fatalf("expected GitHub error %q, actual %v", "Pull request review thread line must be part of the diff", actualErr)
	}
}

func TestAddPullRequestReviewThread_GivenAnInvalidTarget_WhenSubmitting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewReviewServiceWithRunner(&fakeRunner{})

	actualErr := subject.AddPullRequestReviewThread(" ", "Please add context", PullRequestReviewThreadTarget{})

	if !errors.Is(actualErr, ErrInvalidPullRequestReviewThreadTarget) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestReviewThreadTarget, actualErr)
	}
}
