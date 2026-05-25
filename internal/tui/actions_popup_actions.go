package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) currentActionsPopupActions() []actionsPopupAction {
	if program.assigneePickerVisible() {
		return program.currentAssigneePickerActions()
	}
	if program.themePickerVisible() {
		return program.currentThemePickerActions()
	}
	if program.reactionPickerVisible() {
		return program.currentReactionPickerActions()
	}
	if program.pullRequestBuildRunPopupVisible() {
		return nil
	}

	actions := program.currentContextualActionsPopupActions()
	return append(actions, program.currentGlobalActionsPopupActions()...)
}

func (program *Program) currentGlobalActionsPopupActions() []actionsPopupAction {
	actions := []actionsPopupAction{}
	if action, ok := program.currentRecentErrorsActionsPopupAction(); ok {
		actions = append(actions, actionsPopupGrouped(actionsPopupGroupErrors, action)...)
	}
	actions = append(actions, actionsPopupGrouped(actionsPopupGroupTheme, program.changeThemeActionsPopupAction())...)
	cacheActions := program.currentCacheActionsPopupActions(program.actionContext())
	if len(cacheActions) > 0 {
		actions = append(actions, actionsPopupGrouped(actionsPopupGroupCache, cacheActions...)...)
	}
	return actions
}

func (program *Program) currentCacheActionsPopupActions(actionContext ActionContext) []actionsPopupAction {
	actions := make([]actionsPopupAction, 0, 3)
	if actionContext.IsPullRequestContext() {
		actions = append(actions, program.refreshPullRequestAction())
	}
	actions = append(actions, program.refreshPullRequestListAction())
	if program.pullRequestCache != nil {
		actions = append(actions, program.clearCacheActionsPopupAction())
	}
	return actions
}

func (program *Program) currentContextualActionsPopupActions() []actionsPopupAction {
	actionContext := program.actionContext()
	if actionContext.IsNotificationContext() {
		return program.currentNotificationActionsPopupActions()
	}
	if !actionContext.IsPullRequestContext() {
		return nil
	}

	reviewActions, pullRequestActions := program.currentPullRequestActionsPopupGroupActions(actionContext)
	if reRequestReviewAction, ok := program.currentReRequestPullRequestReviewAction(); ok {
		reviewActions = append(reviewActions, reRequestReviewAction)
	}
	if reactionAction, ok := program.currentReactionAction(); ok {
		reviewActions, pullRequestActions = appendActionsPopupActionToMatchingGroup(reviewActions, pullRequestActions, reactionAction)
	}
	if reactionRemovalAction, ok := program.currentReactionRemovalAction(); ok {
		reviewActions, pullRequestActions = appendActionsPopupActionToMatchingGroup(reviewActions, pullRequestActions, reactionRemovalAction)
	}
	for _, action := range program.currentPullRequestCommentEditActions() {
		reviewActions = append(reviewActions, action)
	}
	if replyAction, ok := program.currentInlineCommentReplyAction(); ok {
		reviewActions = append(reviewActions, replyAction)
	}
	for _, action := range program.currentInlineCommentEditActions() {
		reviewActions = append(reviewActions, action)
	}
	if inlineCommentAction, ok := program.currentInlineCommentResolutionAction(); ok {
		reviewActions = append(reviewActions, inlineCommentAction)
	}

	actions := make([]actionsPopupAction, 0, len(reviewActions)+len(pullRequestActions)+2)
	if len(reviewActions) > 0 {
		actions = append(actions, actionsPopupGrouped(actionsPopupGroupReview, reviewActions...)...)
	}
	if len(pullRequestActions) > 0 {
		actions = append(actions, actionsPopupGrouped(actionsPopupGroupPullRequest, pullRequestActions...)...)
	}
	if program.detailCursorActionsAvailable() && program.detailCursorHasBuildLink() {
		actions = append(actions,
			actionsPopupGrouped(actionsPopupGroupNavigation,
				program.pullRequestBuildRunActionsPopupAction(),
				program.pullRequestBuildRunLogsActionsPopupAction(),
			)...,
		)
	}
	if program.detailCursorActionsAvailable() && program.detailCursorHasLink() {
		actions = append(actions, program.openLinkUnderCursorActionsPopupAction().withGroup(actionsPopupGroupNavigation))
	}
	return actions
}

