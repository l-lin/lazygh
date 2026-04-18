package app

type Runner interface {
	Run() error
}

type App struct {
	runner Runner
}

func New(runner Runner) *App {
	return &App{runner: runner}
}

func (app *App) Run() error {
	if app.runner == nil {
		return nil
	}

	return app.runner.Run()
}
