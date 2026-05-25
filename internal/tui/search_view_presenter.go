package tui

import "fmt"

type searchViewPresenter struct {
	mode                       ScreenMode
	mainContentKind            MainContentKind
	showsPullRequestDetailTabs bool
	searchText                 string
	searchCursor               int
	notificationRows           []NotificationRow
}

func (presenter searchViewPresenter) promptText() string {
	return presenter.searchText
}

func (presenter searchViewPresenter) promptCursor() int {
	return presenter.searchCursor
}

func (presenter searchViewPresenter) userViewTitle() string {
	if presenter.mode != ScreenModeBrowser {
		return reviewModeMetadataTitle
	}
	return "[1]-" + detailAuthorIcon + " Connected user"
}

func (presenter searchViewPresenter) detailViewTitle() string {
	switch presenter.mainContentKind {
	case MainContentKindReviewDescription:
		return reviewModeDescriptionTitle
	case MainContentKindStoryChapter:
		return reviewModeChapterTitle
	case MainContentKindReviewDiff:
		return reviewModeDiffTitle
	default:
		if presenter.showsPullRequestDetailTabs {
			return ""
		}
		return "[0]-Detail"
	}
}

func (presenter searchViewPresenter) notificationsViewTitle() string {
	count, ok := presenter.notificationsCount()
	if !ok {
		return "Notifications"
	}
	return fmt.Sprintf("Notifications (%d)", count)
}

func (presenter searchViewPresenter) notificationsCount() (int, bool) {
	rows := presenter.notificationRows
	if len(rows) == 0 {
		return 0, false
	}
	if len(rows) == 1 && rows[0].Notification == nil {
		item := rows[0].Item
		if isNotificationLoadingPlaceholder(item) || isNotificationErrorPlaceholder(item) {
			return 0, false
		}
		if item.Title == notificationsEmptyTitle && item.Detail == notificationsEmptyDetail {
			return 0, true
		}
	}

	count := 0
	for _, row := range rows {
		if row.Notification == nil {
			return 0, false
		}
		count++
	}
	return count, true
}

func (presenter searchViewPresenter) pullRequestsViewTitle() string {
	switch presenter.mode {
	case ScreenModeStoryReview:
		return reviewModeChaptersTitle
	case ScreenModeReview:
		return reviewModeFilesTitle
	default:
		return ""
	}
}

func isNotificationLoadingPlaceholder(item Item) bool {
	return item.Title == notificationsLoadingTitle && item.Detail == notificationsLoadingDetail
}

func isNotificationErrorPlaceholder(item Item) bool {
	switch item.Title {
	case notificationsUnauthenticatedTitle, notificationsUnavailableTitle, notificationsGenericErrorTitle:
		return true
	default:
		return false
	}
}
