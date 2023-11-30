package errorhandle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"forum/application"
	"forum/controllers/chat"
	"forum/view"
	"forum/wsmodel"
)

// Opens a beautiful HTML 404 web page instead of the status 404 "Page not found"
func NotFound(app *application.Application, w http.ResponseWriter, r *http.Request) {
	app.ErrLog.Output(2, fmt.Sprintf("wrong path: %s", r.URL.Path))

	w.WriteHeader(http.StatusNotFound) // Sets status code at 404
	if err := view.ExecuteError(w, r, http.StatusNotFound); err != nil {
		app.ErrLog.Output(2, fmt.Sprintf("Execute NotFound page failed: %v", err))
		http.NotFound(w, r)
	}
}

func ServerError(app *application.Application, w http.ResponseWriter, r *http.Request, message string, err error) {
	app.ErrLog.Output(2, fmt.Sprintf("fail handling the page %s: %s. ERR: %v\nDebug Stack:  %s", r.URL.Path, message, err, debug.Stack()))
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func ClientError(app *application.Application, w http.ResponseWriter, r *http.Request, errStatus int, logTexterr string) {
	app.ErrLog.Output(2, logTexterr)
	http.Error(w, "ERROR: "+http.StatusText(errStatus), errStatus)
}

func MethodNotAllowed(app *application.Application, w http.ResponseWriter, r *http.Request, allowedMethods ...string) {
	if allowedMethods == nil {
		panic("no methods is given to func MethodNotAllowed")
	}
	allowdeString := allowedMethods[0]
	for i := 1; i < len(allowedMethods); i++ {
		allowdeString += ", " + allowedMethods[i]
	}

	w.Header().Set("Allow", allowdeString)
	ClientError(app, w, r, http.StatusMethodNotAllowed, fmt.Sprintf("using the method %s to go to a page %s", r.Method, r.URL))
}

func Forbidden(app *application.Application, w http.ResponseWriter, r *http.Request) {
	app.ErrLog.Output(2, fmt.Sprintf("access was forbidden: %s", r.URL.Path))

	w.WriteHeader(http.StatusForbidden) // Sets status code at 403
	if err := view.ExecuteError(w, r, http.StatusForbidden); err != nil {
		app.ErrLog.Output(2, fmt.Sprintf("Execute Forbidden page failed: %v", err))
		http.Error(w, fmt.Sprintf("ERROR: %s. ", http.StatusText(http.StatusForbidden)), http.StatusForbidden)
	}
}

// WebSocketError sends to the front-end side websocket connection `conn` a message of  type 'ERROR'  with the Payload= `errmessage`. It also logs the `errmessage` and `err` to the app.ErrLog logger.
func WebSocketError(app *application.Application, client *chat.Client, errmessage string, err error) {
	app.ErrLog.Output(2, fmt.Sprintf("websocket:: ERROR: %s: %v\nDebug Stack:  %s", errmessage, err, debug.Stack()))

	message, err := wsmodel.CreateMessage(wsmodel.ERROR, "serverError", errmessage)
	if err != nil {
		app.ErrLog.Output(2, fmt.Sprintf("websocket:: can't serialize error message to JSON: %v\nDebug Stack:  %s", err, debug.Stack()))
		return
	}

	wsMessage, err := json.Marshal(message)
	if err != nil {
		errText := fmt.Sprintf("websocket:: can't serialize the message to JSON: %#v : %v", message, err)
		client.WriteMessage([]byte(errText))
		app.ErrLog.Output(2, fmt.Sprintf("%s\nDebug Stack:  %s", errText, debug.Stack()))
		return
	}
	client.WriteMessage(wsMessage)
}

// WebSocketBadRequest sends to the front-end side websocket connection `conn` a message of `messageType` with the result = "error" and Data= `messageText`. It also logs the messageText to the app.InfoLog logger.
func WebSocketBadRequest(app *application.Application, client *chat.Client, messageRequest wsmodel.WSMessage, messageText string) { // TODO Naming: WebSocketClientDataError  , WebSocketWarning
	app.InfoLog.Printf("websocket:: send reply '%s' to: '%s'\n", messageText, messageRequest.Type)

	message, err := messageRequest.CreateMessageReply("error", messageText)
	if err != nil {
		app.ErrLog.Output(2, fmt.Sprintf("websocket:: can't serialize error message to JSON: %v\nDebug Stack:  %s", err, debug.Stack()))
		return
	}

	wsMessage, err := json.Marshal(message)
	if err != nil {
		errText := fmt.Sprintf("websocket:: can't serialize the message to JSON: %#v : %v", message, err)
		client.WriteMessage([]byte(errText))
		app.ErrLog.Output(2, fmt.Sprintf("%s\nDebug Stack:  %s", errText, debug.Stack()))
		return
	}
	client.WriteMessage(wsMessage)
}
