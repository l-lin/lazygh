package cache

import (
	"reflect"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestStore_Clear_GivenStoredListsDetailsAndDiffs_WhenClearing_ThenItWipesThePersistentCache(t *testing.T) {
	subject := given_cacheStore(t)
	search := appconfig.PullRequestSearch{Label: "Mine", Command: []string{"search", "prs", "--author", "@me"}}
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}
	expectedPullRequests := []githubcli.PullRequest{summary}
	expectedDetail := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Rich body", State: "OPEN"}
	expectedDiff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}}
	expectedNotifications := []githubcli.Notification{{
		ID:         "1001",
		Unread:     true,
		Reason:     "review_requested",
		UpdatedAt:  "2026-05-08T16:53:11Z",
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		Subject: githubcli.NotificationSubject{
			Title: "ship notifications",
			Type:  githubcli.NotificationSubjectTypePullRequest,
			URL:   "https://api.github.com/repos/acme/widgets/pulls/42",
		},
	}}

	then_noError(t, subject.SavePullRequests(search, expectedPullRequests))
	then_noError(t, subject.SavePullRequestDetail(summary, expectedDetail))
	then_noError(t, subject.SavePullRequestDiff(summary, expectedDiff))
	then_noError(t, subject.SaveNotifications(expectedNotifications))

	actualErr := subject.Clear()

	then_noError(t, actualErr)
	actualPullRequests, pullRequestsOK, actualErr := subject.PullRequests(search)
	then_noError(t, actualErr)
	if pullRequestsOK {
		t.Fatalf("expected cleared pull requests cache miss, actual %+v", actualPullRequests)
	}
	actualDetail, detailOK, actualErr := subject.PullRequestDetail("acme/widgets", 42)
	then_noError(t, actualErr)
	if detailOK {
		t.Fatalf("expected cleared detail cache miss, actual %+v", actualDetail)
	}
	actualDiff, diffOK, actualErr := subject.PullRequestDiff("acme/widgets", 42)
	then_noError(t, actualErr)
	if diffOK {
		t.Fatalf("expected cleared diff cache miss, actual %+v", actualDiff)
	}
	if !reflect.DeepEqual(actualPullRequests, []githubcli.PullRequest(nil)) {
		t.Fatalf("expected no cached pull requests after clear, actual %+v", actualPullRequests)
	}
	actualNotifications, notificationsOK, actualErr := subject.Notifications()
	then_noError(t, actualErr)
	if notificationsOK {
		t.Fatalf("expected cleared notifications cache miss, actual %+v", actualNotifications)
	}
	if !reflect.DeepEqual(actualNotifications, []githubcli.Notification(nil)) {
		t.Fatalf("expected no cached notifications after clear, actual %+v", actualNotifications)
	}
}
