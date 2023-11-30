package controllers

import (
	"fmt"

	"forum/application"
	"forum/model"
	"forum/wsmodel"
	"forum/wsmodel/parse"
)

func createPostsPreview(posts []*model.Post) {
	for _, post := range posts {
		if len(post.Message.Content) < POST_PREVIEW_LENGTH {
			continue
		}
		for i := POST_PREVIEW_LENGTH - 10; i < len(post.Message.Content); i++ { // find first space after 440 char
			if string(post.Message.Content[i]) == " " {
				post.Message.Content = post.Message.Content[0:i] + "..."
				break
			}
		}
	}
}

func replyFullPostAndComments(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any, error) {

	postId, err := parse.PayloadToInt(message.Payload)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Invalid postID '%s'", message.Payload), err)
	}

	post, err := getPost(app, currConnection, postId, message) // TODO should it close the connection if a post with the given ID wasn't found?
	if err != nil {
		return nil, err
	}
	return post, nil
}
