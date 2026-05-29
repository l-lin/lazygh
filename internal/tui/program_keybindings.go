package tui

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
	appconfig "github.com/l-lin/lazygh/internal/config"
)

type keybindingSpec struct {
	viewName string
	key      any
	mod      gocui.Modifier
	handler  func(*gocui.Gui, *gocui.View) error
}

type keybindingDefinition struct {
	key     any
	mod     gocui.Modifier
	handler func(*gocui.Gui, *gocui.View) error
}

type configuredKeybinding struct {
	value any
	mod   gocui.Modifier
	label string
}

type configuredKeySequence struct {
	keys  []configuredKeybinding
	label string
}

type keybindingActionID struct {
	scope  string
	action string
}

type keybindingBindingSlice int

const (
	keybindingBindingSliceAll keybindingBindingSlice = iota
	keybindingBindingSliceFirst
	keybindingBindingSliceRest
)

type keybindingAction struct {
	id              keybindingActionID
	configID        keybindingActionID
	configurable    bool
	viewNames       []string
	defaultBindings []configuredKeySequence
	handler         func(*gocui.Gui, *gocui.View) error
	allowSequences  bool
	bindingSlice    keybindingBindingSlice
	registerOnSetup bool
}

type resolvedKeybindingAction struct {
	action     keybindingAction
	bindings   []configuredKeySequence
	overridden bool
}

type keybindingDispatchEntry struct {
	directHandler        func(*gocui.Gui, *gocui.View) error
	prefixTarget         keySequenceTarget
	continuationHandlers map[keySequenceTarget]func(*gocui.Gui, *gocui.View) error
}

type keybindingSequenceTarget struct {
	viewName  string
	first     any
	firstMod  gocui.Modifier
	second    any
	secondMod gocui.Modifier
}

func bindingsForViews(viewNames []string, definitions ...keybindingDefinition) []keybindingSpec {
	specs := make([]keybindingSpec, 0, len(viewNames)*len(definitions))
	for _, viewName := range viewNames {
		for _, definition := range definitions {
			specs = append(specs, keybindingSpec{
				viewName: viewName,
				key:      definition.key,
				mod:      definition.mod,
				handler:  definition.handler,
			})
		}
	}

	return specs
}

func (program *Program) ApplyKeymapOverrides(overrides appconfig.KeymapOverrides) {
	_ = program.dispatchRuntimeMessage(MsgKeymapOverridesApplied{Overrides: overrides})
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
	if gui == nil {
		return nil
	}

	specs := program.registeredKeybindingSpecs()
	fingerprint := fingerprintKeybindingSpecs(specs)
	if fingerprint == program.keybindingRuntime.registeredFingerprint {
		return nil
	}

	gui.DeleteAllKeybindings()
	for _, binding := range specs {
		if err := gui.SetKeybinding(binding.viewName, binding.key, binding.mod, binding.handler); err != nil {
			return err
		}
	}
	program.keybindingRuntime.registeredFingerprint = fingerprint
	return nil
}

func (program *Program) reloadRegisteredKeybindings(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	specs := program.registeredKeybindingSpecs()
	fingerprint := fingerprintKeybindingSpecs(specs)
	if fingerprint == program.keybindingRuntime.registeredFingerprint {
		return nil
	}

	gui.DeleteAllKeybindings()
	for _, binding := range specs {
		if err := gui.SetKeybinding(binding.viewName, binding.key, binding.mod, binding.handler); err != nil {
			return err
		}
	}
	program.keybindingRuntime.registeredFingerprint = fingerprint
	return nil
}

func fingerprintKeybindingSpecs(specs []keybindingSpec) string {
	if len(specs) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, spec := range specs {
		builder.WriteString(spec.viewName)
		builder.WriteRune(':')
		builder.WriteString(keybindingValueID(spec.key, spec.mod))
		builder.WriteRune('|')
	}
	return builder.String()
}

