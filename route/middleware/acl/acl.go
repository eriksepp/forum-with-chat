package acl

import (
	"context"
	"fmt"
	"net/http"

	"forum/application"
	"forum/errorhandle"
	"forum/logger"
	"forum/session"
)

type key string

const SessionKey key = "usersession"

// DisallowAuth does not allow authenticated users to access the page
func DisallowAuth(app *application.Application) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sess := r.Context().Value(SessionKey); sess.(*session.Session).IsLoggedin() {
				app.InfoLog.Printf("Unauthorized access for authenticated user")
				errorhandle.Forbidden(app, w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// DisallowAnon does not allow anonymous users to access the page
func DisallowAnon(app *application.Application) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sess := r.Context().Value(SessionKey); !sess.(*session.Session).IsLoggedin() {
				app.InfoLog.Printf("Unauthorized access for anonymous user")
				errorhandle.Forbidden(app, w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// Add user to the context if any
func AddUser(app *application.Application) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := session.Get(app, w, r)
			if err != nil {
				errorhandle.ServerError(app, w, r, fmt.Sprintf("func %s failed ", logger.GetCurrentFuncName()), err)
				return
			}

			app.InfoLog.Printf("session for user '%v' is added to the context", sess.User)
			ctx := context.WithValue(r.Context(), SessionKey, sess)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
