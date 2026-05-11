package tui

import (
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) withPullRequestDiffFileTeamOwners(repository string, number int, rawDiff githubcli.PullRequestDiff) githubcli.PullRequestDiff {
	if rawDiff.FileTeamOwnersAttempted || program.githubLoader == nil {
		return rawDiff
	}

	rawDiff.FileTeamOwnersAttempted = true
	filePaths := pullRequestDiffFilePaths(rawDiff.Files)
	if len(filePaths) == 0 {
		return rawDiff
	}

	teamOwnersByPath, err := program.githubLoader.GetPullRequestFileTeamOwners(repository, number, filePaths)
	if err != nil {
		return rawDiff
	}
	if len(teamOwnersByPath) == 0 {
		return rawDiff
	}

	rawDiff.Files = pullRequestDiffFilesWithTeamOwners(rawDiff.Files, teamOwnersByPath)
	return rawDiff
}

func pullRequestDiffFilePaths(files []githubcli.PullRequestDiffFile) []string {
	if len(files) == 0 {
		return nil
	}

	paths := make([]string, 0, len(files))
	seenPaths := map[string]bool{}
	for _, file := range files {
		trimmedPath := strings.TrimSpace(file.Path)
		if trimmedPath == "" || seenPaths[trimmedPath] {
			continue
		}
		seenPaths[trimmedPath] = true
		paths = append(paths, trimmedPath)
	}
	if len(paths) == 0 {
		return nil
	}
	return paths
}

func pullRequestDiffFilesWithTeamOwners(files []githubcli.PullRequestDiffFile, teamOwnersByPath map[string][]string) []githubcli.PullRequestDiffFile {
	if len(files) == 0 {
		return nil
	}

	updatedFiles := make([]githubcli.PullRequestDiffFile, 0, len(files))
	for _, file := range files {
		updatedFile := file
		updatedFile.TeamOwners = normalizeReviewDiffTeamOwners(teamOwnersByPath[strings.TrimSpace(file.Path)])
		updatedFiles = append(updatedFiles, updatedFile)
	}
	return updatedFiles
}
