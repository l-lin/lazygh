package tui

type modalEditorStatusLineError struct {
	err            error
	feedbackTarget Focus
}

func newModalEditorStatusLineError(feedbackTarget Focus, err error) error {
	if err == nil {
		return nil
	}
	return modalEditorStatusLineError{err: err, feedbackTarget: feedbackTarget}
}

func (err modalEditorStatusLineError) Error() string {
	if err.err == nil {
		return ""
	}
	return err.err.Error()
}

func (err modalEditorStatusLineError) Unwrap() error {
	return err.err
}
