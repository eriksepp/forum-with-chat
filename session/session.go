package session

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"forum/application"
	"forum/model"

	"github.com/gofrs/uuid"
)

type LoginStatus byte

const (
	Loggedin LoginStatus = iota
	Experied
	Notloggedin
)

const (
	EXP_SESSION               = 24 * time.Hour
	TIME_BEFORE_AFTER_REFRESH = 30 * time.Second
)

const SESSION_TOKEN = "forum_session_id"

type Session struct {
	loginStatus LoginStatus 
	User        *model.User `json:"user"`
}

func (s *Session) isExpired() bool {
	exp := s.User.ExpirySession
	return exp.Before(time.Now())
}

func (s *Session) timeToExpired() time.Duration {
	exp := s.User.ExpirySession
	return time.Until(exp)
}

func (s *Session) IsLoggedin() bool {
	return s != nil && s.loginStatus == Loggedin && !s.isExpired()
}

/*
checks if a loggedin session is expired and change the session status
or if the session is expired then delete the session and change the session status to loggedout.
Returns the new status of the session.
*/
func (s *Session) Tidy(app *application.Application) (LoginStatus, error) {
	if s == nil {
		return Notloggedin, nil
	}
	switch s.loginStatus {
	case Loggedin:
		if s.User == nil {
			s.loginStatus = Notloggedin
			return Notloggedin, errors.New("exception for the session status: loggedIn status and nil User")
		}
		if s.isExpired() {
			s.loginStatus = Experied
			return Experied, nil
		}
	case Experied:
		if s.User == nil {
			s.loginStatus = Notloggedin
			return Notloggedin, errors.New("exception for the session status: expired status and nil User")
		}
		err := app.ForumData.DeleteUsersSession(s.User.Uuid)
		if err != nil {
			return Experied, fmt.Errorf("deleting the expired session failed: %w", err)
		}
		s.loginStatus = Notloggedin
		s.User = nil
		return Notloggedin, nil
	}
	return  s.loginStatus , nil
}

func (s *Session) GetStatus() string {
	var status string
	switch s.loginStatus {
	case Loggedin:
		status = "logged"
	case Experied:
		status = "experied"
	case Notloggedin:
		status = "not logged in"
	}
	return status
}

func GetNotloggedinSession() *Session {
	return &Session{Notloggedin, nil}
}

/*
returns session which contains status of login and uses's data if it's logged in.
If it is left lrss than 30 sec to expiried time, it will refresh the session
If an error occurs it will response to the client with error status and return the error
*/
func Get(app *application.Application, w http.ResponseWriter, r *http.Request) (*Session, error) {
	session := &Session{Notloggedin, nil}
	cook, err := r.Cookie(SESSION_TOKEN)
	if err != nil && err != http.ErrNoCookie {
		return nil, fmt.Errorf("getting cookie failed: '%s', url: '%s'", err, r.URL)
	}
	if err == http.ErrNoCookie || cook.Value == "" {
		return session, nil // session status = notloggedin
	}

	// there is a sessionToken
	sessionToken := cook.Value

	user, err := app.ForumData.GetUserByUUID(sessionToken)
	if err != nil {
		if err == model.ErrNoRecord {
			return session, nil
		}
		return nil, fmt.Errorf("getting a user by uuid failed: %w", err)
	}
	session.User = user

	if session.isExpired() {
		// delete the session & return expiried status
		session.User = nil
		err := app.ForumData.DeleteUsersSession(user.Uuid)
		if err != nil {
			return nil, fmt.Errorf("deleting the expired session failed: %w", err)
		}

		http.SetCookie(w, &http.Cookie{
			Name:    SESSION_TOKEN,
			Value:   "",
			Expires: time.Now(),
		})
		session.loginStatus = Experied
		return session, nil
	}

	if session.timeToExpired() < TIME_BEFORE_AFTER_REFRESH {
		// refresh the session
		session, err = New(app, w, user)
		if err != nil {
			return nil, fmt.Errorf("session creating failed: %w", err)
		}

		return session, nil
	}

	// user was found and their time was not expired or was renewed:
	session.loginStatus = Loggedin
	return session, nil
}

func New(app *application.Application, w http.ResponseWriter, user *model.User) (*Session, error) {
	expiresAt := time.Now().Add(EXP_SESSION)
	newSessionToken, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("UUID creating failed: %w", err)
	}

	err = app.ForumData.AddUsersSession(user.ID, newSessionToken.String(), expiresAt)
	if err != nil {
		return nil, fmt.Errorf("adding session failed: %w", err)
	}

	app.InfoLog.Printf("session tocken '%s' is added for the user ID '%d'", newSessionToken.String(), user.ID)

	http.SetCookie(w, &http.Cookie{
		Name:    SESSION_TOKEN,
		Value:   newSessionToken.String(),
		Expires: expiresAt,
	})

	user.Uuid = newSessionToken.String()
	user.ExpirySession = expiresAt

	return &Session{loginStatus: Loggedin, User: user}, nil
}
