package githubcli

import (
	"reflect"
	"testing"
)

func TestGetPullRequestFileTeamOwners_GivenBaseBranchCodeowners_WhenFetching_ThenItMatchesChangedFilesToTeams(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"baseRefName":"main"}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"dotgithub":{"text":"* @acme/Core\n!negated @acme/Nope\n/docs/ @acme/P3C\nREADME.md @octocat\n"},"root":null,"docs":null}}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestFileTeamOwners("acme/widgets", 42, []string{" docs/design.md ", "internal/tui/render.go", "README.md", "docs/design.md"})

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestBaseRefNameQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestCodeownersBlobQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "dotgithubExpression=refs/heads/main:.github/CODEOWNERS", "-F", "rootExpression=refs/heads/main:CODEOWNERS", "-F", "docsExpression=refs/heads/main:docs/CODEOWNERS"}},
	})

	expected := map[string][]string{
		"docs/design.md":         {"P3C"},
		"internal/tui/render.go": {"Core"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected team owners %+v, actual %+v", expected, actual)
	}
}

func TestGetPullRequestFileTeamOwners_GivenNoCodeownersFile_WhenFetching_ThenItReturnsNoTeamOwners(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"baseRefName":"main"}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"dotgithub":null,"root":null,"docs":null}}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestFileTeamOwners("acme/widgets", 42, []string{"docs/design.md"})

	then_noError(t, actualErr)
	if len(actual) != 0 {
		t.Fatalf("expected no team owners, actual %+v", actual)
	}
}
