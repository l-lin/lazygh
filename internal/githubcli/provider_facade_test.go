package githubcli

import (
	"reflect"
	"testing"
)

func TestSessionService_GivenConnectedUserPayload_WhenFetching_ThenItUsesTheSharedRESTTransport(t *testing.T) {
	executor := &fakeExecutor{result: CommandResult{Stdout: []byte(`{"login":"octocat","html_url":"https://github.com/octocat"}`)}}
	subject := newSessionService(givenSharedTransport(executor))

	actual, actualErr := subject.GetConnectedUser()

	then_noError(t, actualErr)
	then_executedCommandsAre(t, executor, []Command{{Args: []string{"api", "user"}, DisplayArgs: []string{"api", "user"}}})
	if actual.Login != "octocat" {
		t.Fatalf("expected connected user login %q, actual %q", "octocat", actual.Login)
	}
}

func TestNotificationService_GivenPaginatedNotifications_WhenListing_ThenItUsesTheSharedRESTPaginator(t *testing.T) {
	executor := &fakeExecutor{result: CommandResult{Stdout: []byte(`[[{"id":"1001","subject":{"title":"One","type":"Issue"}}],[{"id":"1002","subject":{"title":"Two","type":"Issue"}}]]`)}}
	subject := newNotificationService(givenSharedTransport(executor))

	actual, actualErr := subject.ListNotifications()

	then_noError(t, actualErr)
	then_executedCommandsAre(t, executor, []Command{{Args: []string{"api", notificationsListAPIPath, "--paginate", "--slurp"}, DisplayArgs: []string{"api", notificationsListAPIPath, "--paginate", "--slurp"}}})
	if len(actual) != 2 {
		t.Fatalf("expected 2 notifications, actual %d", len(actual))
	}
}

func TestBuildService_GivenRunJobsPayload_WhenFetching_ThenItUsesTheSharedCommandExecutor(t *testing.T) {
	executor := &fakeExecutor{result: CommandResult{Stdout: []byte(`{"jobs":[{"databaseId":11,"name":"test"}]}`)}}
	subject := newBuildService(givenSharedTransport(executor))

	actual, actualErr := subject.GetPullRequestBuildRunJobs("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	then_noError(t, actualErr)
	then_executedCommandsAre(t, executor, []Command{{Args: []string{"run", "view", "42", "-R", "acme/widgets", "--json", "jobs"}, DisplayArgs: []string{"run", "view", "42", "-R", "acme/widgets", "--json", "jobs"}}})
	if len(actual) != 1 || actual[0].DatabaseID != 11 {
		t.Fatalf("expected build jobs %v, actual %+v", []int{11}, actual)
	}
}

func TestProviderFacade_GivenConnectedUserRequest_WhenFetching_ThenItDelegatesToTheSessionService(t *testing.T) {
	session := &stubSessionProvider{user: ConnectedUser{Login: "octocat"}}
	subject := ProviderFacade{session: session}

	actual, actualErr := subject.GetConnectedUser()

	then_noError(t, actualErr)
	if !session.called {
		t.Fatalf("expected the facade to delegate to the session service")
	}
	if actual.Login != "octocat" {
		t.Fatalf("expected connected user login %q, actual %q", "octocat", actual.Login)
	}
}

func TestProviderFacade_GivenNotificationMutation_WhenMarkingItDone_ThenItDelegatesToTheNotificationService(t *testing.T) {
	notifications := &stubNotificationProvider{}
	subject := ProviderFacade{notifications: notifications}

	actualErr := subject.MarkNotificationDone("thread-42")

	then_noError(t, actualErr)
	if notifications.markDoneThreadID != "thread-42" {
		t.Fatalf("expected thread id %q, actual %q", "thread-42", notifications.markDoneThreadID)
	}
}

func TestProviderFacade_GivenMarkdownRendering_WhenRenderingHTML_ThenItDelegatesToTheMarkdownService(t *testing.T) {
	markdown := &stubMarkdownProvider{html: "<p>ship it</p>"}
	subject := ProviderFacade{markdown: markdown}

	actual, actualErr := subject.RenderMarkdownHTML("acme/widgets", "ship it")

	then_noError(t, actualErr)
	if markdown.repository != "acme/widgets" || markdown.markdown != "ship it" {
		t.Fatalf("expected markdown renderer inputs %q and %q, actual %q and %q", "acme/widgets", "ship it", markdown.repository, markdown.markdown)
	}
	if actual != "<p>ship it</p>" {
		t.Fatalf("expected rendered HTML %q, actual %q", "<p>ship it</p>", actual)
	}
}

type fakeExecutor struct {
	result   CommandResult
	err      error
	commands []Command
}

func (executor *fakeExecutor) Execute(command Command) (CommandResult, error) {
	executor.commands = append(executor.commands, Command{
		Args:        append([]string(nil), command.Args...),
		Stdin:       append([]byte(nil), command.Stdin...),
		DisplayArgs: append([]string(nil), command.DisplayArgs...),
	})
	return executor.result, executor.err
}

func givenSharedTransport(executor Executor) sharedTransport {
	formatter := NewCommandFormatter()
	classifier := NewErrorClassifier(formatter)
	return sharedTransport{
		executor:   executor,
		formatter:  formatter,
		graphql:    NewGraphQLClient(executor),
		rest:       NewRESTClient(executor),
		paginator:  NewPaginator(),
		decoder:    NewResponseDecoder(),
		classifier: classifier,
	}
}

func then_executedCommandsAre(t *testing.T, executor *fakeExecutor, expected []Command) {
	t.Helper()

	if !reflect.DeepEqual(executor.commands, expected) {
		t.Fatalf("expected executed commands %+v, actual %+v", expected, executor.commands)
	}
}

type stubSessionProvider struct {
	user   ConnectedUser
	called bool
}

func (provider *stubSessionProvider) GetConnectedUser() (ConnectedUser, error) {
	provider.called = true
	return provider.user, nil
}

type stubNotificationProvider struct {
	markDoneThreadID string
}

func (provider *stubNotificationProvider) MarkNotificationDone(threadID string) error {
	provider.markDoneThreadID = threadID
	return nil
}

func (provider *stubNotificationProvider) ListNotifications() ([]Notification, error) {
	return nil, nil
}

func (provider *stubNotificationProvider) GetIssueDetail(string, int) (IssueDetail, error) {
	return IssueDetail{}, nil
}

func (provider *stubNotificationProvider) GetReleaseDetail(string, int) (ReleaseDetail, error) {
	return ReleaseDetail{}, nil
}

func (provider *stubNotificationProvider) MarkNotificationRead(string) error {
	return nil
}

func (provider *stubNotificationProvider) MarkAllNotificationsRead() (NotificationBulkReadResult, error) {
	return NotificationBulkReadResult{}, nil
}

func (provider *stubNotificationProvider) MarkAllNotificationsDone([]Notification) (int, error) {
	return 0, nil
}

type stubMarkdownProvider struct {
	repository string
	markdown   string
	html       string
}

func (provider *stubMarkdownProvider) RenderMarkdownHTML(repository string, markdown string) (string, error) {
	provider.repository = repository
	provider.markdown = markdown
	return provider.html, nil
}
