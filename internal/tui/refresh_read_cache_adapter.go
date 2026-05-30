package tui

func (program *Program) updateRefreshReadCache(transition func(refreshReadCacheState) refreshReadCacheState) {
	if program == nil || transition == nil {
		return
	}
	program.refreshReadCache = transition(program.refreshReadCache)
}

func (program *Program) cacheActionsPopupActions(actions []actionsPopupAction) {
	if program == nil || !program.refreshReadCache.enabled {
		return
	}
	program.updateRefreshReadCache(func(state refreshReadCacheState) refreshReadCacheState {
		return state.withActionsPopupActions(actions)
	})
}

func (program *Program) cacheActionsPopupVisibleLines(visibleLines []actionsPopupVisibleLine) {
	if program == nil || !program.refreshReadCache.enabled {
		return
	}
	program.updateRefreshReadCache(func(state refreshReadCacheState) refreshReadCacheState {
		return state.withActionsPopupVisibleLines(visibleLines)
	})
}

func (program *Program) cacheReviewSessionReadModel(model reviewSessionReadModel) {
	if program == nil || !program.refreshReadCache.enabled {
		return
	}
	program.updateRefreshReadCache(func(state refreshReadCacheState) refreshReadCacheState {
		return state.withReviewSessionReadModel(model)
	})
}

func (program *Program) cacheKeybindingLabelResolver(resolver keybindingLabelResolver) {
	if program == nil || !program.refreshReadCache.enabled {
		return
	}
	program.updateRefreshReadCache(func(state refreshReadCacheState) refreshReadCacheState {
		return state.withKeybindingResolver(resolver)
	})
}

func (program *Program) cacheFooterPresenter(presenter footerPresenter) {
	if program == nil || !program.refreshReadCache.enabled {
		return
	}
	program.updateRefreshReadCache(func(state refreshReadCacheState) refreshReadCacheState {
		return state.withFooterPresenter(presenter)
	})
}

func (program *Program) cacheActionsPopupPresenter(presenter actionsPopupPresenter) {
	if program == nil || !program.refreshReadCache.enabled {
		return
	}
	program.updateRefreshReadCache(func(state refreshReadCacheState) refreshReadCacheState {
		return state.withActionsPopupPresenter(presenter)
	})
}
