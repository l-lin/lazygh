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
		keybindingDefinition{key: 'k', handler: program.moveSelectionUp},
		keybindingDefinition{key: gocui.KeyCtrlD, handler: program.pageDown},
		keybindingDefinition{key: gocui.KeyCtrlU, handler: program.pageUp},
		keybindingDefinition{key: '+', handler: program.growFocusedPane},
		keybindingDefinition{key: '-', handler: program.shrinkFocusedPane},
	)...)

	specs = append(specs, bindingsForViews(sidePaneViewNames,
		keybindingDefinition{key: 'l', handler: program.nextSideView},
		keybindingDefinition{key: 'h', handler: program.previousSideView},
		keybindingDefinition{key: '0', handler: program.focusDetailView},
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
		keybindingDefinition{key: 'h', handler: program.moveDetailCursorLeft},
		keybindingDefinition{key: 'l', handler: program.moveDetailCursorRight},
		keybindingDefinition{key: '0', handler: program.moveDetailCursorToRowStart},
		keybindingDefinition{key: '$', handler: program.moveDetailCursorToRowEnd},
		keybindingDefinition{key: 'g', handler: program.moveDetailCursorToTop},
		keybindingDefinition{key: 'G', handler: program.moveDetailCursorToBottom},
		keybindingDefinition{key: 'w', handler: program.moveDetailCursorToNextWord},
		keybindingDefinition{key: 'b', handler: program.moveDetailCursorToPreviousWord},
		keybindingDefinition{key: 'v', handler: program.enterDetailVisualMode},
		keybindingDefinition{key: '[', handler: program.previousDetailTab},
		keybindingDefinition{key: ']', handler: program.nextDetailTab},
		keybindingDefinition{key: 'y', handler: program.copyPullRequestURL},
		keybindingDefinition{key: 'c', handler: program.openPullRequestCommentComposer},
		keybindingDefinition{key: 'a', handler: program.openActionsPopup},
		keybindingDefinition{key: gocui.KeyEsc, handler: program.closeDetail},
		keybindingDefinition{key: gocui.KeyCtrlLsqBracket, handler: program.closeDetail},
	)...)

	specs = append(specs, bindingsForView(viewSearchName,
		keybindingDefinition{key: gocui.KeyEnter, handler: program.submitSearch},
		keybindingDefinition{key: gocui.KeyCtrlJ, handler: program.submitSearch},
		keybindingDefinition{key: gocui.KeyEsc, handler: program.cancelSearch},
		keybindingDefinition{key: gocui.KeyCtrlLsqBracket, handler: program.cancelSearch},
	)...)

	specs = append(specs, bindingsForView(viewActionsPopupName,
		keybindingDefinition{key: '/', handler: program.focusActionsPopupSearch},
		keybindingDefinition{key: 'j', handler: program.moveActionsPopupSelectionDown},
		keybindingDefinition{key: 'k', handler: program.moveActionsPopupSelectionUp},
		keybindingDefinition{key: gocui.KeyEnter, handler: program.executeSelectedActionsPopupAction},
		keybindingDefinition{key: gocui.KeyEsc, handler: program.closeActionsPopup},
		keybindingDefinition{key: gocui.KeyCtrlLsqBracket, handler: program.closeActionsPopup},
	)...)

	specs = append(specs, bindingsForView(viewActionsPopupSearchName,
		keybindingDefinition{key: gocui.KeyEnter, handler: program.focusActionsPopupList},
		keybindingDefinition{key: gocui.KeyEsc, handler: program.closeActionsPopup},
		keybindingDefinition{key: gocui.KeyCtrlLsqBracket, handler: program.closeActionsPopup},
		keybindingDefinition{key: gocui.KeyTab, handler: program.focusActionsPopupList},
	)...)

	specs = append(specs, bindingsForView(viewModalEditorName,
		keybindingDefinition{key: gocui.KeyAltEnter, handler: program.submitModalEditor},
		keybindingDefinition{key: gocui.KeyEsc, handler: program.closeModalEditor},
		keybindingDefinition{key: gocui.KeyCtrlLsqBracket, handler: program.closeModalEditor},
	)...)

	specs = append(specs, bindingsForView(viewHelpName,
		keybindingDefinition{key: gocui.KeyEsc, handler: program.closeHelp},
		keybindingDefinition{key: gocui.KeyCtrlLsqBracket, handler: program.closeHelp},
	)...)

	return specs
}
