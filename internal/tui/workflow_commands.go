package tui

import "github.com/jesseduffield/gocui"

type workflowCommand interface {
	execute(*Program, *gocui.Gui)
}

type asyncWorkflowCommand struct {
	run func(*Program, *gocui.Gui)
}

func (command asyncWorkflowCommand) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.run == nil {
		return
	}
	program.asyncRunner.Go(func() {
		command.run(program, gui)
	})
}

func newAsyncWorkflowCommand(run func(*Program, *gocui.Gui)) workflowCommand {
	return asyncWorkflowCommand{run: run}
}

func (program *Program) executeWorkflowCommands(gui *gocui.Gui, commands []workflowCommand) {
	if program == nil || len(commands) == 0 {
		return
	}
	for _, command := range commands {
		if command == nil {
			continue
		}
		command.execute(program, gui)
	}
}
