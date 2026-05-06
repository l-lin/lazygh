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
	selection, err := program.selectedInlineReviewCommentSelection(gui, view)
	if err != nil {
		program.setFeedback(FocusDetailView, err.Error())
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}

	return program.openMultilineModalEditor(gui, pullRequestReviewInlineCommentComposerTitle, selection.initialBody, func(body string) error {
		return program.submitInlineReviewComment(selection.target, body)
	}, reviewInlineCommentModalHeight, handleMultilineModalEditorExternalEditKey)
}

func (program *Program) currentReviewInlineCommentAction() (actionsPopupAction, bool) {
	if !program.reviewSession.active || program.model.Focus() != FocusDetailView {
		return actionsPopupAction{}, false
	}
	if _, err := program.selectedInlineReviewCommentSelection(program.gui, nil); err != nil {
		return actionsPopupAction{}, false
	}

	return program.addInlineReviewCommentAction(), true
}

func (program *Program) addInlineReviewCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "add-inline-review-comment",
		title:    "Add inline comment",
		icon:     actionsPopupCommentOnPullRequestIcon,
		keywords: []string{"inline", "comment", "review", "diff", "suggestion"},
		execute:  program.executeAddInlineReviewCommentAction,
	}
}

func (program *Program) executeAddInlineReviewCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	wasVisible := program.modalEditorVisible()
	if err := program.openInlineReviewCommentComposer(gui, nil); err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

type reviewInlineCommentSelection struct {
	target      reviewInlineCommentTarget
	initialBody string
}

func (program *Program) selectedInlineReviewCommentSelection(gui *gocui.Gui, view *gocui.View) (reviewInlineCommentSelection, error) {
	if !program.reviewSession.active {
		return reviewInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	repository := pullRequestRepositoryName(program.reviewSession.summary.Repository)
	if strings.TrimSpace(repository) == "" || program.reviewSession.summary.Number <= 0 || strings.TrimSpace(program.reviewSession.pendingReviewID) == "" {
		return reviewInlineCommentSelection{}, errors.New("missing pull request review context")
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
		return reviewInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	threadTarget, err := reviewDiffThreadTargetForSelection(renderedRows, detailDocument, program.detailViewState)
	if err != nil {
		return reviewInlineCommentSelection{}, err
	}

	return reviewInlineCommentSelection{
		target: reviewInlineCommentTarget{
			repository:    repository,
			number:        program.reviewSession.summary.Number,
			pendingReview: strings.TrimSpace(program.reviewSession.pendingReviewID),
			threadTarget:  threadTarget,
		},
		initialBody: reviewInlineCommentSuggestionBody(selectedFile.Path, reviewDiffSelectedSnippet(renderedRows, detailDocument, program.detailViewState)),
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
