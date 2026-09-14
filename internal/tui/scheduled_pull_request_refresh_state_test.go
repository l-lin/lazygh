package tui

import (
	"errors"
	"testing"
	"time"
)

func TestScheduledPullRequestRefreshBatch_GivenPlannedWork_WhenAllWorkSettles_ThenItFinishesOnceAndAcknowledgesTheGeneration(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	triggeredAt := time.Date(2026, time.January, 1, 22, 8, 0, 0, time.UTC)
	batchID := subject.beginScheduledPullRequestRefreshBatch(7, triggeredAt)
	subject.addScheduledPullRequestRefreshWork(batchID, 2)

	if commands := subject.finishScheduledPullRequestRefreshPlanning(batchID); len(commands) != 0 {
		t.Fatalf("expected planning to wait for work, actual %v", commands)
	}
	if commands := subject.completeScheduledPullRequestRefreshWork(batchID, errors.New("detail failed")); len(commands) != 0 {
		t.Fatalf("expected the first completion to wait, actual %v", commands)
	}
	commands := subject.completeScheduledPullRequestRefreshWork(batchID, nil)
	if len(commands) != 1 {
		t.Fatalf("expected one final acknowledgement, actual %v", commands)
	}
	if _, ok := commands[0].(acknowledgePullRequestRefreshDueCmd); !ok {
		t.Fatalf("expected an acknowledgement command, actual %T", commands[0])
	}
	expectedFailure := iconStatusFailure + " Auto-refreshing PRs at " + triggeredAt.Local().Format("15:04") + ": detail failed"
	if actual := subject.statusLinePresenter().Text(); actual != expectedFailure {
		t.Fatalf("expected the first failure to be retained, actual %q", actual)
	}
}

func TestScheduledPullRequestRefreshBatch_GivenCancelledWork_WhenPlanningAndLateWorkSettle_ThenItAcknowledgesAfterTheTombstoneIsRemoved(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	batchID := subject.beginScheduledPullRequestRefreshBatch(8, time.Date(2026, time.January, 1, 22, 8, 0, 0, time.UTC))
	subject.addScheduledPullRequestRefreshWork(batchID, 1)
	subject.cancelScheduledPullRequestRefreshBatch(batchID)

	if subject.statusLineOperationStarted() {
		t.Fatal("expected cancellation to relinquish status-line ownership immediately")
	}
	if commands := subject.finishScheduledPullRequestRefreshPlanning(batchID); len(commands) != 0 {
		t.Fatalf("expected planning to wait for late work, actual %v", commands)
	}
	commands := subject.completeScheduledPullRequestRefreshWork(batchID, errors.New("late failure"))
	if len(commands) != 1 {
		t.Fatalf("expected one acknowledgement after cancelled work settled, actual %v", commands)
	}
	if _, ok := commands[0].(acknowledgePullRequestRefreshDueCmd); !ok {
		t.Fatalf("expected an acknowledgement command, actual %T", commands[0])
	}
	if _, exists := subject.scheduledPullRequestRefreshBatches[batchID]; exists {
		t.Fatal("expected the cancelled tombstone to be removed after its work settled")
	}
}

func TestScheduledPullRequestRefreshBatch_GivenPlannedWork_WhenANewerStatusOperationStarts_ThenItStillAcknowledgesTheRefreshGeneration(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	batchID := subject.beginScheduledPullRequestRefreshBatch(10, time.Date(2026, time.January, 1, 22, 8, 0, 0, time.UTC))
	subject.addScheduledPullRequestRefreshWork(batchID, 1)
	if commands := subject.finishScheduledPullRequestRefreshPlanning(batchID); len(commands) != 0 {
		t.Fatalf("expected planning to wait for work, actual %v", commands)
	}

	subject.startStatusLineOperation(statusLineRefreshNotificationsOperation())
	commands := subject.completeScheduledPullRequestRefreshWork(batchID, nil)
	if len(commands) != 1 {
		t.Fatalf("expected a newer operation not to strand the scheduler, actual %v", commands)
	}
	if _, ok := commands[0].(acknowledgePullRequestRefreshDueCmd); !ok {
		t.Fatalf("expected an acknowledgement command, actual %T", commands[0])
	}
}

func TestScheduledPullRequestRefreshBatch_GivenNoNetworkWork_WhenPlanningFinishes_ThenItCompletesSuccessfully(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	triggeredAt := time.Date(2026, time.January, 1, 22, 8, 0, 0, time.UTC)
	batchID := subject.beginScheduledPullRequestRefreshBatch(9, triggeredAt)

	commands := subject.finishScheduledPullRequestRefreshPlanning(batchID)
	if len(commands) != 1 {
		t.Fatalf("expected an acknowledgement for an empty batch, actual %v", commands)
	}
	expectedSuccess := iconStatusSuccess + " Auto-refreshing PRs at " + triggeredAt.Local().Format("15:04")
	if actual := subject.statusLinePresenter().Text(); actual != expectedSuccess {
		t.Fatalf("expected retained batch success, actual %q", actual)
	}
}
