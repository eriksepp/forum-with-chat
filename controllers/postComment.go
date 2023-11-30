package controllers

import (
	"fmt"

	"forum/application"
	"forum/wsmodel"
	"forum/wsmodel/parse"
)

// proccess the comment creation: adds the commnet to DB and replyes with the full post
func replyNewComment(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any, error) {
	

	comment, err := parse.PayloadToComment(message.Payload)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Invalid payload for a new comment: %s", message.Payload), err)
	}

	errmessage := comment.Validate()
	if errmessage != "" {
		return nil, badRequestHelper(app, currConnection, message, errmessage)
	}

	err = saveCommentToDB(app, currConnection, comment, currConnection.session.User.ID)
	if err != nil {
		return nil, err
	}

	post, err := getPost(app, currConnection, comment.PostID, message)
	if err != nil {
		return nil, err
	}
	
	return post, nil
}

func saveCommentToDB(app *application.Application, currConnection *usersConnection, comment wsmodel.Comment, authorID int) error {
	dateCreate := comment.Date

	id, err := app.ForumData.InsertComment(comment.PostID, comment.Content, nil, authorID, dateCreate)
	if err != nil {
		return errHelper(app, currConnection, "insert a new comment to DB failed", err)
	}

	app.InfoLog.Printf("A comment is added to DB. id: '%d'", id)
	return nil
}

// Delete the comment -not WS version
/*
func DeleteComment(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int
		var err error
		var comment *model.Comment
		var user *model.User

		// Only authenticate users can delete comments
		if sess := r.Context().Value(acl.SessionKey); !sess.(*session.Session).IsLoggedin() { // we might want to store user id instead
			errMsg := "Authentication needed"
			json.NewEncoder(w).Encode(map[string]string{"status": "failure", "error": errMsg})
			app.InfoLog.Printf("Failed to delete comment: %s", errMsg)
			return
		}

		// Check if comment id is provided
		if r.FormValue("id") == "" {
			errMsg := "Missing Comment Id"
			json.NewEncoder(w).Encode(map[string]string{"status": "failure", "error": errMsg})
			app.InfoLog.Printf("Failed to delete Comment: %s", errMsg)
			return
		}

		if id, err = strconv.Atoi(r.FormValue("id")); err != nil {
			errMsg := "Invalid comment Id"
			json.NewEncoder(w).Encode(map[string]string{"status": "failure", "error": errMsg})
			app.InfoLog.Printf("Failed to delete comment: %s", errMsg)
			return
		}

		// Check if we have to post
		if comment, err = app.ForumData.GetCommentByID(id); err != nil {
			errMsg := "No such comment in DB"
			json.NewEncoder(w).Encode(map[string]string{"status": "failure", "error": errMsg})
			app.InfoLog.Printf("Failed to delete comment: %s", errMsg)
			return
		}

		// Check if we are the author of this post
		if comment.Message.Author.ID != user.ID {
			errMsg := "User is not the author of this comment"
			json.NewEncoder(w).Encode(map[string]string{"status": "failure", "error": errMsg})
			app.InfoLog.Printf("Failed to delete comment: %s", errMsg)
			return
		}

		if err = app.ForumData.Delete(id); err != nil {
			errorhandle.ServerError(app, w, r, fmt.Sprintf("comment delete failed: func %s:", logger.GetCurrentFuncName()), err)
			return
		}

		if err = app.ForumData.DeleteCommentLikeByMessageID(id); err != nil {
			errorhandle.ServerError(app, w, r, fmt.Sprintf("comment delete failed: func %s:", logger.GetCurrentFuncName()), err)
			return
		}

		app.InfoLog.Printf("Post deleted successfully '%v'", comment.ID)
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "redirect": fmt.Sprintf("/post/%v", comment.PostID)})
	}
}
*/
