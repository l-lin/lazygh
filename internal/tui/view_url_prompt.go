package tui

import "github.com/jesseduffield/gocui"

const (
	openPullRequestByURLActionTitle  = "Open PR from URL"
	openPullRequestByURLEditorHeight = lineModalEditorTotalHeight
)

func (program *Program) openPullRequestByURLShortcut(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.mainPaneActionBlocked() || program.actionContext().IsReviewContext() {
		return nil
	}

	return program.openPullRequestByURLEditor(gui)
}

func (program *Program) openPullRequestByURLEditor(gui *gocui.Gui) error {
	if err := program.openLineModalEditorWithHeight(gui, openPullRequestByURLActionTitle, "", program.OpenPullRequestByURL, openPullRequestByURLEditorHeight); err != nil {
		return err
	}
	if program.modalEditor != nil {
		program.modalEditor.submitOnEnter = true
	}
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) openPullRequestByURLActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "open-pull-request-by-url",
		title:   openPullRequestByURLActionTitle,
		icon:    actionsPopupOpenPullRequestByURLIcon,
		execute: program.executeOpenPullRequestByURLAction,
	}
}

func (program *Program) executeOpenPullRequestByURLAction(gui *gocui.Gui) actionsPopupActionResult {
	wasVisible := program.modalEditorVisible()
	if err := program.openPullRequestByURLEditor(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}
