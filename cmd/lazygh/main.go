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
	ApplyPullRequestSearches([]appconfig.PullRequestSearch)
	OpenReviewByURL(string) error
	Run() error
}

func main() {
	if err := run(os.Args[1:], appconfig.LoadDefault, func() configurableRunner {
		return tui.NewProgram(githubcli.NewClient())
	}); err != nil {
		fmt.Fprintf(os.Stderr, "lazygh: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, loadConfig func() (appconfig.Config, error), newRunner func() configurableRunner) error {
	startupOptions, actualErr := parseStartupOptions(args)
	if actualErr != nil {
		return actualErr
	}

	configuration, actualErr := loadConfig()
	if actualErr != nil {
		return actualErr
	}

	runner := newRunner()
	runner.ApplyKeymapOverrides(configuration.Keymaps)
	runner.ApplyPullRequestSearches(configuration.PullRequests)
	if startupOptions.reviewURL != "" {
		if actualErr := runner.OpenReviewByURL(startupOptions.reviewURL); actualErr != nil {
			return actualErr
		}
	}
	return app.New(runner).Run()
}

type startupOptions struct {
	reviewURL string
}

func parseStartupOptions(args []string) (startupOptions, error) {
	if len(args) == 0 {
		return startupOptions{}, nil
	}
	if args[0] != "review" {
		return startupOptions{}, fmt.Errorf("unknown subcommand %q", args[0])
	}
	if len(args) != 2 {
		return startupOptions{}, fmt.Errorf("review expects exactly one pull request URL")
	}
	return startupOptions{reviewURL: args[1]}, nil
}
