package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	actionsPopupStartReviewIcon            = ""
	actionsPopupYankPullRequestURLIcon     = ""
	actionsPopupOpenPullRequestBrowserIcon = ""
	actionsPopupReviewApproveIcon          = ""
	actionsPopupReviewCommentIcon          = ""
	actionsPopupReviewRequestChangesIcon   = ""
	actionsPopupCommentOnPullRequestIcon   = ""
	actionsPopupEditPullRequestIcon        = ""
)

type actionsPopupAction struct {
	id       string
	title    string
	icon     string
	keywords []string
	execute  func(*gocui.Gui) actionsPopupActionResult
}

func (action actionsPopupAction) label() string {
	if strings.TrimSpace(action.icon) == "" {
		return action.title
	}
	return action.icon + " " + action.title
}

type actionsPopupActionResult struct {
	closePopup bool
	err        error
}

var errActionsPopupActionUnavailable = errors.New("action is unavailable")

func (program *Program) openActionsPopup(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return nil
	}

	actions := program.currentActionsPopupActions()
	if len(actions) == 0 {
		return nil
	}

	program.model.OpenActionsPopup(len(actions))
	program.actionsPopupSearchEditor = nil
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) closeActionsPopup(gui *gocui.Gui, _ *gocui.View) error {
	program.model.CloseActionsPopup()
	program.actionsPopupSearchEditor = nil
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) focusActionsPopupSearch(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.model.ClearPaneSearchQueries()
	program.actionsPopupSearchEditor = newLineEditor("")
	program.updateActionsPopupSearch("")
	program.model.FocusActionsPopupSearch()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) focusActionsPopupList(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.model.BlurActionsPopupSearch()
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionDown(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return nil
	}

	program.model.MoveActionsPopupSelectionDown()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionUp(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return nil
	}

	program.model.MoveActionsPopupSelectionUp()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) executeSelectedActionsPopupAction(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	action, ok := program.selectedActionsPopupAction()
	if !ok {
		return nil
	}

	result := action.execute(gui)
	if result.err != nil {
		program.actionsPopupErrorMessage = strings.TrimSpace(result.err.Error())
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}

	if result.closePopup {
		return program.closeActionsPopup(gui, nil)
	}
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) editActionsPopupSearch(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyEnter || key == gocui.KeyEsc || key == gocui.KeyCtrlLsqBracket {
		return false
	}
	if program.actionsPopupSearchEditor == nil {
		program.actionsPopupSearchEditor = newLineEditor(program.model.ActionsPopupSearchQuery())
	}
	if !program.actionsPopupSearchEditor.HandleKey(key, ch, mod) {
		return false
	}

	program.updateActionsPopupSearch(program.actionsPopupSearchEditor.Text())
	program.actionsPopupErrorMessage = ""
	if program.gui != nil {
		_ = program.refreshViews(program.gui)
		return true
	}

	program.configureActionsPopupSearchView(view)
	program.renderActionsPopupSearchView(view)
	return true
}

func (program *Program) currentActionsPopupActions() []actionsPopupAction {
	if !program.isPullRequestContext() {
		return nil
	}
	if program.reviewSession.active {
		return []actionsPopupAction{
			program.yankPullRequestURLActionsPopupAction(),
			program.openPullRequestInBrowserActionsPopupAction(),
			{id: "submit-pending-review-approval", title: pullRequestReviewSubmitApprovalTitle, icon: actionsPopupReviewApproveIcon, keywords: []string{"review", "submit", "approve", "approval", "lgtm", "pending"}, execute: program.executeSubmitPendingReviewApprovalAction},
			{id: "submit-pending-review-comment", title: pullRequestReviewSubmitCommentTitle, icon: actionsPopupReviewCommentIcon, keywords: []string{"review", "submit", "comment", "finish", "pending"}, execute: program.executeSubmitPendingReviewCommentAction},
			{id: "submit-pending-review-request-changes", title: pullRequestReviewSubmitRequestChangesTitle, icon: actionsPopupReviewRequestChangesIcon, keywords: []string{"review", "submit", "request", "changes", "block", "pending"}, execute: program.executeSubmitPendingReviewRequestChangesAction},
		}
	}

	return []actionsPopupAction{
		{id: "start-review", title: "Start review", icon: actionsPopupStartReviewIcon, keywords: []string{"start", "review", "pending", "session", "inline"}, execute: program.executeStartReviewAction},
		program.yankPullRequestURLActionsPopupAction(),
		program.openPullRequestInBrowserActionsPopupAction(),
		{id: "review-approve", title: "Review: Approve PR", icon: actionsPopupReviewApproveIcon, keywords: []string{"review", "approve", "lgtm", "shipit"}, execute: program.executeApprovePullRequestAction},
		{id: "review-comment", title: pullRequestReviewCommentComposerTitle, icon: actionsPopupReviewCommentIcon, keywords: []string{"review", "comment", "feedback"}, execute: program.executeReviewCommentAction},
		{id: "review-request-changes", title: pullRequestRequestChangesComposerTitle, icon: actionsPopupReviewRequestChangesIcon, keywords: []string{"review", "request", "changes", "block"}, execute: program.executeRequestChangesAction},
		{id: "comment-on-pr", title: pullRequestCommentComposerTitle, icon: actionsPopupCommentOnPullRequestIcon, keywords: []string{"comment", "reply", "discussion"}, execute: program.executeCommentOnPullRequestAction},
		{id: "edit-pull-request-title", title: pullRequestTitleEditorTitle, icon: actionsPopupEditPullRequestIcon, keywords: []string{"edit", "title", "rename", "subject"}, execute: program.executeEditPullRequestTitleAction},
		{id: "edit-pull-request-description", title: pullRequestDescriptionEditorTitle, icon: actionsPopupEditPullRequestIcon, keywords: []string{"edit", "description", "body", "summary"}, execute: program.executeEditPullRequestDescriptionAction},
	}
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
	for _, keyword := range action.keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			return true
		}
	}
	return false
}

func (program *Program) yankPullRequestURLActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{id: "yank-pull-request-url", title: "Yank URL to clipboard", icon: actionsPopupYankPullRequestURLIcon, keywords: []string{"yank", "copy", "clipboard", "url", "link"}, execute: program.executeYankPullRequestURLAction}
}

func (program *Program) openPullRequestInBrowserActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{id: "open-pull-request-in-browser", title: "Open PR in browser", icon: actionsPopupOpenPullRequestBrowserIcon, keywords: []string{"open", "browser", "web", "url", "link"}, execute: program.executeOpenPullRequestInBrowserAction}
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
