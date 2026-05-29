package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestPastedPullRequestTabState_GivenExistingPullRequests_WhenAddingAnother_ThenItPrependsAndDeduplicatesWithoutMutatingTheOriginal(t *testing.T) {
	first := githubdomain.PullRequest{Title: "First", Number: 1, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/1"}
	second := githubdomain.PullRequest{Title: "Second", Number: 2, Repository: githubdomain.Repository{NameWithOwner: "acme/rocket"}, URL: "https://github.com/acme/rocket/pull/2"}
	updatedFirst := githubdomain.PullRequest{Title: "First updated", Number: 1, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/1"}
	subject := pastedPullRequestTabState{pullRequests: []githubdomain.PullRequest{first, second}}

	actual := subject.withPullRequestAdded(updatedFirst)

	if len(actual.pullRequests) != 2 {
		t.Fatalf("expected two pasted pull requests, actual %+v", actual.pullRequests)
	}
	if actual.pullRequests[0].Title != "First updated" || actual.pullRequests[0].Number != 1 {
		t.Fatalf("expected the updated pasted pull request to move to the front, actual %+v", actual.pullRequests[0])
	}
	if actual.pullRequests[1].Title != "Second" || actual.pullRequests[1].Number != 2 {
		t.Fatalf("expected the unrelated pasted pull request to stay second, actual %+v", actual.pullRequests[1])
	}
	if subject.pullRequests[0].Title != "First" {
		t.Fatalf("expected the original first pasted pull request %q, actual %q", "First", subject.pullRequests[0].Title)
	}
	if subject.pullRequests[1].Title != "Second" {
		t.Fatalf("expected the original second pasted pull request %q, actual %q", "Second", subject.pullRequests[1].Title)
	}
}
