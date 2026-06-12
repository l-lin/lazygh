package tui

import appconfig "github.com/l-lin/lazygh/internal/config"

func (program *Program) ApplyDisplayConfig(config appconfig.DisplayConfig) {
	_ = program.dispatchRuntimeMessage(MsgDisplayConfigApplied{Config: config})
}
