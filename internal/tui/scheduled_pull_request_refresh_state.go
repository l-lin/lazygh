package tui

import (
	"time"

	"github.com/jesseduffield/gocui"
)

type scheduledPullRequestRefreshBatch struct {
	generation        uint64
	statusOperationID uint64
	pending           int
	planning          bool
	cancelled         bool
	firstErr          error
}

func (program *Program) beginScheduledPullRequestRefreshBatch(generation uint64, triggeredAt time.Time) uint64 {
	if program == nil {
		return 0
	}
	if program.scheduledPullRequestRefreshBatches == nil {
		program.scheduledPullRequestRefreshBatches = map[uint64]scheduledPullRequestRefreshBatch{}
	}
	label := "Auto-refreshing PRs at " + triggeredAt.Local().Format("15:04")
	statusOperationID := program.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: label, failureLabel: label})
	if statusOperationID == 0 {
		return 0
	}
	program.scheduledPullRequestRefreshBatches[statusOperationID] = scheduledPullRequestRefreshBatch{
		generation:        generation,
		statusOperationID: statusOperationID,
		pending:           1,
		planning:          true,
	}
	return statusOperationID
}

func (program *Program) scheduledPullRequestRefreshBatchCancelled(batchID uint64) bool {
	if program == nil || batchID == 0 {
		return false
	}
	batch, ok := program.scheduledPullRequestRefreshBatches[batchID]
	return !ok || batch.cancelled
}

func (program *Program) addScheduledPullRequestRefreshWork(batchID uint64, count int) {
	if program == nil || count <= 0 {
		return
	}
	batch, ok := program.scheduledPullRequestRefreshBatches[batchID]
	if !ok || batch.cancelled {
		return
	}
	batch.pending += count
	program.scheduledPullRequestRefreshBatches[batchID] = batch
}

func (program *Program) completeScheduledPullRequestRefreshWork(batchID uint64, operationErr error) []Cmd {
	if program == nil || batchID == 0 {
		return nil
	}
	batch, ok := program.scheduledPullRequestRefreshBatches[batchID]
	if !ok {
		return nil
	}
	if operationErr != nil && batch.firstErr == nil {
		batch.firstErr = operationErr
	}
	if batch.pending > 0 {
		batch.pending--
	}
	if batch.pending > 0 || batch.planning {
		program.scheduledPullRequestRefreshBatches[batchID] = batch
		return nil
	}
	delete(program.scheduledPullRequestRefreshBatches, batchID)
	if batch.cancelled {
		return []Cmd{acknowledgePullRequestRefreshDueCmd{generation: batch.generation}}
	}
	program.finishStatusLineOperation(batch.statusOperationID, batch.firstErr)
	return []Cmd{acknowledgePullRequestRefreshDueCmd{generation: batch.generation}}
}

func (program *Program) finishScheduledPullRequestRefreshPlanning(batchID uint64) []Cmd {
	if program == nil || batchID == 0 {
		return nil
	}
	batch, ok := program.scheduledPullRequestRefreshBatches[batchID]
	if !ok || !batch.planning {
		return nil
	}
	batch.planning = false
	if batch.pending > 0 {
		batch.pending--
	}
	if batch.pending > 0 {
		program.scheduledPullRequestRefreshBatches[batchID] = batch
		return nil
	}
	delete(program.scheduledPullRequestRefreshBatches, batchID)
	if batch.cancelled {
		return []Cmd{acknowledgePullRequestRefreshDueCmd{generation: batch.generation}}
	}
	program.finishStatusLineOperation(batch.statusOperationID, batch.firstErr)
	return []Cmd{acknowledgePullRequestRefreshDueCmd{generation: batch.generation}}
}

func (program *Program) cancelScheduledPullRequestRefreshBatch(batchID uint64) {
	if program == nil || batchID == 0 {
		return
	}
	batch, ok := program.scheduledPullRequestRefreshBatches[batchID]
	if !ok {
		return
	}
	batch.cancelled = true
	// Keep the planning slot until the queued planning command settles; otherwise a cancellation before work starts can strand the scheduler.
	program.scheduledPullRequestRefreshBatches[batchID] = batch
	program.cancelStatusLineOperation(batch.statusOperationID)
}

func (program *Program) cancelScheduledPullRequestRefreshBatches() {
	if program == nil {
		return
	}
	batchIDs := make([]uint64, 0, len(program.scheduledPullRequestRefreshBatches))
	for batchID := range program.scheduledPullRequestRefreshBatches {
		batchIDs = append(batchIDs, batchID)
	}
	for _, batchID := range batchIDs {
		program.cancelScheduledPullRequestRefreshBatch(batchID)
	}
}

type finishScheduledPullRequestRefreshPlanningCmd struct {
	batchID uint64
}

func (command finishScheduledPullRequestRefreshPlanningCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}
	program.executeCmds(gui, program.finishScheduledPullRequestRefreshPlanning(command.batchID))
}

func (program *Program) pullRequestSearchAutoRefreshEnabled(tab PullRequestTab) bool {
	search, ok := program.searchBackedPullRequestSearch(tab)
	return ok && search.AutoRefresh
}
