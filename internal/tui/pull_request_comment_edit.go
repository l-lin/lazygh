package tui

import "strings"

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
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestCommentEditActionTarget()
	if ok {
		requested = MsgModalEditorOpened{Descriptor: newMultilineModalEditorOpenDescriptorWithSubmitDescriptor(pullRequestCommentUpdateEditorTitle, target.body, newPullRequestCommentUpdateSubmitDescriptor(target), reviewInlineCommentModalHeight)}
	}
	return actionsPopupAction{
		id:        "update-pull-request-comment",
		title:     pullRequestCommentUpdateEditorTitle,
		icon:      actionsPopupEditPullRequestIcon,
		requested: requested,
	}
}

func (program *Program) deletePullRequestCommentAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.selectedPullRequestCommentEditActionTarget(); ok {
		requested = MsgPullRequestCommentDeleteRequested{Target: target}
	}
	return actionsPopupAction{
		id:        "delete-pull-request-comment",
		title:     pullRequestCommentDeleteActionTitle,
		icon:      actionsPopupDeleteInlineCommentIcon,
		requested: requested,
	}
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
