package main

import (
	"fmt"
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

// newAppDepsWithRunner is the single app composition root for provider ports.
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
	startupOptions, actualErr := parseStartupOptions(args)
	if actualErr != nil {
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
	if startupOptions.reviewURL != "" {
		if actualErr := runner.OpenReviewByURL(startupOptions.reviewURL); actualErr != nil {
			return actualErr
		}
	}
	if startupOptions.viewURL != "" {
		if actualErr := runner.OpenPullRequestByURL(startupOptions.viewURL); actualErr != nil {
			return actualErr
		}
	}
	if startupOptions.storyReviewURL != "" {
		if actualErr := runner.OpenStoryReviewByURL(startupOptions.storyReviewURL); actualErr != nil {
			return actualErr
		}
	}
	return app.New(runner).Run()
}

type startupOptions struct {
	reviewURL      string
	viewURL        string
	storyReviewURL string
}

func parseStartupOptions(args []string) (startupOptions, error) {
	if len(args) == 0 {
		return startupOptions{}, nil
	}

	switch args[0] {
	case "review":
		if len(args) != 2 {
			return startupOptions{}, fmt.Errorf("review expects exactly one pull request URL")
		}
		return startupOptions{reviewURL: args[1]}, nil
	case "view":
		if len(args) != 2 {
			return startupOptions{}, fmt.Errorf("view expects exactly one pull request URL")
		}
		return startupOptions{viewURL: args[1]}, nil
	case "story-review":
		if len(args) != 2 {
			return startupOptions{}, fmt.Errorf("story-review expects exactly one pull request URL")
		}
		return startupOptions{storyReviewURL: args[1]}, nil
	default:
		return startupOptions{}, fmt.Errorf("unknown subcommand %q", args[0])
	}
}