func (program *Program) currentPullRequestActionsPopupGroupActions(actionContext ActionContext) ([]actionsPopupAction, []actionsPopupAction) {
	reviewActions := []actionsPopupAction{}
	pullRequestActions := []actionsPopupAction{}
	if actionContext.IsReviewContext() {
		reviewActions = append(reviewActions,
			program.submitPendingReviewApprovalAction(),
			program.submitPendingReviewCommentAction(),
			program.submitPendingReviewRequestChangesAction(),
		)
		if inlineCommentAction, ok := program.currentReviewInlineCommentAction(); ok {
			reviewActions = append(reviewActions, inlineCommentAction)
		}
		pullRequestActions = append(pullRequestActions,
			program.yankPullRequestURLActionsPopupAction(),
			program.openPullRequestInBrowserActionsPopupAction(),
		)
		if assignAction, ok := program.currentAssignPullRequestAction(); ok {
			pullRequestActions = append(pullRequestActions, assignAction)
		}
		return reviewActions, pullRequestActions
	}

	reviewActions = append(reviewActions,
		program.startReviewAction(),
		program.reviewStoryAction(),
	)
	if cancelPendingReviewAction, ok := program.currentCancelPendingPullRequestReviewAction(); ok {
		reviewActions = append(reviewActions, cancelPendingReviewAction)
	}
	reviewActions = append(reviewActions,
		program.reviewApproveAction(),
		program.reviewCommentAction(),
		program.reviewRequestChangesAction(),
	)
	if inlineCommentAction, ok := program.currentBrowserChangesInlineCommentAction(); ok {
		reviewActions = append(reviewActions, inlineCommentAction)
	}

	pullRequestActions = append(pullRequestActions,
		program.yankPullRequestURLActionsPopupAction(),
		program.openPullRequestInBrowserActionsPopupAction(),
		program.openPullRequestByURLActionsPopupAction(),
	)
	if actionContext.ActiveView.Focus == FocusPullRequestsView {
		pullRequestActions = append(pullRequestActions, program.pullRequestCustomSearchActionsPopupAction())
	}
	pullRequestActions = append(pullRequestActions,
		program.commendOnPrAction(),
	)
	if assignAction, ok := program.currentAssignPullRequestAction(); ok {
		pullRequestActions = append(pullRequestActions, assignAction)
	}
	pullRequestActions = append(pullRequestActions, program.currentPullRequestStageAndMergeActions()...)
	pullRequestActions = append(pullRequestActions,
		program.editPullRequestTitleAction(),
		program.editPullRequestDescriptionAction(),
	)
	return reviewActions, pullRequestActions
}

func appendActionsPopupActionToMatchingGroup(reviewActions []actionsPopupAction, pullRequestActions []actionsPopupAction, action actionsPopupAction) ([]actionsPopupAction, []actionsPopupAction) {
	if strings.TrimSpace(action.group) == actionsPopupGroupPullRequest {
		return reviewActions, append(pullRequestActions, action)
	}
	return append(reviewActions, action.withGroup(actionsPopupGroupReview)), pullRequestActions
}

func (program *Program) selectedActionsPopupAction() (actionsPopupAction, bool) {
	actions := program.currentActionsPopupActions()
	filteredIndexes := program.model.ActionsPopupFilteredActionIndexes()
	if len(actions) == 0 || len(filteredIndexes) == 0 {
		return actionsPopupAction{}, false
	}

	selectedIndex := program.model.ActionsPopupSelectedActionIndex()
	if selectedIndex < 0 || selectedIndex >= len(actions) {
		return actionsPopupAction{}, false
	}
	if indexOfInt(filteredIndexes, selectedIndex) < 0 {
		return actionsPopupAction{}, false
	}

	return actions[selectedIndex], true
}

func (program *Program) currentActionsPopupMatchingIndexes(query string) []int {
	if program.assigneePickerVisible() {
		return program.matchingAssigneePickerIndexes(query)
	}
	return matchingActionsPopupIndexes(program.currentActionsPopupActions(), query)
}

func (program *Program) syncActionsPopupSearch() {
	if !program.model.ActionsPopupVisible() {
		return
	}

	program.updateActionsPopupSearch(program.model.ActionsPopupSearchQuery())
}

func matchingActionsPopupIndexes(actions []actionsPopupAction, query string) []int {
	trimmedQuery := strings.ToLower(strings.TrimSpace(query))
	if trimmedQuery == "" {
		return actionIndexes(len(actions))
	}

	matchingIndexes := make([]int, 0, len(actions))
	for index, action := range actions {
		if actionsPopupActionMatchesQuery(action, trimmedQuery) {
			matchingIndexes = append(matchingIndexes, index)
		}
	}
	return matchingIndexes
}

func actionsPopupActionMatchesQuery(action actionsPopupAction, query string) bool {
	if query == "" {
		return true
	}
	for _, term := range actionsPopupActionSearchTerms(action) {
		if strings.Contains(strings.ToLower(term), query) {
			return true
		}
	}
	return false
}

func actionsPopupActionSearchTerms(action actionsPopupAction) []string {
	terms := []string{action.title}
	terms = append(terms, action.keywords...)
	terms = append(terms, actionsPopupDefaultKeywords(action)...)
	return filterEmptyStrings(terms)
}

