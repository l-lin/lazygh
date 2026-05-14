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
		OutOfDateWithBase:    true,
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
	issue := IssueDetail{Title: "Broken inbox", Number: 7, URL: "https://github.com/acme/widgets/issues/7", Body: "Fix it", Author: &PullRequestAuthor{Login: "octocat"}, State: "open", Labels: []PullRequestLabel{{Name: "bug"}}, Assignees: []PullRequestAuthor{{Login: "reviewer-one"}}, Comments: 3}
	release := ReleaseDetail{Name: "v1.2.3", TagName: "v1.2.3", URL: "https://github.com/acme/widgets/releases/tag/v1.2.3", Body: "Release notes", PreRelease: true, Author: &PullRequestAuthor{Login: "release-bot"}}
	connectedUser := ConnectedUser{Login: "octocat", Name: "Octo Cat", Bio: "Mascot", Company: "GitHub", Location: "The Internet", PublicRepos: 8, Followers: 42, URL: "https://github.com/octocat"}
	buildRunJobs := []PullRequestBuildRunJob{{DatabaseID: 99, Name: "linux", Status: "completed", Conclusion: "success", URL: "https://github.com/acme/widgets/actions/runs/42/job/99"}}
	bulkReadResult := NotificationBulkReadResult{Accepted: true}

	actualSummary := ToDomainPullRequestSummary(summary)
	actualDetail := ToDomainPullRequestDetail(detail)
	actualDiff := ToDomainPullRequestDiff(diff)
	actualNotification := ToDomainNotification(notification)
	actualIssue := ToDomainIssueDetail(issue)
	actualRelease := ToDomainReleaseDetail(release)
	actualConnectedUser := ToDomainConnectedUser(connectedUser)
	actualBuildRunJobs := ToDomainPullRequestBuildRunJobs(buildRunJobs)
	actualBulkReadResult := ToDomainNotificationBulkReadResult(bulkReadResult)

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
	if !reflect.DeepEqual(IssueDetailFromDomain(actualIssue), issue) {
		t.Fatalf("expected issue detail round-trip %+v, actual %+v", issue, IssueDetailFromDomain(actualIssue))
	}
	if !reflect.DeepEqual(ReleaseDetailFromDomain(actualRelease), release) {
		t.Fatalf("expected release detail round-trip %+v, actual %+v", release, ReleaseDetailFromDomain(actualRelease))
	}
	if !reflect.DeepEqual(ConnectedUserFromDomain(actualConnectedUser), connectedUser) {
		t.Fatalf("expected connected user round-trip %+v, actual %+v", connectedUser, ConnectedUserFromDomain(actualConnectedUser))
	}
	if !reflect.DeepEqual(PullRequestBuildRunJobsFromDomain(actualBuildRunJobs), buildRunJobs) {
		t.Fatalf("expected build run jobs round-trip %+v, actual %+v", buildRunJobs, PullRequestBuildRunJobsFromDomain(actualBuildRunJobs))
	}
	if !reflect.DeepEqual(NotificationBulkReadResultFromDomain(actualBulkReadResult), bulkReadResult) {
		t.Fatalf("expected bulk read result round-trip %+v, actual %+v", bulkReadResult, NotificationBulkReadResultFromDomain(actualBulkReadResult))
	}

	expectedSummary := githubdomain.PullRequestSummary{ID: "PR_42", Title: "Ship notifications", Number: 42, Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"}, URL: summary.URL, Body: "Ship it", State: "OPEN", UpdatedAt: "2026-05-05T10:00:00Z", ReviewRequests: []githubdomain.PullRequestReviewRequest{{RequestedReviewer: githubdomain.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}}}
	if !reflect.DeepEqual(actualSummary, expectedSummary) {
		t.Fatalf("expected provider-neutral pull request summary %+v, actual %+v", expectedSummary, actualSummary)
	}
	if actualNotification.Subject.Type != githubdomain.NotificationSubjectTypePullRequest {
		t.Fatalf("expected notification subject type %q, actual %q", githubdomain.NotificationSubjectTypePullRequest, actualNotification.Subject.Type)
	}
	if !actualBulkReadResult.Accepted {
		t.Fatal("expected provider-neutral notification bulk-read result to preserve acceptance")
	}
	if len(actualBuildRunJobs) != 1 || actualBuildRunJobs[0].DatabaseID != 99 {
		t.Fatalf("expected provider-neutral build run jobs %+v, actual %+v", []int{99}, actualBuildRunJobs)
	}
}
