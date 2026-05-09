package githubcli

import (
	"encoding/json"
	"strings"
)

type renderMarkdownHTMLRequest struct {
	Text    string `json:"text"`
	Mode    string `json:"mode"`
	Context string `json:"context,omitempty"`
}

func (client *Client) RenderMarkdownHTML(repository string, markdown string) (string, error) {
	requestBody, err := json.Marshal(renderMarkdownHTMLRequest{
		Text:    strings.TrimSpace(markdown),
		Mode:    "gfm",
		Context: strings.TrimSpace(repository),
	})
	if err != nil {
		return "", err
	}

	result, err := client.runGHWithInput("gh api markdown", requestBody, "api", "markdown", "--method", "POST", "--input", "-")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(result.Stdout)), nil
}

func (client *Client) GetAuthToken() (string, error) {
	result, err := client.runGH("gh auth token", "auth", "token")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(result.Stdout)), nil
}
