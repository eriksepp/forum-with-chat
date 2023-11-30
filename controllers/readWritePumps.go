// TODO handle (add handling) error of writing to Conn
// TODO rename this file
package controllers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"forum/application"
	"forum/controllers/chat"
	"forum/wsmodel"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var NewLine = []byte("\n")

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (uc *usersConnection) ReadPump(app *application.Application, w http.ResponseWriter) {
	defer func() {
		// (online changes)app.Hub.UnRegisterFromHub(uc.Client)
		err := uc.deleteClientAndSendUserOffline(app, uc.Client)
		if err != nil {
			app.ErrLog.Printf("ReadPump: error client delete: %v", err)
		}

		close(uc.Client.ReceivedMessages)
		err = uc.Client.Conn.Close()
		if err != nil {
			app.ErrLog.Printf("ReadPump: error closing connection %p: %v", uc.Client.Conn, err)
		}
		app.InfoLog.Printf("ReadPump closed connection %p", uc.Client.Conn)
	}()

	uc.Client.Conn.SetReadLimit(maxMessageSize)
	uc.Client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	uc.Client.Conn.SetPongHandler(func(string) error { uc.Client.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil }) // TODO what if the ping has not sent in the pingPeriod cause of sending not control messages during that period?
	for {
		var message wsmodel.WSMessage
		err := uc.Client.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				app.ErrLog.Printf("websocket connection %p to '%s' was unexpected closed: %#v", uc.Client.Conn, uc.Client.Conn.LocalAddr(), err)
			}
			app.InfoLog.Printf("ReadPump is closing connection %p  of client  '%s' : %#v", uc.Client.Conn, uc.Client, err)
			break
		}
		if message.IsAuthentification() {
			// (online changes) oldUser := uc.Client.User
			err = replierAuthenticators[message.Type](app, w, uc, message)
			if err != nil {
				if errors.Is(err, wsmodel.ErrWarning) {
					continue
				} else {
					break
				}
			}
			// (online changes)
			// if uc.session.IsLoggedin() {
			// 	err = sendOnlineUsers(app, uc)
			// 	if err != nil && !errors.Is(err, wsmodel.ErrWarning) {
			// 		break
			// 	}
			// } else {
			// 	err = sendOfflineUserToUsers(app, uc, oldUser)
			// 	if err != nil && !errors.Is(err, wsmodel.ErrWarning) {
			// 		break
			// 	}
			// }

		} else {
			replier, ok := repliers[message.Type]
			if !ok {
				app.ErrLog.Printf("unknown type message received: %s", message.Type)
				continue
			}

			err := replier(app, uc, message)
			if err != nil && !errors.Is(err, wsmodel.ErrWarning) {
				break
			}

		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (uc *usersConnection) WritePump(app *application.Application) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := uc.Client.Conn.Close()
		if err != nil {
			app.ErrLog.Printf("WritePump: error closing connection: %v", err)
		}
		app.InfoLog.Printf("WritePump closed connection %p", uc.Client.Conn)
	}()
	for {
		chann := uc.Client.ReceivedMessages
		select {
		case message, ok := <-chann:
			if !ok {
				// The hub closed the channel.
				uc.Client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				app.InfoLog.Printf("WritePump is closing connection because the hub closed the channel %p ", chann)
				app.InfoLog.Printf("WritePump is closing connection %p of client '%s' because the hub closed the channel %p ", uc.Client.Conn, uc.Client, uc.Client.ReceivedMessages)
				return
			}

			// currentChannel := uc.Client.ReceivedMessages
			uc.Client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := uc.Client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				app.ErrLog.Printf("cannot create the NextWriter on the connenction %p : %v", uc.Client.Conn, err)
				return
			}
			writeMessage(app, w, message, uc.Client)

			// Add queued chat messages to the current websocket message.
			n := len(chann)
			for i := 0; i < n; i++ {
				message = <-chann
				writeMessage(app, w, message, uc.Client)
			}

			if err := w.Close(); err != nil {
				app.ErrLog.Printf("cannot close the writer on the connenction %p : %v", uc.Client.Conn, err)
				return
			}
		case <-ticker.C:
			uc.Client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := uc.Client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				app.ErrLog.Printf("ping the connenction %s failed: %v", uc.Client.Conn.LocalAddr(), err)
				return
			}
		}
	}
}

func writeMessage(app *application.Application, w io.WriteCloser, message []byte, currentClient *chat.Client) error {
	_, err := w.Write(message)
	if err != nil {
		return err
	}
	app.InfoLog.Printf("Websocket: send message: '%s' to client %s", shortMessage(message), currentClient)
	_, err = w.Write(NewLine)
	if err != nil {
		return err
	}
	return nil
}
