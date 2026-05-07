package tui

import (
	"path"
	"strings"
)

const (
	reviewDiffHeaderPathIcon  = detailInlineCommentLocationIcon
	reviewDiffDirectoryIcon   = ""
	reviewDiffDefaultFileIcon = ""
)

var reviewDiffFileIconsByName = map[string]string{
	"go.mod":     "",
	"go.sum":     "",
	"Gemfile":    "",
	"Rakefile":   "",
	"README":     "",
	"README.md":  "",
	"Makefile":   "",
	"Dockerfile": "󰡨",
}

var reviewDiffFileIconsByExtension = map[string]string{
	".go":     "",
	".rb":     "",
	".md":     "",
	".yaml":   "",
	".yml":    "",
	".json":   "",
	".toml":   "",
	".xml":    "󰗀",
	".html":   "",
	".css":    "",
	".js":     "",
	".ts":     "",
	".tsx":    "",
	".jsx":    "",
	".sh":     "",
	".bash":   "",
	".zsh":    "",
	".sql":    "",
	".proto":  "󱘖",
	".png":    "",
	".jpg":    "",
	".jpeg":   "",
	".gif":    "",
	".svg":    "󰜡",
	".kt":     "󱈙",
	".java":   "",
	".docker": "󰡨",
}

func reviewDiffFileIcon(filePath string) string {
	trimmedPath := strings.TrimSpace(strings.TrimSuffix(filePath, "/"))
	if trimmedPath == "" {
		return reviewDiffDefaultFileIcon
	}

	baseName := path.Base(trimmedPath)
	if icon, ok := reviewDiffFileIconsByName[baseName]; ok {
		return icon
	}
	if icon, ok := reviewDiffFileIconsByExtension[strings.ToLower(path.Ext(baseName))]; ok {
		return icon
	}
	return reviewDiffDefaultFileIcon
}

func reviewDiffTreeRowIcon(row reviewDiffTreeRow) string {
	if row.Kind == reviewDiffTreeRowKindChapter {
		return reviewDiffDirectoryIcon
	}
	if row.FileIndex >= 0 {
		return reviewDiffFileIcon(row.Label)
	}
	return reviewDiffDirectoryIcon
}