func (program *Program) keybindingSpecs() []keybindingSpec {
	actions := program.resolvedKeybindingActions()
	dispatches := map[keybindingTarget]*keybindingDispatchEntry{}
	orderedTargets := make([]keybindingTarget, 0)

	ensureDispatch := func(viewName string, key any, mod gocui.Modifier) *keybindingDispatchEntry {
		target := keybindingTarget{viewName: viewName, key: key, mod: mod}
		if entry, ok := dispatches[target]; ok {
			return entry
		}
		entry := &keybindingDispatchEntry{continuationHandlers: map[keySequenceTarget]func(*gocui.Gui, *gocui.View) error{}}
		dispatches[target] = entry
		orderedTargets = append(orderedTargets, target)
		return entry
	}

	for _, action := range actions {
		if !action.action.registerOnSetup {
			continue
		}
		for _, binding := range action.bindings {
			for _, viewName := range bindingViewNames(action.action, binding) {
				if len(binding.keys) == 0 {
					continue
				}
				if len(binding.keys) == 1 {
					ensureDispatch(viewName, binding.keys[0].value, binding.keys[0].mod).directHandler = action.action.handler
					continue
				}

				prefixTarget := keySequencePrefixTarget(viewName, binding.keys[0])
				ensureDispatch(viewName, binding.keys[0].value, binding.keys[0].mod).prefixTarget = prefixTarget
				ensureDispatch(viewName, binding.keys[1].value, binding.keys[1].mod).continuationHandlers[prefixTarget] = action.action.handler
			}
		}
	}

	specs := make([]keybindingSpec, 0, len(orderedTargets))
	for _, target := range orderedTargets {
		entry := dispatches[target]
		handler := entry.directHandler
		if entry.hasDispatchLogic() {
			handler = program.dispatchingKeybindingHandler(target.viewName, target.key, *entry)
		}
		specs = append(specs, keybindingSpec{viewName: target.viewName, key: target.key, mod: target.mod, handler: handler})
	}

	return specs
}

func (entry keybindingDispatchEntry) hasDispatchLogic() bool {
	return entry.prefixTarget != (keySequenceTarget{}) || len(entry.continuationHandlers) > 0
}

func (program *Program) dispatchingKeybindingHandler(viewName string, _ any, entry keybindingDispatchEntry) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		pendingTarget := program.pendingKeySequenceTargetForView(viewName)
		if pendingTarget != (keySequenceTarget{}) && len(entry.continuationHandlers) > 0 {
			program.clearPendingKeySequenceForView(viewName)
			if pendingHandler, ok := entry.continuationHandlers[pendingTarget]; ok {
				return pendingHandler(gui, view)
			}
		}
		if entry.prefixTarget != (keySequenceTarget{}) {
			program.armPendingKeySequenceForView(viewName, entry.prefixTarget)
			if entry.directHandler == nil {
				return nil
			}
		}
		if entry.directHandler != nil {
			return entry.directHandler(gui, view)
		}
		return nil
	}
}

func (program *Program) resolvedKeybindingActions() []resolvedKeybindingAction {
	defaults := program.keybindingActions()
	resolved := make([]resolvedKeybindingAction, 0, len(defaults))
	for _, action := range defaults {
		bindings := action.selectBindings(copyConfiguredKeySequences(action.defaultBindings))
		overridden := false
		if overrideBindings, ok := program.overrideBindings(action); ok {
			bindings = overrideBindings
			overridden = true
		}
		resolved = append(resolved, resolvedKeybindingAction{action: action, bindings: bindings, overridden: overridden})
	}

	conflictingOverrides := conflictingOverrideIndexes(resolved)
	for index := range conflictingOverrides {
		resolved[index].bindings = resolved[index].action.selectBindings(copyConfiguredKeySequences(resolved[index].action.defaultBindings))
		resolved[index].overridden = false
	}

	return resolved
}

func (program *Program) resolvedBindingsForActionID(actionID keybindingActionID) []configuredKeySequence {
	for _, action := range program.resolvedKeybindingActions() {
		if action.action.id != actionID {
			continue
		}
		return copyConfiguredKeySequences(action.bindings)
	}
	return nil
}

func copyConfiguredKeySequences(bindings []configuredKeySequence) []configuredKeySequence {
	if len(bindings) == 0 {
		return nil
	}

	copied := make([]configuredKeySequence, 0, len(bindings))
	for _, binding := range bindings {
		copied = append(copied, configuredKeySequence{
			keys:  append([]configuredKeybinding(nil), binding.keys...),
			label: binding.label,
		})
	}
	return copied
}

var multiStepGlobalBindingViewNames = []string{
	viewUserName,
	viewPullRequestsName,
	viewNotificationsName,
	viewDetailName,
	viewActionsPopupName,
	viewPullRequestBuildInfoName,
	viewHelpName,
}

