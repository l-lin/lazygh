package githubcli

import "strings"

type PullRequestAutoMergeRequest struct {
	EnabledAt string `json:"enabledAt"`
}

func (request PullRequestAutoMergeRequest) normalized() PullRequestAutoMergeRequest {
	request.EnabledAt = strings.TrimSpace(request.EnabledAt)
	return request
}
