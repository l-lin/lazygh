package tui

func (program *Program) setBrowserDetailSectionCollapsed(sectionID string, collapsed bool) bool {
	if program == nil {
		return false
	}

	changed := false
	program.updateDetailStore(func(store detailStore) detailStore {
		updatedStore, storeChanged := store.withBrowserDetailSectionCollapsed(sectionID, collapsed)
		changed = storeChanged
		return updatedStore
	})
	if !changed {
		return false
	}

	program.invalidatePullRequestDetailDocumentCache()
	return true
}

func (program *Program) setBrowserDetailSectionsCollapsed(sectionIDs []string, collapsed bool) bool {
	if program == nil {
		return false
	}

	changed := false
	program.updateDetailStore(func(store detailStore) detailStore {
		updatedStore, storeChanged := store.withBrowserDetailSectionsCollapsed(sectionIDs, collapsed)
		changed = storeChanged
		return updatedStore
	})
	if !changed {
		return false
	}

	program.invalidatePullRequestDetailDocumentCache()
	return true
}
