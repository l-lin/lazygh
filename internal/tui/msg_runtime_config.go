package tui

import appconfig "github.com/l-lin/lazygh/internal/config"

type MsgPullRequestSearchesApplied struct {
	Searches []appconfig.PullRequestSearch
}

func (MsgPullRequestSearchesApplied) isMsg() {}
