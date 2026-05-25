package tui

import "testing"

func TestUpdate_GivenMsgPageNavigationRequested_WhenApplying_ThenItReturnsATypedPageNavigationCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgPageNavigationRequested{Kind: pageNavigationKindHalfDown})

	if len(actual) != 1 {
		t.Fatalf("expected one page-navigation command, actual %d", len(actual))
	}
	if _, ok := actual[0].(pageNavigationCmd); !ok {
		t.Fatalf("expected a pageNavigationCmd, actual %T", actual[0])
	}
}
