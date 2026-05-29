package tui

import (
	"fmt"
	"strings"
)

func (coordinator optimisticMutationCoordinator) nextOptimisticMutationID(kind string) (optimisticMutationCoordinator, string) {
	coordinator.optimisticMutationSequence++
	return coordinator, fmt.Sprintf("%s%s:%d", optimisticPullRequestMutationIDPrefix, strings.TrimSpace(kind), coordinator.optimisticMutationSequence)
}
