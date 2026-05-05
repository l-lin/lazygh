package tui

import (
	persistcache "codeberg.org/l-lin/lazygh/internal/cache"
	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

func (program *Program) ApplyCacheConfig(config appconfig.CacheConfig) error {
	if program.pullRequestCache != nil {
		_ = program.pullRequestCache.Close()
		program.pullRequestCache = nil
	}
	if config.Path == "" {
		return nil
	}

	store, actualErr := persistcache.Open(config.Path)
	if actualErr != nil {
		return actualErr
	}

	program.pullRequestCache = store
	return nil
}
