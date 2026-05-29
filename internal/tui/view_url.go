package tui

import (
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) OpenPullRequestByURL(rawURL string) error {
	summary, err := pullRequestSummaryForURL(newModalEditorSubmitCommandDeps(program), rawURL)
	if err != nil {
		return err
	}
	return program.openPullRequestInBrowser(summary)
}

func (program *Program) openPullRequestInBrowser(summary githubdomain.PullRequest) error {
	return program.dispatchStartupMessage(MsgOpenPullRequestInBrowserView{Summary: summary})
}

func (program *Program) openPullRequestInPastedTabByURL(rawURL string) error {
	summary, err := pullRequestSummaryForURL(newModalEditorSubmitCommandDeps(program), rawURL)
	if err != nil {
		return err
	}
	return program.dispatchStartupMessage(MsgOpenPullRequestInPastedTabView{Summary: summary})
}

func (program *Program) dispatchStartupMessage(msg Msg) error {
	return program.dispatchRuntimeMessage(msg)
}

func samePullRequestIdentity(left any, right any) bool {
	leftSummary, ok := toDomainPullRequestSummary(left)
	if !ok {
		return false
	}
	rightSummary, ok := toDomainPullRequestSummary(right)
	if !ok {
		return false
	}
	leftKey := pullRequestDetailKey(leftSummary.Repository, leftSummary.Number)
	rightKey := pullRequestDetailKey(rightSummary.Repository, rightSummary.Number)
	return leftKey != "" && leftKey == rightKey
}
