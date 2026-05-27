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
	return program.dispatch(gui, MsgOpenPullRequestCommentComposerRequested{})
}

func (program *Program) openDetailPullRequestCommentShortcut(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgOpenDetailPullRequestCommentRequested{})
}

func newPullRequestCommentComposerOpenDescriptor(target pullRequestCommentTarget, feedbackTarget Focus) modalEditorOpenDescriptor {
	return newModalEditorOpenDescriptorWithSubmitDescriptor(pullRequestCommentComposerTitle, "", newPullRequestCommentSubmitDescriptor(target, feedbackTarget))
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
