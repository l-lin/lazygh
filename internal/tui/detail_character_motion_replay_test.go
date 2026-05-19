package tui

import "testing"

func TestLayout_GivenDetailFocus_WhenRendering_ThenTheDetailPaneStaysReadOnly(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Editable {
		t.Fatal("expected the detail pane to stay read-only")
	}
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenRendering_ThenItStaysReadOnly(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Line 1"}))
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	if popupView.Editable {
		t.Fatal("expected the build popup to stay read-only")
	}
}
