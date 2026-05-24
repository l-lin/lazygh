package tui

import "github.com/jesseduffield/gocui"

type Cmd interface {
	execute(*Program, *gocui.Gui)
}

func (program *Program) executeCmds(gui *gocui.Gui, cmds []Cmd) {
	if program == nil || len(cmds) == 0 {
		return
	}
	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		cmd.execute(program, gui)
	}
}

func (program *Program) executeWorkflowCommands(gui *gocui.Gui, commands []Cmd) {
	program.executeCmds(gui, commands)
}
