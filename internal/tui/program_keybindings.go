package tui

import (
	"strings"
	"unicode/utf8"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"github.com/jesseduffield/gocui"
)

type keybindingSpec struct {
	viewName string
	key      any
	handler  func(*gocui.Gui, *gocui.View) error
}

type keybindingDefinition struct {
	key     any
	handler func(*gocui.Gui, *gocui.View) error
}

type configuredKeybinding struct {
	value any
	label string
}

type keybindingActionID struct {
	scope  string
	action string
}

type keybindingAction struct {
	id              keybindingActionID
	viewNames       []string
	defaultBindings []configuredKeybinding
	handler         func(*gocui.Gui, *gocui.View) error
}

type resolvedKeybindingAction struct {
	action     keybindingAction
	bindings   []configuredKeybinding
	overridden bool
}

const (
	keymapScopeGlobal             = "global"
	keymapScopeMain               = "main"
	keymapScopeSide               = "side"
	keymapScopeUser               = "user"
	keymapScopePullRequests       = "pull_requests"
	keymapScopeDetail             = "detail"
	keymapScopeSearch             = "search"
	keymapScopeActionsPopup       = "actions_popup"
	keymapScopeActionsPopupSearch = "actions_popup_search"
	keymapScopeModalEditor        = "modal_editor"
	keymapScopeHelp               = "help"
)

var mainPaneViewNames = []string{viewUserName, viewPullRequestsName, viewDetailName}
var sidePaneViewNames = []string{viewUserName, viewPullRequestsName}

func escapeKeybindingDefinitions(handler func(*gocui.Gui, *gocui.View) error) []keybindingDefinition {
	return []keybindingDefinition{
		{key: gocui.KeyEsc, handler: handler},
		{key: gocui.KeyCtrlLsqBracket, handler: handler},
	}
}

func dismissKeybindingDefinitions(handler func(*gocui.Gui, *gocui.View) error) []keybindingDefinition {
	return append(escapeKeybindingDefinitions(handler), keybindingDefinition{key: 'q', handler: handler})
}

func bindingsForView(viewName string, definitions ...keybindingDefinition) []keybindingSpec {
	return bindingsForViews([]string{viewName}, definitions...)
}

func bindingsForViews(viewNames []string, definitions ...keybindingDefinition) []keybindingSpec {
	specs := make([]keybindingSpec, 0, len(viewNames)*len(definitions))
	for _, viewName := range viewNames {
		for _, definition := range definitions {
			specs = append(specs, keybindingSpec{
				viewName: viewName,
				key:      definition.key,
				handler:  definition.handler,
			})
		}
	}

	return specs
}

func (program *Program) ApplyKeymapOverrides(overrides appconfig.KeymapOverrides) {
	program.keymapOverrides = copyKeymapOverrides(overrides)
}

func copyKeymapOverrides(overrides appconfig.KeymapOverrides) appconfig.KeymapOverrides {
	if len(overrides) == 0 {
		return nil
	}

	copiedScopes := make(appconfig.KeymapOverrides, len(overrides))
	for scopeName, actions := range overrides {
		copiedActions := make(map[string][]string, len(actions))
		for actionName, bindings := range actions {
			copiedActions[actionName] = append([]string(nil), bindings...)
		}
		copiedScopes[scopeName] = copiedActions
	}

	return copiedScopes
}

func (program *Program) setKeybindings(gui *gocui.Gui) error {
	for _, binding := range program.keybindingSpecs() {
		if err := gui.SetKeybinding(binding.viewName, binding.key, gocui.ModNone, binding.handler); err != nil {
			return err
		}
	}

	return nil
}

func (program *Program) keybindingSpecs() []keybindingSpec {
	actions := program.resolvedKeybindingActions()
	specs := make([]keybindingSpec, 0)
	for _, action := range actions {
		for _, viewName := range action.action.viewNames {
			for _, binding := range action.bindings {
				specs = append(specs, keybindingSpec{
					viewName: viewName,
					key:      binding.value,
					handler:  action.action.handler,
				})
			}
		}
	}

	return specs
}

func (program *Program) resolvedKeybindingActions() []resolvedKeybindingAction {
	defaults := program.keybindingActions()
	resolved := make([]resolvedKeybindingAction, 0, len(defaults))
	for _, action := range defaults {
		bindings := append([]configuredKeybinding(nil), action.defaultBindings...)
		overridden := false
		if overrideBindings, ok := program.overrideBindings(action.id); ok {
			bindings = overrideBindings
			overridden = true
		}
		resolved = append(resolved, resolvedKeybindingAction{action: action, bindings: bindings, overridden: overridden})
	}

	conflictingOverrides := conflictingOverrideIndexes(resolved)
	for index := range conflictingOverrides {
		resolved[index].bindings = append([]configuredKeybinding(nil), resolved[index].action.defaultBindings...)
		resolved[index].overridden = false
	}

	return resolved
}

