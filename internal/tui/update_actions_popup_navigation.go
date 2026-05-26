package tui

func (program *Program) applyActionsPopupPageRequested(message MsgActionsPopupPageRequested) []Cmd {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}
	return []Cmd{resolveActionsPopupPageSizeCmd{Kind: message.Kind}}
}

func (program *Program) applyActionsPopupPageResolved(message MsgActionsPopupPageResolved) []Cmd {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	switch message.Kind {
	case pageNavigationKindHalfDown:
		program.model.PageActionsPopupDown(message.PageSize)
	case pageNavigationKindHalfUp:
		program.model.PageActionsPopupUp(message.PageSize)
	case pageNavigationKindFullDown:
		program.model.FullPageActionsPopupDown(message.PageSize)
	case pageNavigationKindFullUp:
		program.model.FullPageActionsPopupUp(message.PageSize)
	default:
		return nil
	}
	program.actionsPopupWidget.errorMessage = ""
	return []Cmd{actionsPopupViewportCmd{Placement: viewportPlacementCenter}}
}

func (program *Program) applyActionsPopupViewportRequested(message MsgActionsPopupViewportRequested) []Cmd {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}
	return []Cmd{actionsPopupViewportCmd{Placement: message.Placement}}
}
