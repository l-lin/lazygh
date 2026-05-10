package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestParsePullRequestURL_GivenCanonicalGitHubPullRequestURL_WhenParsing_ThenItReturnsTheRepositoryNumberAndCanonicalURL(t *testing.T) {
	actual, actualErr := ParsePullRequestURL(" https://github.com/acme/widgets/pull/42 ")

	then_noError(t, actualErr)
	expected := PullRequest{
		Number:     42,
		Repository: Repository{Name: "widgets", NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected pull request %+v, actual %+v", expected, actual)
	}
}

func TestParsePullRequestURL_GivenGitHubFilesTabURL_WhenParsing_ThenItKeepsThePullRequestIdentityAndCanonicalURL(t *testing.T) {
	actual, actualErr := ParsePullRequestURL("https://github.com/acme/widgets/pull/77/files#diff-1")

	then_noError(t, actualErr)
	if actual.Repository.NameWithOwner != "acme/widgets" {
		t.Fatalf("expected repository %q, actual %q", "acme/widgets", actual.Repository.NameWithOwner)
	}
	if actual.Number != 77 {
		t.Fatalf("expected pull request number %d, actual %d", 77, actual.Number)
	}
	if actual.URL != "https://github.com/acme/widgets/pull/77" {
		t.Fatalf("expected canonical url %q, actual %q", "https://github.com/acme/widgets/pull/77", actual.URL)
	}
}

func TestParsePullRequestURL_GivenGitHubPullRequestsPath_WhenParsing_ThenItNormalizesToTheCanonicalPullURL(t *testing.T) {
	actual, actualErr := ParsePullRequestURL("https://github.com/acme/widgets/pulls/77")

	then_noError(t, actualErr)
	if actual.Repository.NameWithOwner != "acme/widgets" {
		t.Fatalf("expected repository %q, actual %q", "acme/widgets", actual.Repository.NameWithOwner)
	}
	if actual.Number != 77 {
		t.Fatalf("expected pull request number %d, actual %d", 77, actual.Number)
	}
	if actual.URL != "https://github.com/acme/widgets/pull/77" {
		t.Fatalf("expected canonical url %q, actual %q", "https://github.com/acme/widgets/pull/77", actual.URL)
	}
}

func TestParsePullRequestURL_GivenInvalidGitHubURL_WhenParsing_ThenItReturnsAValidationError(t *testing.T) {
	_, actualErr := ParsePullRequestURL("https://github.com/acme/widgets/issues/42")

	if !errors.Is(actualErr, ErrInvalidPullRequestURL) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestURL, actualErr)
	}
}
