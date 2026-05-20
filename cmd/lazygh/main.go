package main

import (
	"fmt"
	"io"
	"os"

	"github.com/l-lin/lazygh/internal/app"
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
	"github.com/l-lin/lazygh/internal/theme"
	"github.com/l-lin/lazygh/internal/tui"
)

type configurableRunner interface {
	ApplyKeymapOverrides(appconfig.KeymapOverrides)
	ApplyPullRequestSearches([]appconfig.PullRequestSearch)
	ApplyLinksConfig(appconfig.LinksConfig)
	ApplyStoryReviewConfig(story.Config)
	ApplyCacheConfig(appconfig.CacheConfig) error
	OpenReviewByURL(string) error
	OpenPullRequestByURL(string) error
	OpenStoryReviewByURL(string) error
	Run() error
}

func main() {
	if err := run(os.Args[1:], appconfig.LoadDefault, func() configurableRunner {
		return newRunner()
	}); err != nil {
		fmt.Fprintf(os.Stderr, "lazygh: %v\n", err)
		os.Exit(1)
	}
}

func newRunner() configurableRunner {
	return tui.NewProgramWithModelAndDeps(nil, newAppDepsWithRunner(nil))
}

// newAppDepsWithRunner is the single app composition root for provider ports:
// `internal/github` stays provider-neutral, while `internal/githubcli` stays the
// `gh` adapter that implements those ports for the TUI.
func newAppDepsWithRunner(runner githubcli.Runner) tui.AppDeps {
	notifications := githubcli.NewNotificationAdapterWithRunner(runner)
	return tui.AppDeps{
		SessionQueries:        githubcli.NewSessionAdapterWithRunner(runner),
		PullRequestList:       githubcli.NewPullRequestListAdapterWithRunner(runner),
		NotificationQueries:   notifications,
		DetailQueries:         githubcli.NewPullRequestDetailAdapterWithRunner(runner),
		PullRequestMutations:  githubcli.NewPullRequestMutationAdapterWithRunner(runner),
		ReviewMutations:       githubcli.NewReviewAdapterWithRunner(runner),
		NotificationMutations: notifications,
		ReactionMutations:     githubcli.NewReactionAdapterWithRunner(runner),
		BuildQueries:          githubcli.NewBuildAdapterWithRunner(runner),
		MarkdownHTMLRenderer:  githubcli.NewMarkdownServiceWithRunner(runner),
		AuthTokenProvider:     githubcli.NewAuthServiceWithRunner(runner),
	}
}

func run(args []string, loadConfig func() (appconfig.Config, error), newRunner func() configurableRunner) error {
	return runWithIO(args, os.Stdout, resolvedVersion(), loadConfig, newRunner)
}

func runWithIO(args []string, stdout io.Writer, version string, loadConfig func() (appconfig.Config, error), newRunner func() configurableRunner) error {
	startupCommand, actualErr := parseStartupCommand(args, version)
	if actualErr != nil {
		return actualErr
	}
	if startupCommand.exitAfterOutput {
		_, actualErr = fmt.Fprintln(stdout, startupCommand.output)
		return actualErr
	}

	configuration, actualErr := loadConfig()
	if actualErr != nil {
		return actualErr
	}

	theme.ApplyPalette(configuration.ResolvedTheme())

	resolvedCache, actualErr := configuration.ResolvedCache()
	if actualErr != nil {
		return actualErr
	}

	runner := newRunner()
	runner.ApplyKeymapOverrides(configuration.Keymaps)
	runner.ApplyPullRequestSearches(configuration.PullRequests)
	runner.ApplyLinksConfig(configuration.ResolvedLinks())
	runner.ApplyStoryReviewConfig(configuration.ResolvedStoryReview())
	if actualErr := runner.ApplyCacheConfig(resolvedCache); actualErr != nil {
		return actualErr
	}
	if startupCommand.reviewURL != "" {
		if actualErr := runner.OpenReviewByURL(startupCommand.reviewURL); actualErr != nil {
			return actualErr
		}
	}
	if startupCommand.viewURL != "" {
		if actualErr := runner.OpenPullRequestByURL(startupCommand.viewURL); actualErr != nil {
			return actualErr
		}
	}
	if startupCommand.storyReviewURL != "" {
		if actualErr := runner.OpenStoryReviewByURL(startupCommand.storyReviewURL); actualErr != nil {
			return actualErr
		}
	}
	return app.New(runner).Run()
}

const versionFlag = "--version"

type startupOptions struct {
	reviewURL       string
	viewURL         string
	storyReviewURL  string
	exitAfterOutput bool
	output          string
}

func parseStartupCommand(args []string, version string) (startupOptions, error) {
	if len(args) == 1 {
		switch args[0] {
		case versionFlag:
			return startupOptions{exitAfterOutput: true, output: formatVersionOutput(version)}, nil
		case shortHelpFlag, longHelpFlag:
			return startupOptions{exitAfterOutput: true, output: topLevelHelpOutput()}, nil
		}
	}
	return parseStartupOptions(args)
}

func parseStartupOptions(args []string) (startupOptions, error) {
	if len(args) == 0 {
		return startupOptions{}, nil
	}

	switch args[0] {
	case "review":
		if len(args) == 2 && isHelpFlag(args[1]) {
			return startupOptions{exitAfterOutput: true, output: subcommandHelpOutput("review")}, nil
		}
		if len(args) != 2 {
			return startupOptions{}, fmt.Errorf("review expects exactly one pull request URL")
		}
		return startupOptions{reviewURL: args[1]}, nil
	case "view":
		if len(args) == 2 && isHelpFlag(args[1]) {
			return startupOptions{exitAfterOutput: true, output: subcommandHelpOutput("view")}, nil
		}
		if len(args) != 2 {
			return startupOptions{}, fmt.Errorf("view expects exactly one pull request URL")
		}
		return startupOptions{viewURL: args[1]}, nil
	case "story-review":
		if len(args) == 2 && isHelpFlag(args[1]) {
			return startupOptions{exitAfterOutput: true, output: subcommandHelpOutput("story-review")}, nil
		}
		if len(args) != 2 {
			return startupOptions{}, fmt.Errorf("story-review expects exactly one pull request URL")
		}
		return startupOptions{storyReviewURL: args[1]}, nil
	default:
		return startupOptions{}, fmt.Errorf("unknown subcommand %q", args[0])
	}
}
