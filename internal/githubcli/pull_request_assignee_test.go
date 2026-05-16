package githubcli

import (
	"reflect"
	"testing"
)

func TestListAssignableUsers_GivenRepository_WhenListing_ThenItLoadsAndNormalizesTheAssignableUsers(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[[{"login":" alice ","name":" Alice ","is_bot":false},{"login":"bob","name":" Bob "}],[{"login":"alice","name":"Duplicate"},{"login":"charlie","name":"  Charlie  ","is_bot":true}]]`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actual, actualErr := subject.ListAssignableUsers("acme/widgets")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "repos/acme/widgets/assignees?per_page=100", "--paginate", "--slurp"})
	expected := []PullRequestAuthor{
		{Login: "alice", Name: "Alice", IsBot: false},
		{Login: "bob", Name: "Bob", IsBot: false},
		{Login: "charlie", Name: "Charlie", IsBot: true},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected assignable users %+v, actual %+v", expected, actual)
	}
}

func TestSearchAssignableUsers_GivenRepositoryAndQuery_WhenSearching_ThenItRunsGraphQLAndNormalizesTheAssignableUsers(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"repository":{"assignableUsers":{"nodes":[{"login":" bob ","name":" Bob ","is_bot":false},{"login":"alice","name":" Alice "},{"login":"bob","name":"Duplicate"}]}}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actual, actualErr := subject.SearchAssignableUsers("acme/widgets", "bob")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=query($owner:String!,$name:String!,$first:Int!,$search:String){repository(owner:$owner,name:$name){assignableUsers(first:$first,query:$search){nodes{login name is_bot:isBot}}}}",
		"-F",
		"owner=acme",
		"-F",
		"name=widgets",
		"-F",
		"first=20",
		"-F",
		"search=bob",
	})
	expected := []PullRequestAuthor{{Login: "bob", Name: "Bob", IsBot: false}, {Login: "alice", Name: "Alice", IsBot: false}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected assignable user search results %+v, actual %+v", expected, actual)
	}
}

func TestUpdatePullRequestAssignees_GivenAddedLogins_WhenUpdating_ThenItRunsGhPrEditWithAddAssignee(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestAssignees("acme/widgets", 42, []string{"alice", " bob "}, nil)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "edit", "42", "-R", "acme/widgets", "--add-assignee", "alice,bob"})
}

func TestUpdatePullRequestAssignees_GivenRemovedLogins_WhenUpdating_ThenItRunsGhPrEditWithRemoveAssignee(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestAssignees("acme/widgets", 42, nil, []string{"alice", " bob "})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "edit", "42", "-R", "acme/widgets", "--remove-assignee", "alice,bob"})
}
