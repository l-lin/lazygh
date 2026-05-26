package tui

import "testing"

func TestReadOnlyScrollCommand_GivenHelpViewFullPageDown_WhenExecuting_ThenItScrollsTheReadOnlyView(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGuiWithSize(t, 80, 10)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.toggleHelp(gui, nil))
	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	lineCount := len(helpView.BufferLines())
	expected := clampInt(fullPageDelta(helpView.InnerHeight()), 0, maxInt(0, lineCount-helpView.InnerHeight()))

	readOnlyScrollCmd{View: helpView, FallbackName: viewHelpName, Kind: pageNavigationKindFullDown}.execute(subject, gui)

	then_viewOriginYIs(t, helpView, expected)
}
