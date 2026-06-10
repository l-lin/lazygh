package tui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
)

func (program *Program) currentDetailReadModel(width int) detailReadModel {
	return program.buildDetailReadModel(width, true)
}

func (program *Program) buildDetailReadModel(width int, includeDocumentInputs bool) detailReadModel {
	model := detailReadModel{width: maxInt(width, 1), detailHeaderSource: "Connected user"}
	if program == nil {
		return model
	}

	model.activeTab = program.detailState.activeTab
	model.commitDiffTab = program.detailState.commitDiffTab
	model.wordWrapEnabled = program.detailWordWrapEnabled()
	model.loadingSpinner = program.loadingSpinnerFrame()
	model.markdownRenderer = program.markdownRenderer
	model.connectedUserLogin = program.currentConnectedUserLogin()
	model.reviewSession = program.reviewSessionReadModel()

	if program.model != nil {
		model.currentSideFocus = program.model.currentSideFocus()
		model.detailHeaderSource = program.currentDetailHeaderSource()
		model.activePullRequestTab = program.model.ActivePullRequestTab()
		model.selectedPullRequestIndex = program.model.SelectedPullRequestIndex(model.activePullRequestTab)
		model.selectedNotificationIndex = program.model.SelectedNotificationIndex()
		model.selectedUserIndex = program.model.SelectedUserIndex()
		if item, ok := program.model.detailItem(); ok {
			model.fallbackItem = item
			model.fallbackItemKnown = true
			model.fallbackItemLoading = program.isPullRequestLoadingItem(item) || program.isNotificationLoadingItem(item)
		}
		if model.currentSideFocus == FocusNotificationsView {
			if notification, ok := program.model.SelectedNotification(); ok {
				model.notification = notification
				model.notificationKnown = true
				if result, ok := program.issueDetailForNotification(notification); ok {
					model.issueDetailResult = result
					model.issueDetailResultKnown = true
				}
				if result, ok := program.releaseDetailForNotification(notification); ok {
					model.releaseDetailResult = result
					model.releaseDetailResultKnown = true
				}
			}
		}
	}

	if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
		model.pullRequestSummary = summary
		model.pullRequestSummaryKnown = true
		if result, ok := program.pullRequestDetailForSummary(summary); ok {
			model.pullRequestDetailResult = result
			model.pullRequestDetailResultKnown = true
		}
		if model.activeTab == ChangesDetailTab || model.reviewSession.isActive() {
			if result, ok := program.pullRequestDiffForSummary(summary); ok {
				model.pullRequestDiffResult = result
				model.pullRequestDiffResultKnown = true
			}
		}
		pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
		if model.commitDiffTab.visibleForPullRequestKey(pullRequestKey) {
			if result, ok := program.commitDiffResultForTarget(pullRequestKey, model.commitDiffTab.commitOID); ok {
				model.commitDiffResult = result
				model.commitDiffResultKnown = true
			}
		}
	}

	if !includeDocumentInputs {
		return model
	}
	if model.reviewSession.isActive() {
		if !model.reviewSession.showsDescription() && !model.reviewSession.showsStoryChapter() {
			if selectedFile, ok := model.reviewSession.selectedDiffFile(); ok {
				model.reviewDiffDocument = program.currentReviewDiffDocument(selectedFile, model.width)
				model.reviewDiffDocumentKnown = true
			}
		}
		return model
	}
	if !model.pullRequestSummaryKnown || !model.pullRequestDetailResultKnown || model.pullRequestDetailResult.err != nil {
		return model
	}

	summary := model.pullRequestSummary
	detail := model.pullRequestDetailResult.detail
	switch model.activeTab {
	case CommentsDetailTab:
		model.pullRequestConversationDocument = program.currentPullRequestConversationDocument(summary, detail, model.width)
		model.pullRequestConversationKnown = true
	case ChangesDetailTab:
		if model.pullRequestDiffResultKnown && model.pullRequestDiffResult.err == nil {
			model.pullRequestChangesRenderedRows = program.currentPullRequestChangesRenderedRows(summary, model.pullRequestDiffResult.data.Files, model.width)
			model.pullRequestChangesKnown = true
		}
	case CommitChangesDetailTab:
		pullRequestKey := pullRequestDetailKey(summary.Repository, summary.Number)
		if model.commitDiffResultKnown && model.commitDiffResult.err == nil && pullRequestKey != "" {
			model.commitDiffRenderedRows = program.currentCommitDiffRenderedRows(pullRequestKey, model.commitDiffTab.commitOID, model.commitDiffResult.data.Files, model.width)
			model.commitDiffKnown = true
		}
	case DescriptionDetailTab:
		model.pullRequestOverview = program.renderCurrentPullRequestOverview(summary, detail, model.width)
		model.pullRequestOverviewKnown = true
	}
	return model
}

func (program *Program) currentDetailDocument(view *gocui.View) detailDocument {
	width := detailReadWidth(view, program.detailState.wrapWidth)
	if program.reviewModeActive() && !program.reviewSessionShowsDescription() && !program.reviewSessionShowsStoryChapter() {
		if selectedFile, ok := program.selectedReviewSessionDiffFile(); ok {
			return program.currentReviewDiffDocument(selectedFile, width)
		}
	}

	model := program.currentDetailReadModel(width)
	if cacheKey, ok := model.documentCacheKey(); ok {
		if document, ok := program.pullRequestDetailDocumentForKey(cacheKey); ok {
			return document
		}

		document := model.document()
		program.cachePullRequestDetailDocument(cacheKey, document)
		return document
	}

	return model.document()
}

func (program *Program) detailViewContent() string {
	return program.currentDetailReadModel(maxInt(program.detailState.wrapWidth, 1)).content()
}

func (program *Program) buildCurrentDetailDocument(width int) detailDocument {
	return program.currentDetailReadModel(width).document()
}

func (program *Program) currentDetailIdentity() string {
	return program.buildDetailReadModel(0, false).identity()
}

func detailReadWidth(view *gocui.View, fallback int) int {
	width := fallback
	if view != nil && view.InnerWidth() > 0 {
		width = view.InnerWidth()
	}
	if width < 1 {
		return 1
	}
	return width
}

func (program *Program) currentDetailHeaderSource() string {
	if program == nil || program.model == nil {
		return "Connected user"
	}

	switch program.model.currentSideFocus() {
	case FocusPullRequestsView:
		return fmt.Sprintf("%s tab", program.model.PullRequestTabLabel(program.model.ActivePullRequestTab()))
	case FocusNotificationsView:
		return "Notifications"
	default:
		return "Connected user"
	}
}
