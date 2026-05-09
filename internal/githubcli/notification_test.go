package githubcli

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestListNotifications_GivenPagedNotificationFixtures_WhenListing_ThenItFlattensAndNormalizesTheThreads(t *testing.T) {
	runner := &fakeRunner{stdout: given_notificationFixtureBytes(t, "threads_pr_issue_release.json")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListNotifications()

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "/notifications?all=true&per_page=100", "--paginate", "--slurp"})
	if len(actual) != 3 {
		t.Fatalf("expected 3 notifications, actual %d", len(actual))
	}

	expected := []struct {
		id         string
		kind       string
		repository string
		title      string
		unread     bool
		reason     string
	}{
		{id: "1001", kind: NotificationSubjectTypePullRequest, repository: "acme/widgets", title: "feat: ship notifications", unread: true, reason: "review_requested"},
		{id: "1002", kind: NotificationSubjectTypeIssue, repository: "acme/opencode", title: "Support for skills", unread: false, reason: "manual"},
		{id: "1003", kind: NotificationSubjectTypeRelease, repository: "acme/doctoboot", title: "v3.5.0", unread: false, reason: "subscribed"},
	}
	for index, expectedNotification := range expected {
		actualNotification := actual[index]
		if actualNotification.ID != expectedNotification.id {
			t.Fatalf("expected notification id %q at index %d, actual %q", expectedNotification.id, index, actualNotification.ID)
		}
		if actualNotification.Subject.Type != expectedNotification.kind {
			t.Fatalf("expected notification type %q at index %d, actual %q", expectedNotification.kind, index, actualNotification.Subject.Type)
		}
		if actualNotification.Repository.NameWithOwner != expectedNotification.repository {
			t.Fatalf("expected repository %q at index %d, actual %q", expectedNotification.repository, index, actualNotification.Repository.NameWithOwner)
		}
		if actualNotification.Subject.Title != expectedNotification.title {
			t.Fatalf("expected title %q at index %d, actual %q", expectedNotification.title, index, actualNotification.Subject.Title)
		}
		if actualNotification.Unread != expectedNotification.unread {
			t.Fatalf("expected unread=%t at index %d, actual %t", expectedNotification.unread, index, actualNotification.Unread)
		}
		if actualNotification.Reason != expectedNotification.reason {
			t.Fatalf("expected reason %q at index %d, actual %q", expectedNotification.reason, index, actualNotification.Reason)
		}
	}

	pullRequestSummary, ok := actual[0].PullRequestSummary()
	if !ok {
		t.Fatal("expected pull request notification to resolve a pull request summary")
	}
	if pullRequestSummary.Repository.NameWithOwner != "acme/widgets" || pullRequestSummary.Number != 42 {
		t.Fatalf("expected pull request summary acme/widgets#42, actual %+v", pullRequestSummary)
	}
	issueRepository, issueNumber, ok := actual[1].IssueIdentity()
	if !ok {
		t.Fatal("expected issue notification to resolve an issue identity")
	}
	if issueRepository != "acme/opencode" || issueNumber != 3235 {
		t.Fatalf("expected issue identity acme/opencode#3235, actual %s#%d", issueRepository, issueNumber)
	}
	releaseRepository, releaseID, ok := actual[2].ReleaseIdentity()
	if !ok {
		t.Fatal("expected release notification to resolve a release identity")
	}
	if releaseRepository != "acme/doctoboot" || releaseID != 317927281 {
		t.Fatalf("expected release identity acme/doctoboot#317927281, actual %s#%d", releaseRepository, releaseID)
	}
}

func TestGetIssueDetail_GivenARealFixture_WhenFetching_ThenItReturnsTheNormalizedIssueDetail(t *testing.T) {
	runner := &fakeRunner{stdout: given_notificationFixtureBytes(t, "issue_detail.json")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetIssueDetail("acme/opencode", 3235)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "repos/acme/opencode/issues/3235"})
	if actual.Title != "Support for skills" {
		t.Fatalf("expected title %q, actual %q", "Support for skills", actual.Title)
	}
	if actual.Number != 3235 {
		t.Fatalf("expected issue number %d, actual %d", 3235, actual.Number)
	}
	if actual.URL != "https://github.com/acme/opencode/issues/3235" {
		t.Fatalf("expected issue url %q, actual %q", "https://github.com/acme/opencode/issues/3235", actual.URL)
	}
	if actual.Author == nil || actual.Author.Login != "octocat" {
		t.Fatalf("expected author login %q, actual %+v", "octocat", actual.Author)
	}
	if actual.State != "open" {
		t.Fatalf("expected state %q, actual %q", "open", actual.State)
	}
	if actual.Comments != 51 {
		t.Fatalf("expected comment count %d, actual %d", 51, actual.Comments)
	}
	if len(actual.Labels) != 2 || actual.Labels[0].Name != "enhancement" || actual.Labels[1].Name != "ai" {
		t.Fatalf("expected normalized labels, actual %+v", actual.Labels)
	}
	if len(actual.Assignees) != 1 || actual.Assignees[0].Login != "monalisa" {
		t.Fatalf("expected normalized assignees, actual %+v", actual.Assignees)
	}
}

