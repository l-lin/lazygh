package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"
)

type clearPersistentCacheCmd struct{}

func (clearPersistentCacheCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	err := error(nil)
	if program.pullRequestCache == nil {
		err = errors.New("persistent cache is unavailable")
	} else {
		err = program.pullRequestCache.Clear()
	}
	_ = program.executeRuntimeMessage(gui, MsgPersistentCacheCleared{Err: err})
}
