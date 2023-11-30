package parse

import (
	"encoding/json"

	"forum/wsmodel"
)

func PayloadToInt(payload json.RawMessage) (int, error) {
	var number int
	err := json.Unmarshal(payload, &number)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func PayloadToString(payload json.RawMessage) (string, error) {
	var str string
	err := json.Unmarshal(payload, &str)
	if err != nil {
		return "", err
	}

	return str, nil
}

func PayloadToUserCredential(payload json.RawMessage) (wsmodel.UserCredentials, error) {
	var uC wsmodel.UserCredentials
	err := json.Unmarshal(payload, &uC)
	return uC, err
}

func PayloadToPost(payload json.RawMessage) (wsmodel.Post, error) {
	var post wsmodel.Post
	err := json.Unmarshal(payload, &post)
	return post, err
}

func PayloadToComment(payload json.RawMessage) (wsmodel.Comment, error) {
	var comment wsmodel.Comment
	err := json.Unmarshal(payload, &comment)
	return comment, err
}

func PayloadToChatMessage(payload json.RawMessage) (wsmodel.ChatMessage, error) {
	var message wsmodel.ChatMessage
	err := json.Unmarshal(payload, &message)
	return message, err
}

func PayloadToReaction(payload json.RawMessage) (wsmodel.Reaction, error) {
	var react wsmodel.Reaction
	err := json.Unmarshal(payload, &react)
	return react, err
}
