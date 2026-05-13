package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestLayout_GivenCachedNotifications_WhenRendering_ThenItShowsThemBeforeTheBackgroundRefreshFinishes(t *testing.T) {
	cachedNotification := given_cachedNotification("n-cached", "Cached notification")
	loader := &fakePullRequestDetailLoader{notifications: []githubcli.Notification{given_cachedNotification("n-fresh", "Fresh notification")}}
	cache := &fakePersistentPullRequestCache{notifications: []githubcli.Notification{cachedNotification}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	if !strings.Contains(notificationsView.Buffer(), "Cached notification") {
		t.Fatalf("expected notifications buffer to contain %q, actual %q", "Cached notification", notificationsView.Buffer())
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued notifications refresh, actual %d", len(asyncRunner.runs))
	}
}

func TestLayout_GivenCachedNotificationsAndBackgroundRefreshFailure_WhenRendering_ThenItKeepsTheCachedRowsVisible(t *testing.T) {
	cachedNotification := given_cachedNotification("n-cached", "Cached notification")
	loader := &fakePullRequestDetailLoader{notificationsErr: errors.New("boom")}
	cache := &fakePersistentPullRequestCache{notifications: []githubcli.Notification{cachedNotification}}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)

	then_noError(t, actualErr)
	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	if !strings.Contains(notificationsView.Buffer(), "Cached notification") {
		t.Fatalf("expected notifications buffer to contain %q after refresh failure, actual %q", "Cached notification", notificationsView.Buffer())
	}
	if strings.Contains(notificationsView.Buffer(), notificationsGenericErrorTitle) {
		t.Fatalf("expected cached notifications to stay visible instead of %q, actual %q", notificationsGenericErrorTitle, notificationsView.Buffer())
	}
}

func TestLoadNotifications_GivenAFreshLiveResult_WhenLoading_ThenItStoresTheResultInThePersistentCache(t *testing.T) {
	expected := []githubcli.Notification{given_cachedNotification("n-fresh", "Fresh notification")}
	loader := &fakePullRequestDetailLoader{notifications: expected}
	cache := &fakePersistentPullRequestCache{}
	subject := given_programWithTestGitHubDeps(NewModel(DefaultSeedData()), loader)
	subject.pullRequestCache = cache
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	subject.loadNotifications(gui)

	if !reflect.DeepEqual(cache.savedNotifications, expected) {
		t.Fatalf("expected cached notifications %+v, actual %+v", expected, cache.savedNotifications)
	}
}

func given_cachedNotification(id string, title string) githubcli.Notification {
	return githubcli.Notification{
		ID:         id,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		Reason:     "review_requested",
		Unread:     true,
		UpdatedAt:  "2026-05-08T16:53:11Z",
		Subject: githubcli.NotificationSubject{
			Type:  githubcli.NotificationSubjectTypePullRequest,
			Title: title,
			URL:   "https://api.github.com/repos/acme/widgets/pulls/42",
		},
	}
}
