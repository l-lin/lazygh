package tui

import (
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestNotificationRow_GivenDefaultRepositoryStyle_WhenBuilding_ThenItUsesAFullReferenceAndGreyReferenceSegment(t *testing.T) {
	notification := githubdomain.Notification{
		ID:         "thread-pr",
		Unread:     true,
		Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
		Subject: githubdomain.NotificationSubject{
			Type:  githubdomain.NotificationSubjectTypePullRequest,
			Title: "Ship notifications",
			URL:   "https://api.github.com/repos/acme/widgets/pulls/42",
		},
	}

	actual := notificationRow(notification)

	expected := iconNotificationUnread + " " + iconNotificationPullRequest + " acme/widgets#42 Ship notifications"
	if actual.Item.Title != expected {
		t.Fatalf("expected title %q, actual %q", expected, actual.Item.Title)
	}
	if len(actual.Item.TitleSegments) != 4 {
		t.Fatalf("expected 4 title segments, actual %d", len(actual.Item.TitleSegments))
	}
	if actual.Item.TitleSegments[2].Text != "acme/widgets#42" {
		t.Fatalf("expected reference segment %q, actual %q", "acme/widgets#42", actual.Item.TitleSegments[2].Text)
	}
	if actual.Item.TitleSegments[2].ForegroundHex != theme.PullRequestReferenceHex {
		t.Fatalf("expected reference segment foreground %q, actual %q", theme.PullRequestReferenceHex, actual.Item.TitleSegments[2].ForegroundHex)
	}
}

func TestNotificationRow_GivenShortRepositoryStyle_WhenBuilding_ThenItUsesAShortReferenceAndGreyReferenceSegment(t *testing.T) {
	notification := githubdomain.Notification{
		ID:         "thread-pr",
		Unread:     true,
		Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
		Subject: githubdomain.NotificationSubject{
			Type:  githubdomain.NotificationSubjectTypePullRequest,
			Title: "Ship notifications",
			URL:   "https://api.github.com/repos/acme/widgets/pulls/42",
		},
	}

	actual := notificationRowWithRepositoryStyle(appconfig.RepositoryStyleName, notification)

	expected := iconNotificationUnread + " " + iconNotificationPullRequest + " widgets#42 Ship notifications"
	if actual.Item.Title != expected {
		t.Fatalf("expected title %q, actual %q", expected, actual.Item.Title)
	}
	if len(actual.Item.TitleSegments) != 4 {
		t.Fatalf("expected 4 title segments, actual %d", len(actual.Item.TitleSegments))
	}
	if actual.Item.TitleSegments[2].Text != "widgets#42" {
		t.Fatalf("expected reference segment %q, actual %q", "widgets#42", actual.Item.TitleSegments[2].Text)
	}
	if actual.Item.TitleSegments[2].ForegroundHex != theme.PullRequestReferenceHex {
		t.Fatalf("expected reference segment foreground %q, actual %q", theme.PullRequestReferenceHex, actual.Item.TitleSegments[2].ForegroundHex)
	}
}

func TestNotificationListReference_GivenDefaultRepositoryStyle_WhenBuilding_ThenItUsesFullRepositoryLabels(t *testing.T) {
	testCases := []struct {
		name         string
		notification githubdomain.Notification
		expected     string
	}{
		{
			name: "pull request",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
				Subject:    githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypePullRequest, URL: "https://api.github.com/repos/acme/widgets/pulls/42"},
			},
			expected: "acme/widgets#42",
		},
		{
			name: "issue",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "opencode", NameWithOwner: "acme/opencode"},
				Subject:    githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypeIssue, URL: "https://api.github.com/repos/acme/opencode/issues/3235"},
			},
			expected: "acme/opencode#3235",
		},
		{
			name: "release",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "doctoboot", NameWithOwner: "acme/doctoboot"},
				Subject:    githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypeRelease, URL: "https://api.github.com/repos/acme/doctoboot/releases/317927281"},
			},
			expected: "acme/doctoboot",
		},
		{
			name: "fallback repository",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
				Subject:    githubdomain.NotificationSubject{Type: "Push"},
			},
			expected: "acme/widgets",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := notificationListReference(appconfig.RepositoryStyleOwnerName, testCase.notification)
			if actual != testCase.expected {
				t.Fatalf("expected list reference %q, actual %q", testCase.expected, actual)
			}
		})
	}
}

func TestNotificationListReference_GivenShortRepositoryStyle_WhenBuilding_ThenItUsesShortRepositoryLabels(t *testing.T) {
	testCases := []struct {
		name         string
		notification githubdomain.Notification
		expected     string
	}{
		{
			name: "pull request",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
				Subject:    githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypePullRequest, URL: "https://api.github.com/repos/acme/widgets/pulls/42"},
			},
			expected: "widgets#42",
		},
		{
			name: "issue",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "opencode", NameWithOwner: "acme/opencode"},
				Subject:    githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypeIssue, URL: "https://api.github.com/repos/acme/opencode/issues/3235"},
			},
			expected: "opencode#3235",
		},
		{
			name: "release",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "doctoboot", NameWithOwner: "acme/doctoboot"},
				Subject:    githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypeRelease, URL: "https://api.github.com/repos/acme/doctoboot/releases/317927281"},
			},
			expected: "doctoboot",
		},
		{
			name: "fallback repository",
			notification: githubdomain.Notification{
				Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
				Subject:    githubdomain.NotificationSubject{Type: "Push"},
			},
			expected: "widgets",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := notificationListReference(appconfig.RepositoryStyleName, testCase.notification)
			if actual != testCase.expected {
				t.Fatalf("expected list reference %q, actual %q", testCase.expected, actual)
			}
		})
	}
}

func TestNotificationDetailReference_GivenNotificationKinds_WhenBuilding_ThenItKeepsFullRepositoryLabels(t *testing.T) {
	notification := githubdomain.Notification{
		Repository: githubdomain.RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
		Subject:    githubdomain.NotificationSubject{Type: githubdomain.NotificationSubjectTypePullRequest, URL: "https://api.github.com/repos/acme/widgets/pulls/42"},
	}

	actual := notificationDetailReference(notification)
	expected := "acme/widgets#42"
	if actual != expected {
		t.Fatalf("expected detail reference %q, actual %q", expected, actual)
	}
}
