package tui

type refreshReadCacheState struct {
	enabled                   bool
	actionsPopupActions       []actionsPopupAction
	actionsPopupActionsKnown  bool
	actionsPopupVisibleLines  []actionsPopupVisibleLine
	actionsPopupVisibleKnown  bool
	reviewSessionReadModel    reviewSessionReadModel
	reviewSessionReadModelSet bool
	keybindingResolver        keybindingLabelResolver
	keybindingResolverSet     bool
	footerPresenter           footerPresenter
	footerPresenterSet        bool
	actionsPopupPresenter     actionsPopupPresenter
	actionsPopupPresenterSet  bool
}

func (program *Program) withRefreshReadCache(run func()) {
	if run == nil {
		return
	}
	if program == nil {
		run()
		return
	}

	previous := program.refreshReadCache
	program.refreshReadCache = refreshReadCacheState{enabled: true}
	defer func() {
		program.refreshReadCache = previous
	}()
	run()
}
