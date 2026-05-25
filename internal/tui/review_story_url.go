package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) OpenStoryReviewByURL(rawURL string) error {
	if actualErr := program.validateStoryReviewAvailability(); actualErr != nil {
		return actualErr
	}

	summary, actualErr := githubdomain.ParsePullRequestURL(rawURL)
	if actualErr != nil {
		return actualErr
	}

	return program.dispatchStartupMessage(MsgReviewStoryRequested{Summary: summary})
}