func actionsPopupDefaultKeywords(action actionsPopupAction) []string {
	switch {
	case action.id == "yank-pull-request-url":
		return []string{"copy", "copy link"}
	case action.id == "open-pull-request-in-browser", action.id == "open-notification-in-browser":
		return []string{"web", "github"}
	case action.id == "comment-on-pr", action.id == "review-comment", action.id == "submit-pending-review-comment":
		return []string{"reply", "message", "feedback"}
	case action.id == "start-review":
		return []string{"begin", "review mode"}
	case action.id == "cancel-pending-review":
		return []string{"discard", "abandon"}
	case action.id == "review-pr-as-story":
		return []string{"story mode", "narrative"}
	case action.id == "refresh-current-pull-request-information":
		return []string{"reload", "sync"}
	case action.id == "refresh-pull-request-list":
		return []string{"reload", "sync", "list"}
	case action.id == "edit-pull-request-title":
		return []string{"rename", "headline"}
	case action.id == "edit-pull-request-description":
		return []string{"body", "details"}
	case action.id == "review-approve", action.id == "submit-pending-review-approval":
		return []string{"lgtm", "shipit"}
	case action.id == "review-request-changes", action.id == "submit-pending-review-request-changes":
		return []string{"changes requested", "block"}
	case action.id == "assign-pull-request":
		return []string{"owner", "assignee"}
	case action.id == "mark-pull-request-ready-for-review":
		return []string{"ready", "undraft"}
	case action.id == "convert-pull-request-to-draft":
		return []string{"draft", "wip"}
	case action.id == "close-pull-request":
		return []string{"archive", "decline"}
	case action.id == "reopen-pull-request":
		return []string{"open again", "restore"}
	case action.id == "squash-merge-pull-request":
		return []string{"merge", "land"}
	case action.id == "update-pull-request-branch":
		return []string{"rebase", "sync base"}
	case action.id == "change-theme":
		return []string{"colors", "palette"}
	case action.id == "clear-cache":
		return []string{"reset", "wipe", "cleanup"}
	case action.id == "mark-notification-read", action.id == "mark-all-notifications-read":
		return []string{"ack", "seen"}
	case action.id == "mark-notification-done", action.id == "mark-all-notifications-done":
		return []string{"archive", "dismiss"}
	case action.id == "reply-to-inline-comment":
		return []string{"respond"}
	case action.id == "update-inline-comment", action.id == "update-pull-request-comment":
		return []string{"edit"}
	case action.id == "delete-inline-comment", action.id == "delete-pull-request-comment":
		return []string{"remove"}
	case action.id == "resolve-inline-comment":
		return []string{"fix", "done"}
	case action.id == "unresolve-inline-comment":
		return []string{"reopen"}
	case action.id == "add-inline-comment":
		return []string{"note"}
	case action.id == "open-link-under-cursor":
		return []string{"browse", "visit"}
	case action.id == "view-build-run":
		return []string{"checks", "workflow", "ci"}
	case action.id == "view-build-run-job-logs":
		return []string{"logs", "output", "ci"}
	case action.id == "view-recent-errors":
		return []string{"history", "failures", "logs"}
	case action.id == "add-reaction":
		return []string{"emoji", "react"}
	case strings.HasPrefix(action.id, "remove-reaction-"):
		return []string{"emoji", "react", "unreact"}
	case strings.HasPrefix(action.id, "reaction-"):
		return []string{"emoji", "react"}
	case strings.HasPrefix(action.id, "theme-"):
		return []string{"color", "palette"}
	case strings.HasPrefix(action.id, "re-request-review-"):
		return []string{"request again", "ping reviewer"}
	default:
		return nil
	}
}

func (program *Program) yankPullRequestURLActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "yank-pull-request-url",
		title:   "Yank URL to clipboard",
		icon:    actionsPopupYankPullRequestURLIcon,
		execute: actionsPopupExecuteErr(program.executeYankPullRequestURLAction),
	}
}

func (program *Program) openPullRequestInBrowserActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "open-pull-request-in-browser",
		title:   "Open PR in browser",
		icon:    actionsPopupOpenPullRequestBrowserIcon,
		execute: program.executeOpenPullRequestInBrowserAction,
	}
}

func (program *Program) commendOnPrAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "comment-on-pr",
		title:   pullRequestCommentComposerTitle,
		icon:    actionsPopupCommentOnPullRequestIcon,
		execute: actionsPopupExecuteErr(program.executeCommentOnPullRequestAction),
	}
}

func (program *Program) executeCommentOnPullRequestAction(gui *gocui.Gui) error {
	return program.openModalEditorFromActionsPopup(gui, func(gui *gocui.Gui) error {
		return program.openPullRequestCommentComposer(gui, nil)
	})
}

func (program *Program) executeYankPullRequestURLAction(gui *gocui.Gui) error {
	err := program.copySelectedPullRequestURL()
	switch {
	case err == nil:
		return program.dispatch(gui, MsgActionsPopupClosedWithFeedback{Target: program.model.Focus(), Message: yankSuccessMessage})
	case errors.Is(err, ErrNoPullRequestURL):
		return errors.New(yankUnavailableMessage)
	default:
		return errors.New(yankFailureMessage)
	}
}
