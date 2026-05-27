package tui

func given_programWithTestGitHubDeps(model *Model, githubLoader any) *Program {
	subject := NewProgramWithModelAndDeps(model, given_testAppDeps(githubLoader))
	subject.timingState.transientErrorPopupDuration = 0
	subject.actionsPopupWidget.assigneePickerSearchDebounceDelay = 0
	return subject
}

func given_testAppDeps(loader any) AppDeps {
	if loader == nil {
		return AppDeps{}
	}
	if deps, ok := loader.(AppDeps); ok {
		return deps
	}

	deps := AppDeps{}
	if actual, ok := loader.(SessionQueries); ok {
		deps.SessionQueries = actual
	}
	if actual, ok := loader.(PullRequestListQueries); ok {
		deps.PullRequestList = actual
	}
	if actual, ok := loader.(NotificationQueries); ok {
		deps.NotificationQueries = actual
	}
	if actual, ok := loader.(DetailQueries); ok {
		deps.DetailQueries = actual
	}
	if actual, ok := loader.(PullRequestMutations); ok {
		deps.PullRequestMutations = actual
	}
	if actual, ok := loader.(ReviewMutations); ok {
		deps.ReviewMutations = actual
	}
	if actual, ok := loader.(NotificationMutations); ok {
		deps.NotificationMutations = actual
	}
	if actual, ok := loader.(ReactionMutations); ok {
		deps.ReactionMutations = actual
	}
	if actual, ok := loader.(BuildQueries); ok {
		deps.BuildQueries = actual
	}
	if actual, ok := loader.(MarkdownHTMLRenderer); ok {
		deps.MarkdownHTMLRenderer = actual
	}
	if actual, ok := loader.(AuthTokenProvider); ok {
		deps.AuthTokenProvider = actual
	}
	if actual, ok := loader.(ClipboardReader); ok {
		deps.ClipboardReader = actual
	}
	if actual, ok := loader.(ClipboardWriter); ok {
		deps.ClipboardWriter = actual
	}
	if actual, ok := loader.(ExternalEditor); ok {
		deps.ExternalEditor = actual
	}
	if actual, ok := loader.(LinkOpener); ok {
		deps.LinkOpener = actual
	}
	if actual, ok := loader.(ThemePresetStore); ok {
		deps.ThemePresetStore = actual
	}
	return deps
}

var (
	_ SessionQueries         = (*fakePullRequestDetailLoader)(nil)
	_ PullRequestListQueries = (*fakePullRequestDetailLoader)(nil)
	_ NotificationQueries    = (*fakePullRequestDetailLoader)(nil)
	_ DetailQueries          = (*fakePullRequestDetailLoader)(nil)
	_ PullRequestMutations   = (*fakePullRequestDetailLoader)(nil)
	_ ReviewMutations        = (*fakePullRequestDetailLoader)(nil)
	_ NotificationMutations  = (*fakePullRequestDetailLoader)(nil)
	_ ReactionMutations      = (*fakePullRequestDetailLoader)(nil)
	_ BuildQueries           = (*fakePullRequestDetailLoader)(nil)
)
