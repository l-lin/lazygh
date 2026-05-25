package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

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

	submittedTitle := target.title
	feedbackTarget := program.model.Focus()
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		if err := program.openLineModalEditor(gui, pullRequestTitleEditorTitle, target.title, func(title string) error {
			submittedTitle = title
			return program.submitPullRequestTitleEdit(target, title)
		}); err != nil {
			return err
		}
		if program.overlayState.modalEditor != nil {
			program.overlayState.modalEditor.afterSubmit = func(gui *gocui.Gui) {
				_ = program.dispatch(gui, MsgPullRequestTitleEditApplied{Target: target, Title: submittedTitle, FeedbackTarget: feedbackTarget})
			}
		}
		return nil
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

	submittedBody := target.body
	feedbackTarget := program.model.Focus()
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		if err := program.openMultilineModalEditor(gui, pullRequestDescriptionEditorTitle, target.body, func(body string) error {
			submittedBody = body
			return program.submitPullRequestDescriptionEdit(target, body)
		}, pullRequestDescriptionEditorHeight); err != nil {
			return err
		}
		if program.overlayState.modalEditor != nil {
			program.overlayState.modalEditor.afterSubmit = func(gui *gocui.Gui) {
				_ = program.dispatch(gui, MsgPullRequestDescriptionEditApplied{Target: target, Body: submittedBody, FeedbackTarget: feedbackTarget})
			}
		}
		return nil
	})
}

func (program *Program) submitPullRequestTitleEdit(target pullRequestActionTarget, title string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if !program.hasPullRequestMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.EditPullRequestTitle(target.repository, target.number, title); err != nil {
		return newTransientErrorPopupActionError(err)
	}
	program.optimisticallyUpdatePullRequestTitle(target.repository, target.number, title)
	return nil
}

func (program *Program) submitPullRequestDescriptionEdit(target pullRequestActionTarget, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if !program.hasPullRequestMutations() {
		return errors.New("github loader is unavailable")
	}
	if err := program.pullRequestMutations.EditPullRequestDescription(target.repository, target.number, body); err != nil {
		return newTransientErrorPopupActionError(err)
	}
	program.optimisticallyUpdatePullRequestDescription(target.repository, target.number, body)
	return nil
}
