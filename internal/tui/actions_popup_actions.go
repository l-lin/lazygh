package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) currentActionsPopupActions() []actionsPopupAction {
	if program.assigneePickerLoading() {
		return nil
	}
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
	if program.isNotificationContext() {
		return program.currentNotificationActionsPopupActions()
	}
	if !program.isPullRequestContext() {
		return nil
	}

	actions := []actionsPopupAction{}
	if program.reviewSession.active {
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
		actions = append(actions,
			actionsPopupGrouped(actionsPopupGroupPullRequest,
				program.startReviewAction(),
				program.reviewStoryAction(),
				program.yankPullRequestURLActionsPopupAction(),
				program.openPullRequestInBrowserActionsPopupAction(),
				program.refreshPullRequestAction(),
				program.commendOnPrAction(),
			)...,
		)
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
	if reactionAction, ok := program.currentReactionAction(); ok {
		actions = append(actions, reactionAction.withGroup(actionsPopupGroupReview))
	}
	if reactionRemovalAction, ok := program.currentReactionRemovalAction(); ok {
		actions = append(actions, reactionRemovalAction.withGroup(actionsPopupGroupReview))
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
	if strings.Contains(strings.ToLower(action.title), query) {
		return true
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(action.group)), query)
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
