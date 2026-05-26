package tui

func (program *Program) dispatchRuntimeMessage(msg Msg) error {
	if program == nil || msg == nil {
		return nil
	}
	if program.gui != nil {
		return program.dispatch(program.gui, msg)
	}
	program.executeCmds(nil, Update(program, msg))
	return nil
}
