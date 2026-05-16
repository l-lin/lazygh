package tui

import (
	"fmt"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestFormatPullRequestSearchCommand_GivenExistingJSONFlag_WhenFormatting_ThenItUsesTheExpectedSearchPayload(t *testing.T) {
	actual := formatPullRequestSearchCommand([]string{"pr", "list", "--search", "author:@me", "--json", "title"})

	expected := appconfig.FormatGHCommand([]string{"pr", "list", "--search", "author:@me", "--json", "title,number,repository,url,body,state,isDraft,updatedAt,id"})
	if actual != expected {
		t.Fatalf("expected formatted command %q, actual %q", expected, actual)
	}
}

func TestFormatAssignableUsersCommand_GivenRepository_WhenFormatting_ThenItUsesThePaginatedRESTCommand(t *testing.T) {
	actual := formatAssignableUsersCommand("acme/widgets")

	expected := appconfig.FormatGHCommand([]string{"api", "repos/acme/widgets/assignees?per_page=100", "--paginate", "--slurp"})
	if actual != expected {
		t.Fatalf("expected formatted command %q, actual %q", expected, actual)
	}
}

func TestFormatAssigneeSearchCommand_GivenRepositoryAndQuery_WhenFormatting_ThenItUsesTheGraphQLDisplayCommand(t *testing.T) {
	actual := formatAssigneeSearchCommand("acme/widgets", "bob")

	expected := appconfig.FormatGHCommand([]string{"api", "graphql", "-F", "owner=acme", "-F", "name=widgets", "-F", "first=20", "-F", "search=bob"})
	if actual != expected {
		t.Fatalf("expected formatted command %q, actual %q", expected, actual)
	}
}

func TestFormatPullRequestBuildRunCommand_GivenAttemptURL_WhenFormatting_ThenItIncludesTheAttemptFlag(t *testing.T) {
	actual := formatPullRequestBuildRunCommand("acme/widgets", githubdomain.PullRequestStatusCheck{Link: "https://github.com/acme/widgets/actions/runs/42/attempts/3"})

	expected := appconfig.FormatGHCommand([]string{"run", "view", "42", "-R", "acme/widgets", "--attempt", "3", "--verbose"})
	if actual != expected {
		t.Fatalf("expected formatted command %q, actual %q", expected, actual)
	}
}

func TestIsProviderUnauthenticatedError_GivenWrappedDomainError_WhenChecking_ThenItMatches(t *testing.T) {
	actual := isProviderUnauthenticatedError(fmt.Errorf("wrap: %w", githubdomain.ErrUnauthenticated))

	if !actual {
		t.Fatal("expected the wrapped unauthenticated provider error to match")
	}
}

func TestIsProviderUnavailableError_GivenWrappedDomainError_WhenChecking_ThenItMatches(t *testing.T) {
	actual := isProviderUnavailableError(fmt.Errorf("wrap: %w", githubdomain.ErrUnavailable))

	if !actual {
		t.Fatal("expected the wrapped unavailable provider error to match")
	}
}

func TestIsProviderEmptyConnectedUserError_GivenWrappedDomainError_WhenChecking_ThenItMatches(t *testing.T) {
	actual := isProviderEmptyConnectedUserError(fmt.Errorf("wrap: %w", githubdomain.ErrEmptyConnectedUser))

	if !actual {
		t.Fatal("expected the wrapped empty connected user error to match")
	}
}