func bindingViewNames(action keybindingAction, binding configuredKeySequence) []string {
	viewNames := append([]string(nil), action.viewNames...)
	if len(binding.keys) <= 1 || !slices.Contains(viewNames, "") {
		return viewNames
	}

	viewNames = append(viewNames, multiStepGlobalBindingViewNames...)
	uniqueViewNames := make([]string, 0, len(viewNames))
	seen := map[string]struct{}{}
	for _, viewName := range viewNames {
		if _, ok := seen[viewName]; ok {
			continue
		}
		seen[viewName] = struct{}{}
		uniqueViewNames = append(uniqueViewNames, viewName)
	}
	return uniqueViewNames
}

func conflictingOverrideIndexes(actions []resolvedKeybindingAction) map[int]bool {
	conflicting := map[int]bool{}
	seenTargets := map[keybindingTarget]int{}
	seenSequences := map[keybindingSequenceTarget]int{}
	seenPrefixes := map[keybindingTarget][]int{}

	markConflict := func(indexes ...int) {
		for _, index := range indexes {
			if index < 0 || index >= len(actions) {
				continue
			}
			if actions[index].overridden {
				conflicting[index] = true
			}
		}
	}

	for actionIndex, action := range actions {
		if !action.action.registerOnSetup {
			continue
		}
		for _, binding := range action.bindings {
			for _, viewName := range bindingViewNames(action.action, binding) {
				switch len(binding.keys) {
				case 0:
					continue
				case 1:
					target := keybindingTarget{viewName: viewName, key: binding.keys[0].value, mod: binding.keys[0].mod}
					if previousActionIndex, alreadySeen := seenTargets[target]; alreadySeen {
						markConflict(actionIndex, previousActionIndex)
					}
					for _, previousActionIndex := range seenPrefixes[target] {
						markConflict(actionIndex, previousActionIndex)
					}
					seenTargets[target] = actionIndex
				case 2:
					prefix := keybindingTarget{viewName: viewName, key: binding.keys[0].value, mod: binding.keys[0].mod}
					if previousActionIndex, alreadySeen := seenTargets[prefix]; alreadySeen {
						markConflict(actionIndex, previousActionIndex)
					}
					sequence := keybindingSequenceTarget{viewName: viewName, first: binding.keys[0].value, firstMod: binding.keys[0].mod, second: binding.keys[1].value, secondMod: binding.keys[1].mod}
					if previousActionIndex, alreadySeen := seenSequences[sequence]; alreadySeen {
						markConflict(actionIndex, previousActionIndex)
					}
					seenPrefixes[prefix] = append(seenPrefixes[prefix], actionIndex)
					seenSequences[sequence] = actionIndex
				default:
					markConflict(actionIndex)
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
	mod      gocui.Modifier
}

func (program *Program) overrideBindings(action keybindingAction) ([]configuredKeySequence, bool) {
	if len(program.runtimeConfig.keymapOverrides) == 0 || !action.configurable {
		return nil, false
	}

	bindings, ok := program.parseOverrideBindings(action.configID.scope, action.configID.action, action.allowSequences)
	if !ok {
		return nil, false
	}
	return action.selectBindings(bindings), true
}

func (program *Program) parseOverrideBindings(scope string, action string, allowSequences bool) ([]configuredKeySequence, bool) {
	actions, ok := program.runtimeConfig.keymapOverrides[scope]
	if !ok {
		return nil, false
	}

	rawBindings, ok := actions[action]
	if !ok {
		return nil, false
	}

	parsedBindings, ok := parseConfiguredBindings(rawBindings)
	if !ok {
		return nil, false
	}
	if !allowSequences && containsMultiStepBinding(parsedBindings) {
		return nil, false
	}
	return parsedBindings, true
}

func containsMultiStepBinding(bindings []configuredKeySequence) bool {
	for _, binding := range bindings {
		if len(binding.keys) > 1 {
			return true
		}
	}
	return false
}

func (action keybindingAction) selectBindings(bindings []configuredKeySequence) []configuredKeySequence {
	switch action.bindingSlice {
	case keybindingBindingSliceFirst:
		if len(bindings) == 0 {
			return nil
		}
		return copyConfiguredKeySequences(bindings[:1])
	case keybindingBindingSliceRest:
		if len(bindings) <= 1 {
			return nil
		}
		return copyConfiguredKeySequences(bindings[1:])
	default:
		return copyConfiguredKeySequences(bindings)
	}
}

func parseConfiguredBindings(values []string) ([]configuredKeySequence, bool) {
	if len(values) == 0 {
		return nil, false
	}

	bindings := make([]configuredKeySequence, 0, len(values))
	for _, value := range values {
		binding, ok := parseConfiguredKeySequence(value)
		if !ok {
			return nil, false
		}
		bindings = append(bindings, binding)
	}

	return bindings, true
}

func parseConfiguredKeySequence(value string) (configuredKeySequence, bool) {
	if binding, ok := parseConfiguredKey(value); ok {
		return keySequenceBinding(binding), true
	}

	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return configuredKeySequence{}, false
	}
	if utf8.RuneCountInString(trimmedValue) != 2 {
		return configuredKeySequence{}, false
	}

	runes := []rune(trimmedValue)
	for _, runeValue := range runes {
		if runeValue == ' ' || runeValue == '\t' || runeValue == '\n' || runeValue == '\r' {
			return configuredKeySequence{}, false
		}
	}

	return keySequenceBinding(runeBinding(runes[0]), runeBinding(runes[1])), true
}

func parseConfiguredKey(value string) (configuredKeybinding, bool) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return configuredKeybinding{}, false
	}

	if binding, ok := parseConfiguredAltKey(trimmedValue); ok {
		return binding, true
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
	case "pageup", "page-up", "pgup", "<pageup>", "<pgup>":
		return namedBinding(gocui.KeyPgup, "pageup"), true
	case "pagedown", "page-down", "pgdown", "pgdn", "<pagedown>", "<pgdown>", "<pgdn>":
		return namedBinding(gocui.KeyPgdn, "pagedown"), true
	case "space", "<space>":
		return runeBinding(' '), true
	}

	binding, ok := parseConfiguredControlKey(trimmedValue)
	if !ok {
		return configuredKeybinding{}, false
	}

	return binding, true
}

func parseConfiguredAltKey(value string) (configuredKeybinding, bool) {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	normalizedValue = strings.TrimPrefix(normalizedValue, "<")
	normalizedValue = strings.TrimSuffix(normalizedValue, ">")
	normalizedValue = strings.ReplaceAll(normalizedValue, "alt-", "alt+")
	if !strings.HasPrefix(normalizedValue, "alt+") {
		return configuredKeybinding{}, false
	}

	keyName := strings.TrimPrefix(normalizedValue, "alt+")
	switch keyName {
	case "":
		return configuredKeybinding{}, false
	case "enter":
		return namedBinding(gocui.KeyAltEnter, "alt+enter"), true
	}

	binding, ok := parseConfiguredKey(keyName)
	if !ok || binding.mod != gocui.ModNone {
		return configuredKeybinding{}, false
	}

	return modifiedBinding(binding.value, gocui.ModAlt, "alt+"+configuredKeyLabelForModifier(binding.label)), true
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

	return configuredKeybinding{value: value, mod: gocui.ModNone, label: label}
}

func namedBinding(value any, label string) configuredKeybinding {
	return configuredKeybinding{value: value, mod: gocui.ModNone, label: label}
}

func modifiedBinding(value any, mod gocui.Modifier, label string) configuredKeybinding {
	return configuredKeybinding{value: value, mod: mod, label: label}
}

func configuredKeyLabelForModifier(label string) string {
	trimmedLabel := strings.TrimSpace(label)
	trimmedLabel = strings.TrimPrefix(trimmedLabel, "<")
	trimmedLabel = strings.TrimSuffix(trimmedLabel, ">")
	return trimmedLabel
}

func keySequenceBinding(keys ...configuredKeybinding) configuredKeySequence {
	copiedKeys := append([]configuredKeybinding(nil), keys...)
	labels := make([]string, 0, len(copiedKeys))
	for _, key := range copiedKeys {
		labels = append(labels, key.label)
	}
	return configuredKeySequence{keys: copiedKeys, label: strings.Join(labels, "")}
}

func mustConfiguredKeySequences(values ...string) []configuredKeySequence {
	bindings, ok := parseConfiguredBindings(values)
	if !ok {
		panic(fmt.Sprintf("invalid default key binding sequence %v", values))
	}
	return bindings
}

func keySequencePrefixTarget(viewName string, key configuredKeybinding) keySequenceTarget {
	return keySequenceTarget{
		viewName: viewName,
		actionID: keybindingActionID{scope: keymapScopePrefix, action: keybindingValueID(key.value, key.mod)},
	}
}

func keybindingValueID(value any, mod gocui.Modifier) string {
	return fmt.Sprintf("%T:%v:%d", value, value, mod)
}
