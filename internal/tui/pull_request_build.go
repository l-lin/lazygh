package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	pullRequestBuildRunUnavailableMessage  = iconWarning + " Build run unavailable"
	pullRequestBuildLogsUnavailableMessage = iconWarning + " Build logs unavailable"
	pullRequestBuildRunActionTitle         = "View build run"
	pullRequestBuildRunLogsActionTitle     = "View job logs"
)

type pullRequestBuildRunTarget struct {
	summary      githubdomain.PullRequest
	check        githubdomain.PullRequestStatusCheck
	popupContent pullRequestBuildRunPopupContent
}

func (program *Program) detailCursorHasBuildLink() bool {
	_, ok := program.currentPullRequestBuildRunTargetAtDetailCursor()
	return ok
}

func (program *Program) currentPullRequestBuildRunTargetAtDetailCursor() (pullRequestBuildRunTarget, bool) {
	model := program.currentDescriptionCursorActionReadModel()
	entry, ok := model.buildActionEntryAtCursor()
	if !ok {
		return pullRequestBuildRunTarget{}, false
	}

	check, ok := pullRequestStatusCheckMatchingEntry(model.detail.StatusCheckRollup, entry)
	if !ok {
		return pullRequestBuildRunTarget{}, false
	}
	repository := pullRequestRepositoryName(model.summary.Repository)
	return pullRequestBuildRunTarget{
		summary: model.summary,
		check:   check,
		popupContent: pullRequestBuildRunPopupContent{
			checkTitle: checkTitleForPullRequestBuildRunPopup(check),
			runURL:     strings.TrimSpace(check.Link),
			repository: repository,
		},
	}, true
}

func (program *Program) pullRequestBuildRunActionsPopupAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.currentPullRequestBuildRunTargetAtDetailCursor(); ok {
		requested = MsgPullRequestBuildRunLoadRequested{Target: target}
	}
	return actionsPopupAction{
		id:        "view-build-run",
		title:     pullRequestBuildRunActionTitle,
		icon:      actionsPopupBuildRunIcon,
		requested: requested,
	}
}

func (program *Program) pullRequestBuildRunLogsActionsPopupAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.currentPullRequestBuildRunTargetAtDetailCursor(); ok {
		requested = MsgPullRequestBuildRunJobLogLoadRequested{Summary: target.summary, Check: target.check}
	}
	return actionsPopupAction{
		id:        "view-build-run-job-logs",
		title:     pullRequestBuildRunLogsActionTitle,
		icon:      actionsPopupBuildRunLogsIcon,
		requested: requested,
	}
}

func pullRequestStatusCheckMatchingEntry(checks []githubdomain.PullRequestStatusCheck, entry pullRequestOverviewEntry) (githubdomain.PullRequestStatusCheck, bool) {
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
		return githubdomain.PullRequestStatusCheck{}, false
	}
	for _, check := range checks {
		if strings.EqualFold(strings.TrimSpace(buildPullRequestBuildEntry(check).Label), trimmedLabel) {
			return check, true
		}
	}
	return githubdomain.PullRequestStatusCheck{}, false
}

func checkTitleForPullRequestBuildRunPopup(check githubdomain.PullRequestStatusCheck) string {
	return pullRequestOverviewCheckDisplayName(check)
}
