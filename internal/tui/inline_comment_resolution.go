package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	inlineCommentResolvedSuccessMessage       = "Inline comment resolved"
	inlineCommentUnresolvedSuccessMessage     = "Inline comment marked unresolved"
	inlineCommentResolutionUnavailableMessage = "Inline comment resolution unavailable here"
)

type pullRequestReviewThreadActionTarget struct {
	repository string
	number     int
	threadID   string
	resolved   bool
}

func inlineCommentResolutionShortcutDescription(resolved bool) string {
	if resolved {
		return "Unresolve inline comment"
	}
	return "Resolve inline comment"
}

func inlineCommentResolutionShortcutHintLabel(resolved bool) string {
	if resolved {
		return "unresolve"
	}
	return "resolve"
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

func (program *Program) toggleInlineCommentResolutionShortcut(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgToggleInlineCommentResolutionRequested{})
}

func (program *Program) inlineCommentResolutionShortcutAvailable() bool {
	_, ok := program.selectedPullRequestReviewThreadActionTarget()
	return ok
}

func (program *Program) inlineCommentResolutionShortcutDescription() string {
	target, ok := program.selectedPullRequestReviewThreadActionTarget()
	if !ok {
		return ""
	}
	return inlineCommentResolutionShortcutDescription(target.resolved)
}

func (program *Program) inlineCommentResolutionShortcutHintLabel() string {
	target, ok := program.selectedPullRequestReviewThreadActionTarget()
	if !ok {
		return ""
	}
	return inlineCommentResolutionShortcutHintLabel(target.resolved)
}

func (program *Program) resolveInlineCommentAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.selectedPullRequestReviewThreadActionTarget(); ok {
		requested = MsgInlineCommentResolutionRequested{Target: target, Resolved: true}
	}
	return actionsPopupAction{
		id:        "resolve-inline-comment",
		title:     "Mark inline comment as resolved",
		icon:      actionsPopupResolveInlineCommentIcon,
		requested: requested,
	}
}

func (program *Program) unresolveInlineCommentAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.selectedPullRequestReviewThreadActionTarget(); ok {
		requested = MsgInlineCommentResolutionRequested{Target: target, Resolved: false}
	}
	return actionsPopupAction{
		id:        "unresolve-inline-comment",
		title:     "Mark inline comment as unresolved",
		icon:      actionsPopupResolveInlineCommentIcon,
		requested: requested,
	}
}

func (program *Program) selectedPullRequestReviewThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReviewThreadActionTarget{}, false
	}
	if program.reviewModeActive() {
		return program.selectedReviewDiffReviewThreadActionTarget()
	}
	return program.selectedBrowserInlineCommentThreadActionTarget()
}

func (program *Program) selectedBrowserInlineCommentThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	if !program.shouldShowPullRequestDetailTabs() {
		return pullRequestReviewThreadActionTarget{}, false
	}

	switch program.detailState.activeTab {
	case CommentsDetailTab:
		return program.selectedBrowserCommentsInlineCommentThreadActionTarget()
	case ChangesDetailTab:
		return program.selectedBrowserChangesInlineCommentThreadActionTarget()
	default:
		return pullRequestReviewThreadActionTarget{}, false
	}
}

func (program *Program) selectedBrowserCommentsInlineCommentThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestReviewThreadActionTarget{}, false
	}

	sectionAtCursor, ok := program.browserConversationSectionAtCursor(summary, result.detail, program.detailState.wrapWidth, program.detailState.viewState.cursor.line)
	if !ok || sectionAtCursor.section.inlineThread == nil {
		return pullRequestReviewThreadActionTarget{}, false
	}
	thread := *sectionAtCursor.section.inlineThread
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadActionTarget{}, false
	}

	return pullRequestReviewThreadActionTarget{
		repository: repository,
		number:     summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
		resolved:   thread.IsResolved,
	}, true
}

func (program *Program) selectedBrowserChangesInlineCommentThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	context, ok := program.currentBrowserChangesCursorContext()
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}
	thread, ok := reviewDiffThreadAtCursor(context.renderedRows, context.selection.document, context.selection.state)
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || context.summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadActionTarget{}, false
	}

	return pullRequestReviewThreadActionTarget{
		repository: repository,
		number:     context.summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
		resolved:   thread.IsResolved,
	}, true
}

