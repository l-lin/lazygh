package tui

import (
	"time"

	"github.com/jesseduffield/gocui"
)

type transientErrorPopupExpiryCmd struct {
	Generation uint64
	Delay      time.Duration
}

func (command transientErrorPopupExpiryCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.Delay <= 0 || program.timingState.after == nil || program.asyncRunner == nil {
		return
	}
	if program.captureGUI(gui) == nil {
		return
	}

	delay := program.timingState.after(command.Delay)
	program.asyncRunner.Go(func() {
		if delay != nil {
			<-delay
		}
		program.dispatchAsyncMessage(MsgTransientErrorPopupExpired{Generation: command.Generation})
	})
}
