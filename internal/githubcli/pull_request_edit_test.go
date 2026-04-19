package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestEditPullRequestTitle_GivenRepositoryNumberAndTitle_WhenEditing_ThenItRunsGhPrEditTitle(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.EditPullRequestTitle("acme/widgets", 42, "New title")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "edit", "42", "-R", "acme/widgets", "--title", "New title"})
}

func TestEditPullRequestDescription_GivenRepositoryNumberAndBody_WhenEditing_ThenItRunsGhPrEditBodyFileWithStdin(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)
	body := "Updated body\n\n- more detail"

	actualErr := subject.EditPullRequestDescription("acme/widgets", 42, body)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "edit", "42", "-R", "acme/widgets", "--body-file", "-"})
	then_stdinIs(t, runner, body)
}

func TestEditPullRequestTitle_GivenMissingPullRequestIdentity_WhenEditing_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.EditPullRequestTitle(" ", 0, "Ship it")

	if !errors.Is(actualErr, ErrMissingPullRequestIdentity) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestIdentity, actualErr)
	}
}

func TestEditPullRequestTitle_GivenEmptyTitle_WhenEditing_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.EditPullRequestTitle("acme/widgets", 42, " \n\t ")

	if !errors.Is(actualErr, ErrEmptyPullRequestTitle) {
		t.Fatalf("expected error %v, actual %v", ErrEmptyPullRequestTitle, actualErr)
	}
}

func TestEditPullRequestDescription_GivenCommandFailure_WhenEditing_ThenItReturnsTheGhPrEditError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewClientWithRunner(runner)

	actualErr := subject.EditPullRequestDescription("acme/widgets", 42, "Updated body")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh pr edit") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr edit", actualErr)
	}
}
