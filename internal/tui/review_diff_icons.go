package tui

func reviewDiffFileIcon(filePath string) string {
	return iconFileForPath(filePath)
}

func reviewDiffTreeRowIcon(row reviewDiffTreeRow) string {
	if row.Kind == reviewDiffTreeRowKindChapter {
		return reviewModeChapterIcon
	}
	if row.FileIndex >= 0 {
		return reviewDiffFileIcon(row.Label)
	}
	return reviewDiffDirectoryIcon
}
