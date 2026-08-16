package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type detailReadModel struct {
	width                     int
	activeTab                 DetailTab
	wordWrapEnabled           bool
	loadingSpinner            string
	markdownRenderer          MarkdownRenderer
	connectedUserLogin        string
	currentSideFocus          Focus
	detailHeaderSource        string
	activePullRequestTab      PullRequestTab
	selectedPullRequestIndex  int
	selectedNotificationIndex int
	selectedUserIndex         int

	reviewSession           reviewSessionReadModel
	reviewDiffDocument      detailDocument
	reviewDiffDocumentKnown bool

	pullRequestSummary              githubdomain.PullRequest
	pullRequestSummaryKnown         bool
	pullRequestDetailResult         pullRequestDetailResult
	pullRequestDetailResultKnown    bool
	pullRequestDiffResult           pullRequestDiffResult
	pullRequestDiffResultKnown      bool
	pullRequestConversationDocument browserConversationDocument
	pullRequestConversationKnown    bool
	pullRequestOverview             string
	pullRequestOverviewKnown        bool
	pullRequestChangesRenderedRows  []reviewDiffRenderedRow
	pullRequestChangesKnown         bool
	commitDiffTab                   commitDiffTabState
	commitDiffResult                commitDiffResult
	commitDiffResultKnown           bool
	commitDiffRenderedRows          []reviewDiffRenderedRow
	commitDiffKnown                 bool

	notification             githubdomain.Notification
	notificationKnown        bool
	issueDetailResult        issueDetailResult
	issueDetailResultKnown   bool
	releaseDetailResult      releaseDetailResult
	releaseDetailResultKnown bool

	fallbackItem        Item
	fallbackItemKnown   bool
	fallbackItemLoading bool
}

func (model detailReadModel) identity() string {
	if model.reviewSession.isActive() {
		return model.reviewSession.detailIdentity()
	}

	if model.pullRequestSummaryKnown {
		if key := pullRequestDetailKey(model.pullRequestSummary.Repository, model.pullRequestSummary.Number); key != "" {
			if model.activeTab == CommitChangesDetailTab && model.commitDiffTab.visibleForPullRequestKey(key) {
				return fmt.Sprintf("pr:%s:commit:%s", key, strings.TrimSpace(model.commitDiffTab.commitOID))
			}
			if model.currentSideFocus == FocusNotificationsView {
				return fmt.Sprintf("notification-pr:%s:tab:%d", key, model.activeTab)
			}
			return fmt.Sprintf("pr:%s:tab:%d", key, model.activeTab)
		}
		return fmt.Sprintf("pr-state:%d:%d", model.activePullRequestTab, model.selectedPullRequestIndex)
	}
	if model.currentSideFocus == FocusNotificationsView {
		if model.notificationKnown {
			return fmt.Sprintf("notification:%s", strings.TrimSpace(model.notification.ID))
		}
		return fmt.Sprintf("notification-state:%d", model.selectedNotificationIndex)
	}
	return fmt.Sprintf("user:%d", model.selectedUserIndex)
}

