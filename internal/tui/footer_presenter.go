package tui

import "strings"

type footerPresenter struct {
	model                            *Model
	screenState                      ScreenState
	keyResolver                      keybindingLabelResolver
	helpVisible                      bool
	modalEditorVisible               bool
	searchPromptVisible              bool
	pullRequestBuildPopupVisible     bool
	assigneePickerVisible            bool
	notificationSelectionVisible     bool
	commentShortcutAvailable         bool
	inlineCommentResolutionAvailable bool
	inlineCommentResolutionHintLabel string
	pullRequestBrowserAvailable      bool
	actionsPopupAvailable            bool
	modalEditorSubmitAction          string
	modalEditorSubmitFallback        string
	paneSearchSummaries              map[Focus]string
}

func (presenter footerPresenter) paneFooterStateFor(focus Focus) paneFooterState {
	if presenter.model == nil {
		return paneFooterState{}
	}
	if presenter.model.SearchActive() && presenter.model.SearchTarget() == focus {
		return paneFooterState{}
	}

	return paneFooterState{searchSummary: strings.TrimSpace(presenter.paneSearchSummaries[focus])}
}

func (presenter footerPresenter) statusLineKeyHintsText() string {
	switch presenter.screenState.KeyHintContext() {
	case KeyHintContextModalEditor:
		return strings.TrimSpace(presenter.modalEditorKeyHintsText())
	case KeyHintContextActionsPopupSearch:
		return strings.TrimSpace(presenter.actionsPopupSearchKeyHintsText())
	case KeyHintContextActionsPopup:
		return strings.TrimSpace(presenter.actionsPopupKeyHintsText())
	case KeyHintContextSearch:
		return strings.TrimSpace(presenter.searchKeyHintsText())
	case KeyHintContextBuildInfo:
		return strings.TrimSpace(presenter.pullRequestBuildRunPopupKeyHintsText())
	case KeyHintContextMainPanel, KeyHintContextSidePanel:
		focus := presenter.screenState.ActiveView().Focus
		if !presenter.shouldShowStatusLineKeyHints(focus) {
			return ""
		}
		return presenter.paneFooterKeyHintsText(focus)
	default:
		return ""
	}
}

func (presenter footerPresenter) shouldShowStatusLineKeyHints(focus Focus) bool {
	if presenter.model == nil || focus != presenter.model.Focus() || !presenter.model.PaneVisible(focus) {
		return false
	}
	if presenter.helpVisible || presenter.model.SearchActive() || presenter.model.ActionsPopupVisible() || presenter.modalEditorVisible || presenter.pullRequestBuildPopupVisible {
		return false
	}
	return isMainPaneFocus(focus)
}

func (presenter footerPresenter) modalEditorKeyHintsText() string {
	if !presenter.shouldShowModalEditorStatusLineKeyHints() {
		return ""
	}

	submitHint := statusLineHintSpec{label: "submit", fallback: presenter.modalEditorSubmitFallback, actionIDs: []keybindingActionID{{scope: keymapScopeModalEditor, action: presenter.modalEditorSubmitAction}}}

	return presenter.statusLineKeyHints(
		submitHint,
		statusLineHintSpec{label: "editor", fallback: "Ctrl+G", actionIDs: []keybindingActionID{{scope: keymapScopeModalEditor, action: "open_external_editor"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeModalEditor, action: "cancel"}}},
	)
}

func (presenter footerPresenter) shouldShowModalEditorStatusLineKeyHints() bool {
	if !presenter.modalEditorVisible {
		return false
	}
	if presenter.helpVisible || presenter.model.SearchActive() || presenter.model.ActionsPopupVisible() || presenter.pullRequestBuildPopupVisible {
		return false
	}
	return true
}

