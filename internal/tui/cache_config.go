package tui

import appconfig "codeberg.org/l-lin/lazygh/internal/config"

func (program *Program) ApplyCacheConfig(config appconfig.CacheConfig) error {
	_ = config
	return nil
}
