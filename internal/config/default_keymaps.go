package config

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/BurntSushi/toml"
)

//go:embed default_keymaps.toml
var defaultKeymapsTOML string

var (
	defaultKeymapsOnce  sync.Once
	cachedDefaultKeymap KeymapOverrides
)

type rawKeymapConfig struct {
	Keymaps map[string]map[string]any `toml:"keymaps"`
}

func DefaultKeymaps() KeymapOverrides {
	defaultKeymapsOnce.Do(func() {
		cachedDefaultKeymap = mustLoadDefaultKeymaps(defaultKeymapsTOML)
	})
	return copyKeymapOverrides(cachedDefaultKeymap)
}

func mustLoadDefaultKeymaps(contents string) KeymapOverrides {
	var raw rawKeymapConfig
	if _, actualErr := toml.Decode(contents, &raw); actualErr != nil {
		panic(fmt.Errorf("parse embedded default keymaps: %w", actualErr))
	}

	normalized := normalizeKeymapOverrides(raw.Keymaps)
	if len(normalized) == 0 {
		panic("embedded default keymaps are empty")
	}
	return normalized
}

func copyKeymapOverrides(overrides KeymapOverrides) KeymapOverrides {
	if len(overrides) == 0 {
		return nil
	}

	copiedScopes := make(KeymapOverrides, len(overrides))
	for scopeName, actions := range overrides {
		copiedActions := make(map[string][]string, len(actions))
		for actionName, bindings := range actions {
			copiedActions[actionName] = append([]string(nil), bindings...)
		}
		copiedScopes[scopeName] = copiedActions
	}
	return copiedScopes
}
