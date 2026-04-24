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
		id:       "edit-pull-request-title",
		title:    pullRequestTitleEditorTitle,
		icon:     actionsPopupEditPullRequestIcon,
		keywords: []string{"edit", "title", "rename", "subject"},
		execute:  program.executeEditPullRequestTitleAction,
	}
}

func (program *Program) executeEditPullRequestTitleAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	wasVisible := program.modalEditorVisible()
	err := program.openLineModalEditor(gui, pullRequestTitleEditorTitle, target.title, func(title string) error {
		return program.submitPullRequestTitleEdit(target, title)
	})
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	if program.modalEditor != nil {
		program.modalEditor.afterSubmit = func(gui *gocui.Gui) {
			program.reloadActivePullRequestsTab(gui)
		}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) editPullRequestDescriptionAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "edit-pull-request-description",
		title:    pullRequestDescriptionEditorTitle,
		icon:     actionsPopupEditPullRequestIcon,
		keywords: []string{"edit", "description", "body", "summary"},
		execute:  program.executeEditPullRequestDescriptionAction,
	}
}

func (program *Program) executeEditPullRequestDescriptionAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	wasVisible := program.modalEditorVisible()
	err := program.openMultilineModalEditor(gui, pullRequestDescriptionEditorTitle, target.body, func(body string) error {
		return program.submitPullRequestDescriptionEdit(target, body)
	}, pullRequestDescriptionEditorHeight, handlePullRequestDescriptionEditorKey)
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	if program.modalEditor != nil {
		program.modalEditor.afterSubmit = func(gui *gocui.Gui) {
			program.reloadActivePullRequestsTab(gui)
		}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) submitPullRequestTitleEdit(target pullRequestActionTarget, title string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}
	if err := program.githubLoader.EditPullRequestTitle(target.repository, target.number, title); err != nil {
		return err
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.setFeedback(program.model.Focus(), pullRequestTitleEditSuccessMessage)
	return nil
}

func (program *Program) submitPullRequestDescriptionEdit(target pullRequestActionTarget, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}
	if err := program.githubLoader.EditPullRequestDescription(target.repository, target.number, body); err != nil {
		return err
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.setFeedback(program.model.Focus(), pullRequestDescriptionEditSuccessMessage)
	return nil
}

func handlePullRequestDescriptionEditorKey(program *Program, view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	return handleMultilineModalEditorExternalEditKey(program, view, key, ch, mod)
}
