package cache

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestOpen_GivenANestedCachePath_WhenOpening_ThenItCreatesTheParentDirectory(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "xdg", "lazygh", "cache.sqlite3")

	subject, actualErr := Open(cachePath)

	then_noError(t, actualErr)
	if subject == nil {
		t.Fatal("expected a cache store")
	}
	defer func() {
		then_noError(t, subject.Close())
	}()
	if _, actualErr := os.Stat(filepath.Dir(cachePath)); actualErr != nil {
		t.Fatalf("expected cache directory to exist, actual %v", actualErr)
	}
}

func TestStore_PullRequests_GivenAStoredSearchResult_WhenReading_ThenItReturnsTheCachedPullRequests(t *testing.T) {
	subject := given_cacheStore(t)
	search := appconfig.PullRequestSearch{Label: "Mine", Command: []string{"search", "prs", "--author", "@me"}}
	expected := []githubcli.PullRequest{{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", Body: "First body", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}, {Title: "Second PR", Number: 99, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/99", Body: "Second body", State: "OPEN", UpdatedAt: "2026-05-05T11:00:00Z"}}

	actualErr := subject.SavePullRequests(search, expected)
	then_noError(t, actualErr)

	actual, ok, actualErr := subject.PullRequests(search)

	then_noError(t, actualErr)
	if !ok {
		t.Fatal("expected cached pull requests")
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected pull requests %+v, actual %+v", expected, actual)
	}
}

func TestStore_Notifications_GivenStoredNotifications_WhenReading_ThenItReturnsTheCachedNotifications(t *testing.T) {
	subject := given_cacheStore(t)
	expected := []githubcli.Notification{{
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

	actualErr := subject.SaveNotifications(expected)
	then_noError(t, actualErr)

	actual, ok, actualErr := subject.Notifications()

	then_noError(t, actualErr)
	if !ok {
		t.Fatal("expected cached notifications")
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected notifications %+v, actual %+v", expected, actual)
	}
}

func TestStore_PullRequestDetail_GivenAStoredDetail_WhenReading_ThenItReturnsTheDetailAndSummaryVersion(t *testing.T) {
	subject := given_cacheStore(t)
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:00:00Z"}
	expected := githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Rich body", BaseRefName: "main", HeadRefName: "feature/cache", State: "OPEN"}

	actualErr := subject.SavePullRequestDetail(summary, expected)
	then_noError(t, actualErr)

	actual, ok, actualErr := subject.PullRequestDetail("acme/widgets", 42)

	then_noError(t, actualErr)
	if !ok {
		t.Fatal("expected a cached detail")
	}
	if actual.SourceUpdatedAt != summary.UpdatedAt {
		t.Fatalf("expected summary version %q, actual %q", summary.UpdatedAt, actual.SourceUpdatedAt)
	}
	if !reflect.DeepEqual(actual.Detail, expected) {
		t.Fatalf("expected detail %+v, actual %+v", expected, actual.Detail)
	}
}

func TestStore_PullRequestDiff_GivenAStoredDiff_WhenReading_ThenItReturnsTheDiffAndSummaryVersion(t *testing.T) {
	subject := given_cacheStore(t)
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-05T10:00:00Z"}
	expected := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubcli.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}, Threads: []githubcli.PullRequestReviewThread{{ID: "thread-1", Path: "main.go", Line: 12, Comments: []githubcli.PullRequestComment{{Body: "ship it"}}}}}

	actualErr := subject.SavePullRequestDiff(summary, expected)
	then_noError(t, actualErr)

	actual, ok, actualErr := subject.PullRequestDiff("acme/widgets", 42)

	then_noError(t, actualErr)
	if !ok {
		t.Fatal("expected a cached diff")
	}
	if actual.SourceUpdatedAt != summary.UpdatedAt {
		t.Fatalf("expected summary version %q, actual %q", summary.UpdatedAt, actual.SourceUpdatedAt)
	}
	if !reflect.DeepEqual(actual.Diff, expected) {
		t.Fatalf("expected diff %+v, actual %+v", expected, actual.Diff)
	}
}

func TestStore_InvalidatePullRequest_GivenStoredRichData_WhenInvalidating_ThenItDropsTheCachedListDetailAndDiffEntries(t *testing.T) {
	subject := given_cacheStore(t)
	search := appconfig.PullRequestSearch{Label: "Mine", Command: []string{"search", "prs", "--author", "@me"}}
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z"}

	then_noError(t, subject.SavePullRequests(search, []githubcli.PullRequest{summary}))
	then_noError(t, subject.SavePullRequestDetail(summary, githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Rich body", State: "OPEN"}))
	then_noError(t, subject.SavePullRequestDiff(summary, githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached"}))

	actualErr := subject.InvalidatePullRequest("acme/widgets", 42)

	then_noError(t, actualErr)
	actualDetail, detailOK, actualErr := subject.PullRequestDetail("acme/widgets", 42)
	then_noError(t, actualErr)
	if detailOK {
		t.Fatalf("expected detail cache miss, actual %+v", actualDetail)
	}
	actualDiff, diffOK, actualErr := subject.PullRequestDiff("acme/widgets", 42)
	then_noError(t, actualErr)
	if diffOK {
		t.Fatalf("expected diff cache miss, actual %+v", actualDiff)
	}
	actualPullRequests, pullRequestsOK, actualErr := subject.PullRequests(search)
	then_noError(t, actualErr)
	if pullRequestsOK {
		t.Fatalf("expected cached pull-request list miss, actual %+v", actualPullRequests)
	}
}

func given_cacheStore(t *testing.T) *Store {
	t.Helper()

	subject, actualErr := Open(filepath.Join(t.TempDir(), "lazygh", "cache.sqlite3"))
	then_noError(t, actualErr)
	t.Cleanup(func() {
		then_noError(t, subject.Close())
	})
	return subject
}

func then_noError(t *testing.T, actualErr error) {
	t.Helper()

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
