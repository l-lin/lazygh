package tui

import "github.com/jesseduffield/gocui"

const (
	pullRequestTitleEditorTitle              = "Edit PR title"
	pullRequestDescriptionEditorTitle        = "Edit PR description"
	pullRequestDescriptionEditorHeight       = 20
	pullRequestTitleEditSuccessMessage       = "PR title updated"
	pullRequestDescriptionEditSuccessMessage = "PR description updated"
)

func (program *Program) editPullRequestTitleAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "edit-pull-request-title",
		title:   pullRequestTitleEditorTitle,
		icon:    actionsPopupEditPullRequestIcon,
		execute: program.executeEditPullRequestTitleAction,
	}
}

func (program *Program) executeEditPullRequestTitleAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	feedbackTarget := program.model.Focus()
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openLineModalEditorWithSubmitRequested(gui, pullRequestTitleEditorTitle, target.title, func(title string) Msg {
			return MsgPullRequestTitleEditRequested{Target: target, Title: title, FeedbackTarget: feedbackTarget}
		})
	})
}

func (program *Program) editPullRequestDescriptionAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "edit-pull-request-description",
		title:   pullRequestDescriptionEditorTitle,
		icon:    actionsPopupEditPullRequestIcon,
		execute: program.executeEditPullRequestDescriptionAction,
	}
}

func (program *Program) executeEditPullRequestDescriptionAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	feedbackTarget := program.model.Focus()
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openMultilineModalEditorWithSubmitRequested(gui, pullRequestDescriptionEditorTitle, target.body, func(body string) Msg {
			return MsgPullRequestDescriptionEditRequested{Target: target, Body: body, FeedbackTarget: feedbackTarget}
		}, pullRequestDescriptionEditorHeight)
	})
}