func conflictingOverrideIndexes(actions []resolvedKeybindingAction) map[int]bool {
	conflicting := map[int]bool{}
	seenTargets := map[keybindingTarget]int{}

	for actionIndex, action := range actions {
		for _, viewName := range action.action.viewNames {
			for _, binding := range action.bindings {
				target := keybindingTarget{viewName: viewName, key: binding.value}
				previousActionIndex, alreadySeen := seenTargets[target]
				if !alreadySeen {
					seenTargets[target] = actionIndex
					continue
				}

				if actions[actionIndex].overridden {
					conflicting[actionIndex] = true
				}
				if actions[previousActionIndex].overridden {
					conflicting[previousActionIndex] = true
				}
			}
		}
	}

	if len(conflicting) == 0 {
		return nil
	}

	return conflicting
}

type keybindingTarget struct {
	viewName string
	key      any
}

func (program *Program) overrideBindings(id keybindingActionID) ([]configuredKeybinding, bool) {
	if len(program.keymapOverrides) == 0 {
		return nil, false
	}

	actions, ok := program.keymapOverrides[id.scope]
	if !ok {
		return nil, false
	}

	rawBindings, ok := actions[id.action]
	if !ok {
		return nil, false
	}

	parsedBindings, ok := parseConfiguredBindings(rawBindings)
	if !ok {
		return nil, false
	}

	return parsedBindings, true
}

func parseConfiguredBindings(values []string) ([]configuredKeybinding, bool) {
	if len(values) == 0 {
		return nil, false
	}

	bindings := make([]configuredKeybinding, 0, len(values))
	for _, value := range values {
		binding, ok := parseConfiguredKey(value)
		if !ok {
			return nil, false
		}
		bindings = append(bindings, binding)
	}

	return bindings, true
}

func parseConfiguredKey(value string) (configuredKeybinding, bool) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return configuredKeybinding{}, false
	}

	if utf8.RuneCountInString(trimmedValue) == 1 {
		runeValue, _ := utf8.DecodeRuneInString(trimmedValue)
		return runeBinding(runeValue), true
	}

	switch strings.ToLower(trimmedValue) {
	case "enter", "<enter>":
		return namedBinding(gocui.KeyEnter, "<enter>"), true
	case "esc", "escape", "<esc>", "<escape>":
		return namedBinding(gocui.KeyEsc, "<esc>"), true
	case "tab", "<tab>":
		return namedBinding(gocui.KeyTab, "tab"), true
	case "shift+tab", "shift-tab", "backtab", "<shift+tab>", "<backtab>":
		return namedBinding(gocui.KeyBacktab, "shift+tab"), true
	case "up", "arrowup", "arrow-up", "<up>":
		return namedBinding(gocui.KeyArrowUp, "<up>"), true
	case "down", "arrowdown", "arrow-down", "<down>":
		return namedBinding(gocui.KeyArrowDown, "<down>"), true
	case "left", "arrowleft", "arrow-left", "<left>":
		return namedBinding(gocui.KeyArrowLeft, "<left>"), true
	case "right", "arrowright", "arrow-right", "<right>":
		return namedBinding(gocui.KeyArrowRight, "<right>"), true
	case "space", "<space>":
		return runeBinding(' '), true
	case "alt+enter", "alt-enter", "<alt+enter>":
		return namedBinding(gocui.KeyAltEnter, "alt+enter"), true
	}

	binding, ok := parseConfiguredControlKey(trimmedValue)
	if !ok {
		return configuredKeybinding{}, false
	}

	return binding, true
}

func parseConfiguredControlKey(value string) (configuredKeybinding, bool) {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	normalizedValue = strings.TrimPrefix(normalizedValue, "<")
	normalizedValue = strings.TrimSuffix(normalizedValue, ">")
	normalizedValue = strings.ReplaceAll(normalizedValue, "control", "ctrl")
	normalizedValue = strings.ReplaceAll(normalizedValue, "ctrl-", "ctrl+")
	normalizedValue = strings.ReplaceAll(normalizedValue, "c-", "ctrl+")
	if !strings.HasPrefix(normalizedValue, "ctrl+") {
		return configuredKeybinding{}, false
	}

	controlKeyName := strings.TrimPrefix(normalizedValue, "ctrl+")
	binding, ok := configuredControlBindings()[controlKeyName]
	if !ok {
		return configuredKeybinding{}, false
	}

	return binding, true
}

