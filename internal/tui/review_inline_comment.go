package tui

import (
	"errors"
	"strings"

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

func newReviewInlineCommentOpenDescriptor(selection pullRequestInlineCommentSelection) modalEditorOpenDescriptor {
	return newMultilineModalEditorOpenDescriptorWithSubmitDescriptor(pullRequestReviewInlineCommentComposerTitle, selection.initialBody, newReviewInlineCommentSubmitDescriptor(selection.target), reviewInlineCommentModalHeight)
}

func (program *Program) currentReviewInlineCommentAction() (actionsPopupAction, bool) {
	if !program.reviewModeActive() || program.model.Focus() != FocusDetailView {
		return actionsPopupAction{}, false
	}
	if _, err := program.selectedReviewInlineCommentSelection(); err != nil {
		return actionsPopupAction{}, false
	}

	return program.addInlineCommentAction(), true
}

func (program *Program) currentBrowserChangesInlineCommentAction() (actionsPopupAction, bool) {
	if program.reviewModeActive() || program.model.Focus() != FocusDetailView || !program.browserChangesInlineCommentShortcutActive() {
		return actionsPopupAction{}, false
	}
	if _, err := program.selectedBrowserChangesInlineCommentSelection(); err != nil {
		return actionsPopupAction{}, false
	}

	return program.addInlineCommentAction(), true
}

func (program *Program) addInlineCommentAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errReviewThreadTargetUnavailable)
	selection, err := program.selectedReviewInlineCommentSelection()
	if err != nil {
		selection, err = program.selectedBrowserChangesInlineCommentSelection()
	}
	if err == nil {
		requested = MsgModalEditorOpened{Descriptor: newReviewInlineCommentOpenDescriptor(selection)}
	}
	return actionsPopupAction{
		id:        "add-inline-comment",
		title:     "Add inline comment",
		icon:      actionsPopupCommentOnPullRequestIcon,
		requested: requested,
	}
}

func (program *Program) selectedReviewInlineCommentSelection() (pullRequestInlineCommentSelection, error) {
	if !program.reviewModeActive() {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(program.navigationState.reviewSession.summary.Repository))
	pendingReviewID := strings.TrimSpace(program.navigationState.reviewSession.pendingReviewID)
	if repository == "" || program.navigationState.reviewSession.summary.Number <= 0 || pendingReviewID == "" {
		return pullRequestInlineCommentSelection{}, errors.New("missing pull request review context")
	}

	context, ok := program.currentReviewDiffCursorContext()
	if !ok {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}
	return pullRequestInlineCommentSelectionFromRenderedRows(repository, context.summary.Number, pendingReviewID, false, context.renderedRows, context.selection.document, context.selection.state)
}

func (program *Program) selectedBrowserChangesInlineCommentSelection() (pullRequestInlineCommentSelection, error) {
	if program.reviewModeActive() || !program.browserChangesInlineCommentShortcutActive() {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}

	context, ok := program.currentBrowserChangesCursorContext()
	if !ok {
		return pullRequestInlineCommentSelection{}, errReviewThreadTargetUnavailable
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || context.summary.Number <= 0 {
		return pullRequestInlineCommentSelection{}, errors.New("missing pull request identity")
	}

	pendingReviewID := ""
	if pendingState, ok := program.pendingPullRequestReviewStateForSummary(context.summary); ok {
		pendingReviewID = strings.TrimSpace(pendingState.id)
	}
	return pullRequestInlineCommentSelectionFromRenderedRows(repository, context.summary.Number, pendingReviewID, true, context.renderedRows, context.selection.document, context.selection.state)
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

func (program *Program) browserChangesInlineCommentShortcutActive() bool {
	return program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == ChangesDetailTab
}
