package controllers

import (
	"encoding/json"
	"errors"
	"fmt"

	"forum/application"
	"forum/controllers/chat"
	"forum/model"
	"forum/wsmodel"
)

/*
sends list of online users
*/
func sendOnlineUsersToCurrentUser(app *application.Application, currConnection *usersConnection, onlineUsers chat.MapID) error {
	users, err := app.ForumData.GetFilteredUsersOrderedByMessagesToGivenUser(onlineUsers, currConnection.session.User.ID)
	if err != nil {
		return errHelper(app, currConnection, "get the users from DB failed", err)
	}

	return sendSuccessMessage(app, currConnection, wsmodel.OnlineUsers, users)
}

/*
sends current user to the given client's user
with the date of the last message sent from the current user to the client's user
*/
func sendNewOnlineUserToClient(app *application.Application, currConnection *usersConnection, recipient *chat.Client) error {
	user := wsmodel.UserWithMessageDate{ID: currConnection.session.User.ID, Name: currConnection.session.User.Name}

	date, err := app.ForumData.GetLastMessageDateFromUserToRecipient(currConnection.session.User.ID, recipient.User.ID)
	if err != nil && !errors.Is(err, model.ErrNoRecord){
		return errHelper(app, currConnection, fmt.Sprintf("failed get from DB the date of the last message from '%s' to '%s'", currConnection.session.User, recipient.User), err)
	}
	if err == nil {
		user.LastMessageDate = date
	}else{	//if errors.Is(err, model.ErrNoRecord) 
		user.LastMessageDate = "" // TODO or put zero date?
	}

	err = sendMessageToOtherClient(app, currConnection, recipient, wsmodel.NewOnlineUser, user)
	if err != nil {
		return err
	}

	return nil
}

/*
sends the list of online users to the current user
and the current user's online status to the other users
*/
func sendOnlineUsers(app *application.Application, currConnection *usersConnection) error {
	onlineUsers := app.Hub.GetOnlineUsers()

	err := sendOnlineUsersToCurrentUser(app, currConnection, onlineUsers)
	if err != nil && !errors.Is(err, wsmodel.ErrWarning) {
		return errHelper(app, currConnection, fmt.Sprintf("send list of online users to the user %s faild", currConnection.session.User), err)
	}

	// send the new online user to the other users
	var errs error

	for userID, client := range onlineUsers {
		if userID != currConnection.session.User.ID {
			err = sendNewOnlineUserToClient(app, currConnection, client)
			if err != nil {
				errs = errors.Join(errs, errHelper(app, currConnection, fmt.Sprintf("send new online user to the client %s faild\n", client), err))
			}
		}
	}
	if errs != nil {
		return errors.Join(wsmodel.ErrWarning, errs)
	}

	return nil
}

/*
sends online status of the user (userOff) to the all online users.
*/
func sendOfflineUserToUsers(app *application.Application, currConnection *usersConnection, userOff *model.User) error {
	onlineUsers := app.Hub.GetOnlineUsers()

	var errs error
	for _, client := range onlineUsers {
		err := sendMessageToOtherClient(app, currConnection, client, wsmodel.OfflineUser, userOff)
		if err != nil {
			errs = errors.Join(errs, errMarshalJSON(app, currConnection, err))
		}
	}
	if errs != nil {
		return errors.Join(wsmodel.ErrWarning, errs)
	}

	return nil
}

func sendMessageToOtherClient(app *application.Application, currConnection *usersConnection, recipient *chat.Client, messageType string, data any) error {
	message, err := wsmodel.CreateMessage(messageType, "success", data)
	if err != nil {
		return errCreateMessage(app, currConnection, err)
	}

	wsMessage, err := json.Marshal(message)
	if err != nil {
		return errMarshalJSON(app, currConnection, err)
	}

	recipient.WriteMessage(wsMessage)
	return nil
}
