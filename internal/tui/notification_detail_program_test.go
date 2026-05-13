package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestNotificationDetailRouting_GivenPullRequestIssueAndReleaseNotifications_WhenChangingSelection_ThenDetailPaneTracksTheNotificationType(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusNotificationsView()
	model.SetNotificationRows(given_notificationRowsForDetailRouting())
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Add notifications",
				Number:      42,
				Body:        "Pull request body",
				State:       "OPEN",
				BaseRefName: "main",
				HeadRefName: "feature/notifications",
			},
		},
		issueDetails: map[string]githubcli.IssueDetail{
			"acme/opencode#3235": {
				Title:    "Support notifications in issue detail",
				Number:   3235,
				Body:     "Issue body",
				State:    "open",
				Comments: 7,
			},
		},
		releaseDetails: map[string]githubcli.ReleaseDetail{
			"acme/doctoboot#317927281": {
				Name:       "Notifications 3.5.0",
				TagName:    "v3.5.0",
				Body:       "Release notes",
				PreRelease: true,
			},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView := given_notificationDetailView(t, gui)
	if !strings.Contains(detailView.Buffer(), "Pull request body") {
		t.Fatalf("expected pull request detail body %q, actual %q", "Pull request body", detailView.Buffer())
	}
	if len(detailView.Tabs) != 4 {
		t.Fatalf("expected pull request notification detail to expose 4 browser tabs, actual %v", detailView.Tabs)
	}
	if detailView.TabIndex != 0 {
		t.Fatalf("expected the pull request notification detail tab index %d, actual %d", 0, detailView.TabIndex)
	}
	if !slices.Contains(loader.detailCalls, "acme/widgets#42") {
		t.Fatalf("expected pull request detail load for %q, actual %v", "acme/widgets#42", loader.detailCalls)
	}

	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, notificationsView)
	then_noError(t, actualErr)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	detailView = given_notificationDetailView(t, gui)
	if !strings.Contains(detailView.Buffer(), "Issue body") {
		t.Fatalf("expected issue detail body %q, actual %q", "Issue body", detailView.Buffer())
	}
	if len(detailView.Tabs) != 0 {
		t.Fatalf("expected issue detail to avoid pull request tabs, actual %v", detailView.Tabs)
	}
	if !slices.Contains(loader.issueDetailCalls, "acme/opencode#3235") {
		t.Fatalf("expected issue detail load for %q, actual %v", "acme/opencode#3235", loader.issueDetailCalls)
	}

	actualErr = subject.moveSelectionDown(gui, notificationsView)
	then_noError(t, actualErr)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	detailView = given_notificationDetailView(t, gui)
	if !strings.Contains(detailView.Buffer(), "Release notes") {
		t.Fatalf("expected release detail body %q, actual %q", "Release notes", detailView.Buffer())
	}
	if len(detailView.Tabs) != 0 {
		t.Fatalf("expected release detail to avoid pull request tabs, actual %v", detailView.Tabs)
	}
	if !slices.Contains(loader.releaseDetailCalls, "acme/doctoboot#317927281") {
		t.Fatalf("expected release detail load for %q, actual %v", "acme/doctoboot#317927281", loader.releaseDetailCalls)
	}
}

func TestNotificationDetailRouting_GivenAnUnsupportedNotificationType_WhenRendering_ThenDetailPaneShowsAnUnsupportedMessage(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusNotificationsView()
	model.SetNotificationRows([]NotificationRow{{
		Item: Item{Title: iconWarning + " acme/widgets push", Detail: "Repository: acme/widgets\nType: Unsupported (Push)"},
		Notification: &githubdomain.Notification{
			ID:         "n-unsupported",
			Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
			Subject:    githubdomain.NotificationSubject{Type: "Push", Title: "A push happened"},
		},
	}})
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView := given_notificationDetailView(t, gui)
	if !strings.Contains(detailView.Buffer(), "Unsupported notification type") {
		t.Fatalf("expected unsupported notification message, actual %q", detailView.Buffer())
	}
}

func TestNotificationDetailRouting_GivenAnIssueLoaderFailure_WhenRendering_ThenDetailPaneShowsAFallbackErrorDocument(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusNotificationsView()
	model.SetNotificationRows([]NotificationRow{given_issueNotificationRow()})
	loader := &fakePullRequestDetailLoader{issueDetailErrors: map[string]error{"acme/opencode#3235": githubcli.ErrUnavailable}}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView := given_notificationDetailView(t, gui)
	if !strings.Contains(detailView.Buffer(), "Could not load issue detail.") {
		t.Fatalf("expected issue detail error heading, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Repository: acme/opencode") {
		t.Fatalf("expected fallback notification detail, actual %q", detailView.Buffer())
	}
}

func given_notificationRowsForDetailRouting() []NotificationRow {
	return []NotificationRow{
		given_pullRequestNotificationRow(),
		given_issueNotificationRow(),
		given_releaseNotificationRow(),
	}
}

func given_pullRequestNotificationRow() NotificationRow {
	notification := githubcli.Notification{
		ID:         "n-pr",
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		Reason:     "review_requested",
		Unread:     true,
		UpdatedAt:  "2026-05-08T16:53:11Z",
		Subject: githubcli.NotificationSubject{
			Type:  githubcli.NotificationSubjectTypePullRequest,
			Title: "Add notifications",
			URL:   "https://api.github.com/repos/acme/widgets/pulls/42",
		},
	}
	return notificationRow(notification)
}

func given_issueNotificationRow() NotificationRow {
	notification := githubcli.Notification{
		ID:         "n-issue",
		Repository: githubcli.Repository{NameWithOwner: "acme/opencode"},
		Reason:     "manual",
		Unread:     false,
		UpdatedAt:  "2025-12-23T04:29:51Z",
		Subject: githubcli.NotificationSubject{
			Type:             githubcli.NotificationSubjectTypeIssue,
			Title:            "Support notifications in issue detail",
			URL:              "https://api.github.com/repos/acme/opencode/issues/3235",
			LatestCommentURL: "https://api.github.com/repos/acme/opencode/issues/comments/999",
		},
	}
	return notificationRow(notification)
}

func given_releaseNotificationRow() NotificationRow {
	notification := githubcli.Notification{
		ID:         "n-release",
		Repository: githubcli.Repository{NameWithOwner: "acme/doctoboot"},
		Reason:     "subscribed",
		Unread:     false,
		UpdatedAt:  "2026-05-05T16:38:09Z",
		Subject: githubcli.NotificationSubject{
			Type:  githubcli.NotificationSubjectTypeRelease,
			Title: "v3.5.0",
			URL:   "https://api.github.com/repos/acme/doctoboot/releases/317927281",
		},
	}
	return notificationRow(notification)
}

func given_notificationDetailView(t *testing.T, gui *gocui.Gui) *gocui.View {
	t.Helper()

	actual, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	return actual
}
