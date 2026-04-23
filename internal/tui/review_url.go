package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	actionsPopupReviewPullRequestURLIcon = ""
	reviewPullRequestURLEditorTitle      = "Review PR from URL"
)

func (program *Program) reviewPullRequestURLActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "review-pull-request-from-url",
		title:    reviewPullRequestURLEditorTitle,
		icon:     actionsPopupReviewPullRequestURLIcon,
		keywords: []string{"review", "pull", "request", "url", "paste", "open", "link"},
		execute:  program.executeReviewPullRequestURLAction,
	}
}

func (program *Program) executeReviewPullRequestURLAction(gui *gocui.Gui) actionsPopupActionResult {
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	wasVisible := program.modalEditorVisible()
	err := program.openLineModalEditor(gui, reviewPullRequestURLEditorTitle, "", func(rawURL string) error {
		return program.OpenReviewByURL(rawURL)
	})
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	if program.modalEditor != nil {
		program.modalEditor.handleKey = handleReviewPullRequestURLEditorKey
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) OpenReviewByURL(rawURL string) error {
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}

	summary, err := githubcli.ParsePullRequestURL(rawURL)
	if err != nil {
		return err
	}
	return program.openPullRequestReview(summary)
}

func handleReviewPullRequestURLEditorKey(program *Program, view *gocui.View, key gocui.Key, _ rune, _ gocui.Modifier) bool {
	if key != gocui.KeyEnter || program == nil {
		return false
	}
	_ = program.submitModalEditor(program.gui, view)
	return true
}

func (program *Program) openPullRequestReview(summary githubcli.PullRequest) error {
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return errors.New("missing pull request identity")
	}

	pendingReviewID, err := program.githubLoader.StartPendingPullRequestReview(repository, summary.Number)
	if err != nil {
		return err
	}

	program.startReviewSession(summary, pendingReviewID)
	return nil
}
