package tui

import "testing"

func TestUpdate_GivenMsgSideListViewportRequested_WhenApplying_ThenItReturnsATypedSideListViewportCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgSideListViewportRequested{Placement: viewportPlacementCenter})

	if len(actual) != 1 {
		t.Fatalf("expected one side-list viewport command, actual %d", len(actual))
	}
	if _, ok := actual[0].(sideListViewportCmd); !ok {
		t.Fatalf("expected a sideListViewportCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgDetailViewportRequested_WhenApplying_ThenItReturnsATypedDetailViewportCommand(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	actual := Update(subject, MsgDetailViewportRequested{Operation: detailViewportOperationRecenter})

	if len(actual) != 1 {
		t.Fatalf("expected one detail viewport command, actual %d", len(actual))
	}
	if _, ok := actual[0].(detailViewportCmd); !ok {
		t.Fatalf("expected a detailViewportCmd, actual %T", actual[0])
	}
}
