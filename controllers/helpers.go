package controllers

import (
	"encoding/json"
	"errors"
	"fmt"

	"forum/application"
	"forum/controllers/chat"
	"forum/errorhandle"
	"forum/model"
	"forum/wsmodel"

	"github.com/gorilla/websocket"
)

// All functions of this package send an error message to the websocket connection given in 'currConnection' parameter,
// so it is not necessary to send an error message using the function inside of another function in this package.

func errHelper(app *application.Application, currConnection *usersConnection, errMessage string, err error) error {
	errorhandle.WebSocketError(app, currConnection.Client, errMessage, err)
	return fmt.Errorf("%s: %w", errMessage, err)
}

func errCreateMessage(app *application.Application, currConnection *usersConnection, err error) error {
	return errHelper(app, currConnection, "creating message to websocket failed", err)
}

func errMarshalJSON(app *application.Application, currConnection *usersConnection, err error) error {
	return errHelper(app, currConnection, "Marshaling to JSON failed", err)
}

func badRequestHelper(app *application.Application, currConnection *usersConnection, requestMessage wsmodel.WSMessage, errMessage string) error {
	errorhandle.WebSocketBadRequest(app, currConnection.Client, requestMessage, errMessage)
	return errors.Join(wsmodel.ErrWarning, errors.New(errMessage))
}

/*
checks the session status, changes it if necessary. Returns nil only if the session has Loggedin status,
otherwise sends an error message to the websocket connection ('conn') and returns an error
*/
func checkLoggedStatus(app *application.Application, currConnection *usersConnection, requestMessage wsmodel.WSMessage) error {
	_, err := currConnection.session.Tidy(app)
	if err != nil {
		return errHelper(app, currConnection, "invalid session status", err)
	}
	if !currConnection.session.IsLoggedin() {
		return badRequestHelper(app, currConnection, requestMessage, "not logged in")
	}

	return nil
}

func sendReply(app *application.Application, currConnection *usersConnection, requestMessage wsmodel.WSMessage, data any) error {
	message, err := requestMessage.CreateMessageReply("success", data)
	if err != nil {
		return errCreateMessage(app, currConnection, err)
	}

	wsMessage, err := json.Marshal(message)
	if err != nil {
		return errMarshalJSON(app, currConnection, err)
	}

	currConnection.Client.WriteMessage(wsMessage)
	app.InfoLog.Printf("send message %s to channel of client %p", shortMessage(wsMessage), currConnection.Client)
	return nil
}

func sendSuccessMessage(app *application.Application, currConnection *usersConnection, messageType string, data any) error {
	message, err := wsmodel.CreateMessage(messageType, "success", data)
	if err != nil {
		return errCreateMessage(app, currConnection, err)
	}

	wsMessage, err := json.Marshal(message)
	if err != nil {
		return errMarshalJSON(app, currConnection, err)
	}

	currConnection.Client.WriteMessage(wsMessage)
	return nil
}

/*
gets a post from DB by its ID
*/
func getPost(app *application.Application, currConnection *usersConnection, postId int, message wsmodel.WSMessage) (*model.Post, error) {
	post, err := app.ForumData.GetPostByID(postId, currConnection.session.User.ID)
	if errors.Is(err, model.ErrNoRecord) {
		return nil, badRequestHelper(app, currConnection, message, fmt.Sprintf("cannot find a post with id '%d'", postId))
	}
	if err != nil {
		return nil, errHelper(app, currConnection, "get post from DB failed", err)
	}
	if post == nil {
		return nil, badRequestHelper(app, currConnection, message, fmt.Sprintf("cannot find a post with id '%d'", postId))
	}

	return post, nil
}

/*
gets 'postNumbers' posts from DB with ids less than 'beforeId'
*/
func getPosts(app *application.Application, currConnection *usersConnection, beforeId, postsNumber int) ([]*model.Post, error) {
	posts, err := app.ForumData.GetPosts(beforeId, postsNumber, &model.Filter{}, currConnection.session.User.ID)
	if err != nil {
		return nil, errHelper(app, currConnection, "getting posts from DB failed", err)
	}

	createPostsPreview(posts)
	return posts, nil
}

func (uc *usersConnection) createNewClientAndSendUserOnline(app *application.Application, conn *websocket.Conn, receivedMessages chan []byte, clientRegistered chan struct{})  error {
	uc.Client = chat.NewClient(app.Hub, uc.session.User,conn, receivedMessages, clientRegistered)
	if uc.session.IsLoggedin() {
		return sendOnlineUsers(app, uc)
	}
	return nil
}

func (uc *usersConnection) deleteClientAndSendUserOffline(app *application.Application, client *chat.Client) error {
	app.Hub.UnRegisterFromHub(client)
	if client.User != nil {
		return sendOfflineUserToUsers(app, uc, client.User)
	}
	return nil
}

func shortMessage(message []byte) []byte {
	if len(message) > 100 {
		shortMessage := make([]byte, 100)
		copy(shortMessage, message[:97])
		shortMessage[97] = '.'
		shortMessage[98] = '.'
		shortMessage[99] = '.'
		return append(shortMessage, []byte("...")...)
	}
	return message
}
