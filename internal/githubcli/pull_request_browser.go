package githubcli

import "strconv"

func (client *Client) OpenPullRequestInBrowser(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.runGH("gh pr view", "pr", "view", strconv.Itoa(number), "-R", trimmedRepository, "--web"); err != nil {
		return err
	}

	return nil
}
