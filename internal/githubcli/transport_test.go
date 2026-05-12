package githubcli

import (
	"errors"
	"os/exec"
	"testing"
)

func TestCommandFormatter_GivenDisplayArguments_WhenFormatting_ThenItUsesThemInsteadOfTheRawCommand(t *testing.T) {
	subject := NewCommandFormatter()

	actual := subject.Format(Command{
		Args:        []string{"api", "graphql", "-f", "query=query($owner:String!){viewer{login}}", "-F", "owner=acme"},
		DisplayArgs: []string{"api", "graphql"},
	})

	if actual != "gh api graphql" {
		t.Fatalf("expected formatted command %q, actual %q", "gh api graphql", actual)
	}
}

func TestExecutor_GivenARawCommand_WhenExecuting_ThenItRunsGhAndReturnsTheResult(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("ok")}
	subject := NewExecutor(runner, NewCommandFormatter(), NewErrorClassifier(NewCommandFormatter()))

	actual, actualErr := subject.Execute(Command{Args: []string{"api", "user"}})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "user"})
	if string(actual.Stdout) != "ok" {
		t.Fatalf("expected stdout %q, actual %q", "ok", string(actual.Stdout))
	}
}

func TestExecutor_GivenAStdinCommand_WhenExecuting_ThenItDelegatesToTheInputRunner(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("ok")}
	subject := NewExecutor(runner, NewCommandFormatter(), NewErrorClassifier(NewCommandFormatter()))

	actual, actualErr := subject.Execute(Command{Args: []string{"api", "markdown", "--input", "-"}, Stdin: []byte("# Ship it")})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "markdown", "--input", "-"})
	then_stdinIs(t, runner, "# Ship it")
	if string(actual.Stdout) != "ok" {
		t.Fatalf("expected stdout %q, actual %q", "ok", string(actual.Stdout))
	}
}

func TestGraphQLClient_GivenTypedVariables_WhenQuerying_ThenItBuildsTheGraphQLCommand(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"viewer":{"login":"octocat"}}}`)}
	executor := NewExecutor(runner, NewCommandFormatter(), NewErrorClassifier(NewCommandFormatter()))
	subject := NewGraphQLClient(executor)

	_, actualErr := subject.Query(GraphQLRequest{
		Query: `query($owner:String!,$name:String!){repository(owner:$owner,name:$name){id}}`,
		Variables: []GraphQLVariable{
			{Name: "owner", Value: "acme", Typed: true},
			{Name: "name", Value: "widgets", Typed: true},
		},
	})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=query($owner:String!,$name:String!){repository(owner:$owner,name:$name){id}}",
		"-F",
		"owner=acme",
		"-F",
		"name=widgets",
	})
}

func TestRESTClient_GivenAPaginatedRequest_WhenExecuting_ThenItAppendsPaginateAndSlurp(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[[{"id":"1001"}]]`)}
	executor := NewExecutor(runner, NewCommandFormatter(), NewErrorClassifier(NewCommandFormatter()))
	subject := NewRESTClient(executor)

	_, actualErr := subject.Do(RESTRequest{Path: "/notifications?all=true", Paginate: true, Slurp: true})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "/notifications?all=true", "--paginate", "--slurp"})
}

func TestResponseDecoder_GivenPlainJSON_WhenDecoding_ThenItPopulatesTheTarget(t *testing.T) {
	subject := NewResponseDecoder()
	var actual struct {
		Login string `json:"login"`
	}

	actualErr := subject.DecodeJSON([]byte(`{"login":"octocat"}`), &actual)

	then_noError(t, actualErr)
	if actual.Login != "octocat" {
		t.Fatalf("expected decoded login %q, actual %q", "octocat", actual.Login)
	}
}

func TestPaginator_GivenSlurpedPages_WhenDecoding_ThenItFlattensThePages(t *testing.T) {
	subject := NewPaginator()
	type pageItem struct {
		ID string `json:"id"`
	}
	actual := make([]pageItem, 0)

	actualErr := subject.DecodeSlurpedJSON([]byte(`[[{"id":"1001"}],[{"id":"1002"},{"id":"1003"}]]`), &actual)

	then_noError(t, actualErr)
	if len(actual) != 3 {
		t.Fatalf("expected 3 flattened items, actual %d", len(actual))
	}
	if actual[0].ID != "1001" || actual[1].ID != "1002" || actual[2].ID != "1003" {
		t.Fatalf("expected flattened ids %v, actual %v", []string{"1001", "1002", "1003"}, []string{actual[0].ID, actual[1].ID, actual[2].ID})
	}
}

func TestResponseDecoder_GivenAGraphQLErrorPayload_WhenDecoding_ThenItReturnsTheGraphQLError(t *testing.T) {
	subject := NewResponseDecoder()
	var actual struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}

	actualErr := subject.DecodeGraphQL([]byte(`{"errors":[{"message":"viewer is forbidden"}]}`), &actual)

	if actualErr == nil || actualErr.Error() != "viewer is forbidden" {
		t.Fatalf("expected GraphQL error %q, actual %v", "viewer is forbidden", actualErr)
	}
}

func TestErrorClassifier_GivenUnavailableGh_WhenClassifying_ThenItReturnsUnavailable(t *testing.T) {
	subject := NewErrorClassifier(NewCommandFormatter())

	actual := subject.Classify(Command{Args: []string{"api", "user"}}, CommandResult{}, exec.ErrNotFound)

	if !errors.Is(actual, ErrUnavailable) {
		t.Fatalf("expected error %v, actual %v", ErrUnavailable, actual)
	}
}

func TestErrorClassifier_GivenAnAuthenticationPrompt_WhenClassifying_ThenItReturnsUnauthenticated(t *testing.T) {
	subject := NewErrorClassifier(NewCommandFormatter())

	actual := subject.Classify(
		Command{Args: []string{"api", "user"}},
		CommandResult{Stderr: []byte("To get started with GitHub CLI, please run: gh auth login")},
		errors.New("exit status 4"),
	)

	if !errors.Is(actual, ErrUnauthenticated) {
		t.Fatalf("expected error %v, actual %v", ErrUnauthenticated, actual)
	}
}

func TestErrorClassifier_GivenAGenericFailure_WhenClassifying_ThenItIncludesTheFormattedCommand(t *testing.T) {
	subject := NewErrorClassifier(NewCommandFormatter())

	actual := subject.Classify(
		Command{Args: []string{"api", "graphql", "-f", "query=..."}, DisplayArgs: []string{"api", "graphql"}},
		CommandResult{Stderr: []byte("boom")},
		errors.New("exit status 1"),
	)

	if actual == nil || actual.Error() != "run `gh api graphql`: exit status 1: boom" {
		t.Fatalf("expected formatted command error %q, actual %v", "run `gh api graphql`: exit status 1: boom", actual)
	}
}
