package tui

func (program *Program) applyConnectedUserLoadPlanned() {
	program.planConnectedUserLoad()
	program.setConnectedUserStatusOperationID(program.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: "Refreshing connected user", failureLabel: "Refreshing connected user"}))
}

func (program *Program) applyPullRequestsLoadPlanned(message MsgPullRequestsLoadPlanned) {
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		store, _ = store.withNextPullRequestLoadGeneration(message.Tab)
		return store.withPullRequestsLoadStarted(message.Tab, true).withPullRequestsLoading(message.Tab, true)
	})
	if message.ScheduledRefreshBatchID == 0 && !program.manualRefreshInProgress() {
		program.setPullRequestListStatusOperationID(message.Tab, program.startStatusLineOperation(statusLineRefreshPullRequestListOperation()))
	}
}

func (program *Program) applyNotificationsLoadPlanned() {
	program.planNotificationsLoad()
	if !program.manualRefreshInProgress() {
		program.setNotificationsStatusLineOperationID(program.startStatusLineOperation(statusLineRefreshNotificationsOperation()))
	}
}

func (program *Program) applyPullRequestDetailLoadPlanned(message MsgPullRequestDetailLoadPlanned) {
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withPullRequestDetailLoadPlanned(message.Key)
	})
	if message.ScheduledRefreshBatchID == 0 && !program.manualRefreshInProgress() {
		program.setPullRequestDetailStatusOperationID(message.Key, program.startStatusLineOperation(statusLineRefreshPullRequestKeyOperation(program, message.Key)))
	}
}

func (program *Program) applyPullRequestDiffLoadPlanned(message MsgPullRequestDiffLoadPlanned) {
	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withPullRequestDiffLoadPlanned(message.Key)
	})
	if message.ScheduledRefreshBatchID == 0 && !program.manualRefreshInProgress() {
		program.setPullRequestDiffStatusOperationID(message.Key, program.startStatusLineOperation(statusLineRefreshPullRequestKeyOperation(program, message.Key)))
	}
}

func (program *Program) applyIssueDetailLoadPlanned(message MsgIssueDetailLoadPlanned) {
	key := notificationDetailKey(message.Repository, message.Number)
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withIssueDetailLoadPlanned(key)
	})
	program.setIssueDetailStatusOperationID(key, program.startStatusLineOperation(statusLineRefreshIssueOperation(message.Repository, message.Number)))
}

func (program *Program) applyReleaseDetailLoadPlanned(message MsgReleaseDetailLoadPlanned) {
	key := notificationDetailKey(message.Repository, message.ID)
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withReleaseDetailLoadPlanned(key)
	})
	program.setReleaseDetailStatusOperationID(key, program.startStatusLineOperation(statusLineRefreshReleaseOperation(message.Repository, message.ID)))
}

func (program *Program) applyCurrentDetailImageHTMLLoadPlanned(message MsgCurrentDetailImageHTMLLoadPlanned) {
	program.markDetailImageHTMLLoadPlanned(message.SourceKey)
}

func (program *Program) applyCurrentDetailImageLoadPlanned(message MsgCurrentDetailImageLoadPlanned) {
	program.markDetailImageLoadPlanned(message.ImageURL)
}

func (program *Program) applyPullRequestDetailCacheHydrated(message MsgPullRequestDetailCacheHydrated) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}

	hydrated := false
	program.updateDetailStore(func(store detailStore) detailStore {
		if detailResultKnown(store.pullRequestDetailCache, key) {
			return store
		}
		hydrated = true
		return store.withPullRequestDetailCached(key, message.Result)
	})
	if !hydrated {
		return
	}
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) applyPullRequestDiffCacheHydrated(message MsgPullRequestDiffCacheHydrated) {
	key := pullRequestDetailKey(message.Summary.Repository, message.Summary.Number)
	if key == "" {
		return
	}

	hydrated := false
	program.updateReviewStore(func(store reviewStore) reviewStore {
		if diffResultKnown(store.pullRequestDiffCache, key) {
			return store
		}
		hydrated = true
		return store.withPullRequestDiffCached(key, message.Result)
	})
	if !hydrated {
		return
	}
	program.invalidateReviewDiffRenderCache()
	program.clampReviewSessionSelection()
}
