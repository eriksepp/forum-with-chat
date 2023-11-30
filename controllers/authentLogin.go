package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"

	"forum/application"
	"forum/model"
	"forum/session"
	"forum/wsmodel"
	"forum/wsmodel/parse"

	"golang.org/x/crypto/bcrypt"
)

// Process the login form

func replyLogin(app *application.Application, w http.ResponseWriter, currConnection *usersConnection, message wsmodel.WSMessage) error {
	sessionStatus, err := currConnection.session.Tidy(app)
	if err != nil {
		return errHelper(app, currConnection, "invalid session status", err)
	}
	if sessionStatus == session.Loggedin {
		// return badRequestHelper(app, currConnection, message, "login is forbidden: a user has already logged in")
		return errHelper(app, currConnection, "login is forbidden", errors.New("a user has already logged in"))
	}

	userCredentials, err := parse.PayloadToUserCredential(message.Payload)
	if err != nil {
		return errHelper(app, currConnection, fmt.Sprintf("Invalid payload for user credential '%s'", message.Payload), err)
	}

	err = validateCredentials(app, currConnection, userCredentials, message)
	if err != nil {
		return err
	}

	user, err := getUserFromDB(app, currConnection, userCredentials, message)
	if err != nil {
		return err
	}

	newSess, err := session.New(app, w, user)
	if err != nil {
		return errHelper(app, currConnection, "session creation failed", err)
	}

	
	err = sendReply(app, currConnection, message, newSess)
	if err != nil {
		return err
	}
	
	app.InfoLog.Printf("User '%v' logged in", userCredentials.Username)
	
	currConnection.session = newSess
	return currConnection.renewClientForUser(app, newSess)
}

func validateCredentials(app *application.Application, currConnection *usersConnection, userCredentials wsmodel.UserCredentials, message wsmodel.WSMessage) error {
	if userCredentials.Username == "" || userCredentials.Username == "undefined" {
		return badRequestHelper(app, currConnection, message, "Username missing")
	}
	if userCredentials.Password == "" || userCredentials.Password == "undefined" {
		return badRequestHelper(app, currConnection, message, "Password missing")
	}
	return nil
}

func getUserFromDB(app *application.Application, currConnection *usersConnection, userCredentials wsmodel.UserCredentials, message wsmodel.WSMessage) (*model.User, error) {
	var user *model.User

	address, err := mail.ParseAddress(userCredentials.Username)
	if err == nil {
		user, err = app.ForumData.GetUserByEmail(address.Address)
	} else { // get the user by the name
		user, err = app.ForumData.GetUserByName(userCredentials.Username)
	}

	// Did the given user exist in DB?
	if err != nil {
		if err == model.ErrNoRecord {
			return nil, badRequestHelper(app, currConnection, message, fmt.Sprintf("User '%s' doesn't exist", userCredentials.Username))
		}
		return nil, errHelper(app, currConnection, fmt.Sprintf("get the user '%s' from DB failed", userCredentials.Username), err)
	}

	// Does the password match?
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userCredentials.Password)); err != nil {
		return nil, badRequestHelper(app, currConnection, message, "Wrong password")
	}

	return user, nil
}

func replyLogout(app *application.Application, w http.ResponseWriter, currConnection *usersConnection, message wsmodel.WSMessage) error {
	sessionStatus, err := currConnection.session.Tidy(app)
	if err != nil {
		return errHelper(app, currConnection, "invalid session status", err)
	}
	if sessionStatus == session.Notloggedin {
		// return badRequestHelper(app, currConnection, message, "logout is forbidden: no logged user")
		return errHelper(app, currConnection, "logout is forbidden", errors.New("no logged user"))
	}

	err = app.ForumData.DeleteUsersSession(currConnection.session.User.Uuid)
	if err != nil {
		return errHelper(app, currConnection, "session delete failed", err)
	}

	newSess := session.GetNotloggedinSession()

	err = sendReply(app, currConnection, message, newSess)
	if err != nil {
		return err
	}

	app.InfoLog.Printf("User '%s' logged out", currConnection.session.User.Name)

	currConnection.session = newSess
	return currConnection.renewClientForUser(app, newSess)
}
