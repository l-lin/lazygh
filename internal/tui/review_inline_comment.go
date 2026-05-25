package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	pullRequestReviewInlineCommentComposerTitle  = "Review: Add inline comment"
	pullRequestReviewInlineCommentSuccessMessage = "Inline comment added"
	reviewInlineCommentModalHeight               = 10
)

type pullRequestInlineCommentTarget struct {
	repository    string
	number        int
	pendingReview string
	threadTarget  githubdomain.PullRequestReviewThreadTarget
	updateDetail  bool
}

type pullRequestInlineCommentSelection struct {
	target      pullRequestInlineCommentTarget
	initialBody string
}

func (program *Program) openInlineReviewCommentComposer(gui *gocui.Gui, view *gocui.View) error {
	selection, err := program.selectedReviewInlineCommentSelection(gui, view)
	if err != nil {
		return program.handleInlineCommentSelectionError(gui, err)
	}
	return program.openPullRequestInlineCommentComposer(gui, selection)
}

func (program *Program) openBrowserChangesInlineCommentComposer(gui *gocui.Gui, view *gocui.View) error {
	selection, err := program.selectedBrowserChangesInlineCommentSelection(gui, view)
	if err != nil {
		return program.handleInlineCommentSelectionError(gui, err)
	}
	return program.openPullRequestInlineCommentComposer(gui, selection)
}

func (program *Program) handleInlineCommentSelectionError(gui *gocui.Gui, err error) error {
	return program.dispatch(gui, MsgFeedbackSet{Target: FocusDetailView, Message: strings.TrimSpace(err.Error())})
}

func (program *Program) openPullRequestInlineCommentComposer(gui *gocui.Gui, selection pullRequestInlineCommentSelection) error {
	if err := program.openMultilineModalEditor(gui, pullRequestReviewInlineCommentComposerTitle, selection.initialBody, func(body string) error {
		return program.submitPullRequestInlineComment(selection.target, body)
	}, reviewInlineCommentModalHeight); err != nil {
		return err
	}
	if program.overlayState.modalEditor != nil {
		program.overlayState.modalEditor.afterSubmit = func(*gocui.Gui) {
			program.detailState.viewState.exitVisualMode()
		}
	}
	return nil
}

func (program *Program) currentReviewInlineCommentAction() (actionsPopupAction, bool) {
	if !program.reviewModeActive() || program.model.Focus() != FocusDetailView {
		return actionsPopupAction{}, false
	}
	if _, err := program.selectedReviewInlineCommentSelection(program.gui, nil); err != nil {
		return actionsPopupAction{}, false
	}

	return program.addInlineCommentAction(), true
}

func (program *Program) currentBrowserChangesInlineCommentAction() (actionsPopupAction, bool) {
	if program.reviewModeActive() || program.model.Focus() != FocusDetailView || !program.browserChangesInlineCommentShortcutActive() {
		return actionsPopupAction{}, false
	}
	if _, err := program.selectedBrowserChangesInlineCommentSelection(program.gui, nil); err != nil {
		return actionsPopupAction{}, false
	}

	return program.addInlineCommentAction(), true
}

func (program *Program) addInlineCommentAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "add-inline-comment",
		title:   "Add inline comment",
		icon:    actionsPopupCommentOnPullRequestIcon,
		execute: program.executeAddInlineCommentAction,
	}
}

func (program *Program) executeAddInlineCommentAction(gui *gocui.Gui) actionsPopupActionResult {
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		if program.reviewModeActive() {
			return program.openInlineReviewCommentComposer(gui, nil)
		}
		return program.openBrowserChangesInlineCommentComposer(gui, nil)
	})
}

func (program *Program) selectedReviewInlineCommentSelection(gui *gocui.Gui, view *gocui.View) (pullRequestInlineCommentSelection, error) {
	if !program.reviewModeActive() {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(program.navigationState.reviewSession.summary.Repository))
	pendingReviewID := strings.TrimSpace(program.navigationState.reviewSession.pendingReviewID)
	if repository == "" || program.navigationState.reviewSession.summary.Number <= 0 || pendingReviewID == "" {
		return pullRequestInlineCommentSelection{}, errors.New("missing pull request review context")
	}

	detailDocument := program.inlineCommentDetailDocument(gui, view)
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	renderedRows := program.currentReviewDiffRenderedRows(selectedFile, detailDocument.width)
	return pullRequestInlineCommentSelectionFromRenderedRows(repository, program.navigationState.reviewSession.summary.Number, pendingReviewID, false, renderedRows, detailDocument, program.detailState.viewState)
}

