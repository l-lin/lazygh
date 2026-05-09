package cache

import (
	"strconv"
	"testing"
	"time"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestOpen_GivenAFreshCachePath_WhenOpening_ThenItEnablesIncrementalAutoVacuum(t *testing.T) {
	subject := given_cacheStore(t)

	actual := given_autoVacuumMode(t, subject)

	if actual != 2 {
		t.Fatalf("expected auto_vacuum mode %d, actual %d", 2, actual)
	}
}

func TestStore_SavePullRequests_GivenAnUnreferencedMergedPullRequestOlderThanTheDiffTTL_WhenSavingAnotherList_ThenItPrunesTheCachedDiff(t *testing.T) {
	subject := given_cacheStore(t)
	staleSummary := given_mergedPullRequestSummary(42)
	then_noError(t, subject.SavePullRequestDetail(staleSummary, githubcli.PullRequestDetail{Title: staleSummary.Title, Number: staleSummary.Number, Body: "Merged body", State: staleSummary.State}))
	then_noError(t, subject.SavePullRequestDiff(staleSummary, githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+stale"}))
	given_pullRequestCacheRowUpdatedAt(t, subject, "acme/widgets", 42, staleMergedClosedPullRequestDiffTTL+time.Hour)

	actualErr := subject.SavePullRequests(given_pullRequestSearch("Mine"), []githubcli.PullRequest{given_openPullRequestSummary(7)})

	then_noError(t, actualErr)
	actualDiff, diffOK, actualErr := subject.PullRequestDiff("acme/widgets", 42)
	then_noError(t, actualErr)
	if diffOK {
		t.Fatalf("expected stale diff cache miss, actual %+v", actualDiff)
	}
	_, detailOK, actualErr := subject.PullRequestDetail("acme/widgets", 42)
	then_noError(t, actualErr)
	if !detailOK {
		t.Fatal("expected merged pull request detail to stay cached after diff pruning")
	}
}

func TestStore_SaveNotifications_GivenAnUnreferencedMergedPullRequestOlderThanTheDetailTTL_WhenSavingNotifications_ThenItPrunesTheCachedDetail(t *testing.T) {
	subject := given_cacheStore(t)
	staleSummary := given_mergedPullRequestSummary(42)
	then_noError(t, subject.SavePullRequestDetail(staleSummary, githubcli.PullRequestDetail{Title: staleSummary.Title, Number: staleSummary.Number, Body: "Merged body", State: staleSummary.State}))
	given_pullRequestCacheRowUpdatedAt(t, subject, "acme/widgets", 42, staleMergedClosedPullRequestDetailTTL+time.Hour)

	actualErr := subject.SaveNotifications([]githubcli.Notification{given_pullRequestNotification(7)})

	then_noError(t, actualErr)
	actualDetail, detailOK, actualErr := subject.PullRequestDetail("acme/widgets", 42)
	then_noError(t, actualErr)
	if detailOK {
		t.Fatalf("expected stale detail cache miss, actual %+v", actualDetail)
	}
	if actual := given_pullRequestRowCount(t, subject, "acme/widgets", 42); actual != 1 {
		t.Fatalf("expected the cached pull request summary row count %d, actual %d", 1, actual)
	}
}

func TestStore_SaveNotifications_GivenAMergedPullRequestStillReferencedByNotifications_WhenPruningRuns_ThenItKeepsTheCachedDetailAndDiff(t *testing.T) {
	subject := given_cacheStore(t)
	staleSummary := given_mergedPullRequestSummary(42)
	then_noError(t, subject.SavePullRequestDetail(staleSummary, githubcli.PullRequestDetail{Title: staleSummary.Title, Number: staleSummary.Number, Body: "Merged body", State: staleSummary.State}))
	then_noError(t, subject.SavePullRequestDiff(staleSummary, githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+stale"}))
	given_pullRequestCacheRowUpdatedAt(t, subject, "acme/widgets", 42, staleMergedClosedPullRequestDetailTTL+time.Hour)

	actualErr := subject.SaveNotifications([]githubcli.Notification{given_pullRequestNotification(42)})

	then_noError(t, actualErr)
	_, detailOK, actualErr := subject.PullRequestDetail("acme/widgets", 42)
	then_noError(t, actualErr)
	if !detailOK {
		t.Fatal("expected notification-referenced detail cache to stay available")
	}
	_, diffOK, actualErr := subject.PullRequestDiff("acme/widgets", 42)
	then_noError(t, actualErr)
	if !diffOK {
		t.Fatal("expected notification-referenced diff cache to stay available")
	}
}

func TestStore_SavePullRequests_GivenAnUnreferencedMergedPullRequestOlderThanTheSummaryTTLWithoutHeavyBlobs_WhenSavingAnotherList_ThenItDeletesTheCachedRow(t *testing.T) {
	subject := given_cacheStore(t)
	staleSummary := given_mergedPullRequestSummary(42)
	then_noError(t, subject.SavePullRequests(given_pullRequestSearch("Mine"), []githubcli.PullRequest{staleSummary}))
	given_pullRequestCacheRowUpdatedAt(t, subject, "acme/widgets", 42, staleMergedClosedPullRequestSummaryTTL+time.Hour)

	actualErr := subject.SavePullRequests(given_pullRequestSearch("Mine"), []githubcli.PullRequest{given_openPullRequestSummary(7)})

	then_noError(t, actualErr)
	if actual := given_pullRequestRowCount(t, subject, "acme/widgets", 42); actual != 0 {
		t.Fatalf("expected the stale cached pull request row count %d, actual %d", 0, actual)
	}
}

func given_autoVacuumMode(t *testing.T, subject *Store) int {
	t.Helper()

	var actual int
	then_noError(t, subject.db.QueryRow("PRAGMA auto_vacuum").Scan(&actual))
	return actual
}

func given_pullRequestCacheRowUpdatedAt(t *testing.T, subject *Store, repository string, number int, age time.Duration) {
	t.Helper()

	updatedAt := time.Now().UTC().Add(-age).Format("2006-01-02 15:04:05")
	_, actualErr := subject.db.Exec(`
		UPDATE pull_requests
		SET updated_at = ?
		WHERE repository = ? AND number = ?
	`, updatedAt, repository, number)
	then_noError(t, actualErr)
}

func given_pullRequestRowCount(t *testing.T, subject *Store, repository string, number int) int {
	t.Helper()

	var actual int
	then_noError(t, subject.db.QueryRow(`
		SELECT COUNT(*)
		FROM pull_requests
		WHERE repository = ? AND number = ?
	`, repository, number).Scan(&actual))
	return actual
}

func given_pullRequestSearch(label string) appconfig.PullRequestSearch {
	return appconfig.PullRequestSearch{Label: label, Command: []string{"search", "prs", "--search", label}}
}

func given_mergedPullRequestSummary(number int) githubcli.PullRequest {
	return githubcli.PullRequest{
		Title:      "Merged PR",
		Number:     number,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/" + strconv.Itoa(number),
		Body:       "Merged body",
		State:      "MERGED",
		UpdatedAt:  "2026-05-05T10:00:00Z",
	}
}

func given_openPullRequestSummary(number int) githubcli.PullRequest {
	return githubcli.PullRequest{
		Title:      "Open PR",
		Number:     number,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/" + strconv.Itoa(number),
		Body:       "Open body",
		State:      "OPEN",
		UpdatedAt:  "2026-05-05T11:00:00Z",
	}
}

func given_pullRequestNotification(number int) githubcli.Notification {
	return githubcli.Notification{
		ID:         "notification-" + strconv.Itoa(number),
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		Reason:     "review_requested",
		Unread:     true,
		UpdatedAt:  "2026-05-08T16:53:11Z",
		Subject: githubcli.NotificationSubject{
			Type:  githubcli.NotificationSubjectTypePullRequest,
			Title: "Notification PR",
			URL:   "https://api.github.com/repos/acme/widgets/pulls/" + strconv.Itoa(number),
		},
	}
}
