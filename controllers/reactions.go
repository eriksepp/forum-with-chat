package controllers

import (
	"errors"
	"fmt"

	"forum/application"
	"forum/controllers/liker"
	"forum/model"
	"forum/wsmodel"
	"forum/wsmodel/parse"
)

// type reactionData struct {
// 	UserID      int64 `json:"userID"`
// 	PostID      int64 `json:"postID"`
// 	Reaction    bool  `json:"reaction"`    // True for like, false for dislike
// 	AddOrRemove bool  `json:"addOrRemove"` // True to add, false to remove
// 	IsPost      bool  `json:"ispost"`
// }

type Response struct {
	AmountOfReactions int `json:"reactionAmount"` // Likes or dislikes, corresponding on request
}

func replyReaction(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any ,error) {
	err := checkLoggedStatus(app, currConnection, message)
	if err != nil {
		return nil, err
	}

	reactionData, err := parse.PayloadToReaction(message.Payload)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Invalid payload in reaction request: %s", message.Payload), err)
	}

	errmessage := reactionData.Validate()
	if errmessage != "" {
		return nil, badRequestHelper(app, currConnection, message, errmessage)
	}

	var react liker.Liker
	switch reactionData.MessageType {
	case model.POST:
		react = liker.NewLikePost(currConnection.session.User, reactionData)
	case model.COMMENT:
		react = liker.NewLikePost(currConnection.session.User, reactionData)
	default:
		return nil, errHelper(app, currConnection, fmt.Sprintf("unexpected message type in reaction request %s", reactionData.MessageType), errors.New("unexpected message type"))
	}

	if err = liker.SetLike(app.ForumData, react, reactionData.Reaction); err != nil {
		return nil, errHelper(app, currConnection, "DB error during reaction handling", err)
	}

	// get the new number of likes/dislikes
	newReactions, err := react.GetLikesNumbers(app.ForumData)
	if err != nil {
		return nil, errHelper(app, currConnection, "get new reactions failed", err)
	}

	return newReactions, nil
}
