package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	inlineCommentUpdateEditorTitle     = "Update inline comment"
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
	}, reviewInlineCommentModalHeight, handleMultilineModalEditorExternalEditKey)
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
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}
	if err := program.githubLoader.UpdatePullRequestReviewComment(target.commentID, body); err != nil {
		return err
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setFeedback(FocusDetailView, inlineCommentUpdatedSuccessMessage)
	return nil
}

func (program *Program) deleteInlineComment(target pullRequestReviewCommentActionTarget) error {
	if strings.TrimSpace(target.commentID) == "" {
		return errors.New("missing inline comment identity")
	}
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}
	if err := program.githubLoader.DeletePullRequestReviewComment(target.commentID); err != nil {
		return err
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setFeedback(FocusDetailView, inlineCommentDeletedSuccessMessage)
	return nil
}

func (program *Program) selectedPullRequestReviewCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReviewCommentActionTarget{}, false
	}
	if program.reviewSession.active {
		return program.selectedReviewDiffInlineCommentActionTarget()
	}
	return program.selectedBrowserInlineCommentActionTarget()
}

func (program *Program) selectedBrowserInlineCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() || program.activeDetailTab != CommentsDetailTab {
		return pullRequestReviewCommentActionTarget{}, false
	}

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
	if target.repository == "" || target.number <= 0 || strings.TrimSpace(target.commentID) == "" {
		return pullRequestReviewCommentActionTarget{}, false
	}
	return target, true
}

func (program *Program) selectedReviewDiffInlineCommentActionTarget() (pullRequestReviewCommentActionTarget, bool) {
	if !program.reviewSession.active {
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
	if repository == "" || program.reviewSession.summary.Number <= 0 || strings.TrimSpace(comment.ID) == "" || !comment.ViewerDidAuthor {
		return pullRequestReviewCommentActionTarget{}, false
	}

	return pullRequestReviewCommentActionTarget{
		repository: repository,
		number:     program.reviewSession.summary.Number,
		commentID:  strings.TrimSpace(comment.ID),
		body:       comment.Body,
	}, true
}

func pullRequestInlineThreadCommentActionTargetAtBodyCursor(thread githubcli.PullRequestReviewThread, renderer MarkdownRenderer, width int, cursorLine int) (pullRequestReviewCommentActionTarget, bool) {
	threadComment, ok := pullRequestInlineThreadCommentAtBodyCursor(thread, renderer, width, cursorLine)
	if !ok {
		return pullRequestReviewCommentActionTarget{}, false
	}
	return pullRequestInlineThreadCommentActionTarget(threadComment)
}

func pullRequestInlineThreadCommentActionTarget(threadComment githubcli.PullRequestComment) (pullRequestReviewCommentActionTarget, bool) {
	if strings.TrimSpace(threadComment.ID) == "" || !threadComment.ViewerDidAuthor {
		return pullRequestReviewCommentActionTarget{}, false
	}
	return pullRequestReviewCommentActionTarget{commentID: strings.TrimSpace(threadComment.ID), body: threadComment.Body}, true
}

func pullRequestInlineThreadCommentAtBodyCursor(thread githubcli.PullRequestReviewThread, renderer MarkdownRenderer, width int, cursorLine int) (githubcli.PullRequestComment, bool) {
	lineIndex := cursorLine
	if diffPreview := renderPullRequestInlineCommentThreadDiffPreview(pullRequestInlineCommentFromThread(thread)); diffPreview != "" {
		lineIndex -= renderedTextLineCount(diffPreview)
		if lineIndex < 0 {
			return githubcli.PullRequestComment{}, false
		}
	}

	for commentIndex, threadComment := range thread.Comments {
		commentLineCount := renderedTextLineCount(renderInlineThreadCommentBlock(threadComment, renderer, width, commentIndex, len(thread.Comments)))
		if lineIndex < commentLineCount {
			return threadComment, true
		}
		lineIndex -= commentLineCount
	}

	return githubcli.PullRequestComment{}, false
}

func reviewDiffCommentAtCursor(renderedRows []reviewDiffRenderedRow, document detailDocument, state detailViewState) (reviewDiffThread, githubcli.PullRequestComment, bool) {
	if len(renderedRows) == 0 || len(document.rows) == 0 {
		return reviewDiffThread{}, githubcli.PullRequestComment{}, false
	}

	renderedRowIndex := document.rows[document.rowIndexForPosition(state.cursor)].line
	if renderedRowIndex < 0 || renderedRowIndex >= len(renderedRows) {
		return reviewDiffThread{}, githubcli.PullRequestComment{}, false
	}
	row := renderedRows[renderedRowIndex]
	if row.Thread == nil || row.Comment == nil {
		return reviewDiffThread{}, githubcli.PullRequestComment{}, false
	}
	return *row.Thread, *row.Comment, true
}
