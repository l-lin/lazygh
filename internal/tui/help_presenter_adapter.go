package tui

func (program *Program) helpPresenter() helpPresenter {
	if program == nil {
		return helpPresenter{}
	}

	return helpPresenter{
		actionContext:               program.actionContext(),
		keyResolver:                 program.keybindingLabelResolver(),
		inlineCommentReplyAvailable: program.inlineCommentReplyShortcutAvailable(),
	}
}

func (program *Program) helpViewSize(maxX int, maxY int) (int, int) {
	return program.helpPresenter().viewSize(maxX, maxY)
}
