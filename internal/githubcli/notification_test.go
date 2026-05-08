package githubcli

import (
	"os"
	"path/filepath"
	"testing"
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

func given_notificationFixtureBytes(t *testing.T, name string) []byte {
	t.Helper()

	actual, actualErr := os.ReadFile(filepath.Join("testdata", "notifications", name))
	then_noError(t, actualErr)
	return actual
}
