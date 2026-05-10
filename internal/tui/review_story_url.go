package tui

import "github.com/l-lin/lazygh/internal/githubcli"

func (program *Program) OpenStoryReviewByURL(rawURL string) error {
	if actualErr := program.validateStoryReviewAvailability(); actualErr != nil {
		return actualErr
	}

	summary, actualErr := githubcli.ParsePullRequestURL(rawURL)
	if actualErr != nil {
		return actualErr
	}

	prepared, actualErr := program.prepareStoryReview(summary)
	if actualErr != nil {
		return actualErr
	}

	program.feedbackMessage = ""
	program.storyReviewLoading = false
	program.applyPreparedStoryReview(prepared)
	if program.gui != nil {
		return program.layout(program.gui)
	}
	return nil
}
