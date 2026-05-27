package tui

import "strings"

func (program *Program) setBrowserDetailSectionCollapsed(sectionID string, collapsed bool) bool {
	if program == nil {
		return false
	}

	updatedStates, changed := browserCollapsedSectionStatesWithSectionCollapsed(program.browserCollapsedSectionStates, sectionID, collapsed)
	if !changed {
		return false
	}
	program.browserCollapsedSectionStates = updatedStates
	program.invalidatePullRequestDetailDocumentCache()
	return true
}

func (program *Program) setBrowserDetailSectionsCollapsed(sectionIDs []string, collapsed bool) bool {
	if program == nil {
		return false
	}

	updatedStates, changed := browserCollapsedSectionStatesWithAllSectionsCollapsed(program.browserCollapsedSectionStates, sectionIDs, collapsed)
	if !changed {
		return false
	}
	program.browserCollapsedSectionStates = updatedStates
	program.invalidatePullRequestDetailDocumentCache()
	return true
}

func browserCollapsedSectionStatesWithSectionCollapsed(current map[string]bool, sectionID string, collapsed bool) (map[string]bool, bool) {
	trimmedSectionID := strings.TrimSpace(sectionID)
	if trimmedSectionID == "" {
		return current, false
	}

	updatedStates := copyCollapsedSectionStates(current)
	if actualCollapsed, ok := updatedStates[trimmedSectionID]; ok && actualCollapsed == collapsed {
		return current, false
	}
	updatedStates[trimmedSectionID] = collapsed
	return updatedStates, true
}

func browserCollapsedSectionStatesWithAllSectionsCollapsed(current map[string]bool, sectionIDs []string, collapsed bool) (map[string]bool, bool) {
	if len(sectionIDs) == 0 {
		return current, false
	}

	updatedStates := copyCollapsedSectionStates(current)
	changed := false
	for _, sectionID := range sectionIDs {
		trimmedSectionID := strings.TrimSpace(sectionID)
		if trimmedSectionID == "" {
			continue
		}
		if actualCollapsed, ok := updatedStates[trimmedSectionID]; !ok || actualCollapsed != collapsed {
			changed = true
		}
		updatedStates[trimmedSectionID] = collapsed
	}
	if !changed {
		return current, false
	}
	return updatedStates, true
}

func copyCollapsedSectionStates(source map[string]bool) map[string]bool {
	copied := make(map[string]bool, len(source))
	for sectionID, collapsed := range source {
		copied[sectionID] = collapsed
	}
	return copied
}
