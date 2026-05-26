package tui

const (
	pullRequestTitleEditorTitle              = "Edit PR title"
	pullRequestDescriptionEditorTitle        = "Edit PR description"
	pullRequestDescriptionEditorHeight       = 20
	pullRequestTitleEditSuccessMessage       = "PR title updated"
	pullRequestDescriptionEditSuccessMessage = "PR description updated"
)

func (program *Program) editPullRequestTitleAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestActionTarget()
	if ok {
		feedbackTarget := program.model.Focus()
		requested = MsgModalEditorOpened{State: newLineModalEditorStateWithSubmitRequested(pullRequestTitleEditorTitle, target.title, func(title string) Msg {
			return MsgPullRequestTitleEditRequested{Target: target, Title: title, FeedbackTarget: feedbackTarget}
		})}
	}
	return actionsPopupAction{
		id:        "edit-pull-request-title",
		title:     pullRequestTitleEditorTitle,
		icon:      actionsPopupEditPullRequestIcon,
		requested: requested,
	}
}

func (program *Program) editPullRequestDescriptionAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestActionTarget()
	if ok {
		feedbackTarget := program.model.Focus()
		requested = MsgModalEditorOpened{State: newMultilineModalEditorStateWithSubmitRequested(pullRequestDescriptionEditorTitle, target.body, func(body string) Msg {
			return MsgPullRequestDescriptionEditRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
		}, pullRequestDescriptionEditorHeight)}
	}
	return actionsPopupAction{
		id:        "edit-pull-request-description",
		title:     pullRequestDescriptionEditorTitle,
		icon:      actionsPopupEditPullRequestIcon,
		requested: requested,
	}
}
