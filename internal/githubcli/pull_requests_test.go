package githubcli

import (
	"testing"
)

func TestListPullRequests_GivenExplicitCommandArguments_WhenFetching_ThenItRunsThoseArguments(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[]`)}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"})

	then_noError(t, actualErr)
	if len(actual) != 0 {
		t.Fatalf("expected no pull requests, actual %+v", actual)
	}
	then_commandIs(t, runner, "gh", []string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"})
}
