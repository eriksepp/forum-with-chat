// TODO if adding filters, change name="author" to "authorID" in home.tmpl
// TODO if the server sends an error message "not logged in" the connection will not be closed, so JS can ask the user to login
// TODO simplify the templates, cause javascript now rules the view
// TODO encode the password in the message from front end to base64 (or the all message content)
// TODO server must send OnlineUserUpdate message after login/out any of the users (keep connections in a  map, don't use sessions, just current connections)
// TODO delete fmt.Print...
package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"forum/application"
	"forum/errorhandle"
	"forum/model"
	"forum/route/middleware/acl"
	"forum/session"
	"forum/wsmodel"

	"github.com/gorilla/websocket"
)

type viewVarsMap map[string]any

// Home/Index page handler/controller
func Index(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var categories []*model.Category
		viewVars := make(viewVarsMap, 5)

		if r.URL.Path != "/" {
			errorhandle.NotFound(app, w, r)
			return
		}

		sess, userID := checkAuthenticated(r, &viewVars)

		// get all categories that have posts
		allCategories, err := getCategories(app, &viewVars)
		if err != nil {
			errorhandle.ServerError(app, w, r, "getting categories failed", err)
			return
		}

		filter, err := getFilters(r, &viewVars, allCategories, sess.User)
		if err != nil {
			errorhandle.ClientError(app, w, r, http.StatusBadRequest, fmt.Sprintf("getting filter failed: %v", err))
			return
		}

		posts, err := app.ForumData.GetPosts(0, 0, filter, userID)
		if err != nil {
			errorhandle.ServerError(app, w, r, "getting data from DB failed", err)
			return
		}

		createPostsPreview(posts)

		// set the data to the view
		viewVars["Posts"] = posts

		// render the view
		if err = app.View.Execute(w, viewVars); err != nil {
			errorhandle.ServerError(app, w, r, "template executing failed", err)
			return
		}
	}
}

// IndexWs handles websocket requests from the main page. URL: /ws
func IndexWs(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		currentConnection := &usersConnection{}

		currentConnection.session, err = session.Get(app, w, r)
		if err != nil {
			errorhandle.ServerError(app, w, r, "getting a session failed", err)
			return
		}

		// TODO Erik: Delete later? Added it here to use Browsersync for easier frontend development
		app.Upgrader.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			fmt.Println(app.Server.Addr)
			return origin == "http://www."+app.Server.Addr || origin == "http://"+app.Server.Addr || origin == "http://localhost:3000"
		}

		conn, err := app.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			errorhandle.ServerError(app, w, r, "Upgrade failed:", err)
			return
		}
		app.InfoLog.Printf("connection %p to '%s' is upgraded to the WebSocket protocol", conn, r.URL.Path)
		// the connection will be closed in WritePump or ReadPump functions

		// (online changes)currentConnection.Client = chat.NewClient(app.Hub, currentConnection.session.User, conn, nil, nil)
		err=currentConnection.createNewClientAndSendUserOnline(app, conn, nil, nil)
		if err != nil && !errors.Is(err, wsmodel.ErrWarning) {
			logErrorAndCloseConn(app, conn, "send online users list failed", err)
			return
		}
		app.InfoLog.Printf("registered new client: %s", currentConnection.Client)

		go currentConnection.WritePump(app)

		err = sendSession(app, currentConnection)
		if err != nil && !errors.Is(err, wsmodel.ErrWarning) {
			logErrorAndCloseConn(app, conn, "send session failed", err)
			return
		}

		// (online changes)
		// if currentConnection.session.IsLoggedin() {
		// 	err = sendOnlineUsers(app, currentConnection)
		// 	if err != nil && !errors.Is(err, wsmodel.ErrWarning) {
		// 		logErrorAndCloseConn(app, conn, "send online users list failed", err)
		// 		return
		// 	}
		// }

		go currentConnection.ReadPump(app, w)
	}
}

func checkAuthenticated(r *http.Request, viewVars *viewVarsMap) (sess *session.Session, userID int) {
	userID = 0
	sess = r.Context().Value(acl.SessionKey).(*session.Session)
	(*viewVars)["Session"] = sess
	if sess.IsLoggedin() {
		userID = sess.User.ID
	}
	return
}

func getCategories(app *application.Application, viewVars *viewVarsMap) (allCategories []*model.Category, err error) {
	allCategories, err = app.ForumData.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("getting data (set of categories) from DB failed: %w", err)
	}
	(*viewVars)["AllCategories"] = allCategories
	return
}

func getFilters(r *http.Request, viewVars *viewVarsMap, allCategories []*model.Category, user *model.User) (filter *model.Filter, err error) {
	// get category filters
	uQ := r.URL.Query()
	var categoryID []int
	if len(uQ[F_CATEGORIESID]) > 0 {
		for _, c := range uQ[F_CATEGORIESID] {
			id, err := strconv.Atoi(c)
			if err != nil || id <= 0 {
				return nil, fmt.Errorf("wrong category id in the filter request: '%s', err: %s", c, err)
			}
			categoryID = append(categoryID, id)
		}
	}

	filter = &model.Filter{
		CategoryID: categoryID,
	}

	// get author's filters
	if user != nil {
		if uQ.Get(F_AUTHORID) != "" {
			filter.AuthorID = user.ID
		}
		if uQ.Get(F_LIKEBY) != "" {
			filter.LikedByUserID = user.ID
		}
		if uQ.Get(F_DISLIKEBY) != "" {
			filter.DisLikedByUserID = user.ID
		}

	}
	return
}

func sendSession(app *application.Application, currConn *usersConnection) error {
	return sendSuccessMessage(app, currConn, wsmodel.CurrentSession, currConn.session)
}

func logErrorAndCloseConn(app *application.Application, conn *websocket.Conn, errMessage string, err error) {
	app.ErrLog.Printf("%s: %v", errMessage, err)

	err = conn.Close()
	if err != nil {
		app.ErrLog.Printf("error closing connection: %v", err)
	}
}
