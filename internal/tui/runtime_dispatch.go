package tui

func (program *Program) dispatchRuntimeMessage(msg Msg) error {
	if program == nil || msg == nil {
		return nil
	}
	if gui := program.captureGUI(nil); gui != nil {
		return program.dispatch(gui, msg)
	}
	program.executeCmds(nil, Update(program, msg))
	return nil
}
