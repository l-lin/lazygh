package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	addReactionMutation    = `mutation($subjectId:ID!,$content:ReactionContent!){addReaction(input:{subjectId:$subjectId,content:$content}){reaction{content}subject{id}}}`
	removeReactionMutation = `mutation($subjectId:ID!,$content:ReactionContent!){removeReaction(input:{subjectId:$subjectId,content:$content}){reaction{content}subject{id}}}`
)

var (
	ErrInvalidReactionTarget   = errors.New("invalid reaction target")
	ErrInvalidReactionContent  = errors.New("invalid reaction content")
	ErrInvalidReactionResponse = errors.New("invalid reaction response")
)

func (client *Client) AddReaction(subjectID string, content ReactionContent) error {
	return client.mutateReaction(addReactionMutation, subjectID, content, parseAddedReaction)
}

func (client *Client) RemoveReaction(subjectID string, content ReactionContent) error {
	return client.mutateReaction(removeReactionMutation, subjectID, content, parseRemovedReaction)
}

func (client *Client) mutateReaction(mutation string, subjectID string, content ReactionContent, parse func([]byte) error) error {
	trimmedSubjectID := strings.TrimSpace(subjectID)
	if trimmedSubjectID == "" {
		return ErrInvalidReactionTarget
	}

	reactionEnum, err := reactionContentGraphQLEnum(content)
	if err != nil {
		return err
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: mutation, Variables: []GraphQLVariable{literalGraphQLVariable("subjectId", trimmedSubjectID), literalGraphQLVariable("content", reactionEnum)}})
	if err != nil {
		return err
	}

	return parse(result.Stdout)
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

type reactionMutationPayload struct {
	Reaction *struct {
		Content string `json:"content"`
	} `json:"reaction"`
	Subject *struct {
		ID string `json:"id"`
	} `json:"subject"`
}

type reactionMutationResponse struct {
	Data struct {
		AddReaction    *reactionMutationPayload `json:"addReaction"`
		RemoveReaction *reactionMutationPayload `json:"removeReaction"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func parseAddedReaction(stdout []byte) error {
	return parseReactionMutation(stdout, func(response reactionMutationResponse) *reactionMutationPayload {
		return response.Data.AddReaction
	})
}

func parseRemovedReaction(stdout []byte) error {
	return parseReactionMutation(stdout, func(response reactionMutationResponse) *reactionMutationPayload {
		return response.Data.RemoveReaction
	})
}

func parseReactionMutation(stdout []byte, selectPayload func(reactionMutationResponse) *reactionMutationPayload) error {
	var response reactionMutationResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidReactionResponse, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}

	payload := selectPayload(response)
	if payload == nil || payload.Subject == nil || payload.Reaction == nil {
		return ErrInvalidReactionResponse
	}
	if strings.TrimSpace(payload.Subject.ID) == "" {
		return ErrInvalidReactionResponse
	}
	if _, err := reactionContentGraphQLEnum(ReactionContent(payload.Reaction.Content)); err != nil {
		return ErrInvalidReactionResponse
	}
	return nil
}
