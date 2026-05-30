package tui

func (state refreshReadCacheState) withActionsPopupActions(actions []actionsPopupAction) refreshReadCacheState {
	state.actionsPopupActions = copyActionsPopupActions(actions)
	state.actionsPopupActionsKnown = true
	return state
}

func (state refreshReadCacheState) withActionsPopupVisibleLines(visibleLines []actionsPopupVisibleLine) refreshReadCacheState {
	state.actionsPopupVisibleLines = copyActionsPopupVisibleLines(visibleLines)
	state.actionsPopupVisibleKnown = true
	return state
}

func (state refreshReadCacheState) withReviewSessionReadModel(model reviewSessionReadModel) refreshReadCacheState {
	state.reviewSessionReadModel = model
	state.reviewSessionReadModelSet = true
	return state
}

func (state refreshReadCacheState) withKeybindingResolver(resolver keybindingLabelResolver) refreshReadCacheState {
	state.keybindingResolver = resolver
	state.keybindingResolverSet = true
	return state
}

func (state refreshReadCacheState) withFooterPresenter(presenter footerPresenter) refreshReadCacheState {
	state.footerPresenter = presenter
	state.footerPresenterSet = true
	return state
}

func (state refreshReadCacheState) withActionsPopupPresenter(presenter actionsPopupPresenter) refreshReadCacheState {
	state.actionsPopupPresenter = presenter
	state.actionsPopupPresenterSet = true
	return state
}

func copyActionsPopupActions(actions []actionsPopupAction) []actionsPopupAction {
	if len(actions) == 0 {
		return nil
	}

	copied := append([]actionsPopupAction(nil), actions...)
	for index := range copied {
		copied[index].keywords = append([]string(nil), copied[index].keywords...)
	}
	return copied
}

func copyActionsPopupVisibleLines(visibleLines []actionsPopupVisibleLine) []actionsPopupVisibleLine {
	if len(visibleLines) == 0 {
		return nil
	}

	copied := append([]actionsPopupVisibleLine(nil), visibleLines...)
	for index := range copied {
		copied[index].titleSegments = append([]ItemTitleSegment(nil), copied[index].titleSegments...)
	}
	return copied
}
