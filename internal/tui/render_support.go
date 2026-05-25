package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) detailViewContent() string {
	if program.reviewModeActive() {
		return program.reviewSessionDetailContent()
	}
	if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
		if result, ok := program.pullRequestDetailForSummary(summary); ok {
			if result.err != nil {
				return renderPullRequestDetailError(summary, result.err)
			}

			switch program.detailState.activeTab {
			case CommentsDetailTab:
				return program.renderCurrentPullRequestConversationsTab(summary, result.detail, program.detailState.wrapWidth)
			case CommitsDetailTab:
				return renderPullRequestCommitsTab(result.detail.Commits, program.markdownRenderer, program.detailState.wrapWidth)
			case ChangesDetailTab:
				return program.renderCurrentPullRequestChangesTab(summary, program.detailState.wrapWidth)
			default:
				header := renderPullRequestBrowserHeader(summary, result.detail)
				overview := program.renderCurrentPullRequestOverview(summary, result.detail, program.detailState.wrapWidth)
				content := renderPullRequestDescription(summary, result.detail, program.markdownRenderer, program.detailState.wrapWidth)
				return renderPullRequestBrowserDetailContent(header, overview, content, program.detailState.wrapWidth)
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
					return renderIssueDetail(repository, result.detail, program.markdownRenderer, program.detailState.wrapWidth)
				}
				return renderNotificationDetailLoading(notification, repository, program.loadingSpinnerFrame())
			}
			if repository, _, ok := notification.ReleaseIdentity(); ok {
				if result, ok := program.releaseDetailForNotification(notification); ok {
					if result.err != nil {
						return renderReleaseDetailError(notification, repository, result.err)
					}
					return renderReleaseDetail(repository, result.detail, program.markdownRenderer, program.detailState.wrapWidth)
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

func (program *Program) renderCurrentPullRequestChangesTab(summary githubdomain.PullRequest, width int) string {
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

func (program *Program) buildCurrentDetailDocument(width int) detailDocument {
	if !program.reviewModeActive() && program.detailState.activeTab == ChangesDetailTab {
		if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
			if result, ok := program.pullRequestDiffForSummary(summary); ok && result.err == nil {
				return newReviewDiffDetailDocument(program.currentPullRequestChangesRenderedRows(summary, result.data.Files, width), width)
			}
		}
	}
	return newDetailDocumentWithWrap(program.detailViewContent(), width, program.detailViewWraps())
}

func (program *Program) detailViewWraps() bool {
	return false
}

func (program *Program) currentConnectedUserLogin() string {
	return strings.TrimSpace(program.connectedUserLogin)
}

func (program *Program) currentConnectedUserName() string {
	return strings.TrimSpace(program.connectedUserName)
}

func (program *Program) syncDetailViewState(detailDocument detailDocument, viewportHeight int) {
	identity := program.currentDetailIdentity()
	if identity != program.detailState.lastIdentity {
		program.detailState.lastIdentity = identity
		program.detailState.viewState.reset()
	}

	program.detailState.viewState.sync(detailDocument, viewportHeight)
	program.detailState.viewState.syncSearch(detailDocument, program.model.DetailSearchQuery())
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
