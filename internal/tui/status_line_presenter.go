package tui

import "strings"

type statusLinePresenter struct {
	feedbackMessage                       string
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
	if message := strings.TrimSpace(presenter.feedbackMessage); message != "" {
		return message
	}
	if message := strings.TrimSpace(presenter.loadingText()); message != "" {
		return message
	}
	return ""
}

func (presenter statusLinePresenter) loadingText() string {
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
