package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	pullRequestCommentUpdateEditorTitle     = "Update PR comment"
	pullRequestCommentDeleteActionTitle     = "Delete PR comment"
	pullRequestCommentUpdatedSuccessMessage = "PR comment updated"
	pullRequestCommentDeletedSuccessMessage = "PR comment deleted"
)

type pullRequestCommentEditActionTarget struct {
	repository string
	number     int
	commentID  string
	body       string
}

func (program *Program) currentPullRequestCommentEditActions() []actionsPopupAction {
	if _, ok := program.selectedPullRequestCommentEditActionTarget(); !ok {
		return nil
	}
	return []actionsPopupAction{program.updatePullRequestCommentAction(), program.deletePullRequestCommentAction()}
}

func (program *Program) updatePullRequestCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "update-pull-request-comment",
		title:   pullRequestCommentUpdateEditorTitle,
		icon:    actionsPopupEditPullRequestIcon,
		execute: program.executeUpdatePullRequestCommentAction,
	}
}

func (program *Program) deletePullRequestCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "delete-pull-request-comment",
		title:   pullRequestCommentDeleteActionTitle,
		icon:    actionsPopupDeleteInlineCommentIcon,
		execute: program.executeDeletePullRequestCommentAction,
	}
}

func (program *Program) executeUpdatePullRequestCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestCommentEditActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		if err := program.openMultilineModalEditor(gui, pullRequestCommentUpdateEditorTitle, target.body, func(body string) error {
			return program.submitPullRequestCommentUpdate(target, body)
		}, reviewInlineCommentModalHeight); err != nil {
			return err
		}
		if program.overlayState.modalEditor != nil {
			program.overlayState.modalEditor.afterSubmit = func(gui *gocui.Gui) {
				_ = program.dispatch(gui, MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestCommentUpdatedSuccessMessage})
			}
		}
		return nil
	})
}

func (program *Program) executeDeletePullRequestCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestCommentEditActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.deletePullRequestComment(target); err != nil {
		return actionsPopupActionResult{err: err}
	}
	if err := program.dispatch(gui, MsgActionsPopupClosedWithFeedback{Target: FocusDetailView, Message: pullRequestCommentDeletedSuccessMessage}); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{}
}

func (program *Program) submitPullRequestCommentUpdate(target pullRequestCommentEditActionTarget, body string) error {
	if strings.TrimSpace(target.commentID) == "" {
		return errors.New("missing pull request comment identity")
	}
	if !program.hasPullRequestMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.UpdatePullRequestComment(target.commentID, body); err != nil {
		return newTransientErrorPopupActionError(err)
	}

	program.optimisticallyUpdatePullRequestComment(target, body)
	return nil
}

func (program *Program) deletePullRequestComment(target pullRequestCommentEditActionTarget) error {
	if strings.TrimSpace(target.commentID) == "" {
		return errors.New("missing pull request comment identity")
	}
	if !program.hasPullRequestMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.DeletePullRequestComment(target.commentID); err != nil {
		return newTransientErrorPopupActionError(err)
	}

	program.optimisticallyDeletePullRequestComment(target)
	return nil
}

func (program *Program) selectedPullRequestCommentEditActionTarget() (pullRequestCommentEditActionTarget, bool) {
	if program.model.Focus() != FocusDetailView || program.reviewModeActive() {
		return pullRequestCommentEditActionTarget{}, false
	}
	if !program.shouldShowPullRequestDetailTabs() || program.detailState.activeTab != CommentsDetailTab {
		return pullRequestCommentEditActionTarget{}, false
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestCommentEditActionTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestCommentEditActionTarget{}, false
	}

	sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, program.detailState.wrapWidth, program.detailState.viewState.cursor.line)
	if !ok || sectionAtCursor.section.comment == nil {
		return pullRequestCommentEditActionTarget{}, false
	}
	comment := *sectionAtCursor.section.comment
	if !hasUsablePullRequestMutationID(comment.ID) {
		return pullRequestCommentEditActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 {
		return pullRequestCommentEditActionTarget{}, false
	}
	return pullRequestCommentEditActionTarget{
		repository: repository,
		number:     summary.Number,
		commentID:  strings.TrimSpace(comment.ID),
		body:       comment.Body,
	}, true
}