func TestGetReleaseDetail_GivenARealFixture_WhenFetching_ThenItReturnsTheNormalizedReleaseDetail(t *testing.T) {
	runner := &fakeRunner{stdout: given_notificationFixtureBytes(t, "release_detail.json")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetReleaseDetail("acme/doctoboot", 317927281)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "repos/acme/doctoboot/releases/317927281"})
	if actual.Name != "Notifications 3.5.0" {
		t.Fatalf("expected release name %q, actual %q", "Notifications 3.5.0", actual.Name)
	}
	if actual.TagName != "v3.5.0" {
		t.Fatalf("expected tag %q, actual %q", "v3.5.0", actual.TagName)
	}
	if actual.URL != "https://github.com/acme/doctoboot/releases/tag/v3.5.0" {
		t.Fatalf("expected release url %q, actual %q", "https://github.com/acme/doctoboot/releases/tag/v3.5.0", actual.URL)
	}
	if actual.Author == nil || actual.Author.Login != "release-bot" {
		t.Fatalf("expected author login %q, actual %+v", "release-bot", actual.Author)
	}
	if actual.Draft {
		t.Fatal("expected the release not to be a draft")
	}
	if !actual.PreRelease {
		t.Fatal("expected the release to be a prerelease")
	}
}

func TestMarkNotificationRead_GivenThreadID_WhenMarking_ThenItCallsTheThreadReadEndpoint(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.MarkNotificationRead("1001")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "/notifications/threads/1001", "--method", "PATCH"})
}

func TestMarkNotificationDone_GivenThreadID_WhenMarking_ThenItCallsTheThreadDoneEndpoint(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.MarkNotificationDone("1001")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "/notifications/threads/1001", "--method", "DELETE"})
}

func TestMarkAllNotificationsRead_GivenAcceptedResponse_WhenMarking_ThenItReturnsAcceptedStatus(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(strings.Join([]string{
		"HTTP/2.0 202 Accepted",
		"X-Poll-Interval: 60",
		"",
		`{"message":"Notifications are being marked as read in the background."}`,
	}, "\n"))}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.MarkAllNotificationsRead()

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "/notifications", "--method", "PUT", "--include"})
	if !actual.Accepted {
		t.Fatal("expected the bulk read response to be marked as accepted")
	}
}

func TestMarkAllNotificationsDone_GivenLoadedNotifications_WhenMarking_ThenItDeletesEachLoadedThread(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)
	notifications := []Notification{{ID: "1001"}, {ID: "1002"}, {ID: ""}, {ID: "1003"}}

	actualCount, actualErr := subject.MarkAllNotificationsDone(notifications)

	then_noError(t, actualErr)
	if actualCount != 3 {
		t.Fatalf("expected marked count %d, actual %d", 3, actualCount)
	}
	then_commandsMatchIgnoringOrder(t, runner.calls, []fakeCommandCall{
		{name: "gh", args: []string{"api", "/notifications/threads/1001", "--method", "DELETE"}, stdin: nil},
		{name: "gh", args: []string{"api", "/notifications/threads/1002", "--method", "DELETE"}, stdin: nil},
		{name: "gh", args: []string{"api", "/notifications/threads/1003", "--method", "DELETE"}, stdin: nil},
	})
}

func TestMarkAllNotificationsDone_GivenMoreNotificationsThanTheWorkerLimit_WhenMarking_ThenItCapsConcurrency(t *testing.T) {
	runner := newBlockingNotificationDoneRunner()
	subject := NewClientWithRunner(runner)
	notifications := []Notification{{ID: "1001"}, {ID: "1002"}, {ID: "1003"}, {ID: "1004"}, {ID: "1005"}}
	results := make(chan struct {
		count int
		err   error
	}, 1)

	go func() {
		count, err := subject.MarkAllNotificationsDone(notifications)
		results <- struct {
			count int
			err   error
		}{count: count, err: err}
	}()

	for index := 0; index < notificationsBulkDoneConcurrency; index++ {
		select {
		case <-runner.started:
		case <-time.After(time.Second):
			t.Fatalf("expected worker %d to start within the concurrency limit", index+1)
		}
	}
	select {
	case <-runner.started:
		t.Fatalf("expected at most %d concurrent notification-done requests", notificationsBulkDoneConcurrency)
	default:
	}

	runner.releaseAll()
	result := <-results
	then_noError(t, result.err)
	if result.count != 5 {
		t.Fatalf("expected marked count %d, actual %d", 5, result.count)
	}
	if runner.maxActive() != notificationsBulkDoneConcurrency {
		t.Fatalf("expected max concurrency %d, actual %d", notificationsBulkDoneConcurrency, runner.maxActive())
	}
}

