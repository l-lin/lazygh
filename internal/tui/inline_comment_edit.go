package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	inlineCommentUpdateEditorTitle     = "Edit inline comment"
	inlineCommentDeleteActionTitle     = "Delete inline comment"
	inlineCommentUpdatedSuccessMessage = "Inline comment updated"
	inlineCommentDeletedSuccessMessage = "Inline comment deleted"
)

type pullRequestReviewCommentActionTarget struct {
	repository string
	number     int
	commentID  string
	body       string
}

func (program *Program) currentInlineCommentEditActions() []actionsPopupAction {
	if _, ok := program.selectedPullRequestReviewCommentActionTarget(); !ok {
		return nil
	}
	return []actionsPopupAction{program.updateInlineCommentAction(), program.deleteInlineCommentAction()}
}

func (program *Program) updateInlineCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "update-inline-comment",
		title:   inlineCommentUpdateEditorTitle,
		icon:    actionsPopupEditPullRequestIcon,
		execute: program.executeUpdateInlineCommentAction,
	}
}

func (program *Program) deleteInlineCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "delete-inline-comment",
		title:   inlineCommentDeleteActionTitle,
		icon:    actionsPopupDeleteInlineCommentIcon,
		execute: program.executeDeleteInlineCommentAction,
	}
}

func (program *Program) executeUpdateInlineCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestReviewCommentActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	wasVisible := program.modalEditorVisible()
	err := program.openMultilineModalEditor(gui, inlineCommentUpdateEditorTitle, target.body, func(body string) error {
		return program.submitInlineCommentUpdate(target, body)
	}, reviewInlineCommentModalHeight)
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) executeDeleteInlineCommentAction(_ *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestReviewCommentActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.deleteInlineComment(target); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) submitInlineCommentUpdate(target pullRequestReviewCommentActionTarget, body string) error {
	if strings.TrimSpace(target.commentID) == "" {
		return errors.New("missing inline comment identity")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.UpdatePullRequestReviewComment(target.commentID, body); err != nil {
		return err
	}

	program.optimisticallyUpdateReviewComment(target, body)
	program.setFeedback(FocusDetailView, inlineCommentUpdatedSuccessMessage)
	return nil
}

func (program *Program) deleteInlineComment(target pullRequestReviewCommentActionTarget) error {
	if strings.TrimSpace(target.commentID) == "" {
		return errors.New("missing inline comment identity")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.reviewMutations.DeletePullRequestReviewComment(target.commentID); err != nil {
		return err
	}

	program.optimisticallyDeleteReviewComment(target)
	program.setFeedback(FocusDetailView, inlineCommentDeletedSuccessMessage)
	return nil
}

func (program *Program) selectedPullRequestReviewCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReviewCommentActionTarget{}, false
	}
	if program.reviewModeActive() {
		return program.selectedReviewDiffInlineCommentActionTarget()
	}
	return program.selectedBrowserInlineCommentActionTarget()
}

func (program *Program) selectedBrowserInlineCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() {
		return pullRequestReviewCommentActionTarget{}, false
	}

	switch program.activeDetailTab {
	case CommentsDetailTab:
		return program.selectedBrowserCommentsInlineCommentActionTarget()
	case ChangesDetailTab:
		return program.selectedBrowserChangesInlineCommentActionTarget()
	default:
		return pullRequestReviewCommentActionTarget{}, false
	}
}

func (program *Program) selectedBrowserCommentsInlineCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReviewCommentActionTarget{}, false
	}

	sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, program.detailWrapWidth, program.detailViewState.cursor.line)
	if !ok || sectionAtCursor.section.inlineThread == nil || !sectionAtCursor.inBody {
		return pullRequestReviewCommentActionTarget{}, false
	}

	threadComment, ok := browserConversationInlineThreadCommentAtCursor(sectionAtCursor)
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}
	target, ok := pullRequestInlineThreadCommentActionTarget(threadComment)
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}
	target.repository = strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	target.number = summary.Number
	if target.repository == "" || target.number <= 0 || !hasUsablePullRequestMutationID(target.commentID) {
		return pullRequestReviewCommentActionTarget{}, false
	}
	return target, true
}

func (program *Program) selectedBrowserChangesInlineCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}
	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReviewCommentActionTarget{}, false
	}

	detailDocument := program.currentDetailDocument(nil)
	renderedRows := program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width)
	_, comment, ok := reviewDiffCommentAtCursor(renderedRows, detailDocument, program.detailViewState)
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 || !hasUsablePullRequestMutationID(comment.ID) {
		return pullRequestReviewCommentActionTarget{}, false
	}
	return pullRequestReviewCommentActionTarget{
		repository: repository,
		number:     summary.Number,
		commentID:  strings.TrimSpace(comment.ID),
		body:       comment.Body,
	}, true
}

func (program *Program) selectedReviewDiffInlineCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	if !program.reviewModeActive() {
		return pullRequestReviewCommentActionTarget{}, false
	}

	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, program.detailWrapWidth)
	document := program.currentReviewDiffDocument(selectedFile, program.detailWrapWidth)
	_, comment, ok := reviewDiffCommentAtCursor(renderedRows, document, program.detailViewState)
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(program.reviewSession.summary.Repository))
	if repository == "" || program.reviewSession.summary.Number <= 0 || !hasUsablePullRequestMutationID(comment.ID) {
		return pullRequestReviewCommentActionTarget{}, false
	}

	return pullRequestReviewCommentActionTarget{
		repository: repository,
		number:     program.reviewSession.summary.Number,
		commentID:  strings.TrimSpace(comment.ID),
		body:       comment.Body,
	}, true
}

func pullRequestInlineThreadCommentActionTarget(threadComment githubdomain.PullRequestComment) (pullRequestReviewCommentActionTarget, bool) {
	if !hasUsablePullRequestMutationID(threadComment.ID) {
		return pullRequestReviewCommentActionTarget{}, false
	}
	return pullRequestReviewCommentActionTarget{commentID: strings.TrimSpace(threadComment.ID), body: threadComment.Body}, true
}

func pullRequestInlineThreadCommentAtBodyCursor(thread any, renderer MarkdownRenderer, width int, cursorLine int) (githubdomain.PullRequestComment, bool) {
	threadValue, ok := toDomainPullRequestReviewThread(thread)
	if !ok {
		return githubdomain.PullRequestComment{}, false
	}
	lineIndex := cursorLine
	if diffPreview := renderPullRequestInlineCommentThreadDiffPreview(pullRequestInlineCommentFromThread(threadValue)); diffPreview != "" {
		lineIndex -= renderedTextLineCount(diffPreview)
		if lineIndex < 0 {
			return githubdomain.PullRequestComment{}, false
		}
	}

	for commentIndex, threadComment := range threadValue.Comments {
		commentLineCount := renderedTextLineCount(renderInlineThreadCommentBlock(threadComment, renderer, width, commentIndex, len(threadValue.Comments)))
		if lineIndex < commentLineCount {
			return threadComment, true
		}
		lineIndex -= commentLineCount
	}

	return githubdomain.PullRequestComment{}, false
}

func reviewDiffCommentAtCursor(renderedRows []reviewDiffRenderedRow, document detailDocument, state detailViewState) (reviewDiffThread, githubdomain.PullRequestComment, bool) {
	if len(renderedRows) == 0 || len(document.rows) == 0 {
		return reviewDiffThread{}, githubdomain.PullRequestComment{}, false
	}

	renderedRowIndex := document.rows[document.rowIndexForPosition(state.cursor)].line
	if renderedRowIndex < 0 || renderedRowIndex >= len(renderedRows) {
		return reviewDiffThread{}, githubdomain.PullRequestComment{}, false
	}
	row := renderedRows[renderedRowIndex]
	if row.Thread == nil || row.Comment == nil {
		return reviewDiffThread{}, githubdomain.PullRequestComment{}, false
	}
	comment, ok := toDomainPullRequestComment(row.Comment)
	if !ok {
		return reviewDiffThread{}, githubdomain.PullRequestComment{}, false
	}
	return *row.Thread, comment, true
}
