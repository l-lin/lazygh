package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) withPullRequestDiffFileTeamOwners(repository string, number int, rawDiff githubdomain.PullRequestDiff) githubdomain.PullRequestDiff {
	if rawDiff.FileTeamOwnersAttempted || !program.hasDetailQueries() || !program.shouldLoadPullRequestDiffTeamOwners() {
		return rawDiff
	}

	rawDiff.FileTeamOwnersAttempted = true
	filePaths := pullRequestDiffFilePaths(rawDiff.Files)
	if len(filePaths) == 0 {
		return rawDiff
	}

	teamOwnersByPath, err := program.detailQueries.GetPullRequestFileTeamOwners(repository, number, filePaths)
	if err != nil {
		return rawDiff
	}
	if len(teamOwnersByPath) == 0 {
		return rawDiff
	}

	rawDiff.Files = pullRequestDiffFilesWithTeamOwners(rawDiff.Files, teamOwnersByPath)
	return rawDiff
}

func pullRequestDiffFilePaths(files []githubdomain.PullRequestDiffFile) []string {
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

func pullRequestDiffFilesWithTeamOwners(files []githubdomain.PullRequestDiffFile, teamOwnersByPath map[string][]string) []githubdomain.PullRequestDiffFile {
	if len(files) == 0 {
		return nil
	}

	updatedFiles := make([]githubdomain.PullRequestDiffFile, 0, len(files))
	for _, file := range files {
		updatedFile := file
		updatedFile.TeamOwners = normalizeReviewDiffTeamOwners(teamOwnersByPath[strings.TrimSpace(file.Path)])
		updatedFiles = append(updatedFiles, updatedFile)
	}
	return updatedFiles
}

func (program *Program) shouldLoadPullRequestDiffTeamOwners() bool {
	if program.reviewModeActive() {
		return true
	}
	return program.shouldShowPullRequestDetailTabs() && program.detailState.activeTab == ChangesDetailTab
}
