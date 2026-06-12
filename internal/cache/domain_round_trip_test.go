package cache

import (
	"reflect"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestStore_GivenProviderNeutralModels_WhenSavingAndLoading_ThenTheRoundTripsStayLossless(t *testing.T) {
	subject := given_cacheStore(t)
	search := appconfig.PullRequestSearch{Label: "Mine", Command: []string{"search", "prs", "--author", "@me"}}
	expectedPullRequests := []githubdomain.PullRequestSummary{{
		Title:               "First PR",
		Number:              42,
		Repository:          githubdomain.RepositoryRef{NameWithOwner: "acme/widgets"},
		URL:                 "https://github.com/acme/widgets/pull/42",
		Body:                "First body",
		State:               "OPEN",
		UpdatedAt:           "2026-05-05T10:00:00Z",
		IsMergeQueueEnabled: true,
		IsInMergeQueue:      true,
		MergeQueueEntry: &githubdomain.PullRequestMergeQueueEntry{
			ID:                   "MQE_1",
			State:                "QUEUED",
			Position:             2,
			EstimatedTimeToMerge: 17,
		},
	}}
	expectedNotifications := []githubdomain.Notification{{
		ID:         "1001",
		Unread:     true,
		Reason:     "review_requested",
		UpdatedAt:  "2026-05-08T16:53:11Z",
		Repository: githubdomain.RepositoryRef{NameWithOwner: "acme/widgets"},
		Subject:    githubdomain.NotificationSubject{Title: "ship notifications", Type: githubdomain.NotificationSubjectTypePullRequest, URL: "https://api.github.com/repos/acme/widgets/pulls/42"},
	}}
	summary := githubdomain.PullRequestSummary{
		Title:               "First PR",
		Number:              42,
		Repository:          githubdomain.RepositoryRef{NameWithOwner: "acme/widgets"},
		UpdatedAt:           "2026-05-05T10:00:00Z",
		IsMergeQueueEnabled: true,
		IsInMergeQueue:      true,
		MergeQueueEntry: &githubdomain.PullRequestMergeQueueEntry{
			ID:                   "MQE_1",
			State:                "QUEUED",
			Position:             2,
			EstimatedTimeToMerge: 17,
		},
	}
	expectedDetail := githubdomain.PullRequestDetail{
		Title:               "First PR",
		Number:              42,
		Body:                "Rich body",
		BaseRefName:         "main",
		HeadRefName:         "feature/cache",
		State:               "OPEN",
		IsMergeQueueEnabled: true,
		IsInMergeQueue:      true,
		MergeQueueEntry: &githubdomain.PullRequestMergeQueueEntry{
			ID:                   "MQE_1",
			State:                "QUEUED",
			Position:             2,
			EstimatedTimeToMerge: 17,
		},
		InlineCommentThreads: []githubdomain.ReviewThread{{ID: "thread-1", Path: "main.go", Line: 12, Comments: []githubdomain.PullRequestComment{{Body: "ship it"}}}},
	}
	expectedDiff := githubdomain.PullRequestDiff{UnifiedDiff: "diff --git a/main.go b/main.go\n+cached", Files: []githubdomain.PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1}}, Threads: []githubdomain.ReviewThread{{ID: "thread-1", Path: "main.go", Line: 12, Comments: []githubdomain.PullRequestComment{{Body: "ship it"}}}}}

	then_noError(t, subject.SavePullRequests(search, expectedPullRequests))
	then_noError(t, subject.SaveNotifications(expectedNotifications))
	then_noError(t, subject.SavePullRequestDetail(summary, expectedDetail))
	then_noError(t, subject.SavePullRequestDiff(summary, expectedDiff))

	actualPullRequests, ok, actualErr := subject.PullRequests(search)
	then_noError(t, actualErr)
	if !ok || !reflect.DeepEqual(actualPullRequests, expectedPullRequests) {
		t.Fatalf("expected pull requests %+v, actual %+v ok=%t", expectedPullRequests, actualPullRequests, ok)
	}

	actualNotifications, ok, actualErr := subject.Notifications()
	then_noError(t, actualErr)
	if !ok || !reflect.DeepEqual(actualNotifications, expectedNotifications) {
		t.Fatalf("expected notifications %+v, actual %+v ok=%t", expectedNotifications, actualNotifications, ok)
	}

	actualDetail, ok, actualErr := subject.PullRequestDetail("acme/widgets", 42)
	then_noError(t, actualErr)
	if !ok || !reflect.DeepEqual(actualDetail.Detail, expectedDetail) {
		t.Fatalf("expected detail %+v, actual %+v ok=%t", expectedDetail, actualDetail.Detail, ok)
	}

	actualDiff, ok, actualErr := subject.PullRequestDiff("acme/widgets", 42)
	then_noError(t, actualErr)
	if !ok || !reflect.DeepEqual(actualDiff.Diff, expectedDiff) {
		t.Fatalf("expected diff %+v, actual %+v ok=%t", expectedDiff, actualDiff.Diff, ok)
	}
}
