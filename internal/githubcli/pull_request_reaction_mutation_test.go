package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestAddReaction_GivenSubjectIDAndContent_WhenAdding_ThenItRunsGhApiGraphQLWithTheAddReactionMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"addReaction":{"reaction":{"content":"THUMBS_UP"},"subject":{"id":"PR_kwDOA"}}}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.AddReaction("PR_kwDOA", ReactionContentThumbsUp)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + addReactionMutation,
		"-f",
		"subjectId=PR_kwDOA",
		"-f",
		"content=THUMBS_UP",
	})
}

func TestAddReaction_GivenInvalidContent_WhenAdding_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.AddReaction("PR_kwDOA", ReactionContent("wat"))

	if !errors.Is(actualErr, ErrInvalidReactionContent) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidReactionContent, actualErr)
	}
}

func TestAddReaction_GivenGraphQLErrorPayload_WhenAdding_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"Reactions are disabled"}],"data":{"addReaction":null}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.AddReaction("PR_kwDOA", ReactionContentThumbsUp)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "Reactions are disabled" {
		t.Fatalf("expected GitHub error %q, actual %v", "Reactions are disabled", actualErr)
	}
}

func TestAddReaction_GivenCommandFailure_WhenAdding_ThenItReturnsTheGhApiGraphQLError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewClientWithRunner(runner)

	actualErr := subject.AddReaction("PR_kwDOA", ReactionContentThumbsUp)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh api graphql") {
		t.Fatalf("expected error to mention %q, actual %v", "gh api graphql", actualErr)
	}
}
