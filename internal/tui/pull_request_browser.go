package tui

import "fmt"

const pullRequestBrowserOpenSuccessMessage = "PR opened in browser"

func openPullRequestInBrowserCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "view", fmt.Sprintf("%d", number), "-R", repository, "--web")
}
