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

func (program *Program) openPullRequestCommentComposer(gui *gocui.Gui, view *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return nil
	}
	if program.reviewSession.active && program.model.Focus() == FocusDetailView {
		return program.openInlineReviewCommentComposer(gui, view)
	}

	target, ok := program.selectedPullRequestCommentTarget()
	if !ok {
		return nil
	}

	return program.openModalEditor(gui, pullRequestCommentComposerTitle, "", func(body string) error {
		return program.submitPullRequestComment(target, body)
	})
}

func (program *Program) submitPullRequestComment(target pullRequestCommentTarget, body string) error {
	if target.repository == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}

	if err := program.githubLoader.CommentOnPullRequest(target.repository, target.number, body); err != nil {
		return err
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.setFeedback(program.model.Focus(), pullRequestCommentSuccessMessage)
	return nil
}

func (program *Program) selectedPullRequestCommentTarget() (pullRequestCommentTarget, bool) {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return pullRequestCommentTarget{}, false
	}

	return pullRequestCommentTarget{repository: target.repository, number: target.number}, true
}
