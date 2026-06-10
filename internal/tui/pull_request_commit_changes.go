package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const displayCommitChangesActionTitle = "Display commit changes"

type pullRequestCommitActionTarget struct {
	summary githubdomain.PullRequest
	commit  githubdomain.PullRequestCommit
}

func (target pullRequestCommitActionTarget) commitChangesURL() (string, bool) {
	return githubdomain.PullRequestCommitChangesURL(target.summary.Repository, target.summary.Number, target.commit.OID)
}

func (program *Program) selectedPullRequestCommitActionTarget() (pullRequestCommitActionTarget, bool) {
	if program == nil || program.model == nil {
		return pullRequestCommitActionTarget{}, false
	}

	actionContext := program.actionContext()
	if actionContext.ActiveView.Focus != FocusDetailView || actionContext.IsReviewContext() || actionContext.MainView.ContentKind != MainContentKindPullRequestDetail || actionContext.ActiveDetailTab != CommitsDetailTab {
		return pullRequestCommitActionTarget{}, false
	}

	context, ok := program.currentPullRequestCommitsCursorContext()
	if !ok {
		return pullRequestCommitActionTarget{}, false
	}
	commit, ok := commitAtCursor(context)
	if !ok {
		return pullRequestCommitActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(context.summary.Repository))
	if repository == "" || repository == "-" || context.summary.Number <= 0 || strings.TrimSpace(commit.OID) == "" {
		return pullRequestCommitActionTarget{}, false
	}
	return pullRequestCommitActionTarget{summary: context.summary, commit: commit}, true
}

func (program *Program) commitChangesLinkUnderCursor() (string, bool) {
	target, ok := program.selectedPullRequestCommitActionTarget()
	if !ok {
		return "", false
	}
	return target.commitChangesURL()
}

func (program *Program) displayCommitChangesShortcut(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgDisplayCommitChangesRequested{})
}

func (program *Program) displayCommitChangesShortcutAvailable() bool {
	_, ok := program.selectedPullRequestCommitActionTarget()
	return ok
}

func (program *Program) currentDisplayCommitChangesAction() (actionsPopupAction, bool) {
	if !program.displayCommitChangesShortcutAvailable() {
		return actionsPopupAction{}, false
	}
	return actionsPopupAction{
		id:        "display-commit-changes",
		title:     displayCommitChangesActionTitle,
		icon:      actionsPopupDisplayCommitChangesIcon,
		requested: MsgDisplayCommitChangesRequested{},
	}, true
}

func (program *Program) applyDisplayCommitChangesRequested(_ MsgDisplayCommitChangesRequested) []Cmd {
	if program == nil {
		return nil
	}

	program.clearDetailPendingPrefix()
	if !program.displayCommitChangesShortcutAvailable() {
		return nil
	}
	return nil
}
