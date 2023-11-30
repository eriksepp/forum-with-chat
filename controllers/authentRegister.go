package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"forum/application"
	"forum/model"
	"forum/session"
	"forum/wsmodel"
	"forum/wsmodel/parse"

	"golang.org/x/crypto/bcrypt"
)

// Process the register form
func replyRegister(app *application.Application, w http.ResponseWriter, currConnection *usersConnection, message wsmodel.WSMessage) error {
	sessionStatus, err := currConnection.session.Tidy(app)
	if err != nil {
		return errHelper(app, currConnection, "invalid session status", err)
	}
	if sessionStatus == session.Loggedin {
		// return badRequestHelper(app, currConnection, message, "register is forbidden: a user has already logged in")
		return errHelper(app, currConnection, "register is forbidden", errors.New("a user has already logged in"))
	}

	userCredentials, err := parse.PayloadToUserCredential(message.Payload)
	if err != nil {
		return errHelper(app, currConnection, fmt.Sprintf("Invalid payload for user credential '%s'", message.Payload), err)
	}

	errmessage := userCredentials.Validate()
	if errmessage != "" {
		return badRequestHelper(app, currConnection, message, errmessage)
	}

	user, err := saveUserToDB(app, currConnection, userCredentials, message)
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

	app.InfoLog.Printf("User '%s'--'%s' signed up and logged in", user.Name, user.Email)

	currConnection.session = newSess
	return currConnection.renewClientForUser(app, newSess)

}

func createUser(u wsmodel.UserCredentials) (*model.User, error) {
	user := model.User{
		Name:       u.Username,
		Email:      u.Email,
		DateCreate: time.Now(),
		Gender:     u.Gender,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
	}
	var err error

	user.Password, err = bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost) // hash the password
	if err != nil {
		return nil, fmt.Errorf("failed to generate crypto password: %w", err)
	}

	user.DateBirth, err = time.Parse(time.DateOnly, u.DateBirth) // date must by in the format  "2006-01-02"
	if err != nil {
		return nil, fmt.Errorf("failed to parse the date of birth: %w", err)
	}

	return &user, nil
}

func saveUserToDB(app *application.Application, currConnection *usersConnection, userCredentials wsmodel.UserCredentials, message wsmodel.WSMessage) (*model.User, error) {
	user, err := createUser(userCredentials)
	if err != nil {
		errmessage := fmt.Sprintf("failed to create a user: %v", err)
		return nil, badRequestHelper(app, currConnection, message, errmessage)
	}

	id, err := app.ForumData.AddUser(user)
	if err != nil {
		switch err {
		case model.ErrUniqueUserName:
			errmessage := fmt.Sprintf("Username '%s' is already taken", user.Name)
			return nil, badRequestHelper(app, currConnection, message, errmessage)
		case model.ErrUniqueUserEmail:
			errmessage := fmt.Sprintf("Account with email' %s' already exists", user.Email)
			return nil, badRequestHelper(app, currConnection, message, errmessage)
		default:
			return nil, errHelper(app, currConnection, fmt.Sprintf("add a new user '%s' to DB failed", userCredentials.Username), err)
		}
	}

	user.ID = id
	app.InfoLog.Printf("add the user '%s' to DB, email: '%s'", user.Name, user.Email)
	return user, nil
}