func (presenter footerPresenter) actionsPopupSearchKeyHintsText() string {
	if !presenter.shouldShowActionsPopupSearchStatusLineKeyHints() {
		return ""
	}
	if presenter.assigneePickerVisible {
		return presenter.statusLineKeyHints(
			statusLineHintSpec{label: "next", fallback: "Ctrl+N/↓"},
			statusLineHintSpec{label: "previous", fallback: "Ctrl+P/↑"},
			statusLineHintSpec{label: "toggle", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
			statusLineHintSpec{label: "submit", fallback: "Alt+Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "submit_selected_picker"}}},
			statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "cancel"}}},
		)
	}

	return presenter.statusLineKeyHints(
		statusLineHintSpec{label: "next", fallback: "Ctrl+N/↓"},
		statusLineHintSpec{label: "previous", fallback: "Ctrl+P/↑"},
		statusLineHintSpec{label: "execute", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "cancel"}}},
	)
}

func (presenter footerPresenter) shouldShowActionsPopupSearchStatusLineKeyHints() bool {
	if presenter.model == nil || !presenter.model.ActionsPopupVisible() || !presenter.model.ActionsPopupSearchActive() {
		return false
	}
	if presenter.helpVisible || presenter.model.SearchActive() || presenter.modalEditorVisible || presenter.pullRequestBuildPopupVisible {
		return false
	}
	return true
}

func (presenter footerPresenter) actionsPopupKeyHintsText() string {
	if !presenter.shouldShowActionsPopupStatusLineKeyHints() {
		return ""
	}
	if presenter.assigneePickerVisible {
		return presenter.statusLineKeyHints(
			statusLineHintSpec{label: "search", fallback: "/", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "open_search"}}},
			statusLineHintSpec{label: "toggle", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
			statusLineHintSpec{label: "submit", fallback: "Alt+Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "submit_selected_picker"}}},
			statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "close"}}},
		)
	}

	return presenter.statusLineKeyHints(
		statusLineHintSpec{label: "search", fallback: "/", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "open_search"}}},
		statusLineHintSpec{label: "execute", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "execute_selected_action"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeActionsPopup, action: "close"}}},
	)
}

func (presenter footerPresenter) shouldShowActionsPopupStatusLineKeyHints() bool {
	if presenter.model == nil || !presenter.model.ActionsPopupVisible() || presenter.model.ActionsPopupSearchActive() {
		return false
	}
	if presenter.helpVisible || presenter.model.SearchActive() || presenter.modalEditorVisible || presenter.pullRequestBuildPopupVisible {
		return false
	}
	return true
}

func (presenter footerPresenter) searchKeyHintsText() string {
	if !presenter.shouldShowSearchStatusLineKeyHints() {
		return ""
	}

	return presenter.statusLineKeyHints(
		statusLineHintSpec{label: "submit", fallback: "Enter", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "submit"}}},
		statusLineHintSpec{label: "cancel", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopeSearch, action: "cancel"}}},
	)
}

func (presenter footerPresenter) shouldShowSearchStatusLineKeyHints() bool {
	if !presenter.searchPromptVisible {
		return false
	}
	if presenter.helpVisible || presenter.model.ActionsPopupVisible() || presenter.modalEditorVisible {
		return false
	}
	return true
}

func (presenter footerPresenter) pullRequestBuildRunPopupKeyHintsText() string {
	if !presenter.shouldShowPullRequestBuildRunPopupStatusLineKeyHints() {
		return ""
	}

	specs := []statusLineHintSpec{
		{label: "search", fallback: "/", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "open_search"}}},
	}
	if presenter.pullRequestBrowserAvailable {
		specs = append(specs, statusLineHintSpec{label: "browser", fallback: "Alt+B", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "open_pull_request_in_browser"}}})
	}
	specs = append(specs,
		statusLineHintSpec{label: "yank", fallback: "y", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "start_yank"}}},
		statusLineHintSpec{label: "copy", fallback: "Alt+Y", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "copy_content"}}},
		statusLineHintSpec{label: "back", fallback: "Escape", actionIDs: []keybindingActionID{{scope: keymapScopePullRequestBuildInfo, action: "close"}}},
	)
	return presenter.statusLineKeyHints(specs...)
}

func (presenter footerPresenter) shouldShowPullRequestBuildRunPopupStatusLineKeyHints() bool {
	if !presenter.pullRequestBuildPopupVisible || presenter.searchPromptVisible {
		return false
	}
	if presenter.helpVisible || presenter.model.ActionsPopupVisible() || presenter.modalEditorVisible {
		return false
	}
	return true
}

