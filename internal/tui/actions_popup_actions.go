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
	actions := actionsPopupGrouped(actionsPopupGroupTheme, program.changeThemeActionsPopupAction())
	if program.pullRequestCache != nil {
		actions = append(actions, actionsPopupGrouped(actionsPopupGroupCache, program.clearCacheActionsPopupAction())...)
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

	actions := []actionsPopupAction{}
	if actionContext.IsReviewContext() {
		actions = append(actions,
			actionsPopupGrouped(actionsPopupGroupPullRequest,
				program.yankPullRequestURLActionsPopupAction(),
				program.openPullRequestInBrowserActionsPopupAction(),
				program.refreshPullRequestAction(),
			)...,
		)
		if assignAction, ok := program.currentAssignPullRequestAction(); ok {
			actions = append(actions, assignAction.withGroup(actionsPopupGroupPullRequest))
		}
		actions = append(actions,
			actionsPopupGrouped(actionsPopupGroupReview,
				program.submitPendingReviewApprovalAction(),
				program.submitPendingReviewCommentAction(),
				program.submitPendingReviewRequestChangesAction(),
			)...,
		)
		if inlineCommentAction, ok := program.currentReviewInlineCommentAction(); ok {
			actions = append(actions, inlineCommentAction.withGroup(actionsPopupGroupReview))
		}
	} else {
		pullRequestActions := []actionsPopupAction{program.startReviewAction()}
		if cancelPendingReviewAction, ok := program.currentCancelPendingPullRequestReviewAction(); ok {
			pullRequestActions = append(pullRequestActions, cancelPendingReviewAction)
		}
		pullRequestActions = append(pullRequestActions,
			program.reviewStoryAction(),
			program.yankPullRequestURLActionsPopupAction(),
			program.openPullRequestInBrowserActionsPopupAction(),
			program.openPullRequestByURLActionsPopupAction(),
			program.refreshPullRequestAction(),
			program.commendOnPrAction(),
		)
		actions = append(actions, actionsPopupGrouped(actionsPopupGroupPullRequest, pullRequestActions...)...)
		if assignAction, ok := program.currentAssignPullRequestAction(); ok {
			actions = append(actions, assignAction.withGroup(actionsPopupGroupPullRequest))
		}
		if stateActions := program.currentPullRequestStageAndMergeActions(); len(stateActions) > 0 {
			actions = append(actions, actionsPopupGrouped(actionsPopupGroupPullRequest, stateActions...)...)
		}
		actions = append(actions,
			actionsPopupGrouped(actionsPopupGroupPullRequest,
				program.editPullRequestTitleAction(),
				program.editPullRequestDescriptionAction(),
			)...,
		)
		actions = append(actions,
			actionsPopupGrouped(actionsPopupGroupReview,
				program.reviewApproveAction(),
				program.reviewCommentAction(),
				program.reviewRequestChangesAction(),
			)...,
		)
		if inlineCommentAction, ok := program.currentBrowserChangesInlineCommentAction(); ok {
			actions = append(actions, inlineCommentAction.withGroup(actionsPopupGroupReview))
		}
	}
	if reRequestReviewAction, ok := program.currentReRequestPullRequestReviewAction(); ok {
		actions = append(actions, reRequestReviewAction.withGroup(actionsPopupGroupReview))
	}
	if reactionAction, ok := program.currentReactionAction(); ok {
		actions = append(actions, reactionAction.withGroup(actionsPopupGroupReview))
	}
	if reactionRemovalAction, ok := program.currentReactionRemovalAction(); ok {
		actions = append(actions, reactionRemovalAction.withGroup(actionsPopupGroupReview))
	}
	for _, action := range program.currentPullRequestCommentEditActions() {
		actions = append(actions, action.withGroup(actionsPopupGroupReview))
	}
	if replyAction, ok := program.currentInlineCommentReplyAction(); ok {
		actions = append(actions, replyAction.withGroup(actionsPopupGroupReview))
	}
	for _, action := range program.currentInlineCommentEditActions() {
		actions = append(actions, action.withGroup(actionsPopupGroupReview))
	}
	if inlineCommentAction, ok := program.currentInlineCommentResolutionAction(); ok {
		actions = append(actions, inlineCommentAction.withGroup(actionsPopupGroupReview))
	}
	if !actionContext.IsReviewContext() && actionContext.ActiveView.Focus == FocusPullRequestsView {
		actions = append(actions, actionsPopupGrouped(actionsPopupGroupNavigation, program.pullRequestCustomSearchActionsPopupAction())...)
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

func (program *Program) updateActionsPopupSearch(query string) {
	program.model.UpdateActionsPopupSearch(query, matchingActionsPopupIndexes(program.currentActionsPopupActions(), query))
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
		execute: program.executeYankPullRequestURLAction,
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
		execute: program.executeCommentOnPullRequestAction,
	}
}

func (program *Program) executeCommentOnPullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	wasVisible := program.modalEditorVisible()
	if err := program.openPullRequestCommentComposer(gui, nil); err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) executeYankPullRequestURLAction(_ *gocui.Gui) actionsPopupActionResult {
	err := program.copySelectedPullRequestURL()
	switch {
	case err == nil:
		program.setFeedback(program.model.Focus(), yankSuccessMessage)
		return actionsPopupActionResult{closePopup: true}
	case errors.Is(err, ErrNoPullRequestURL):
		return actionsPopupActionResult{err: errors.New(yankUnavailableMessage)}
	default:
		return actionsPopupActionResult{err: errors.New(yankFailureMessage)}
	}
}