func TestMarkNotificationRead_GivenUnsupportedCredentialError_WhenMarking_ThenItReturnsActionableGuidance(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("Resource not accessible by personal access token"), err: errors.New("exit status 1")}
	subject := NewClientWithRunner(runner)

	actualErr := subject.MarkNotificationRead("1001")

	if !errors.Is(actualErr, ErrNotificationEndpointAuthRefused) {
		t.Fatalf("expected error %v, actual %v", ErrNotificationEndpointAuthRefused, actualErr)
	}
	for _, expected := range []string{"personal access token (classic)", "notifications", "repo"} {
		if !strings.Contains(actualErr.Error(), expected) {
			t.Fatalf("expected actionable notification error to mention %q, actual %q", expected, actualErr.Error())
		}
	}
}

func TestNotificationIncludedHTTPStatus_GivenIncludedResponse_WhenParsing_ThenItReturnsTheStatusCode(t *testing.T) {
	actual := notificationIncludedHTTPStatus([]byte("HTTP/2.0 205 Reset Content\r\nX-Poll-Interval: 60\r\n\r\n"))

	if actual != 205 {
		t.Fatalf("expected status code %d, actual %d", 205, actual)
	}
}

func TestNormalizeNotificationEndpointError_GivenScopeError_WhenNormalizing_ThenItReturnsActionableGuidance(t *testing.T) {
	actual := normalizeNotificationEndpointError(errors.New("run `gh api notifications read`: exit status 1: Requires one of the following scopes: ['notifications']"))

	if !errors.Is(actual, ErrNotificationEndpointAuthRefused) {
		t.Fatalf("expected error %v, actual %v", ErrNotificationEndpointAuthRefused, actual)
	}
	if !strings.Contains(actual.Error(), "notifications") || !strings.Contains(actual.Error(), "repo") {
		t.Fatalf("expected actionable scope guidance, actual %q", actual.Error())
	}
}

func TestNotificationThreadAPIPath_GivenBlankThreadID_WhenFormatting_ThenItRejectsTheTarget(t *testing.T) {
	_, actualErr := notificationThreadAPIPath("   ")

	if !errors.Is(actualErr, ErrMissingNotificationThreadID) {
		t.Fatalf("expected error %v, actual %v", ErrMissingNotificationThreadID, actualErr)
	}
}

func TestNormalizedNotifications_GivenNotifications_WhenNormalizing_ThenItReturnsANewSlice(t *testing.T) {
	original := []Notification{{ID: " 1001 "}}

	actual := normalizedNotifications(original)

	actual[0].ID = "changed"
	if original[0].ID != " 1001 " {
		t.Fatalf("expected the original notification id %q to stay unchanged, actual %q", " 1001 ", original[0].ID)
	}
	if actual[0].ID != "changed" {
		t.Fatalf("expected normalized id %q, actual %q", "1001", actual[0].ID)
	}
}

func given_notificationFixtureBytes(t *testing.T, name string) []byte {
	t.Helper()

	actual, actualErr := os.ReadFile(filepath.Join("testdata", "notifications", name))
	then_noError(t, actualErr)
	return actual
}

func then_commandsMatchIgnoringOrder(t *testing.T, actual []fakeCommandCall, expected []fakeCommandCall) {
	t.Helper()

	actualKeys := fakeCommandCallKeys(actual)
	expectedKeys := fakeCommandCallKeys(expected)
	if len(actualKeys) != len(expectedKeys) {
		t.Fatalf("expected command calls %v, actual %v", expected, actual)
	}
	for index := range actualKeys {
		if actualKeys[index] != expectedKeys[index] {
			t.Fatalf("expected command calls %v, actual %v", expected, actual)
		}
	}
}

func fakeCommandCallKeys(calls []fakeCommandCall) []string {
	keys := make([]string, 0, len(calls))
	for _, call := range calls {
		keys = append(keys, call.name+"\x00"+strings.Join(call.args, "\x00")+"\x00"+string(call.stdin))
	}
	sort.Strings(keys)
	return keys
}

type blockingNotificationDoneRunner struct {
	mu         sync.Mutex
	active     int
	maxRunning int
	started    chan struct{}
	release    chan struct{}
}

func newBlockingNotificationDoneRunner() *blockingNotificationDoneRunner {
	return &blockingNotificationDoneRunner{
		started: make(chan struct{}, 32),
		release: make(chan struct{}),
	}
}

func (runner *blockingNotificationDoneRunner) Run(name string, args ...string) (CommandResult, error) {
	runner.mu.Lock()
	runner.active++
	if runner.active > runner.maxRunning {
		runner.maxRunning = runner.active
	}
	runner.mu.Unlock()

	runner.started <- struct{}{}
	<-runner.release

	runner.mu.Lock()
	runner.active--
	runner.mu.Unlock()
	return CommandResult{}, nil
}

func (runner *blockingNotificationDoneRunner) RunWithInput(name string, input []byte, args ...string) (CommandResult, error) {
	return CommandResult{}, errors.New("unexpected RunWithInput call")
}

func (runner *blockingNotificationDoneRunner) releaseAll() {
	close(runner.release)
}

func (runner *blockingNotificationDoneRunner) maxActive() int {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	return runner.maxRunning
}
