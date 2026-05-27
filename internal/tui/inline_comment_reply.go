package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	pullRequestInlineCommentReplyEditorTitle    = "Reply to inline comment"
	pullRequestInlineCommentReplySuccessMessage = "Inline reply posted"
	inlineCommentReplyUnavailableMessage        = "Reply to inline comment unavailable here"
)

type pullRequestReviewThreadReplyTarget struct {
	repository    string
	number        int
	pendingReview string
	threadID      string
}

func (program *Program) currentInlineCommentReplyAction() (actionsPopupAction, bool) {
	if _, ok := program.selectedPullRequestReviewThreadReplyTarget(); !ok {
		return actionsPopupAction{}, false
	}
	return program.replyToInlineCommentAction(), true
}

func (program *Program) replyToInlineCommentAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestReviewThreadReplyTarget()
	if ok {
		requested = MsgModalEditorOpened{Descriptor: newInlineCommentReplyOpenDescriptor(target)}
	}
	return actionsPopupAction{
		id:        "reply-to-inline-comment",
		title:     pullRequestInlineCommentReplyEditorTitle,
		icon:      actionsPopupCommentOnPullRequestIcon,
		requested: requested,
	}
}

func newInlineCommentReplyOpenDescriptor(target pullRequestReviewThreadReplyTarget) modalEditorOpenDescriptor {
	return newMultilineModalEditorOpenDescriptorWithSubmitDescriptor(pullRequestInlineCommentReplyEditorTitle, "", newInlineCommentReplySubmitDescriptor(target), reviewInlineCommentModalHeight)
}

func (program *Program) selectedPullRequestReviewThreadReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	if program.reviewModeActive() {
		return program.selectedReviewInlineCommentReplyTarget()
	}
	return program.selectedBrowserInlineCommentReplyTarget()
}

func (program *Program) selectedBrowserInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	switch program.detailState.activeTab {
	case CommentsDetailTab:
		return program.selectedBrowserConversationsInlineCommentReplyTarget()
	case ChangesDetailTab:
		return program.selectedBrowserChangesInlineCommentReplyTarget()
	default:
		return pullRequestReviewThreadReplyTarget{}, false
	}
}

func (program *Program) selectedBrowserConversationsInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, program.detailState.wrapWidth, program.detailState.viewState.cursor.line)
	if !ok || sectionAtCursor.section.inlineThread == nil || !sectionAtCursor.inBody {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	thread := *sectionAtCursor.section.inlineThread
	if _, ok := browserConversationInlineThreadCommentAtCursor(sectionAtCursor); !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	return pullRequestReviewThreadReplyTarget{
		repository: repository,
		number:     summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
	}, true
}

func (program *Program) selectedBrowserChangesInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	context, ok := program.currentBrowserChangesCursorContext()
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	thread, _, ok := reviewDiffCommentAtCursor(context.renderedRows, context.selection.document, context.selection.state)
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || context.summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	return pullRequestReviewThreadReplyTarget{
		repository: repository,
		number:     context.summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
	}, true
}

func (program *Program) selectedReviewInlineCommentReplyTarget() (pullRequestReviewThreadReplyTarget, bool) {
	if !program.reviewModeActive() {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	context, ok := program.currentReviewDiffCursorContext()
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	thread, _, ok := reviewDiffCommentAtCursor(context.renderedRows, context.selection.document, context.selection.state)
	if !ok {
		return pullRequestReviewThreadReplyTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || context.summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadReplyTarget{}, false
	}
	return pullRequestReviewThreadReplyTarget{
		repository:    repository,
		number:        context.summary.Number,
		pendingReview: strings.TrimSpace(program.navigationState.reviewSession.pendingReviewID),
		threadID:      strings.TrimSpace(thread.ID),
	}, true
}

func (program *Program) replyToInlineCommentShortcut(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgOpenInlineCommentReplyRequested{})
}

func (program *Program) inlineCommentReplyShortcutContextActive() bool {
	if program.model.Focus() != FocusDetailView {
		return false
	}
	if program.reviewModeActive() {
		return true
	}
	if !program.shouldShowPullRequestDetailTabs() {
		return false
	}
	switch program.detailState.activeTab {
	case CommentsDetailTab, ChangesDetailTab:
		return true
	default:
		return false
	}
}

func (program *Program) inlineCommentReplyShortcutAvailable() bool {
	if !program.inlineCommentReplyShortcutContextActive() {
		return false
	}
	_, ok := program.selectedPullRequestReviewThreadReplyTarget()
	return ok
}