func (program *Program) selectedReviewDiffReviewThreadActionTarget() (pullRequestReviewThreadActionTarget, bool) {
	if !program.reviewModeActive() {
		return pullRequestReviewThreadActionTarget{}, false
	}

	context, ok := program.currentReviewDiffCursorContext()
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}
	thread, ok := reviewDiffThreadAtCursor(context.renderedRows, context.selection.document, context.selection.state)
	if !ok {
		return pullRequestReviewThreadActionTarget{}, false
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || context.summary.Number <= 0 || !hasUsablePullRequestMutationID(thread.ID) {
		return pullRequestReviewThreadActionTarget{}, false
	}

	return pullRequestReviewThreadActionTarget{
		repository: repository,
		number:     context.summary.Number,
		threadID:   strings.TrimSpace(thread.ID),
		resolved:   thread.IsResolved,
	}, true
}

func browserConversationThreadSectionID(repository string, number int, threadID string) string {
	pullRequestKey := pullRequestMutationCacheKey(repository, number)
	trimmedThreadID := strings.TrimSpace(threadID)
	if pullRequestKey == "" || trimmedThreadID == "" {
		return ""
	}
	return browserDetailSectionID(pullRequestKey, "thread", 0, trimmedThreadID)
}

func browserChangesThreadSectionIDByIdentity(repository string, number int, threadID string) string {
	pullRequestKey := pullRequestMutationCacheKey(repository, number)
	trimmedThreadID := strings.TrimSpace(threadID)
	if pullRequestKey == "" || trimmedThreadID == "" {
		return ""
	}
	return browserDetailSectionID(pullRequestKey, "changes-thread", 0, trimmedThreadID)
}

func (program *Program) currentInlineCommentResolutionCollapsed(target pullRequestReviewThreadActionTarget) (bool, bool) {
	defaultCollapsed := target.resolved
	switch {
	case program.reviewModeActive() && !program.reviewSessionShowsDescription() && !program.reviewSessionShowsStoryChapter():
		return reviewDiffThreadCollapsed(reviewDiffThread{ID: target.threadID, IsResolved: defaultCollapsed}, program.navigationState.reviewSession.collapsedThreadIDs), true
	case program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == CommentsDetailTab:
		sectionID := browserConversationThreadSectionID(target.repository, target.number, target.threadID)
		if sectionID == "" {
			return false, false
		}
		return program.browserDetailSectionCollapsed(sectionID, defaultCollapsed), true
	case program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == ChangesDetailTab:
		sectionID := browserChangesThreadSectionIDByIdentity(target.repository, target.number, target.threadID)
		if sectionID == "" {
			return false, false
		}
		return program.browserDetailSectionCollapsed(sectionID, defaultCollapsed), true
	default:
		return false, false
	}
}

func (program *Program) applyInlineCommentResolutionCollapsed(target pullRequestReviewThreadActionTarget, collapsed bool) bool {
	switch {
	case program.reviewModeActive() && !program.reviewSessionShowsDescription() && !program.reviewSessionShowsStoryChapter():
		program.setReviewSessionThreadCollapsed(target.threadID, collapsed)
		program.invalidateReviewDiffRenderCache()
		return true
	case program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == CommentsDetailTab:
		return program.setBrowserDetailSectionCollapsed(browserConversationThreadSectionID(target.repository, target.number, target.threadID), collapsed)
	case program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == ChangesDetailTab:
		return program.setBrowserDetailSectionCollapsed(browserChangesThreadSectionIDByIdentity(target.repository, target.number, target.threadID), collapsed)
	default:
		return false
	}
}

type inlineCommentResolutionAsyncError struct {
	err                    error
	target                 pullRequestReviewThreadActionTarget
	previousCollapsed      bool
	rollbackCollapsedState bool
}

func newInlineCommentResolutionAsyncError(err error, target pullRequestReviewThreadActionTarget, previousCollapsed bool, rollbackCollapsedState bool) error {
	if err == nil {
		return nil
	}

	wrappedErr := newTransientErrorPopupActionError(err)
	if !rollbackCollapsedState {
		return wrappedErr
	}
	return inlineCommentResolutionAsyncError{err: wrappedErr, target: target, previousCollapsed: previousCollapsed, rollbackCollapsedState: true}
}

func (err inlineCommentResolutionAsyncError) Error() string {
	if err.err == nil {
		return ""
	}
	return err.err.Error()
}

func (err inlineCommentResolutionAsyncError) Unwrap() error {
	return err.err
}

func inlineCommentResolutionRollback(err error) (pullRequestReviewThreadActionTarget, bool, bool) {
	var actual inlineCommentResolutionAsyncError
	if !errors.As(err, &actual) || !actual.rollbackCollapsedState {
		return pullRequestReviewThreadActionTarget{}, false, false
	}
	return actual.target, actual.previousCollapsed, true
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
