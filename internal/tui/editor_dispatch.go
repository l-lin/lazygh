package tui

func (program *Program) dispatchEditorMessage(msg Msg) bool {
	if program == nil || msg == nil {
		return false
	}
	return program.dispatchRuntimeMessage(msg) == nil
}
