package controllers

import (
	"errors"
	"fmt"

	"forum/application"
	"forum/model"
	"forum/wsmodel"
	"forum/wsmodel/parse"
)

// Proccess the post creation
func replyNewPost(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any ,error) {

	postData, err := parse.PayloadToPost(message.Payload)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Invalid payload for a new post: %s", message.Payload), err)
	}

	errmessage := postData.Validate()
	if errmessage != "" {
		return nil, badRequestHelper(app, currConnection, message, errmessage)
	}

	err = savePostToDB(app, currConnection, postData)
	if err != nil {
		return nil, err
	}

	posts, err := getPosts(app, currConnection, 0, POSTS_ON_POSTSVIEW)
	if err != nil {
		return nil, err
	}
	
	return posts, nil
}

func savePostToDB(app *application.Application, currConnection *usersConnection, postData wsmodel.Post) error {
	dateCreate := postData.Date

	// check if categories ids are valid (exist in DB)
	for _, id := range postData.CategoriesID {
		_, err := app.ForumData.GetCategoryByID(id)
		if err != nil {
			if errors.Is(err, model.ErrNoRecord) {
				return errHelper(app, currConnection, fmt.Sprintf("no cathegory with id: '%d' in DB", id), err)
			} else {
				return errHelper(app, currConnection, fmt.Sprintf("getting a category with id: '%d' from DB faild", id), err)
			}
		}
	}

	id, err := app.ForumData.InsertPost(postData.Theme, postData.Content, nil, currConnection.session.User.ID, dateCreate, postData.CategoriesID)
	if err != nil {
		return errHelper(app, currConnection, "insert a new post to DB failed", err)
	}

	app.InfoLog.Printf("Post is added to DB. id: '%d' categories: '%v' ", id, postData.CategoriesID)
	return nil
}
