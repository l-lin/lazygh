package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	viewUserFooterName          = "user-footer"
	viewPullRequestsFooterName  = "pull-requests-footer"
	viewNotificationsFooterName = "notifications-footer"
	viewDetailFooterName        = "detail-footer"

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

func (program *Program) paneFooterStateFor(focus Focus) paneFooterState {
	if program.model.SearchActive() && program.model.SearchTarget() == focus {
		return paneFooterState{}
	}

	return paneFooterState{searchSummary: strings.TrimSpace(program.appliedSearchFooterText(focus))}
}

func (program *Program) statusLineKeyHintsText() string {
	switch program.screenState().KeyHintContext() {
	case KeyHintContextModalEditor:
		return strings.TrimSpace(program.modalEditorKeyHintsText())
	case KeyHintContextActionsPopupSearch:
		return strings.TrimSpace(program.actionsPopupSearchKeyHintsText())
	case KeyHintContextActionsPopup:
		return strings.TrimSpace(program.actionsPopupKeyHintsText())
	case KeyHintContextSearch:
		return strings.TrimSpace(program.searchKeyHintsText())
	case KeyHintContextBuildInfo:
		return strings.TrimSpace(program.pullRequestBuildRunPopupKeyHintsText())
	case KeyHintContextMainPanel, KeyHintContextSidePanel:
		focus := program.screenState().ActiveView().Focus
		if !program.shouldShowStatusLineKeyHints(focus) {
			return ""
		}
		return program.paneFooterKeyHintsText(focus)
	default:
		return ""
	}
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

func (program *Program) modalEditorKeyHintsText() string {
	if !program.shouldShowModalEditorStatusLineKeyHints() {
		return ""
	}

	submitHint := statusLineHintSpec{label: "submit", fallback: "Alt+Enter", actionIDs: []keybindingActionID{{scope: keymapScopeModalEditor, action: "submit"}}}
	if program.modalEditor != nil && program.modalEditor.submitOnEnter {
		submitHint = statusLineHintSpec{label: "submit", fallback: "Enter"}
	}

	return program.statusLineKeyHints(
		submitHint,
		statusLineHintSpec{label: "editor", fallback: "Ctrl+G", actionIDs: []keybindingActionID{{scope: keymapScopeModalEditor, action: "open_external_editor"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeModalEditor, action: "cancel"}}},
	)
}

func (program *Program) shouldShowModalEditorStatusLineKeyHints() bool {
	if !program.modalEditorVisible() {
		return false
	}
	if program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.pullRequestBuildRunPopupVisible() {
		return false
	}
	return true
}

func (program *Program) actionsPopupSearchKeyHintsText() string {
	if !program.shouldShowActionsPopupSearchStatusLineKeyHints() {
		return ""
	}
	if program.assigneePickerLoading() {
		return program.statusLineKeyHints(
			statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "cancel"}}},
		)
	}
	if program.assigneePickerVisible() {
		return program.statusLineKeyHints(
			statusLineHintSpec{label: "next", fallback: "Ctrl+N/↓"},
			statusLineHintSpec{label: "previous", fallback: "Ctrl+P/↑"},
			statusLineHintSpec{label: "toggle", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
			statusLineHintSpec{label: "submit", fallback: "Alt+Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "submit_selected_picker"}}},
			statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "cancel"}}},
		)
	}

	return program.statusLineKeyHints(
		statusLineHintSpec{label: "next", fallback: "Ctrl+N/↓"},
		statusLineHintSpec{label: "previous", fallback: "Ctrl+P/↑"},
		statusLineHintSpec{label: "execute", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "cancel"}}},
	)
}

func (program *Program) shouldShowActionsPopupSearchStatusLineKeyHints() bool {
	if !program.model.ActionsPopupVisible() || !program.model.ActionsPopupSearchActive() {
		return false
	}
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible() {
		return false
	}
	return true
}

func (program *Program) actionsPopupKeyHintsText() string {
	if !program.shouldShowActionsPopupStatusLineKeyHints() {
		return ""
	}
	if program.assigneePickerLoading() {
		return program.statusLineKeyHints(
			statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "close"}}},
		)
	}
	if program.assigneePickerVisible() {
		return program.statusLineKeyHints(
			statusLineHintSpec{label: "search", fallback: "/", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "open_search"}}},
			statusLineHintSpec{label: "toggle", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
			statusLineHintSpec{label: "submit", fallback: "Alt+Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "submit_selected_picker"}}},
			statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "close"}}},
		)
	}

	return program.statusLineKeyHints(
		statusLineHintSpec{label: "search", fallback: "/", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "open_search"}}},
		statusLineHintSpec{label: "execute", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "close"}}},
	)
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

