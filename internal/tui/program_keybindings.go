package tui

import "github.com/jesseduffield/gocui"

type keybindingSpec struct {
	viewName string
	key      any
	handler  func(*gocui.Gui, *gocui.View) error
}

type keybindingDefinition struct {
	key     any
	handler func(*gocui.Gui, *gocui.View) error
}

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

func (program *Program) setKeybindings(gui *gocui.Gui) error {
	for _, binding := range program.keybindingSpecs() {
		if err := gui.SetKeybinding(binding.viewName, binding.key, gocui.ModNone, binding.handler); err != nil {
			return err
		}
	}

	return nil
}

func (program *Program) keybindingSpecs() []keybindingSpec {
	specs := []keybindingSpec{
		{viewName: "", key: gocui.KeyCtrlC, handler: program.quit},
		{viewName: "", key: gocui.KeyTab, handler: program.nextSideView},
		{viewName: "", key: gocui.KeyBacktab, handler: program.previousSideView},
	}

	specs = append(specs, bindingsForViews(mainPaneViewNames,
		keybindingDefinition{key: '?', handler: program.toggleHelp},
		keybindingDefinition{key: '1', handler: program.focusUserView},
		keybindingDefinition{key: '2', handler: program.focusPullRequestsView},
		keybindingDefinition{key: '/', handler: program.openSearch},
		keybindingDefinition{key: 'j', handler: program.moveSelectionDown},
		keybindingDefinition{key: gocui.KeyArrowDown, handler: program.moveSelectionDown},
		keybindingDefinition{key: 'k', handler: program.moveSelectionUp},
		keybindingDefinition{key: gocui.KeyArrowUp, handler: program.moveSelectionUp},
		keybindingDefinition{key: gocui.KeyCtrlD, handler: program.pageDown},
		keybindingDefinition{key: gocui.KeyCtrlU, handler: program.pageUp},
		keybindingDefinition{key: '+', handler: program.growFocusedPane},
		keybindingDefinition{key: '-', handler: program.shrinkFocusedPane},
	)...)

	specs = append(specs, bindingsForViews(sidePaneViewNames,
		append([]keybindingDefinition{
			{key: 'l', handler: program.nextSideView},
			{key: 'h', handler: program.previousSideView},
			{key: '0', handler: program.focusDetailView},
		}, dismissKeybindingDefinitions(program.exitReviewMode)...)...,
	)...)

	specs = append(specs, bindingsForView(viewUserName,
		keybindingDefinition{key: gocui.KeyEnter, handler: program.openDetail},
		keybindingDefinition{key: 'y', handler: program.copyPullRequestURL},
	)...)

	specs = append(specs, bindingsForView(viewPullRequestsName,
		keybindingDefinition{key: '[', handler: program.previousPullRequestTab},
		keybindingDefinition{key: ']', handler: program.nextPullRequestTab},
		keybindingDefinition{key: gocui.KeyEnter, handler: program.openDetail},
		keybindingDefinition{key: 'y', handler: program.copyPullRequestURL},
		keybindingDefinition{key: 'c', handler: program.openPullRequestCommentComposer},
		keybindingDefinition{key: 'a', handler: program.openActionsPopup},
	)...)

	specs = append(specs, bindingsForView(viewDetailName,
		append([]keybindingDefinition{
			{key: 'h', handler: program.moveDetailCursorLeft},
			{key: 'l', handler: program.moveDetailCursorRight},
			{key: '0', handler: program.moveDetailCursorToRowStart},
			{key: '$', handler: program.moveDetailCursorToRowEnd},
			{key: 'g', handler: program.moveDetailCursorToTop},
			{key: 'G', handler: program.moveDetailCursorToBottom},
			{key: 'w', handler: program.moveDetailCursorToNextWord},
			{key: 'e', handler: program.moveDetailCursorToWordEnd},
			{key: 'b', handler: program.moveDetailCursorToPreviousWord},
			{key: 'n', handler: program.nextDetailSearchMatch},
			{key: 'N', handler: program.previousDetailSearchMatch},
			{key: 'v', handler: program.enterDetailVisualMode},
			{key: 'V', handler: program.enterDetailLineVisualMode},
			{key: '[', handler: program.previousDetailTab},
			{key: ']', handler: program.nextDetailTab},
			{key: 'y', handler: program.copyPullRequestURL},
			{key: 'c', handler: program.openPullRequestCommentComposer},
			{key: 'a', handler: program.openActionsPopup},
		}, dismissKeybindingDefinitions(program.closeDetail)...)...,
	)...)

	specs = append(specs, bindingsForView(viewSearchName,
		append([]keybindingDefinition{
			{key: gocui.KeyEnter, handler: program.submitSearch},
			{key: gocui.KeyCtrlJ, handler: program.submitSearch},
		}, escapeKeybindingDefinitions(program.cancelSearch)...)...,
	)...)

	specs = append(specs, bindingsForView(viewActionsPopupName,
		append([]keybindingDefinition{
			{key: '/', handler: program.focusActionsPopupSearch},
			{key: 'j', handler: program.moveActionsPopupSelectionDown},
			{key: gocui.KeyArrowDown, handler: program.moveActionsPopupSelectionDown},
			{key: 'k', handler: program.moveActionsPopupSelectionUp},
			{key: gocui.KeyArrowUp, handler: program.moveActionsPopupSelectionUp},
			{key: gocui.KeyEnter, handler: program.executeSelectedActionsPopupAction},
		}, dismissKeybindingDefinitions(program.closeActionsPopup)...)...,
	)...)

	specs = append(specs, bindingsForView(viewActionsPopupSearchName,
		append([]keybindingDefinition{
			{key: gocui.KeyEnter, handler: program.focusActionsPopupList},
			{key: gocui.KeyTab, handler: program.focusActionsPopupList},
		}, escapeKeybindingDefinitions(program.closeActionsPopup)...)...,
	)...)

	specs = append(specs, bindingsForView(viewModalEditorName,
		append([]keybindingDefinition{
			{key: gocui.KeyAltEnter, handler: program.submitModalEditor},
		}, escapeKeybindingDefinitions(program.closeModalEditor)...)...,
	)...)

	specs = append(specs, bindingsForView(viewHelpName,
		dismissKeybindingDefinitions(program.closeHelp)...,
	)...)

	return specs
}
