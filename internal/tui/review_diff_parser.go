package tui

import "strings"

func parseUnifiedReviewDiff(unifiedDiff string) []reviewDiffFile {
	normalizedDiff := strings.ReplaceAll(strings.ReplaceAll(unifiedDiff, "\r\n", "\n"), "\r", "\n")
	if strings.TrimSpace(normalizedDiff) == "" {
		return nil
	}

	lines := strings.Split(normalizedDiff, "\n")
	files := make([]reviewDiffFile, 0)
	var currentFile *reviewDiffFile
	var currentHunk *reviewDiffHunk
	leftLineNumber := 0
	rightLineNumber := 0

	flushCurrentFile := func() {
		if currentFile == nil {
			return
		}
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			currentHunk = nil
		}
		if currentFile.Path != "" || currentFile.PreviousPath != "" || len(currentFile.Hunks) > 0 {
			files = append(files, *currentFile)
		}
		currentFile = nil
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			flushCurrentFile()
			path, previousPath := parseReviewDiffGitPaths(line)
			currentFile = &reviewDiffFile{Path: path, PreviousPath: previousPath}
			continue
		}
		if currentFile == nil {
			continue
		}

		if strings.HasPrefix(line, "@@ ") {
			if currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}
			currentHunk = &reviewDiffHunk{Header: line}
			leftLineNumber, rightLineNumber, _ = parseDiffHunkHeader(line)
			continue
		}

		if currentHunk != nil {
			switch {
			case strings.HasPrefix(line, " "):
				currentHunk.Lines = append(currentHunk.Lines, reviewDiffLine{Kind: reviewDiffContextLine, Text: line[1:], LeftLine: leftLineNumber, RightLine: rightLineNumber, Side: reviewDiffLineSideBoth})
				leftLineNumber++
				rightLineNumber++
			case strings.HasPrefix(line, "-"):
				currentHunk.Lines = append(currentHunk.Lines, reviewDiffLine{Kind: reviewDiffDeletionLine, Text: line[1:], LeftLine: leftLineNumber, Side: reviewDiffLineSideLeft})
				leftLineNumber++
			case strings.HasPrefix(line, "+"):
				currentHunk.Lines = append(currentHunk.Lines, reviewDiffLine{Kind: reviewDiffAdditionLine, Text: line[1:], RightLine: rightLineNumber, Side: reviewDiffLineSideRight})
				rightLineNumber++
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "rename from "):
			currentFile.PreviousPath = normalizeReviewDiffPath(strings.TrimPrefix(line, "rename from "))
		case strings.HasPrefix(line, "rename to "):
			currentFile.Path = normalizeReviewDiffPath(strings.TrimPrefix(line, "rename to "))
		case strings.HasPrefix(line, "new file mode "):
			currentFile.ChangeType = reviewDiffChangeTypeAdded
		case strings.HasPrefix(line, "deleted file mode "):
			currentFile.ChangeType = reviewDiffChangeTypeRemoved
		case strings.HasPrefix(line, "--- "):
			currentFile.PreviousPath = normalizeReviewDiffPath(strings.TrimPrefix(line, "--- "))
			if currentFile.Path == "" {
				currentFile.Path = currentFile.PreviousPath
			}
		case strings.HasPrefix(line, "+++ "):
			if path := normalizeReviewDiffPath(strings.TrimPrefix(line, "+++ ")); path != "" {
				currentFile.Path = path
			}
		}
	}

	flushCurrentFile()
	return files
}

func parseReviewDiffGitPaths(line string) (string, string) {
	trimmedLine := strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
	previousPathToken, pathToken, ok := splitReviewDiffPathPair(trimmedLine)
	if !ok {
		return "", ""
	}

	previousPath := normalizeReviewDiffPath(previousPathToken)
	path := normalizeReviewDiffPath(pathToken)
	if path == "" {
		path = previousPath
	}
	return path, previousPath
}

func splitReviewDiffPathPair(text string) (string, string, bool) {
	firstToken, rest, ok := nextReviewDiffPathToken(text)
	if !ok {
		return "", "", false
	}
	secondToken, _, ok := nextReviewDiffPathToken(rest)
	if !ok {
		return "", "", false
	}
	return firstToken, secondToken, true
}

func nextReviewDiffPathToken(text string) (string, string, bool) {
	trimmedText := strings.TrimLeft(text, " ")
	if trimmedText == "" {
		return "", "", false
	}
	if trimmedText[0] != '"' {
		separatorIndex := strings.IndexByte(trimmedText, ' ')
		if separatorIndex < 0 {
			return trimmedText, "", true
		}
		return trimmedText[:separatorIndex], trimmedText[separatorIndex+1:], true
	}

	var builder strings.Builder
	escaped := false
	for index := 1; index < len(trimmedText); index++ {
		currentCharacter := trimmedText[index]
		if escaped {
			builder.WriteByte(currentCharacter)
			escaped = false
			continue
		}
		switch currentCharacter {
		case '\\':
			escaped = true
		case '"':
			return builder.String(), trimmedText[index+1:], true
		default:
			builder.WriteByte(currentCharacter)
		}
	}
	return "", "", false
}

func normalizeReviewDiffPath(path string) string {
	trimmedPath := strings.Trim(strings.TrimSpace(path), `"`)
	if trimmedPath == "" || trimmedPath == "/dev/null" {
		return ""
	}
	if strings.HasPrefix(trimmedPath, "a/") || strings.HasPrefix(trimmedPath, "b/") {
		return trimmedPath[2:]
	}
	return trimmedPath
}
