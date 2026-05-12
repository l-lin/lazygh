package cache

import (
	"path/filepath"
	"reflect"
	"testing"

	githubcli "github.com/l-lin/lazygh/internal/github"
)

func TestNotificationDoneStore_GivenReopenedStore_WhenFilteringPreviouslyHiddenThread_ThenItKeepsTheDoneNotificationHidden(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "notification-done.json")
	subject, actualErr := OpenNotificationDoneStore(storePath)
	then_noErrorCache(t, actualErr)

	hiddenNotification := given_notificationDoneStoreNotification("n-hidden", "2026-05-09T10:00:00Z")
	visibleNotification := given_notificationDoneStoreNotification("n-visible", "2026-05-09T11:00:00Z")
	actualErr = subject.HideNotifications([]githubcli.Notification{hiddenNotification})
	then_noErrorCache(t, actualErr)

	reopened, actualErr := OpenNotificationDoneStore(storePath)
	then_noErrorCache(t, actualErr)
	actual := reopened.FilterNotifications([]githubcli.Notification{hiddenNotification, visibleNotification})

	expected := []githubcli.Notification{visibleNotification}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected visible notifications %+v, actual %+v", expected, actual)
	}
}

func TestNotificationDoneStore_GivenNewerThreadActivity_WhenFiltering_ThenItShowsTheNotificationAgain(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "notification-done.json")
	subject, actualErr := OpenNotificationDoneStore(storePath)
	then_noErrorCache(t, actualErr)

	hiddenNotification := given_notificationDoneStoreNotification("n-hidden", "2026-05-09T10:00:00Z")
	actualErr = subject.HideNotifications([]githubcli.Notification{hiddenNotification})
	then_noErrorCache(t, actualErr)

	refreshedNotification := given_notificationDoneStoreNotification("n-hidden", "2026-05-09T10:05:00Z")
	actual := subject.FilterNotifications([]githubcli.Notification{refreshedNotification})

	expected := []githubcli.Notification{refreshedNotification}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected visible notifications %+v, actual %+v", expected, actual)
	}
}

func given_notificationDoneStoreNotification(threadID string, updatedAt string) githubcli.Notification {
	return githubcli.Notification{
		ID:        threadID,
		UpdatedAt: updatedAt,
		Repository: githubcli.Repository{
			NameWithOwner: "acme/widgets",
		},
		Subject: githubcli.NotificationSubject{
			Type:  githubcli.NotificationSubjectTypePullRequest,
			Title: "Tracked notification",
			URL:   "https://api.github.com/repos/acme/widgets/pulls/42",
		},
	}
}

func then_noErrorCache(t *testing.T, actualErr error) {
	t.Helper()
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
