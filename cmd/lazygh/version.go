package main

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const defaultVersion = "devel"

var version = ""

func resolvedVersion() string {
	if trimmedVersion := strings.TrimSpace(version); trimmedVersion != "" {
		return trimmedVersion
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return defaultVersion
	}

	trimmedVersion := strings.TrimSpace(buildInfo.Main.Version)
	if trimmedVersion == "" {
		return defaultVersion
	}
	if trimmedVersion == "(devel)" {
		return buildInfoVersionFromRevision(buildInfo)
	}
	return trimmedVersion
}

func buildInfoVersionFromRevision(buildInfo *debug.BuildInfo) string {
	revision := ""
	modified := false
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = strings.TrimSpace(setting.Value)
		case "vcs.modified":
			modified = strings.EqualFold(strings.TrimSpace(setting.Value), "true")
		}
	}
	if revision == "" {
		return defaultVersion
	}

	if len(revision) > 12 {
		revision = revision[:12]
	}
	if modified {
		return fmt.Sprintf("devel+%s-dirty", revision)
	}
	return fmt.Sprintf("devel+%s", revision)
}

func formatVersionOutput(version string) string {
	resolved := strings.TrimSpace(version)
	if resolved == "" {
		resolved = defaultVersion
	}
	return fmt.Sprintf("lazygh %s", resolved)
}
