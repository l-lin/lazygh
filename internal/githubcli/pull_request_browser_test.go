package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestOpenPullRequestInBrowser_GivenRepositoryAndNumber_WhenOpening_ThenItRunsGhPrViewWithWeb(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.OpenPullRequestInBrowser("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "view", "42", "-R", "acme/widgets", "--web"})
}

func TestOpenPullRequestInBrowser_GivenMissingPullRequestIdentity_WhenOpening_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.OpenPullRequestInBrowser(" ", 0)

	if !errors.Is(actualErr, ErrMissingPullRequestIdentity) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestIdentity, actualErr)
	}
}

func TestOpenPullRequestInBrowser_GivenCommandFailure_WhenOpening_ThenItReturnsTheGhPrViewError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewClientWithRunner(runner)

	actualErr := subject.OpenPullRequestInBrowser("acme/widgets", 42)

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh pr view") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr view", actualErr)
	}
}
