package tui

func (program *Program) helpPresenter() helpPresenter {
	if program == nil {
		return helpPresenter{}
	}

	return helpPresenter{
		actionContext:                    program.actionContext(),
		keyResolver:                      program.keybindingLabelResolver(),
		inlineCommentReplyAvailable:      program.inlineCommentReplyShortcutAvailable(),
		inlineCommentResolutionHelpLabel: program.inlineCommentResolutionShortcutDescription(),
		pullRequestBrowserAvailable:      program.pullRequestBrowserShortcutAvailable(),
	}
}
