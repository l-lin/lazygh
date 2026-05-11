package githubcli

import "strconv"

func (client *Client) MarkPullRequestReadyForReview(repository string, number int) error {
	return client.runPullRequestReady(repository, number, false)
}

func (client *Client) ConvertPullRequestToDraft(repository string, number int) error {
	return client.runPullRequestReady(repository, number, true)
}

func (client *Client) ClosePullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.runGH("gh pr close", "pr", "close", strconv.Itoa(number), "-R", trimmedRepository); err != nil {
		return err
	}

	return nil
}

func (client *Client) ReopenPullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.runGH("gh pr reopen", "pr", "reopen", strconv.Itoa(number), "-R", trimmedRepository); err != nil {
		return err
	}

	return nil
}

func (client *Client) runPullRequestReady(repository string, number int, undo bool) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	arguments := []string{"pr", "ready", strconv.Itoa(number), "-R", trimmedRepository}
	if undo {
		arguments = append(arguments, "--undo")
	}
	if _, err := client.runGH("gh pr ready", arguments...); err != nil {
		return err
	}

	return nil
}

func (client *Client) SquashMergePullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.runGH("gh pr merge", "pr", "merge", strconv.Itoa(number), "-R", trimmedRepository, "--squash"); err != nil {
		return err
	}

	return nil
}
