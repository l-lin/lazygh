package tui

func (program *Program) plannedWorkflow() workflowPlan {
	if program == nil {
		return workflowPlan{}
	}

	actual := workflowPlan{}
	actual.append(program.sessionLoadPlan())
	actual.append(program.pullRequestListLoadPlan(program.model.ActivePullRequestTab()))
	if !program.reviewModeActive() {
		actual.append(program.notificationLoadPlan())
		actual.append(program.selectedNotificationDetailLoadPlan())
	}
	actual.append(program.selectedPullRequestDetailLoadPlan())
	actual.append(program.selectedPullRequestDiffLoadPlan())
	actual.append(program.currentDetailImageHTMLLoadsPlan())
	actual.append(program.currentDetailImageLoadsPlan())
	return actual
}