func configuredControlBindings() map[string]configuredKeybinding {
	return map[string]configuredKeybinding{
		"a":          namedBinding(gocui.KeyCtrlA, "<c-a>"),
		"b":          namedBinding(gocui.KeyCtrlB, "<c-b>"),
		"c":          namedBinding(gocui.KeyCtrlC, "<c-c>"),
		"d":          namedBinding(gocui.KeyCtrlD, "<c-d>"),
		"e":          namedBinding(gocui.KeyCtrlE, "<c-e>"),
		"f":          namedBinding(gocui.KeyCtrlF, "<c-f>"),
		"g":          namedBinding(gocui.KeyCtrlG, "<c-g>"),
		"h":          namedBinding(gocui.KeyCtrlH, "<c-h>"),
		"i":          namedBinding(gocui.KeyCtrlI, "<c-i>"),
		"j":          namedBinding(gocui.KeyCtrlJ, "<c-j>"),
		"k":          namedBinding(gocui.KeyCtrlK, "<c-k>"),
		"l":          namedBinding(gocui.KeyCtrlL, "<c-l>"),
		"m":          namedBinding(gocui.KeyCtrlM, "<c-m>"),
		"n":          namedBinding(gocui.KeyCtrlN, "<c-n>"),
		"o":          namedBinding(gocui.KeyCtrlO, "<c-o>"),
		"p":          namedBinding(gocui.KeyCtrlP, "<c-p>"),
		"q":          namedBinding(gocui.KeyCtrlQ, "<c-q>"),
		"r":          namedBinding(gocui.KeyCtrlR, "<c-r>"),
		"s":          namedBinding(gocui.KeyCtrlS, "<c-s>"),
		"t":          namedBinding(gocui.KeyCtrlT, "<c-t>"),
		"u":          namedBinding(gocui.KeyCtrlU, "<c-u>"),
		"v":          namedBinding(gocui.KeyCtrlV, "<c-v>"),
		"w":          namedBinding(gocui.KeyCtrlW, "<c-w>"),
		"x":          namedBinding(gocui.KeyCtrlX, "<c-x>"),
		"y":          namedBinding(gocui.KeyCtrlY, "<c-y>"),
		"z":          namedBinding(gocui.KeyCtrlZ, "<c-z>"),
		"[":          namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>"),
		"]":          namedBinding(gocui.KeyCtrlRsqBracket, "<c-]>"),
		"\\":         namedBinding(gocui.KeyCtrlBackslash, "<c-\\>"),
		"/":          namedBinding(gocui.KeyCtrlSlash, "<c-/>"),
		"_":          namedBinding(gocui.KeyCtrlUnderscore, "<c-_>"),
		"space":      namedBinding(gocui.KeyCtrlSpace, "<c-space>"),
		"lsqbracket": namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>"),
		"rsqbracket": namedBinding(gocui.KeyCtrlRsqBracket, "<c-]>"),
		"backslash":  namedBinding(gocui.KeyCtrlBackslash, "<c-\\>"),
		"slash":      namedBinding(gocui.KeyCtrlSlash, "<c-/>"),
		"underscore": namedBinding(gocui.KeyCtrlUnderscore, "<c-_>"),
	}
}

func runeBinding(value rune) configuredKeybinding {
	label := string(value)
	if value == ' ' {
		label = "space"
	}

	return configuredKeybinding{value: value, label: label}
}

func namedBinding(value any, label string) configuredKeybinding {
	return configuredKeybinding{value: value, label: label}
}

func keybindingActionFor(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error, bindings ...configuredKeybinding) keybindingAction {
	return keybindingAction{
		id:              keybindingActionID{scope: scope, action: action},
		viewNames:       append([]string(nil), viewNames...),
		defaultBindings: append([]configuredKeybinding(nil), bindings...),
		handler:         handler,
	}
}

