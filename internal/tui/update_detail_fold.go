package tui

func (program *Program) applyToggleInlineConversationVisibilityResolved(message MsgToggleInlineConversationVisibilityResolved) {
	program.clearDetailPendingPrefix()
	plan, ok := program.toggleInlineConversationVisibilityState(message.Document)
	if !ok {
		return
	}
	program.applyDetailViewSyncPlanResolved(MsgDetailViewSyncPlanResolved{Plan: plan, ViewportHeight: message.ViewportHeight})
}

func (program *Program) applySetAllDetailFoldsResolved(message MsgSetAllDetailFoldsResolved) {
	program.clearDetailPendingPrefix()
	plan, ok := program.setAllDetailFolds(message.Document, message.Collapsed)
	if !ok {
		return
	}
	program.applyDetailViewSyncPlanResolved(MsgDetailViewSyncPlanResolved{Plan: plan, ViewportHeight: message.ViewportHeight})
}
