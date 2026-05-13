package tui

import (
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"testing"
)

func TestRefactorGuard_GivenBrowserAndNotificationFiles_WhenScanning_ThenTheyDoNotImportGithubcli(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGoFiles(t, ".", []string{"github.com/l-lin/lazygh/internal/githubcli"})
	allowedPaths := map[string]bool{}
	browserAndNotificationPaths := []string{
		"connected_user.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"model.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"notification_detail_loader.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"notification_loading.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"notifications.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"persistent_notification_list_cache.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"persistent_pull_request_cache_helpers.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"persistent_pull_request_list_cache.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"program.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"program_loading.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_assignee.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_comment_optimistic.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_detail.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_detail_loader.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_detail_render.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_edit_optimistic.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_request_stage_merge.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"pull_requests.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"view_url.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
		"workflow_stores.go contains \"github.com/l-lin/lazygh/internal/githubcli\"",
	}
	for _, path := range browserAndNotificationPaths {
		allowedPaths[path] = true
	}

	remainingMatches := make([]string, 0)
	for _, match := range actualMatches {
		if allowedPaths[match] {
			remainingMatches = append(remainingMatches, match)
		}
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected browser and notification files to stop importing githubcli, actual %v", remainingMatches)
	}
}

func TestPullRequestRow_GivenDomainSummary_WhenBuilding_ThenItKeepsTheProviderNeutralSummary(t *testing.T) {
	summary := githubdomain.PullRequest{
		Title:      "Ship notifications",
		Number:     42,
		Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		State:      "OPEN",
	}

	actual := pullRequestRow(summary)

	if actual.Summary == nil || actual.Summary.Repository.NameWithOwner != "acme/widgets" {
		t.Fatalf("expected provider-neutral pull request summary, actual %+v", actual.Summary)
	}
}

func TestNotificationRow_GivenDomainNotification_WhenBuilding_ThenItKeepsTheProviderNeutralNotification(t *testing.T) {
	notification := githubdomain.Notification{
		ID:        "thread-42",
		Unread:    true,
		Reason:    "review_requested",
		UpdatedAt: "2026-05-12T10:00:00Z",
		Repository: githubdomain.Repository{
			NameWithOwner: "acme/widgets",
		},
		Subject: githubdomain.NotificationSubject{Title: "Ship notifications", Type: githubdomain.NotificationSubjectTypePullRequest, URL: "https://api.github.com/repos/acme/widgets/pulls/42"},
	}

	actual := notificationRow(notification)

	if actual.Notification == nil || actual.Notification.ID != "thread-42" {
		t.Fatalf("expected provider-neutral notification, actual %+v", actual.Notification)
	}
}

func TestConnectedUserStateItem_GivenDomainUser_WhenRendering_ThenItFormatsTheProviderNeutralUser(t *testing.T) {
	actual := connectedUserStateItem(githubdomain.ConnectedUser{Login: "octocat", URL: "https://github.com/octocat"}, nil)

	if actual.Title != "@octocat" {
		t.Fatalf("expected title %q, actual %q", "@octocat", actual.Title)
	}
}
