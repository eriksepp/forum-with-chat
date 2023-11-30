package controllers

import (
	"fmt"

	"forum/application"
	"forum/wsmodel"
	"forum/wsmodel/parse"
)

func replyPosts(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any, error) {
	beforeID, err := parse.PayloadToInt(message.Payload)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Invalid payload for the portion of messages '%s'", message.Payload), err)
	}

	posts, err := getPosts(app, currConnection, beforeID, POSTS_ON_POSTSVIEW)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
