package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"
)

const (
	pullRequestCommentComposerTitle  = "Comment on PR"
	pullRequestCommentSuccessMessage = "Comment posted"
)

type pullRequestCommentTarget struct {
	repository string
	number     int
}

func (program *Program) openPullRequestCommentComposer(gui *gocui.Gui, _ *gocui.View) error {
	if program.pullRequestCommentComposerBlocked() {
		return nil
	}

	target, ok := program.selectedPullRequestCommentTarget()
	if !ok {
		return nil
	}

	feedbackTarget := program.model.Focus()
	if err := program.openModalEditor(gui, pullRequestCommentComposerTitle, "", func(body string) error {
		return program.submitPullRequestComment(target, body)
	}); err != nil {
		return err
	}
	if program.overlayState.modalEditor != nil {
		program.overlayState.modalEditor.afterSubmit = func(gui *gocui.Gui) {
			_ = program.dispatch(gui, MsgFeedbackSet{Target: feedbackTarget, Message: pullRequestCommentSuccessMessage})
		}
	}
	return nil
}

func (program *Program) openDetailPullRequestCommentShortcut(gui *gocui.Gui, view *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	if program.pullRequestCommentComposerBlocked() {
		return nil
	}

	switch program.inputContext().DetailInputMode {
	case DetailInputModeReviewInlineComment:
		return program.openInlineReviewCommentComposer(gui, view)
	case DetailInputModeBrowserChangesInlineComment:
		return program.openBrowserChangesInlineCommentComposer(gui, view)
	case DetailInputModePullRequestComment:
		return program.openPullRequestCommentComposer(gui, nil)
	default:
		return nil
	}
}

func (program *Program) pullRequestCommentComposerBlocked() bool {
	return program.overlayState.helpVisible || program.model.SearchActive() || program.modalEditorVisible()
}

func (program *Program) submitPullRequestComment(target pullRequestCommentTarget, body string) error {
	if target.repository == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if !program.hasPullRequestMutations() {
		return errors.New("github loader is unavailable")
	}

	if err := program.pullRequestMutations.CommentOnPullRequest(target.repository, target.number, body); err != nil {
		return newTransientErrorPopupActionError(err)
	}

	program.optimisticallyAppendPullRequestComment(target, body)
	return nil
}

func (program *Program) selectedPullRequestCommentTarget() (pullRequestCommentTarget, bool) {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return pullRequestCommentTarget{}, false
	}

	return pullRequestCommentTarget{repository: target.repository, number: target.number}, true
}
