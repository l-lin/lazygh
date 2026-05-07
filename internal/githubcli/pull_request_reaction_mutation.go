package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const addReactionMutation = `mutation($subjectId:ID!,$content:ReactionContent!){addReaction(input:{subjectId:$subjectId,content:$content}){reaction{content}subject{id}}}`

var (
	ErrInvalidReactionTarget   = errors.New("invalid reaction target")
	ErrInvalidReactionContent  = errors.New("invalid reaction content")
	ErrInvalidReactionResponse = errors.New("invalid reaction response")
)

func (client *Client) AddReaction(subjectID string, content ReactionContent) error {
	trimmedSubjectID := strings.TrimSpace(subjectID)
	if trimmedSubjectID == "" {
		return ErrInvalidReactionTarget
	}

	reactionEnum, err := reactionContentGraphQLEnum(content)
	if err != nil {
		return err
	}

	result, err := client.runGH(
		"gh api graphql",
		"api",
		"graphql",
		"-f",
		"query="+addReactionMutation,
		"-f",
		"subjectId="+trimmedSubjectID,
		"-f",
		"content="+reactionEnum,
	)
	if err != nil {
		return err
	}

	return parseAddedReaction(result.Stdout)
}

func reactionContentGraphQLEnum(content ReactionContent) (string, error) {
	switch normalizeReactionContent(string(content)) {
	case ReactionContentThumbsUp:
		return "THUMBS_UP", nil
	case ReactionContentThumbsDown:
		return "THUMBS_DOWN", nil
	case ReactionContentLaugh:
		return "LAUGH", nil
	case ReactionContentHooray:
		return "HOORAY", nil
	case ReactionContentConfused:
		return "CONFUSED", nil
	case ReactionContentHeart:
		return "HEART", nil
	case ReactionContentRocket:
		return "ROCKET", nil
	case ReactionContentEyes:
		return "EYES", nil
	default:
		return "", ErrInvalidReactionContent
	}
}

type addReactionResponse struct {
	Data struct {
		AddReaction *struct {
			Reaction *struct {
				Content string `json:"content"`
			} `json:"reaction"`
			Subject *struct {
				ID string `json:"id"`
			} `json:"subject"`
		} `json:"addReaction"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func parseAddedReaction(stdout []byte) error {
	var response addReactionResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidReactionResponse, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}
	if response.Data.AddReaction == nil || response.Data.AddReaction.Subject == nil || response.Data.AddReaction.Reaction == nil {
		return ErrInvalidReactionResponse
	}
	if strings.TrimSpace(response.Data.AddReaction.Subject.ID) == "" {
		return ErrInvalidReactionResponse
	}
	if _, err := reactionContentGraphQLEnum(ReactionContent(response.Data.AddReaction.Reaction.Content)); err != nil {
		return ErrInvalidReactionResponse
	}
	return nil
}