func (presenter footerPresenter) paneFooterKeyHintsText(focus Focus) string {
	hints := []string{
		presenter.paneFooterKeyHint("help", keybindingActionID{scope: keymapScopeMain, action: "toggle_help"}),
		presenter.paneFooterKeyHint("search", keybindingActionID{scope: keymapScopeMain, action: "open_search"}),
	}
	if focus == FocusNotificationsView && presenter.notificationSelectionVisible {
		hints = append(hints,
			presenter.paneFooterKeyHint("read", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_read"}),
			presenter.paneFooterKeyHint("done", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_done"}),
		)
	}
	if commentHint := presenter.paneFooterCommentHint(focus); commentHint != "" {
		hints = append(hints, commentHint)
	}
	if inlineCommentResolutionHint := presenter.paneFooterInlineCommentResolutionHint(focus); inlineCommentResolutionHint != "" {
		hints = append(hints, inlineCommentResolutionHint)
	}
	if actionsHint := presenter.paneFooterActionsHint(focus); actionsHint != "" {
		hints = append(hints, actionsHint)
	}
	return strings.Join(filterEmptyStrings(hints), ", ")
}

func (presenter footerPresenter) statusLineKeyHints(specs ...statusLineHintSpec) string {
	hints := make([]string, 0, len(specs))
	for _, spec := range specs {
		hints = append(hints, presenter.statusLineKeyHint(spec.label, spec.fallback, spec.actionIDs...))
	}
	return strings.Join(filterEmptyStrings(hints), ", ")
}

func (presenter footerPresenter) statusLineKeyHint(label string, fallback string, actionIDs ...keybindingActionID) string {
	resolvedKeys := strings.TrimSpace(presenter.statusLineKeyHintKeys(fallback, actionIDs...))
	if resolvedKeys == "" {
		return ""
	}
	return resolvedKeys + ": " + label
}

func (presenter footerPresenter) statusLineKeyHintKeys(fallback string, actionIDs ...keybindingActionID) string {
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(presenter.keyResolver.helpKeysOrFallback(fallback, actionIDs...))
	}
	return strings.TrimSpace(presenter.keyResolver.resolvedKeyLabelsText(actionIDs...))
}

func (presenter footerPresenter) paneFooterKeyHint(label string, actionIDs ...keybindingActionID) string {
	return presenter.statusLineKeyHint(label, "", actionIDs...)
}

func (presenter footerPresenter) paneFooterCommentHint(focus Focus) string {
	if !presenter.commentShortcutAvailable {
		return ""
	}
	if focus != FocusPullRequestsView && focus != FocusDetailView {
		return ""
	}
	return presenter.paneFooterOverriddenKeyHint("comment", keybindingActionID{scope: keymapScopePullRequests, action: "comment_on_pull_request"})
}

func (presenter footerPresenter) paneFooterOverriddenKeyHint(label string, actionIDs ...keybindingActionID) string {
	actualLabels, ok, hasOverride := presenter.keyResolver.resolvedKeyLabels(actionIDs...)
	if !ok || !hasOverride || len(actualLabels) == 0 {
		return ""
	}
	return strings.Join(formattedKeySequenceLabelsForDisplay(actualLabels), "/") + ": " + label
}

func (presenter footerPresenter) paneFooterInlineCommentResolutionHint(focus Focus) string {
	if !presenter.inlineCommentResolutionAvailable || focus != FocusDetailView || presenter.inlineCommentResolutionHintLabel == "" {
		return ""
	}
	return presenter.statusLineKeyHint(presenter.inlineCommentResolutionHintLabel, "R", keybindingActionID{scope: keymapScopePullRequests, action: "toggle_inline_comment_resolution"})
}

func (presenter footerPresenter) paneFooterActionsHint(focus Focus) string {
	actionID, ok := paneFooterActionsActionID(focus)
	if !ok || !presenter.actionsPopupAvailable {
		return ""
	}
	return presenter.paneFooterKeyHint("action", actionID)
}
