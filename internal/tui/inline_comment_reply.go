package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	pullRequestInlineCommentReplyEditorTitle    = "Reply to inline comment"
	pullRequestInlineCommentReplySuccessMessage = "Inline reply posted"
	inlineCommentReplyUnavailableMessage        = "Reply to inline comment unavailable here"
)

type pullRequestReviewThreadReplyTarget struct {
	repository    string
	number        int
	pendingReview string
	threadID      string
}

func (program *Program) currentInlineCommentReplyAction() (actionsPopupAction, bool) {
	if _, ok := program.selectedPullRequestReviewThreadReplyTarget(); !ok {
		return actionsPopupAction{}, false
	}
	return program.replyToInlineCommentAction(), true
}

func (program *Program) replyToInlineCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "reply-to-inline-comment",
		title:   pullRequestInlineCommentReplyEditorTitle,
		icon:    actionsPopupCommentOnPullRequestIcon,
		execute: program.executeReplyToInlineCommentAction,
	}
}

func (program *Program) executeReplyToInlineCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestReviewThreadReplyTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openInlineCommentReplyComposer(gui, target)
	})
}

func (program *Program) openInlineCommentReplyComposer(gui *gocui.Gui, target pullRequestReviewThreadReplyTarget) error {
	return program.openMultilineModalEditor(gui, pullRequestInlineCommentReplyEditorTitle, "", func(body string) error {
		return program.submitInlineCommentReply(target, body)
	}, reviewInlineCommentModalHeight)
}

func (program *Program) submitInlineCommentReply(target pullRequestReviewThreadReplyTarget, body string) error {
	if strings.TrimSpace(target.threadID) == "" {
		return errors.New("missing inline comment thread identity")
	}
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.AddPullRequestReviewThreadReply(target.pendingReview, target.threadID, body); err != nil {
		return err
	}

	program.optimisticallyAppendInlineCommentReply(target, body)
	program.setFeedback(FocusDetailView, pullRequestInlineCommentReplySuccessMessage)
	return nil
}

func (program *Program) selectedPullRequestReviewThreadReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	if program.reviewModeActive() {
		return program.selectedReviewInlineCommentReplyTarget()
	}
	return program.selectedBrowserInlineCommentReplyTarget()
}

func (program *Program) selectedBrowserInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	switch program.activeDetailTab {
	case CommentsDetailTab:
		return program.selectedBrowserConversationsInlineCommentReplyTarget()
	case ChangesDetailTab:
		return program.selectedBrowserChangesInlineCommentReplyTarget()
	default:
		return pullRequestReviewThreadReplyTarget{}, false
	}
}

func (program *Program) selectedBrowserConversationsInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, program.detailWrapWidth, program.detailViewState.cursor.line)
	if !ok || sectionAtCursor.section.inlineThread == nil || !sectionAtCursor.inBody {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	thread := *sectionAtCursor.section.inlineThread
	if _, ok := browserConversationInlineThreadCommentAtCursor(sectionAtCursor); !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	return pullRequestReviewThreadReplyTarget{
		repository: repository,
		number:     summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
	}, true
}

func (program *Program) selectedBrowserChangesInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	detailDocument := program.currentDetailDocument(nil)
	renderedRows := program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width)
	thread, _, ok := reviewDiffCommentAtCursor(renderedRows, detailDocument, program.detailViewState)
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	return pullRequestReviewThreadReplyTarget{
		repository: repository,
		number:     summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
	}, true
}

func (program *Program) selectedReviewInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	if !program.reviewModeActive() {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, program.detailWrapWidth)
	document := program.currentReviewDiffDocument(selectedFile, program.detailWrapWidth)
	thread, _, ok := reviewDiffCommentAtCursor(renderedRows, document, program.detailViewState)
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(program.reviewSession.summary.Repository))
	if repository == "" || program.reviewSession.summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	return pullRequestReviewThreadReplyTarget{
		repository:    repository,
		number:        program.reviewSession.summary.Number,
		pendingReview: strings.TrimSpace(program.reviewSession.pendingReviewID),
		threadID:      strings.TrimSpace(thread.ID),
	}, true
}

func (program *Program) replyToInlineCommentShortcut(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return nil
	}
	if !program.inlineCommentReplyShortcutContextActive() {
		return nil
	}

	target, ok := program.selectedPullRequestReviewThreadReplyTarget()
	if !ok {
		program.setFeedback(FocusDetailView, inlineCommentReplyUnavailableMessage)
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}

	return program.openInlineCommentReplyComposer(gui, target)
}

func (program *Program) inlineCommentReplyShortcutContextActive() bool {
	if program.model.Focus() != FocusDetailView {
		return false
	}
	if program.reviewModeActive() {
		return true
	}
	if !program.shouldShowPullRequestDetailTabs() {
		return false
	}
	switch program.activeDetailTab {
	case CommentsDetailTab, ChangesDetailTab:
		return true
	default:
		return false
	}
}

func (program *Program) inlineCommentReplyShortcutAvailable() bool {
	if !program.inlineCommentReplyShortcutContextActive() {
		return false
	}
	_, ok := program.selectedPullRequestReviewThreadReplyTarget()
	return ok
}
