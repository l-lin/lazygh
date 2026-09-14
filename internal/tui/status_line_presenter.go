package tui

import "strings"

type statusLinePresenter struct {
	feedbackMessage                       string
	statusLineOperationLoadingMessage     string
	statusLineOperationFailureMessage     string
	statusLineOperationSuccessMessage     string
	statusLineOperationStarted            bool
	loadingSpinner                        string
	storyReviewLoading                    bool
	storyReviewLoadingMessage             string
	assigneePickerLoadingMessage          string
	pullRequestBuildRunLoadingMessage     string
	ghCommandLoadingMessage               string
	selectedPullRequestDetailLoadingText  string
	selectedPullRequestDiffLoadingText    string
	selectedNotificationDetailLoadingText string
	activePullRequestsLoadingText         string
	notificationsLoadingText              string
}

func (presenter statusLinePresenter) Text() string {
	if message := strings.TrimSpace(presenter.statusLineOperationFailureMessage); message != "" {
		return message
	}
	if presenter.statusLineOperationStarted || strings.TrimSpace(presenter.statusLineOperationLoadingMessage) != "" {
		return presenter.loadingText()
	}
	if message := strings.TrimSpace(presenter.statusLineOperationSuccessMessage); message != "" {
		return message
	}
	if message := strings.TrimSpace(presenter.feedbackMessage); message != "" {
		return message
	}
	if message := strings.TrimSpace(presenter.loadingText()); message != "" {
		return message
	}
	return ""
}

func (presenter statusLinePresenter) IsFailure() bool {
	return strings.TrimSpace(presenter.statusLineOperationFailureMessage) != ""
}

func (presenter statusLinePresenter) loadingText() string {
	if message := strings.TrimSpace(presenter.statusLineOperationLoadingMessage); message != "" {
		return presenter.loadingSpinnerStatus(message)
	}
	if presenter.statusLineOperationStarted {
		return ""
	}
	if presenter.storyReviewLoading {
		return presenter.loadingSpinnerStatus(presenter.storyReviewLoadingMessage)
	}
	for _, message := range []string{
		presenter.assigneePickerLoadingMessage,
		presenter.pullRequestBuildRunLoadingMessage,
		presenter.ghCommandLoadingMessage,
		presenter.selectedPullRequestDetailLoadingText,
		presenter.selectedPullRequestDiffLoadingText,
		presenter.selectedNotificationDetailLoadingText,
		presenter.activePullRequestsLoadingText,
		presenter.notificationsLoadingText,
	} {
		if trimmedMessage := strings.TrimSpace(message); trimmedMessage != "" {
			return presenter.loadingSpinnerStatus(trimmedMessage)
		}
	}
	return ""
}

func (presenter statusLinePresenter) loadingSpinnerStatus(label string) string {
	trimmedSpinner := strings.TrimSpace(presenter.loadingSpinner)
	trimmedLabel := strings.TrimSpace(label)
	switch {
	case trimmedSpinner == "":
		return trimmedLabel
	case trimmedLabel == "":
		return trimmedSpinner
	default:
		return trimmedSpinner + " " + trimmedLabel
	}
}
