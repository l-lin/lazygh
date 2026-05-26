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
		requested = MsgModalEditorOpened{State: newLineModalEditorStateWithSubmitDescriptor(pullRequestTitleEditorTitle, target.title, newPullRequestTitleEditSubmitDescriptor(target, feedbackTarget))}
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
		requested = MsgModalEditorOpened{State: newMultilineModalEditorStateWithSubmitDescriptor(pullRequestDescriptionEditorTitle, target.body, newPullRequestDescriptionEditSubmitDescriptor(target, feedbackTarget), pullRequestDescriptionEditorHeight)}
	}
	return actionsPopupAction{
		id:        "edit-pull-request-description",
		title:     pullRequestDescriptionEditorTitle,
		icon:      actionsPopupEditPullRequestIcon,
		requested: requested,
	}
}
