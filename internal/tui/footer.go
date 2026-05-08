package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	viewUserFooterName         = "user-footer"
	viewPullRequestsFooterName = "pull-requests-footer"
	viewDetailFooterName       = "detail-footer"

	pullRequestDetailLoadingTitle = "Loading pull request detail..."
)

type paneFooterState struct {
	searchSummary string
}

func (state paneFooterState) Text() string {
	return strings.TrimSpace(state.searchSummary)
}

func (state paneFooterState) Visible() bool {
	return state.Text() != ""
}

func (program *Program) layoutPaneFooterViews(gui *gocui.Gui) error {
	for _, focus := range []Focus{FocusUserView, FocusPullRequestsView, FocusDetailView} {
		if err := program.layoutPaneFooterView(gui, focus); err != nil {
			return err
		}
	}

	return nil
}

func (program *Program) layoutPaneFooterView(gui *gocui.Gui, focus Focus) error {
	viewName := paneFooterViewName(focus)
	if !program.model.PaneVisible(focus) {
		return deleteViewIfPresent(gui, viewName)
	}

	state := program.paneFooterStateFor(focus)
	if !state.Visible() {
		return deleteViewIfPresent(gui, viewName)
	}

	view, err := program.layoutPaneBottomOverlayView(gui, viewName, paneViewName(focus))
	if err != nil {
		return err
	}

	program.configurePaneFooterView(view)
	program.renderPaneFooterView(view, state.Text())
	_, err = gui.SetViewOnTop(viewName)
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

func (program *Program) paneFooterStateFor(focus Focus) paneFooterState {
	if program.model.SearchActive() && program.model.SearchTarget() == focus {
		return paneFooterState{}
	}

	return paneFooterState{searchSummary: strings.TrimSpace(program.appliedSearchFooterText(focus))}
}

func (program *Program) statusLineKeyHintsText() string {
	if actionsPopupHints := strings.TrimSpace(program.actionsPopupKeyHintsText()); actionsPopupHints != "" {
		return actionsPopupHints
	}

	focus := program.model.Focus()
	if !program.shouldShowStatusLineKeyHints(focus) {
		return ""
	}

	return program.paneFooterKeyHintsText(focus)
}

func (program *Program) shouldShowStatusLineKeyHints(focus Focus) bool {
	if focus != program.model.Focus() || !program.model.PaneVisible(focus) {
		return false
	}
	if program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible() {
		return false
	}
	return isMainPaneFocus(focus)
}

func (program *Program) actionsPopupKeyHintsText() string {
	if !program.shouldShowActionsPopupStatusLineKeyHints() {
		return ""
	}
	if !program.assigneePickerVisible() && !program.assigneePickerLoading() {
		return ""
	}

	hints := []string{
		program.actionsPopupKeyHint("Search", "/", keybindingActionID{scope: keymapScopeActionsPopup, action: "focus_search"}),
		program.actionsPopupKeyHint("Toggle", "Enter", keybindingActionID{scope: keymapScopeActionsPopup, action: "execute_selected_action"}),
		program.actionsPopupKeyHint("Submit", "Alt+Enter", keybindingActionID{scope: keymapScopeActionsPopup, action: "submit_selected_picker"}),
	}
	return strings.Join(filterEmptyStrings(hints), ", ")
}

func (program *Program) shouldShowActionsPopupStatusLineKeyHints() bool {
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return false
	}
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible() {
		return false
	}
	return true
}

func (program *Program) paneFooterKeyHintsText(focus Focus) string {
	hints := []string{
		program.paneFooterKeyHint("Help", keybindingActionID{scope: keymapScopeMain, action: "toggle_help"}),
		program.paneFooterKeyHint("Search", keybindingActionID{scope: keymapScopeMain, action: "open_search"}),
	}
	if actionsHint := program.paneFooterActionsHint(focus); actionsHint != "" {
		hints = append(hints, actionsHint)
	}
	return strings.Join(filterEmptyStrings(hints), ", ")
}

func (program *Program) paneFooterKeyHint(label string, actionIDs ...keybindingActionID) string {
	resolvedKeys := strings.TrimSpace(program.resolvedKeyLabelsText(actionIDs...))
	if resolvedKeys == "" {
		return ""
	}
	return resolvedKeys + ": " + label
}

func (program *Program) actionsPopupKeyHint(label string, fallback string, actionIDs ...keybindingActionID) string {
	resolvedKeys := strings.TrimSpace(program.helpKeysOrFallback(fallback, actionIDs...))
	if resolvedKeys == "" {
		return ""
	}
	return resolvedKeys + ": " + label
}

func (program *Program) paneFooterActionsHint(focus Focus) string {
	actionID, ok := paneFooterActionsActionID(focus)
	if !ok || len(program.currentActionsPopupActions()) == 0 {
		return ""
	}
	return program.paneFooterKeyHint("Action", actionID)
}

func paneFooterActionsActionID(focus Focus) (keybindingActionID, bool) {
	switch focus {
	case FocusUserView:
		return keybindingActionID{scope: keymapScopeUser, action: "open_actions_popup"}, true
	case FocusPullRequestsView:
		return keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}, true
	case FocusDetailView:
		return keybindingActionID{scope: keymapScopeDetail, action: "open_actions_popup"}, true
	default:
		return keybindingActionID{}, false
	}
}

func (program *Program) appliedSearchFooterText(focus Focus) string {
	if program.reviewSession.active {
		switch focus {
		case FocusPullRequestsView:
			query := program.reviewFileTreeSearchQuery()
			return searchSummaryText(query, program.reviewFileTreeSearchMatchCount(query))
		case FocusDetailView:
			query := program.model.appliedSearchQuery(FocusDetailView, MyPullRequestsTab)
			return searchSummaryText(query, program.detailSearchMatchCount(query))
		default:
			return ""
		}
	}

	switch focus {
	case FocusPullRequestsView:
		query := program.model.appliedSearchQuery(FocusPullRequestsView, program.model.ActivePullRequestTab())
		return searchSummaryText(query, len(program.model.VisiblePullRequests()))
	case FocusDetailView:
		query := program.model.appliedSearchQuery(FocusDetailView, MyPullRequestsTab)
		return searchSummaryText(query, program.detailSearchMatchCount(query))
	default:
		query := program.model.appliedSearchQuery(FocusUserView, MyPullRequestsTab)
		return searchSummaryText(query, len(program.model.VisibleUsers()))
	}
}

func (program *Program) detailSearchMatchCount(query string) int {
	if strings.TrimSpace(query) == "" {
		return 0
	}

	return len(program.currentDetailDocument(nil).searchMatches(query))
}

func searchSummaryText(query string, count int) string {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return ""
	}

	return fmt.Sprintf("/%s (%d %s)", trimmedQuery, count, pluralize(count, "match", "matches"))
}

func (program *Program) configurePaneFooterView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
}

func (program *Program) renderPaneFooterView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}

func paneFooterViewName(focus Focus) string {
	switch focus {
	case FocusPullRequestsView:
		return viewPullRequestsFooterName
	case FocusDetailView:
		return viewDetailFooterName
	default:
		return viewUserFooterName
	}
}

func paneViewName(focus Focus) string {
	switch focus {
	case FocusPullRequestsView:
		return viewPullRequestsName
	case FocusDetailView:
		return viewDetailName
	default:
		return viewUserName
	}
}
