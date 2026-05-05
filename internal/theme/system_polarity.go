package theme

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type systemPolarity int

const (
	systemPolarityUnknown systemPolarity = iota
	systemPolarityLight
	systemPolarityDark
	systemPolarityCommandTimeout = time.Second
)

type systemCommandRunner func(name string, args ...string) (string, error)

func detectSystemPolarity() systemPolarity {
	return detectSystemPolarityForOS(runtime.GOOS, runSystemCommand)
}

func detectSystemPolarityForOS(goos string, run systemCommandRunner) systemPolarity {
	switch goos {
	case "darwin":
		return detectDarwinSystemPolarity(run)
	case "linux":
		return detectLinuxSystemPolarity(run)
	case "windows":
		return detectWindowsSystemPolarity(run)
	default:
		return systemPolarityUnknown
	}
}

func detectDarwinSystemPolarity(run systemCommandRunner) systemPolarity {
	output, actualErr := run("defaults", "read", "-g", "AppleInterfaceStyle")
	normalizedOutput := normalizeSystemPolarityCommandOutput(output)
	switch {
	case strings.Contains(normalizedOutput, "dark"):
		return systemPolarityDark
	case strings.Contains(normalizedOutput, "light"), strings.Contains(normalizedOutput, "does not exist"):
		return systemPolarityLight
	case actualErr == nil && normalizedOutput == "":
		return systemPolarityLight
	default:
		return systemPolarityUnknown
	}
}

func detectLinuxSystemPolarity(run systemCommandRunner) systemPolarity {
	colorSchemeOutput, _ := run("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	if polarity := linuxColorSchemePolarity(colorSchemeOutput); polarity != systemPolarityUnknown {
		return polarity
	}

	gtkThemeOutput, actualErr := run("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme")
	if actualErr == nil {
		return themeNamePolarity(gtkThemeOutput)
	}

	return systemPolarityUnknown
}

func linuxColorSchemePolarity(output string) systemPolarity {
	normalizedOutput := normalizeSystemPolarityCommandOutput(output)
	switch {
	case strings.Contains(normalizedOutput, "prefer-dark"):
		return systemPolarityDark
	case strings.Contains(normalizedOutput, "prefer-light"):
		return systemPolarityLight
	default:
		return systemPolarityUnknown
	}
}

func detectWindowsSystemPolarity(run systemCommandRunner) systemPolarity {
	output, _ := run("reg", "query", `HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize`, "/v", "AppsUseLightTheme")
	normalizedOutput := normalizeSystemPolarityCommandOutput(output)
	switch {
	case strings.Contains(normalizedOutput, "0x0"):
		return systemPolarityDark
	case strings.Contains(normalizedOutput, "0x1"):
		return systemPolarityLight
	default:
		return systemPolarityUnknown
	}
}

func themeNamePolarity(output string) systemPolarity {
	normalizedOutput := normalizeSystemPolarityCommandOutput(output)
	if normalizedOutput == "" {
		return systemPolarityUnknown
	}
	if strings.Contains(normalizedOutput, "dark") {
		return systemPolarityDark
	}

	return systemPolarityLight
}

func normalizeSystemPolarityCommandOutput(output string) string {
	trimmedOutput := strings.TrimSpace(strings.ToLower(output))
	trimmedOutput = strings.Trim(trimmedOutput, "'\"")
	return trimmedOutput
}

func runSystemCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), systemPolarityCommandTimeout)
	defer cancel()

	output, actualErr := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return string(output), actualErr
}
