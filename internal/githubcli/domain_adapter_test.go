package githubcli

import (
	"reflect"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDomainAdapters_GivenTransportPayloads_WhenMappingToProviderNeutralModels_ThenTheyStayLosslessForShippedFeatures(t *testing.T) {
	summary := PullRequest{
		ID:             "PR_42",
		Title:          "Ship notifications",
		Number:         42,
		Repository:     Repository{Name: "widgets", NameWithOwner: "acme/widgets"},
		URL:            "https://github.com/acme/widgets/pull/42",
		Body:           "Ship it",
		State:          "OPEN",
		UpdatedAt:      "2026-05-05T10:00:00Z",
		ReviewRequests: []PullRequestReviewRequest{{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}},
	}
	detail := PullRequestDetail{
		ID:                   "PR_42",
		Title:                "Ship notifications",
		Number:               42,
		URL:                  summary.URL,
		Body:                 "Rich body",
		Author:               &PullRequestAuthor{Login: "octocat"},
		State:                "OPEN",
		BaseRefName:          "main",
		HeadRefName:          "feature/cache",
		Comments:             []PullRequestComment{{ID: "comment-1", Body: "Looks good", Author: &PullRequestCommentAuthor{Login: "reviewer-one"}, ReactionGroups: []ReactionGroup{{Content: ReactionContentEyes, TotalCount: 2}}}},
		StatusCheckRollup:    []PullRequestStatusCheck{{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "SUCCESS", Link: "https://github.com/acme/widgets/actions/runs/42"}},
		InlineCommentThreads: []PullRequestReviewThread{{ID: "thread-1", Path: "main.go", Line: 12, Comments: []PullRequestComment{{ID: "comment-2", Body: "ship it"}}}},
	}
	diff := PullRequestDiff{
		UnifiedDiff: "diff --git a/main.go b/main.go\n+cached",
		Files:       []PullRequestDiffFile{{Path: "main.go", ChangeType: "modified", Additions: 1, TeamOwners: []string{"@acme/maintainers"}}},
		Threads:     []PullRequestReviewThread{{ID: "thread-1", Path: "main.go", Line: 12, Comments: []PullRequestComment{{ID: "comment-2", Body: "ship it"}}}},
	}
	notification := Notification{ID: "1001", Unread: true, Reason: "review_requested", Repository: Repository{NameWithOwner: "acme/widgets"}, Subject: NotificationSubject{Title: "Ship notifications", Type: NotificationSubjectTypePullRequest, URL: "https://api.github.com/repos/acme/widgets/pulls/42"}}

	actualSummary := ToDomainPullRequestSummary(summary)
	actualDetail := ToDomainPullRequestDetail(detail)
	actualDiff := ToDomainPullRequestDiff(diff)
	actualNotification := ToDomainNotification(notification)

	if !reflect.DeepEqual(PullRequestSummaryFromDomain(actualSummary), summary) {
		t.Fatalf("expected pull request summary round-trip %+v, actual %+v", summary, PullRequestSummaryFromDomain(actualSummary))
	}
	if !reflect.DeepEqual(PullRequestDetailFromDomain(actualDetail), detail) {
		t.Fatalf("expected pull request detail round-trip %+v, actual %+v", detail, PullRequestDetailFromDomain(actualDetail))
	}
	if !reflect.DeepEqual(PullRequestDiffFromDomain(actualDiff), diff) {
		t.Fatalf("expected pull request diff round-trip %+v, actual %+v", diff, PullRequestDiffFromDomain(actualDiff))
	}
	if !reflect.DeepEqual(NotificationFromDomain(actualNotification), notification) {
		t.Fatalf("expected notification round-trip %+v, actual %+v", notification, NotificationFromDomain(actualNotification))
	}

	expectedSummary := githubdomain.PullRequestSummary{ID: "PR_42", Title: "Ship notifications", Number: 42, Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"}, URL: summary.URL, Body: "Ship it", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z", ReviewRequests: []githubdomain.PullRequestReviewRequest{{RequestedReviewer: githubdomain.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}}}
	if !reflect.DeepEqual(actualSummary, expectedSummary) {
		t.Fatalf("expected provider-neutral pull request summary %+v, actual %+v", expectedSummary, actualSummary)
	}
	if actualNotification.Subject.Type != githubdomain.NotificationSubjectTypePullRequest {
		t.Fatalf("expected notification subject type %q, actual %q", githubdomain.NotificationSubjectTypePullRequest, actualNotification.Subject.Type)
	}
}
