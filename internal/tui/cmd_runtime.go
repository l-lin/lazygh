package tui

func (program *Program) runAsync(run func()) {
	if program == nil || run == nil {
		return
	}
	if program.asyncRunner == nil {
		run()
		return
	}
	program.asyncRunner.Go(run)
}