func (program *Program) selectedBrowserChangesInlineCommentSelection(gui *gocui.Gui, view *gocui.View) (pullRequestInlineCommentSelection, error) {
	if program.reviewModeActive() || !program.browserChangesInlineCommentShortcutActive() {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 {
		return pullRequestInlineCommentSelection{}, errors.New("missing pull request identity")
	}

	result, ok := program.pullRequestDiffForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	pendingReviewID := ""
	if pendingState, ok := program.pendingPullRequestReviewStateForSummary(summary); ok {
		pendingReviewID = strings.TrimSpace(pendingState.id)
	}
	detailDocument := program.inlineCommentDetailDocument(gui, view)
	renderedRows := program.currentPullRequestChangesRenderedRows(summary, result.data.Files, detailDocument.width)
	return pullRequestInlineCommentSelectionFromRenderedRows(repository, summary.Number, pendingReviewID, true, renderedRows, detailDocument, program.detailState.viewState)
}

func (program *Program) inlineCommentDetailDocument(gui *gocui.Gui, view *gocui.View) detailDocument {
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
	return detailDocument
}

func pullRequestInlineCommentSelectionFromRenderedRows(repository string, number int, pendingReviewID string, updateDetail bool, renderedRows []reviewDiffRenderedRow, detailDocument detailDocument, state detailViewState) (pullRequestInlineCommentSelection, error) {
	threadTarget, err := reviewDiffThreadTargetForSelection(renderedRows, detailDocument, state)
	if err != nil {
		return pullRequestInlineCommentSelection{}, err
	}

	return pullRequestInlineCommentSelection{
		target: pullRequestInlineCommentTarget{
			repository:    strings.TrimSpace(repository),
			number:        number,
			pendingReview: strings.TrimSpace(pendingReviewID),
			threadTarget:  threadTarget,
			updateDetail:  updateDetail,
		},
		initialBody: reviewInlineCommentSuggestionBody(threadTarget.Path, reviewDiffSelectedSnippet(renderedRows, detailDocument, state)),
	}, nil
}

func (program *Program) submitPullRequestInlineComment(target pullRequestInlineCommentTarget, body string) error {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return errors.New("missing pull request identity")
	}
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}

	pendingReviewID, err := program.pendingReviewIDForInlineComment(target)
	if err != nil {
		return err
	}
	if err := program.reviewMutations.AddPullRequestReviewThread(pendingReviewID, body, target.threadTarget); err != nil {
		return newTransientErrorPopupActionError(err)
	}

	target.pendingReview = pendingReviewID
	program.optimisticallyAppendInlineComment(target, body)
	program.setFeedback(FocusDetailView, pullRequestReviewInlineCommentSuccessMessage)
	return nil
}

func (program *Program) pendingReviewIDForInlineComment(target pullRequestInlineCommentTarget) (string, error) {
	if strings.TrimSpace(target.repository) == "" || target.number <= 0 {
		return "", errors.New("missing pull request identity")
	}
	if strings.TrimSpace(target.pendingReview) != "" {
		program.setPendingPullRequestReviewStateByIdentity(target.repository, target.number, target.pendingReview)
		return strings.TrimSpace(target.pendingReview), nil
	}
	if !program.hasReviewMutations() {
		return "", errors.New("github loader is unavailable")
	}

	pendingReviewID, err := program.reviewMutations.StartPendingPullRequestReview(target.repository, target.number)
	if err != nil {
		return "", newTransientErrorPopupActionError(err)
	}
	pendingReviewID = strings.TrimSpace(pendingReviewID)
	if pendingReviewID == "" {
		return "", errors.New("missing pull request review context")
	}
	program.setPendingPullRequestReviewStateByIdentity(target.repository, target.number, pendingReviewID)
	return pendingReviewID, nil
}

func (program *Program) browserChangesInlineCommentShortcutActive() bool {
	return program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == ChangesDetailTab
}
