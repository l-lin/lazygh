package app

import (
	"fmt"
	"io"
)

const bootstrapMessage = "lazygh is bootstrapped. TUI work starts in TODO 02."

type App struct {
	stdout io.Writer
}

func New(stdout io.Writer) *App {
	if stdout == nil {
		stdout = io.Discard
	}

	return &App{stdout: stdout}
}

func (app *App) Run() error {
	_, err := fmt.Fprintln(app.stdout, bootstrapMessage)
	return err
}
