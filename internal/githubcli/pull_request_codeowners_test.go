package githubcli

import (
	"reflect"
	"testing"
)

func TestGetPullRequestFileTeamOwners_GivenBaseBranchCodeowners_WhenFetching_ThenItMatchesChangedFilesToTeams(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"baseRefName":"main","baseRefOid":"base-1"}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"dotgithub":{"text":"* @acme/Core\n!negated @acme/Nope\n/docs/ @acme/P3C\nREADME.md @octocat\n"},"root":null,"docs":null}}}`)},
		},
	}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestFileTeamOwners("acme/widgets", 42, []string{" docs/design.md ", "internal/tui/render.go", "README.md", "docs/design.md"})

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestBaseRevisionQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestCodeownersBlobQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "dotgithubExpression=base-1:.github/CODEOWNERS", "-F", "rootExpression=base-1:CODEOWNERS", "-F", "docsExpression=base-1:docs/CODEOWNERS"}},
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
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"baseRefName":"main","baseRefOid":"base-1"}}}}`)},
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
