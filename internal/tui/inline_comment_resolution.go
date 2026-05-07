package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	inlineCommentResolvedSuccessMessage   = "Inline comment resolved"
	inlineCommentUnresolvedSuccessMessage = "Inline comment marked unresolved"
)

type pullRequestReviewThreadActionTarget struct {
	repository string
	number     int
	threadID   string
	resolved   bool
}

func (program *Program) currentInlineCommentResolutionAction() (actionsPopupAction, bool) {
	target, ok := program.selectedPullRequestReviewThreadActionTarget()
	if !ok {
		return actionsPopupAction{}, false
	}
	if target.resolved {
		return program.unresolveInlineCommentAction(), true
	}
	return program.resolveInlineCommentAction(), true
}

func (program *Program) resolveInlineCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "resolve-inline-comment",
		title:    "Mark inline comment as resolved",
		icon:     actionsPopupResolveInlineCommentIcon,
		keywords: []string{"inline", "comment", "resolve", "thread", "discussion"},
		execute:  program.executeResolveInlineCommentAction,
	}
}

func (program *Program) unresolveInlineCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "unresolve-inline-comment",
		title:    "Mark inline comment as unresolved",
		icon:     actionsPopupResolveInlineCommentIcon,
		keywords: []string{"inline", "comment", "unresolve", "unresolved", "reopen", "thread", "discussion"},
		execute:  program.executeUnresolveInlineCommentAction,
	}
}

func (program *Program) executeResolveInlineCommentAction(_ *gocui.Gui) actionsPopupActionResult {
	return program.executeInlineCommentResolutionAction(true)
}

func (program *Program) executeUnresolveInlineCommentAction(_ *gocui.Gui) actionsPopupActionResult {
	return program.executeInlineCommentResolutionAction(false)
}

func (program *Program) executeInlineCommentResolutionAction(resolved bool) actionsPopupActionResult {
	target, ok := program.selectedPullRequestReviewThreadActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	var err error
	feedbackMessage := inlineCommentResolvedSuccessMessage
	if resolved {
		err = program.githubLoader.ResolvePullRequestReviewThread(target.threadID)
	} else {
		err = program.githubLoader.UnresolvePullRequestReviewThread(target.threadID)
		feedbackMessage = inlineCommentUnresolvedSuccessMessage
	}
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setFeedback(program.model.Focus(), feedbackMessage)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) selectedPullRequestReviewThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReviewThreadActionTarget{}, false
	}
	if program.reviewSession.active {
		return program.selectedReviewDiffReviewThreadActionTarget()
	}
	return program.selectedBrowserInlineCommentThreadActionTarget()
}

func (program *Program) selectedBrowserInlineCommentThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() || program.activeDetailTab != CommentsDetailTab {
		return pullRequestReviewThreadActionTarget{}, false
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReviewThreadActionTarget{}, false
	}

	sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, program.detailWrapWidth, program.detailViewState.cursor.line)
	if !ok || sectionAtCursor.section.inlineThread == nil {
		return pullRequestReviewThreadActionTarget{}, false
	}
	thread := *sectionAtCursor.section.inlineThread
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 || strings.TrimSpace(thread.ID) == "" {
		return pullRequestReviewThreadActionTarget{}, false
	}

	return pullRequestReviewThreadActionTarget{
		repository: repository,
		number:     summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
		resolved:   thread.IsResolved,
	}, true
}

func (program *Program) selectedReviewDiffReviewThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	if !program.reviewSession.active {
		return pullRequestReviewThreadActionTarget{}, false
	}

	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, program.detailWrapWidth)
	document := program.currentReviewDiffDocument(selectedFile, program.detailWrapWidth)
	thread, ok := reviewDiffThreadAtCursor(renderedRows, document, program.detailViewState)
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(program.reviewSession.summary.Repository))
	if repository == "" || program.reviewSession.summary.Number <= 0 || strings.TrimSpace(thread.ID) == "" {
		return pullRequestReviewThreadActionTarget{}, false
	}

	return pullRequestReviewThreadActionTarget{
		repository: repository,
		number:     program.reviewSession.summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
		resolved:   thread.IsResolved,
	}, true
}

func renderedTextLineCount(text string) int {
	if text == "" {
		return 1
	}
	return strings.Count(text, "\n") + 1
}

func reviewDiffThreadAtCursor(renderedRows []reviewDiffRenderedRow, document detailDocument, state detailViewState) (reviewDiffThread, bool) {
	if len(renderedRows) == 0 || len(document.rows) == 0 {
		return reviewDiffThread{}, false
	}

	renderedRowIndex := document.rows[document.rowIndexForPosition(state.cursor)].line
	if renderedRowIndex < 0 || renderedRowIndex >= len(renderedRows) {
		return reviewDiffThread{}, false
	}
	if renderedRows[renderedRowIndex].Thread == nil {
		return reviewDiffThread{}, false
	}
	return *renderedRows[renderedRowIndex].Thread, true
}
