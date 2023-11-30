package chat

import (
	"fmt"

	"forum/model"

	"github.com/gorilla/websocket"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	User *model.User

	// The websocket connection.
	Conn *websocket.Conn
	// Buffered channel of received messages.
	ReceivedMessages chan []byte

	ClientRegistered chan struct{}
	// OnlineUsers      chan  MapID

	OpenedChatWith struct {
		ChatID     int
		ChatName   string
		UserClient *Client
	}
}

func NewClient(hub *Hub, user *model.User, conn *websocket.Conn, receivedMessages chan []byte, clientRegistered chan struct{}) *Client {
	var shortUser *model.User
	if user != nil {
		shortUser = &model.User{ID: user.ID, Name: user.Name}
	}
	client := &Client{
		User: shortUser,
		Conn: conn,
		// OnlineUsers:      make(chan  MapID),
	}

	if receivedMessages == nil {
		client.ReceivedMessages = make(chan []byte, 256)
	} else {
		client.ReceivedMessages = receivedMessages
	}

	if clientRegistered == nil {
		client.ClientRegistered = make(chan struct{})
	} else {
		client.ClientRegistered = clientRegistered
	}

	hub.RegisterToHub(client)
	// Wait for client registration to complete
	<-client.ClientRegistered
	return client
}



func (c *Client) WriteMessage(message []byte) {
	c.ReceivedMessages <- message
}

func (c *Client) String() string {
	return fmt.Sprintf("addr: %p :: User: '%s' | connection: %p | channels: clientRegistered %p  |  ReceivedMessages %p", c, c.User, c.Conn, c.ClientRegistered, c.ReceivedMessages)
}
