package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const pullRequestBuildInfoUnavailableMessage = "󰅚 Build info unavailable"

func (program *Program) handleBrowserOverviewBuildEnter(gui *gocui.Gui, summary githubcli.PullRequest, detail githubcli.PullRequestDetail, _ detailDocument, sectionAtCursor browserDetailSectionCursor) (error, bool) {
	if !sectionAtCursor.inBody || !strings.EqualFold(strings.TrimSpace(sectionAtCursor.section.overviewBlockTitle), "Builds") {
		return nil, false
	}

	entry, ok := pullRequestOverviewEntryAtBodyLine(sectionAtCursor.section, sectionAtCursor.bodyLine)
	if !ok || strings.TrimSpace(entry.Link) == "" {
		return nil, true
	}
	if program.githubLoader == nil {
		return nil, true
	}

	check, ok := pullRequestStatusCheckMatchingEntry(detail.StatusCheckRollup, entry)
	if !ok {
		return nil, true
	}

	buildInfo, err := program.githubLoader.GetPullRequestBuildInfo(pullRequestRepositoryName(summary.Repository), summary.Number, check)
	if err != nil {
		program.setFeedback(program.model.Focus(), pullRequestBuildInfoUnavailableMessage)
		return program.refreshViewsIfGUI(gui), true
	}
	return program.openPullRequestBuildInfoPopup(gui, buildInfo), true
}

func pullRequestStatusCheckMatchingEntry(checks []githubcli.PullRequestStatusCheck, entry pullRequestOverviewEntry) (githubcli.PullRequestStatusCheck, bool) {
	trimmedLink := strings.TrimSpace(entry.Link)
	if trimmedLink != "" {
		for _, check := range checks {
			if strings.EqualFold(strings.TrimSpace(check.Link), trimmedLink) {
				return check, true
			}
		}
	}

	trimmedLabel := strings.TrimSpace(entry.Label)
	if trimmedLabel == "" {
		return githubcli.PullRequestStatusCheck{}, false
	}
	for _, check := range checks {
		if strings.EqualFold(strings.TrimSpace(buildPullRequestBuildEntry(check).Label), trimmedLabel) {
			return check, true
		}
	}
	return githubcli.PullRequestStatusCheck{}, false
}
