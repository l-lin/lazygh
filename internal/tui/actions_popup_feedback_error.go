package tui

type actionsPopupStatusLineError struct {
	err            error
	feedbackTarget Focus
}

func newActionsPopupStatusLineError(feedbackTarget Focus, err error) error {
	if err == nil {
		return nil
	}
	return actionsPopupStatusLineError{err: err, feedbackTarget: feedbackTarget}
}

func (err actionsPopupStatusLineError) Error() string {
	if err.err == nil {
		return ""
	}
	return err.err.Error()
}

func (err actionsPopupStatusLineError) Unwrap() error {
	return err.err
}
