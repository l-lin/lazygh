package tui

func (program *Program) applyToggleInlineConversationVisibilityResolved(message MsgToggleInlineConversationVisibilityResolved) {
	program.detailState.viewState.clearPendingPrefix()
	plan, ok := program.toggleInlineConversationVisibilityState(message.Document)
	if !ok {
		return
	}
	program.applyDetailViewSyncPlanResolved(MsgDetailViewSyncPlanResolved{Plan: plan, ViewportHeight: message.ViewportHeight})
}

func (program *Program) applySetAllDetailFoldsResolved(message MsgSetAllDetailFoldsResolved) {
	program.detailState.viewState.clearPendingPrefix()
	plan, ok := program.setAllDetailFolds(message.Document, message.Collapsed)
	if !ok {
		return
	}
	program.applyDetailViewSyncPlanResolved(MsgDetailViewSyncPlanResolved{Plan: plan, ViewportHeight: message.ViewportHeight})
}
