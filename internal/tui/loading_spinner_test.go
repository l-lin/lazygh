package tui

import "testing"

func TestTickLoadingSpinner_GivenNoLoadingWork_WhenTicking_ThenNoMessageIsDispatched(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.publishLoadingSpinnerAnimating()

	var actual []Msg
	subject.tickLoadingSpinner(func(msg Msg) { actual = append(actual, msg) })

	if len(actual) != 0 {
		t.Fatalf("expected no spinner tick dispatched while idle, actual %v", actual)
	}
}

func TestTickLoadingSpinner_GivenLoadingWork_WhenTicking_ThenTickIsDispatched(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.notificationsLoading = true
	subject.publishLoadingSpinnerAnimating()

	var actual []Msg
	subject.tickLoadingSpinner(func(msg Msg) { actual = append(actual, msg) })

	expected := []Msg{MsgLoadingSpinnerTick{}}
	if len(actual) != len(expected) || actual[0] != expected[0] {
		t.Fatalf("expected spinner tick %v dispatched while loading, actual %v", expected, actual)
	}
}

func TestAfterStateChange_GivenLoadingWorkChanges_WhenStateChanges_ThenSpinnerTickingFollowsIt(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.notificationsLoading = true

	then_noError(t, subject.afterStateChange(nil))

	if actual := subject.loadingSpinnerAnimating.Load(); !actual {
		t.Fatalf("expected spinner ticking enabled while loading, actual %v", actual)
	}

	subject.notificationsLoading = false
	then_noError(t, subject.afterStateChange(nil))

	if actual := subject.loadingSpinnerAnimating.Load(); actual {
		t.Fatalf("expected spinner ticking disabled once loading finished, actual %v", actual)
	}
}
