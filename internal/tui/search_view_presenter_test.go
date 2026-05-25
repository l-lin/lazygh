package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestSearchViewPresenter_GivenActiveEditorAndLoadedNotifications_WhenResolvingPromptAndNotificationTitle_ThenItUsesTheSnapshot(t *testing.T) {
	subject := searchViewPresenter{
		searchText:   "Alpha",
		searchCursor: 2,
		notificationRows: []NotificationRow{
			{Notification: &githubdomain.Notification{}},
			{Notification: &githubdomain.Notification{}},
		},
	}

	if actual := subject.promptText(); actual != "Alpha" {
		t.Fatalf("expected prompt text %q, actual %q", "Alpha", actual)
	}
	if actual := subject.promptCursor(); actual != 2 {
		t.Fatalf("expected prompt cursor %d, actual %d", 2, actual)
	}
	if actual := subject.notificationsViewTitle(); actual != "Notifications (2)" {
		t.Fatalf("expected notifications title %q, actual %q", "Notifications (2)", actual)
	}
}

func TestSearchViewPresenter_GivenBrowserPullRequestDetailWithTabs_WhenResolvingDetailTitle_ThenItHidesTheFallbackTitle(t *testing.T) {
	subject := searchViewPresenter{
		mode:                       ScreenModeBrowser,
		mainContentKind:            MainContentKindPullRequestDetail,
		showsPullRequestDetailTabs: true,
	}

	if actual := subject.detailViewTitle(); actual != "" {
		t.Fatalf("expected hidden detail title %q, actual %q", "", actual)
	}
}

func TestSearchViewPresenter_GivenStoryReviewChapterContext_WhenResolvingPaneTitles_ThenItUsesTheFocusedModeSnapshot(t *testing.T) {
	subject := searchViewPresenter{
		mode:            ScreenModeStoryReview,
		mainContentKind: MainContentKindStoryChapter,
	}

	if actual := subject.userViewTitle(); actual != reviewModeMetadataTitle {
		t.Fatalf("expected user title %q, actual %q", reviewModeMetadataTitle, actual)
	}
	if actual := subject.pullRequestsViewTitle(); actual != reviewModeChaptersTitle {
		t.Fatalf("expected pull requests title %q, actual %q", reviewModeChaptersTitle, actual)
	}
	if actual := subject.detailViewTitle(); actual != reviewModeChapterTitle {
		t.Fatalf("expected detail title %q, actual %q", reviewModeChapterTitle, actual)
	}
}
