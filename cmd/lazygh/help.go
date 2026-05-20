package main

import "strings"

const cliName = "lazygh"
const shortHelpFlag = "-h"
const longHelpFlag = "--help"

func isHelpFlag(arg string) bool {
	return arg == shortHelpFlag || arg == longHelpFlag
}

func topLevelHelpOutput() string {
	return strings.Join([]string{
		"Usage:",
		"  " + cliName,
		"  " + cliName + " view <pull-request-url>",
		"  " + cliName + " review <pull-request-url>",
		"  " + cliName + " story-review <pull-request-url>",
		"",
		"Commands:",
		"  view          Open a pull request in the detail view",
		"  review        Start review mode for a pull request",
		"  story-review  Start story review mode for a pull request",
		"",
		"Options:",
		"  -h, --help   Show help",
		"  --version    Show version",
	}, "\n")
}

func subcommandHelpOutput(command string) string {
	switch command {
	case "review":
		return strings.Join([]string{
			"Usage:",
			"  " + cliName + " review <pull-request-url>",
			"",
			"Start review mode for a pull request URL.",
			"",
			"Options:",
			"  -h, --help  Show help",
		}, "\n")
	case "view":
		return strings.Join([]string{
			"Usage:",
			"  " + cliName + " view <pull-request-url>",
			"",
			"Open a pull request in the detail view.",
			"",
			"Options:",
			"  -h, --help  Show help",
		}, "\n")
	case "story-review":
		return strings.Join([]string{
			"Usage:",
			"  " + cliName + " story-review <pull-request-url>",
			"",
			"Start story review mode for a pull request URL.",
			"",
			"Options:",
			"  -h, --help  Show help",
		}, "\n")
	default:
		return topLevelHelpOutput()
	}
}
