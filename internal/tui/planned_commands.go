package tui

import "github.com/jesseduffield/gocui"

func (program *Program) plannedCommands(gui *gocui.Gui) []Cmd {
	if program == nil || gui == nil {
		return nil
	}

	commands := make([]Cmd, 0, 8)
	commands = append(commands, program.sessionStore.planLoad(program, gui)...)
	commands = append(commands, program.pullRequestListStore.planLoad(program, gui, program.model.ActivePullRequestTab())...)
	if !program.reviewModeActive() {
		commands = append(commands, program.notificationStore.planLoad(program, gui)...)
		commands = append(commands, program.detailStore.planSelectedNotificationDetailLoad(program, gui)...)
	}
	commands = append(commands, program.detailStore.planSelectedPullRequestDetailLoad(program, gui)...)
	commands = append(commands, program.reviewStore.planSelectedPullRequestDiffLoad(program, gui)...)
	commands = append(commands, program.imageLoadCoordinator.planCurrentDetailImageHTMLLoads(program, gui)...)
	commands = append(commands, program.imageLoadCoordinator.planCurrentDetailImageLoads(program, gui)...)
	return commands
}