func (model detailReadModel) documentCacheKey() (pullRequestDetailDocumentCacheKey, bool) {
	if model.width < 1 || !model.pullRequestSummaryKnown {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	pullRequestKey := pullRequestDetailKey(model.pullRequestSummary.Repository, model.pullRequestSummary.Number)
	if pullRequestKey == "" {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	if model.reviewSession.isActive() {
		if model.reviewSession.showsStoryChapter() {
			chapter, ok := model.reviewSession.selectedStoryChapter()
			if !ok {
				return pullRequestDetailDocumentCacheKey{}, false
			}
			chapterIdentity := firstNonEmpty(strings.TrimSpace(chapter.ID), strings.TrimSpace(chapter.Title))
			if chapterIdentity == "" {
				return pullRequestDetailDocumentCacheKey{}, false
			}
			return pullRequestDetailDocumentCacheKey{
				pullRequestKey: pullRequestKey,
				tab:            DescriptionDetailTab,
				variant:        fmt.Sprintf("story:%s:%s", strings.TrimSpace(model.reviewSession.pendingReviewID), chapterIdentity),
				width:          model.width,
			}, true
		}
		if !model.reviewSession.showsDescription() {
			return pullRequestDetailDocumentCacheKey{}, false
		}
	}
	if !model.pullRequestDetailResultKnown || model.pullRequestDetailResult.err != nil {
		return pullRequestDetailDocumentCacheKey{}, false
	}

	tab := model.activeTab
	if model.reviewSession.isActive() {
		tab = DescriptionDetailTab
	}
	cacheKey := pullRequestDetailDocumentCacheKey{pullRequestKey: pullRequestKey, tab: tab, width: model.width}
	if tab == CommitChangesDetailTab {
		cacheKey.variant = strings.TrimSpace(model.commitDiffTab.commitOID)
	}
	return cacheKey, true
}

func (model detailReadModel) content() string {
	if model.reviewSession.isActive() {
		return model.reviewSession.detailContent()
	}
	if model.pullRequestSummaryKnown {
		return model.pullRequestContent()
	}
	if model.currentSideFocus == FocusNotificationsView && model.notificationKnown {
		return model.notificationContent()
	}
	if !model.fallbackItemKnown {
		return "No detail available."
	}
	return model.fallbackContent()
}

func (model detailReadModel) document() detailDocument {
	if model.reviewSession.isActive() && !model.reviewSession.showsDescription() && !model.reviewSession.showsStoryChapter() {
		if model.reviewDiffDocumentKnown {
			return model.reviewDiffDocument
		}
		return newDetailDocumentWithWrap(model.content(), model.width, false)
	}
	if model.activeTab == CommentsDetailTab {
		return newBrowserConversationDetailDocument(model.content(), model.width, model.wordWrapEnabled)
	}
	if model.activeTab == ChangesDetailTab && model.pullRequestSummaryKnown && model.pullRequestDetailResultKnown && model.pullRequestDetailResult.err == nil && model.pullRequestDiffResultKnown && model.pullRequestDiffResult.err == nil && model.pullRequestChangesKnown {
		return newBrowserChangesDetailDocumentWithWordWrap(model.pullRequestChangesRenderedRows, model.width, model.wordWrapEnabled)
	}
	if model.activeTab == CommitChangesDetailTab && model.pullRequestSummaryKnown && model.pullRequestDetailResultKnown && model.pullRequestDetailResult.err == nil && model.commitDiffResultKnown && model.commitDiffResult.err == nil && model.commitDiffKnown {
		return newBrowserChangesDetailDocumentWithWordWrap(model.commitDiffRenderedRows, model.width, model.wordWrapEnabled)
	}
	return newDetailDocumentWithWrap(model.content(), model.width, false)
}

func (model detailReadModel) pullRequestContent() string {
	summary := model.pullRequestSummary
	if !model.pullRequestDetailResultKnown {
		return renderPullRequestDetailLoading(summary, model.loadingSpinner)
	}
	if model.pullRequestDetailResult.err != nil {
		return renderPullRequestDetailError(summary, model.pullRequestDetailResult.err)
	}

	detail := model.pullRequestDetailResult.detail
	switch model.activeTab {
	case CommentsDetailTab:
		if !model.pullRequestConversationKnown || len(model.pullRequestConversationDocument.sections) == 0 {
			return "No comments yet."
		}
		return model.pullRequestConversationDocument.text
	case CommitsDetailTab:
		return renderPullRequestCommitsTabWithWordWrapForSummary(summary, detail.Commits, model.markdownRenderer, model.width, model.wordWrapEnabled)
	case ChangesDetailTab:
		return model.pullRequestChangesContent()
	case CommitChangesDetailTab:
		return model.commitDiffContent()
	default:
		header := renderPullRequestBrowserHeader(summary, detail)
		overview := ""
		if model.pullRequestOverviewKnown {
			overview = model.pullRequestOverview
		}
		content := renderPullRequestDescriptionWithWordWrap(summary, detail, model.markdownRenderer, model.width, model.wordWrapEnabled)
		return renderPullRequestBrowserDetailContent(header, overview, content, model.width)
	}
}

func (model detailReadModel) pullRequestChangesContent() string {
	if !model.pullRequestDiffResultKnown {
		return strings.TrimSpace(model.loadingSpinner)
	}
	if model.pullRequestDiffResult.err != nil {
		return renderPullRequestChangesTabError(model.pullRequestDiffResult.err)
	}
	return renderPullRequestChangesRows(model.pullRequestChangesRenderedRows)
}

func (model detailReadModel) commitDiffContent() string {
	if !model.commitDiffResultKnown {
		return strings.TrimSpace(model.loadingSpinner)
	}
	if model.commitDiffResult.err != nil {
		return renderPullRequestChangesTabError(model.commitDiffResult.err)
	}
	return renderPullRequestChangesRows(model.commitDiffRenderedRows)
}

func (model detailReadModel) notificationContent() string {
	notification := model.notification
	if repository, _, ok := notification.IssueIdentity(); ok {
		if model.issueDetailResultKnown {
			if model.issueDetailResult.err != nil {
				return renderIssueDetailError(notification, repository, model.issueDetailResult.err)
			}
			return renderIssueDetailWithWordWrap(repository, model.issueDetailResult.detail, model.markdownRenderer, model.width, model.wordWrapEnabled)
		}
		return renderNotificationDetailLoading(notification, repository, model.loadingSpinner)
	}
	if repository, _, ok := notification.ReleaseIdentity(); ok {
		if model.releaseDetailResultKnown {
			if model.releaseDetailResult.err != nil {
				return renderReleaseDetailError(notification, repository, model.releaseDetailResult.err)
			}
			return renderReleaseDetailWithWordWrap(repository, model.releaseDetailResult.detail, model.markdownRenderer, model.width, model.wordWrapEnabled)
		}
		return renderNotificationDetailLoading(notification, repository, model.loadingSpinner)
	}
	return renderUnsupportedNotificationDetail(notification)
}

func (model detailReadModel) fallbackContent() string {
	if model.fallbackItemLoading {
		return model.loadingSpinner
	}
	body := strings.TrimSpace(model.fallbackItem.Detail)
	if body == "" {
		body = "No description available. Even the dummy data is disappointed."
	}
	header := detailHeaderText(model.detailHeaderSource, detailDisplayItemTitle(model.fallbackItem))
	return renderPullRequestDetailContent(header, body)
}

func detailDisplayItemTitle(item Item) string {
	return item.Title
}

func detailHeaderText(source string, title string) string {
	return fmt.Sprintf("%s\n%s", source, title)
}
