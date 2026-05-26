package tui

import "github.com/jesseduffield/gocui"

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
	return program.openModalEditorWithSubmitDescriptor(gui, pullRequestCommentComposerTitle, "", newPullRequestCommentSubmitDescriptor(target, feedbackTarget))
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

func (program *Program) selectedPullRequestCommentTarget() (pullRequestCommentTarget, bool) {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return pullRequestCommentTarget{}, false
	}

	return pullRequestCommentTarget{repository: target.repository, number: target.number}, true
}
