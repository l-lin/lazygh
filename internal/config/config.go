package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	configDirectoryName = "lazygh"
	configFileName      = "config.toml"
)

type Config struct {
	Keymaps KeymapOverrides
}

type KeymapOverrides map[string]map[string][]string

type rawConfig struct {
	Keymaps map[string]map[string]any `toml:"keymaps"`
}

func DefaultPath(homeDirectory string) string {
	return filepath.Join(homeDirectory, ".config", configDirectoryName, configFileName)
}

func LoadDefault() (Config, error) {
	homeDirectory, actualErr := os.UserHomeDir()
	if actualErr != nil {
		return Config{}, actualErr
	}

	return Load(DefaultPath(homeDirectory))
}

func Load(configPath string) (Config, error) {
	_, actualErr := os.Stat(configPath)
	if errors.Is(actualErr, os.ErrNotExist) {
		return Config{}, nil
	}
	if actualErr != nil {
		return Config{}, actualErr
	}

	var raw rawConfig
	if _, actualErr = toml.DecodeFile(configPath, &raw); actualErr != nil {
		return Config{}, actualErr
	}

	return Config{Keymaps: normalizeKeymapOverrides(raw.Keymaps)}, nil
}

func normalizeKeymapOverrides(rawScopes map[string]map[string]any) KeymapOverrides {
	if len(rawScopes) == 0 {
		return nil
	}

	normalizedScopes := make(KeymapOverrides)
	for scopeName, rawActions := range rawScopes {
		normalizedActions := normalizeKeymapActions(rawActions)
		if len(normalizedActions) == 0 {
			continue
		}
		normalizedScopes[scopeName] = normalizedActions
	}

	if len(normalizedScopes) == 0 {
		return nil
	}

	return normalizedScopes
}

func normalizeKeymapActions(rawActions map[string]any) map[string][]string {
	normalizedActions := map[string][]string{}
	for actionName, rawValue := range rawActions {
		normalizedValue, ok := normalizeKeymapValue(rawValue)
		if !ok {
			continue
		}
		normalizedActions[actionName] = normalizedValue
	}

	if len(normalizedActions) == 0 {
		return nil
	}

	return normalizedActions
}

func normalizeKeymapValue(rawValue any) ([]string, bool) {
	switch actual := rawValue.(type) {
	case string:
		return normalizeKeymapStrings([]string{actual})
	case []string:
		return normalizeKeymapStrings(actual)
	case []any:
		values := make([]string, 0, len(actual))
		for _, value := range actual {
			stringValue, ok := value.(string)
			if !ok {
				return nil, false
			}
			values = append(values, stringValue)
		}
		return normalizeKeymapStrings(values)
	default:
		return nil, false
	}
}

func normalizeKeymapStrings(values []string) ([]string, bool) {
	if len(values) == 0 {
		return nil, false
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, false
		}
		normalized = append(normalized, value)
	}

	return normalized, true
}