func (program *Program) keybindingActions() []keybindingAction {
	return []keybindingAction{
		keybindingActionFor(keymapScopeGlobal, "quit", []string{""}, program.quit, namedBinding(gocui.KeyCtrlC, "<c-c>")),
		keybindingActionFor(keymapScopeGlobal, "next_side_view", []string{""}, program.nextSideView, namedBinding(gocui.KeyTab, "tab")),
		keybindingActionFor(keymapScopeGlobal, "previous_side_view", []string{""}, program.previousSideView, namedBinding(gocui.KeyBacktab, "shift+tab")),

		keybindingActionFor(keymapScopeMain, "toggle_help", mainPaneViewNames, program.toggleHelp, runeBinding('?')),
		keybindingActionFor(keymapScopeMain, "focus_user_view", mainPaneViewNames, program.focusUserView, runeBinding('1')),
		keybindingActionFor(keymapScopeMain, "focus_pull_requests_view", mainPaneViewNames, program.focusPullRequestsView, runeBinding('2')),
		keybindingActionFor(keymapScopeMain, "open_search", mainPaneViewNames, program.openSearch, runeBinding('/')),
		keybindingActionFor(keymapScopeMain, "move_selection_down", mainPaneViewNames, program.moveSelectionDown, runeBinding('j'), namedBinding(gocui.KeyArrowDown, "<down>")),
		keybindingActionFor(keymapScopeMain, "move_selection_up", mainPaneViewNames, program.moveSelectionUp, runeBinding('k'), namedBinding(gocui.KeyArrowUp, "<up>")),
		keybindingActionFor(keymapScopeMain, "page_down", mainPaneViewNames, program.pageDown, namedBinding(gocui.KeyCtrlD, "<c-d>")),
		keybindingActionFor(keymapScopeMain, "page_up", mainPaneViewNames, program.pageUp, namedBinding(gocui.KeyCtrlU, "<c-u>")),
		keybindingActionFor(keymapScopeMain, "grow_focused_pane", mainPaneViewNames, program.growFocusedPane, runeBinding('+')),
		keybindingActionFor(keymapScopeMain, "shrink_focused_pane", mainPaneViewNames, program.shrinkFocusedPane, runeBinding('-')),

		keybindingActionFor(keymapScopeSide, "next_side_view", sidePaneViewNames, program.nextSideView, runeBinding('l')),
		keybindingActionFor(keymapScopeSide, "previous_side_view", sidePaneViewNames, program.previousSideView, runeBinding('h')),
		keybindingActionFor(keymapScopeSide, "focus_detail_view", sidePaneViewNames, program.focusDetailView, runeBinding('0')),
		keybindingActionFor(keymapScopeSide, "exit_review_mode", sidePaneViewNames, program.exitReviewMode, namedBinding(gocui.KeyEsc, "<esc>"), namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>"), runeBinding('q')),

		keybindingActionFor(keymapScopeUser, "open_detail", []string{viewUserName}, program.openDetail, namedBinding(gocui.KeyEnter, "<enter>")),
		keybindingActionFor(keymapScopeUser, "copy_pull_request_url", []string{viewUserName}, program.copyPullRequestURL, runeBinding('y')),

		keybindingActionFor(keymapScopePullRequests, "previous_tab", []string{viewPullRequestsName}, program.previousPullRequestTab, runeBinding('[')),
		keybindingActionFor(keymapScopePullRequests, "next_tab", []string{viewPullRequestsName}, program.nextPullRequestTab, runeBinding(']')),
		keybindingActionFor(keymapScopePullRequests, "open_detail", []string{viewPullRequestsName}, program.openDetail, namedBinding(gocui.KeyEnter, "<enter>")),
		keybindingActionFor(keymapScopePullRequests, "copy_pull_request_url", []string{viewPullRequestsName}, program.copyPullRequestURL, runeBinding('y')),
		keybindingActionFor(keymapScopePullRequests, "comment_on_pull_request", []string{viewPullRequestsName}, program.openPullRequestCommentComposer, runeBinding('c')),
		keybindingActionFor(keymapScopePullRequests, "open_actions_popup", []string{viewPullRequestsName}, program.openActionsPopup, runeBinding('a')),
		keybindingActionFor(keymapScopePullRequests, "next_search_match", []string{viewPullRequestsName}, program.nextReviewFileTreeSearchMatch, runeBinding('n')),
		keybindingActionFor(keymapScopePullRequests, "previous_search_match", []string{viewPullRequestsName}, program.previousReviewFileTreeSearchMatch, runeBinding('N')),

		keybindingActionFor(keymapScopeDetail, "move_cursor_left", []string{viewDetailName}, program.moveDetailCursorLeft, runeBinding('h')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_right", []string{viewDetailName}, program.moveDetailCursorRight, runeBinding('l')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_row_start", []string{viewDetailName}, program.moveDetailCursorToRowStart, runeBinding('0')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_row_end", []string{viewDetailName}, program.moveDetailCursorToRowEnd, runeBinding('$')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_top", []string{viewDetailName}, program.moveDetailCursorToTop, runeBinding('g')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_bottom", []string{viewDetailName}, program.moveDetailCursorToBottom, runeBinding('G')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_next_word", []string{viewDetailName}, program.moveDetailCursorToNextWord, runeBinding('w')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_word_end", []string{viewDetailName}, program.moveDetailCursorToWordEnd, runeBinding('e')),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_previous_word", []string{viewDetailName}, program.moveDetailCursorToPreviousWord, runeBinding('b')),
		keybindingActionFor(keymapScopeDetail, "next_search_match", []string{viewDetailName}, program.nextDetailSearchMatch, runeBinding('n')),
		keybindingActionFor(keymapScopeDetail, "previous_search_match", []string{viewDetailName}, program.previousDetailSearchMatch, runeBinding('N')),
		keybindingActionFor(keymapScopeDetail, "enter_visual_mode", []string{viewDetailName}, program.enterDetailVisualMode, runeBinding('v')),
		keybindingActionFor(keymapScopeDetail, "enter_line_visual_mode", []string{viewDetailName}, program.enterDetailLineVisualMode, runeBinding('V')),
		keybindingActionFor(keymapScopeDetail, "previous_tab", []string{viewDetailName}, program.previousDetailTab, runeBinding('[')),
		keybindingActionFor(keymapScopeDetail, "next_tab", []string{viewDetailName}, program.nextDetailTab, runeBinding(']')),
		keybindingActionFor(keymapScopeDetail, "copy_pull_request_url", []string{viewDetailName}, program.copyPullRequestURL, runeBinding('y')),
		keybindingActionFor(keymapScopeDetail, "comment_on_pull_request", []string{viewDetailName}, program.openPullRequestCommentComposer, runeBinding('c')),
		keybindingActionFor(keymapScopeDetail, "open_actions_popup", []string{viewDetailName}, program.openActionsPopup, runeBinding('a')),
		keybindingActionFor(keymapScopeDetail, "close", []string{viewDetailName}, program.closeDetail, namedBinding(gocui.KeyEsc, "<esc>"), namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>"), runeBinding('q')),

		keybindingActionFor(keymapScopeSearch, "submit", []string{viewSearchName}, program.submitSearch, namedBinding(gocui.KeyEnter, "<enter>"), namedBinding(gocui.KeyCtrlJ, "<c-j>")),
		keybindingActionFor(keymapScopeSearch, "cancel", []string{viewSearchName}, program.cancelSearch, namedBinding(gocui.KeyEsc, "<esc>"), namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>")),

		keybindingActionFor(keymapScopeActionsPopup, "focus_search", []string{viewActionsPopupName}, program.focusActionsPopupSearch, runeBinding('/')),
		keybindingActionFor(keymapScopeActionsPopup, "move_selection_down", []string{viewActionsPopupName}, program.moveActionsPopupSelectionDown, runeBinding('j'), namedBinding(gocui.KeyArrowDown, "<down>")),
		keybindingActionFor(keymapScopeActionsPopup, "move_selection_up", []string{viewActionsPopupName}, program.moveActionsPopupSelectionUp, runeBinding('k'), namedBinding(gocui.KeyArrowUp, "<up>")),
		keybindingActionFor(keymapScopeActionsPopup, "execute_selected_action", []string{viewActionsPopupName}, program.executeSelectedActionsPopupAction, namedBinding(gocui.KeyEnter, "<enter>")),
		keybindingActionFor(keymapScopeActionsPopup, "close", []string{viewActionsPopupName}, program.closeActionsPopup, namedBinding(gocui.KeyEsc, "<esc>"), namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>"), runeBinding('q')),

		keybindingActionFor(keymapScopeActionsPopupSearch, "focus_list", []string{viewActionsPopupSearchName}, program.focusActionsPopupList, namedBinding(gocui.KeyEnter, "<enter>"), namedBinding(gocui.KeyTab, "tab")),
		keybindingActionFor(keymapScopeActionsPopupSearch, "close", []string{viewActionsPopupSearchName}, program.closeActionsPopup, namedBinding(gocui.KeyEsc, "<esc>"), namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>")),

		keybindingActionFor(keymapScopeModalEditor, "submit", []string{viewModalEditorName}, program.submitModalEditor, namedBinding(gocui.KeyAltEnter, "alt+enter")),
		keybindingActionFor(keymapScopeModalEditor, "close", []string{viewModalEditorName}, program.closeModalEditor, namedBinding(gocui.KeyEsc, "<esc>"), namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>")),

		keybindingActionFor(keymapScopeHelp, "close", []string{viewHelpName}, program.closeHelp, namedBinding(gocui.KeyEsc, "<esc>"), namedBinding(gocui.KeyCtrlLsqBracket, "<c-[>"), runeBinding('q')),
	}
}
