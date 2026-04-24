package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	pullRequestReviewInlineCommentComposerTitle  = "Review: Add inline comment"
	pullRequestReviewInlineCommentSuccessMessage = "Inline comment added"
	reviewInlineCommentModalHeight               = 10
)

type reviewInlineCommentTarget struct {
	repository    string
	number        int
	pendingReview string
	threadTarget  githubcli.PullRequestReviewThreadTarget
}

func (program *Program) openInlineReviewCommentComposer(gui *gocui.Gui, view *gocui.View) error {
	target, err := program.selectedInlineReviewCommentTarget(gui, view)
	if err != nil {
		program.setFeedback(FocusDetailView, err.Error())
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}

	return program.openMultilineModalEditor(gui, pullRequestReviewInlineCommentComposerTitle, "", func(body string) error {
		return program.submitInlineReviewComment(target, body)
	}, reviewInlineCommentModalHeight, handleMultilineModalEditorExternalEditKey)
}

func (program *Program) selectedInlineReviewCommentTarget(gui *gocui.Gui, view *gocui.View) (reviewInlineCommentTarget, error) {
	if !program.reviewSession.active {
		return reviewInlineCommentTarget{}, errReviewThreadTargetUnavailable
	}

	repository := pullRequestRepositoryName(program.reviewSession.summary.Repository)
	if strings.TrimSpace(repository) == "" || program.reviewSession.summary.Number <= 0 || strings.TrimSpace(program.reviewSession.pendingReviewID) == "" {
		return reviewInlineCommentTarget{}, errors.New("missing pull request review context")
	}

	actualView := view
	if actualView == nil && gui != nil {
		detailView, err := gui.View(viewDetailName)
		if err == nil {
			actualView = detailView
		}
	}
	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return reviewInlineCommentTarget{}, errReviewThreadTargetUnavailable
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	threadTarget, err := reviewDiffThreadTargetForSelection(renderedRows, detailDocument, program.detailViewState)
	if err != nil {
		return reviewInlineCommentTarget{}, err
	}

	return reviewInlineCommentTarget{
		repository:    repository,
		number:        program.reviewSession.summary.Number,
		pendingReview: strings.TrimSpace(program.reviewSession.pendingReviewID),
		threadTarget:  threadTarget,
	}, nil
}

func (program *Program) submitInlineReviewComment(target reviewInlineCommentTarget, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 || strings.TrimSpace(target.pendingReview) == "" {
		return errors.New("missing pull request review context")
	}
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}
	if err := program.githubLoader.AddPullRequestReviewThread(target.pendingReview, body, target.threadTarget); err != nil {
		return err
	}

	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setFeedback(FocusDetailView, pullRequestReviewInlineCommentSuccessMessage)
	return nil
}
