package main

import (
	"fmt"
	"os"

	"codeberg.org/l-lin/lazygh/internal/app"
	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/tui"
)

type configurableRunner interface {
	ApplyKeymapOverrides(appconfig.KeymapOverrides)
	Run() error
}

func main() {
	if err := run(appconfig.LoadDefault, func() configurableRunner {
		return tui.NewProgram(githubcli.NewClient())
	}); err != nil {
		fmt.Fprintf(os.Stderr, "lazygh: %v\n", err)
		os.Exit(1)
	}
}

func run(loadConfig func() (appconfig.Config, error), newRunner func() configurableRunner) error {
	configuration, actualErr := loadConfig()
	if actualErr != nil {
		return actualErr
	}

	runner := newRunner()
	runner.ApplyKeymapOverrides(configuration.Keymaps)
	return app.New(runner).Run()
}
