package tui

func (program *Program) currentDetailImageSourceReadModel() detailImageSourceReadModel {
	model := detailImageSourceReadModel{}
	if program == nil {
		return model
	}

	model.activeTab = program.detailState.activeTab
	model.reviewModeActive = program.reviewModeActive()
	if model.reviewModeActive {
		reviewModel := program.reviewSessionReadModel()
		if !reviewModel.isActive() {
			return model
		}
		model.summary = reviewModel.summary
		model.summaryKnown = true
		model.reviewShowsDescription = reviewModel.showsDescription()
		model.reviewShowsStoryChapter = reviewModel.showsStoryChapter()
		if result, ok := program.pullRequestDetailForSummary(reviewModel.summary); ok && result.err == nil {
			model.detail = result.detail
			model.detailKnown = true
		}
		if model.reviewShowsStoryChapter {
			if chapter, ok := reviewModel.selectedStoryChapter(); ok {
				model.reviewStoryChapter = chapter
				model.reviewStoryChapterKnown = true
			}
			return model
		}
		if model.reviewShowsDescription {
			return model
		}
		if result, ok := program.pullRequestDiffForSummary(reviewModel.summary); ok && result.err == nil {
			if selectedFile, ok := reviewModel.selectedDiffFile(); ok {
				model.reviewDiffFile = selectedFile
				model.reviewDiffFileKnown = true
				model.reviewDiffFileIndex = pullRequestDiffFileIndex(result.data.Files, selectedFile.Path)
			}
		}
		return model
	}

	if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
		model.summary = summary
		model.summaryKnown = true
		if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
			model.detail = result.detail
			model.detailKnown = true
		}
		if model.activeTab == ChangesDetailTab {
			if result, ok := program.pullRequestDiffForSummary(summary); ok && result.err == nil {
				model.diffFiles = append([]reviewDiffFile(nil), result.data.Files...)
				model.diffFilesKnown = true
			}
		}
		return model
	}
	if program.model == nil || program.model.currentSideFocus() != FocusNotificationsView {
		return model
	}

	notification, ok := program.model.SelectedNotification()
	if !ok {
		return model
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		model.issueRepository = repository
		model.issueNumber = number
		model.issueKnown = true
		if result, ok := program.issueDetailForNotification(notification); ok && result.err == nil {
			model.issueDetail = result.detail
			model.issueDetailKnown = true
		}
		return model
	}
	if repository, id, ok := notification.ReleaseIdentity(); ok {
		model.releaseRepository = repository
		model.releaseID = id
		model.releaseKnown = true
		if result, ok := program.releaseDetailForNotification(notification); ok && result.err == nil {
			model.releaseDetail = result.detail
			model.releaseDetailKnown = true
		}
	}
	return model
}

func (program *Program) currentDetailImageHTMLSources() []detailImageHTMLSource {
	return program.currentDetailImageSourceReadModel().sources()
}