func (program *Program) searchKeyHintsText() string {
	if !program.shouldShowSearchStatusLineKeyHints() {
		return ""
	}

	return program.statusLineKeyHints(
		statusLineHintSpec{label: "submit", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "submit"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "cancel"}}},
	)
}

func (program *Program) shouldShowSearchStatusLineKeyHints() bool {
	if !program.searchPromptVisible() {
		return false
	}
	if program.helpVisible || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return false
	}
	return true
}

func (program *Program) pullRequestBuildRunPopupKeyHintsText() string {
	if !program.shouldShowPullRequestBuildRunPopupStatusLineKeyHints() {
		return ""
	}

	return program.statusLineKeyHints(
		statusLineHintSpec{label: "search", fallback: "/", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "open_search"}}},
		statusLineHintSpec{label: "copy", fallback: "y", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "copy_content"}}},
		statusLineHintSpec{label: "back", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "close"}}},
	)
}

func (program *Program) shouldShowPullRequestBuildRunPopupStatusLineKeyHints() bool {
	if !program.pullRequestBuildRunPopupVisible() || program.searchPromptVisible() {
		return false
	}
	if program.helpVisible || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return false
	}
	return true
}

func (program *Program) paneFooterKeyHintsText(focus Focus) string {
	hints := []string{
		program.paneFooterKeyHint("help", keybindingActionID{scope: keymapScopeMain, action: "toggle_help"}),
		program.paneFooterKeyHint("search", keybindingActionID{scope: keymapScopeMain, action: "open_search"}),
	}
	if focus == FocusNotificationsView {
		if _, ok := program.selectedNotificationActionTarget(); ok {
			hints = append(hints,
				program.paneFooterKeyHint("read", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_read"}),
				program.paneFooterKeyHint("done", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_done"}),
			)
		}
	}
	if actionsHint := program.paneFooterActionsHint(focus); actionsHint != "" {
		hints = append(hints, actionsHint)
	}
	return strings.Join(filterEmptyStrings(hints), ", ")
}

type statusLineHintSpec struct {
	label     string
	fallback  string
	actionIDs []keybindingActionID
}

func (program *Program) statusLineKeyHints(specs ...statusLineHintSpec) string {
	hints := make([]string, 0, len(specs))
	for _, spec := range specs {
		hints = append(hints, program.statusLineKeyHint(spec.label, spec.fallback, spec.actionIDs...))
	}
	return strings.Join(filterEmptyStrings(hints), ", ")
}

func (program *Program) statusLineKeyHint(label string, fallback string, actionIDs ...keybindingActionID) string {
	resolvedKeys := strings.TrimSpace(program.statusLineKeyHintKeys(fallback, actionIDs...))
	if resolvedKeys == "" {
		return ""
	}
	return resolvedKeys + ": " + label
}

func (program *Program) statusLineKeyHintKeys(fallback string, actionIDs ...keybindingActionID) string {
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(program.helpKeysOrFallback(fallback, actionIDs...))
	}
	return strings.TrimSpace(program.resolvedKeyLabelsText(actionIDs...))
}

func (program *Program) paneFooterKeyHint(label string, actionIDs ...keybindingActionID) string {
	return program.statusLineKeyHint(label, "", actionIDs...)
}

func (program *Program) paneFooterActionsHint(focus Focus) string {
	actionID, ok := paneFooterActionsActionID(focus)
	if !ok || len(program.currentActionsPopupActions()) == 0 {
		return ""
	}
	return program.paneFooterKeyHint("action", actionID)
}

func paneFooterActionsActionID(focus Focus) (keybindingActionID, bool) {
	switch focus {
	case FocusUserView:
		return keybindingActionID{scope: keymapScopeUser, action: "open_actions_popup"}, true
	case FocusPullRequestsView:
		return keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}, true
	case FocusNotificationsView:
		return keybindingActionID{scope: keymapScopeNotifications, action: "open_actions_popup"}, true
	case FocusDetailView:
		return keybindingActionID{scope: keymapScopeGlobal, action: "open_actions_popup"}, true
	default:
		return keybindingActionID{}, false
	}
}

func (program *Program) appliedSearchFooterText(focus Focus) string {
	if program.actionContext().IsReviewContext() {
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
	case FocusNotificationsView:
		query := program.model.appliedSearchQuery(FocusNotificationsView, MyPullRequestsTab)
		return searchSummaryText(query, len(program.model.VisibleNotifications()))
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

func paneViewName(focus Focus) string {
	switch focus {
	case FocusPullRequestsView:
		return viewPullRequestsName
	case FocusNotificationsView:
		return viewNotificationsName
	case FocusDetailView:
		return viewDetailName
	default:
		return viewUserName
	}
}
