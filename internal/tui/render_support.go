package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) detailViewContent() string {
	if program.reviewSession.active {
		return program.reviewSessionDetailContent()
	}
	if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
		if result, ok := program.pullRequestDetailForSummary(summary); ok {
			if result.err != nil {
				return renderPullRequestDetailError(summary, result.err)
			}

			switch program.activeDetailTab {
			case CommentsDetailTab:
				return program.renderCurrentPullRequestConversationsTab(summary, result.detail, program.detailWrapWidth)
			case CommitsDetailTab:
				return renderPullRequestCommitsTab(result.detail.Commits, program.markdownRenderer, program.detailWrapWidth)
			case ChangesDetailTab:
				return program.renderCurrentPullRequestChangesTab(summary, program.detailWrapWidth)
			default:
				header := renderPullRequestBrowserHeader(summary, result.detail)
				overview := program.renderCurrentPullRequestOverview(summary, result.detail, program.detailWrapWidth)
				content := renderPullRequestDescription(summary, result.detail, program.markdownRenderer, program.detailWrapWidth)
				return renderPullRequestBrowserDetailContent(header, overview, content, program.detailWrapWidth)
			}
		}
		return renderPullRequestDetailLoading(summary, program.loadingSpinnerFrame())
	}
	if program.model.currentSideFocus() == FocusNotificationsView {
		if notification, ok := program.model.SelectedNotification(); ok {
			if repository, _, ok := notification.IssueIdentity(); ok {
				if result, ok := program.issueDetailForNotification(notification); ok {
					if result.err != nil {
						return renderIssueDetailError(notification, repository, result.err)
					}
					return renderIssueDetail(repository, result.detail, program.markdownRenderer, program.detailWrapWidth)
				}
				return renderNotificationDetailLoading(notification, repository, program.loadingSpinnerFrame())
			}
			if repository, _, ok := notification.ReleaseIdentity(); ok {
				if result, ok := program.releaseDetailForNotification(notification); ok {
					if result.err != nil {
						return renderReleaseDetailError(notification, repository, result.err)
					}
					return renderReleaseDetail(repository, result.detail, program.markdownRenderer, program.detailWrapWidth)
				}
				return renderNotificationDetailLoading(notification, repository, program.loadingSpinnerFrame())
			}
			return renderUnsupportedNotificationDetail(notification)
		}
	}

	item, ok := program.model.detailItem()
	if !ok {
		return "No detail available."
	}

	return program.fallbackDetailViewContent(item)
}

func (program *Program) renderCurrentPullRequestChangesTab(summary githubcli.PullRequest, width int) string {
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok {
		return strings.TrimSpace(program.loadingSpinnerFrame())
	}
	if result.err != nil {
		return renderPullRequestChangesTabError(result.err)
	}
	return renderPullRequestChangesRows(program.currentPullRequestChangesRenderedRows(summary, result.data.Files, width))
}

func (program *Program) fallbackDetailViewContent(item Item) string {
	if program.isPullRequestLoadingItem(item) || program.isNotificationLoadingItem(item) {
		return program.loadingSpinnerFrame()
	}

	header := program.detailHeader(item)
	body := strings.TrimSpace(item.Detail)
	if body == "" {
		body = "No description available. Even the dummy data is disappointed."
	}

	return renderPullRequestDetailContent(header, body)
}

func (program *Program) detailHeader(item Item) string {
	source := "Connected user"
	switch program.model.currentSideFocus() {
	case FocusPullRequestsView:
		source = fmt.Sprintf("%s tab", program.model.PullRequestTabLabel(program.model.ActivePullRequestTab()))
	case FocusNotificationsView:
		source = "Notifications"
	}

	return fmt.Sprintf("%s\n%s", source, program.displayItemTitle(item))
}

func (program *Program) currentDetailDocument(view *gocui.View) detailDocument {
	width := program.detailWrapWidth
	if view != nil && view.InnerWidth() > 0 {
		width = view.InnerWidth()
	}
	if width < 1 {
		width = 1
	}

	if program.reviewSession.active && !program.reviewSessionShowsDescription() && !program.reviewSessionShowsStoryChapter() {
		if selectedFile, ok := program.selectedReviewSessionDiffFile(); ok {
			return program.currentReviewDiffDocument(selectedFile, width)
		}
	}

	if cacheKey, ok := program.currentPullRequestDetailDocumentCacheKey(width); ok {
		if document, ok := program.pullRequestDetailDocumentForKey(cacheKey); ok {
			return document
		}

		document := newDetailDocumentWithWrap(program.detailViewContent(), width, program.detailViewWraps())
		program.cachePullRequestDetailDocument(cacheKey, document)
		return document
	}

	return newDetailDocumentWithWrap(program.detailViewContent(), width, program.detailViewWraps())
}

func (program *Program) detailViewWraps() bool {
	return false
}

func (program *Program) currentConnectedUserLogin() string {
	return strings.TrimSpace(program.connectedUserLogin)
}

func (program *Program) syncDetailViewState(detailDocument detailDocument, viewportHeight int) {
	identity := program.currentDetailIdentity()
	if identity != program.lastDetailIdentity {
		program.lastDetailIdentity = identity
		program.detailViewState.reset()
	}

	program.detailViewState.sync(detailDocument, viewportHeight)
	program.detailViewState.syncSearch(detailDocument, program.model.DetailSearchQuery())
}

func (program *Program) shouldHighlightSelection(focus Focus, selectable bool) bool {
	if !selectable {
		return false
	}

	if program.model.Focus() == focus {
		return true
	}

	return program.model.Focus() == FocusDetailView && program.model.currentSideFocus() == focus
}

func searchNoMatchesMessage(query string) string {
	return fmt.Sprintf("No matches for %q.", strings.TrimSpace(query))
}

func (program *Program) layoutContentHeight(maxY int) int {
	if maxY < 1 {
		return 1
	}
	if maxY > 1 {
		return maxY - 1
	}
	return maxY
}
