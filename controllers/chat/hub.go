package chat

import (
	"sync"
)

type SafeClientsMap struct {
	sync.RWMutex
	items map[*Client]bool
}

func NewSafeMap() *SafeClientsMap {
	sm := &SafeClientsMap{}
	sm.items = make(map[*Client]bool)
	return sm
}

func (sm *SafeClientsMap) Set(key *Client, value bool) {
	sm.Lock()
	defer sm.Unlock()
	sm.items[key] = value
}

func (sm *SafeClientsMap) Get(key *Client) (bool, bool) {
	sm.RLock()
	defer sm.RUnlock()
	value, ok := sm.items[key]
	return value, ok
}

func (sm *SafeClientsMap) Delete(key *Client) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.items, key)
}

func (sm *SafeClientsMap) RRange(act func(key *Client, value bool)) {
	sm.RLock()
	defer sm.RUnlock()
	for key, value := range sm.items {
		act(key, value)
	}
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered Clients.
	Clients *SafeClientsMap

	// Inbound messages from a client.
	messageForAll chan []byte

	// register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// OnlineUsersRequest chan *Client
}

func NewHub() *Hub {
	return &Hub{
		messageForAll: make(chan []byte),
		Clients:       NewSafeMap(),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		// OnlineUsersRequest: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Clients.Set(client, true)
			client.ClientRegistered <- struct{}{}
		case client := <-h.unregister:
			if _, ok := h.Clients.Get(client); ok {
				h.Clients.Delete(client)
			}
		case message := <-h.messageForAll:
			h.Clients.Lock()
			for client := range h.Clients.items {
				select {
				case client.ReceivedMessages <- message:
				default:
					// If the client's send buffer is full, then the hub assumes that the client is dead or stuck.
					// In this case, the hub unregisters the client.
					close(client.ReceivedMessages)
					delete(h.Clients.items, client)
				}
			}
			h.Clients.Unlock()
			// case client := <-h.OnlineUsersRequest:

		}
	}
}

type MapID map[int]*Client

func (m MapID) CheckID(userID int) bool {
	_, ok := m[userID]
	return ok
}

// RegisterToHub registers the client to its hub
func (h *Hub) RegisterToHub(c *Client) {
	h.register <- c
}

// UnRegisterToHub removes the client from its hub
func (h *Hub) UnRegisterFromHub(c *Client) {
	h.unregister <- c
}

func (h *Hub) GetOnlineUsers() MapID {
	usersID := make(MapID)
	h.Clients.RLock()
	defer h.Clients.RUnlock()
	for client := range h.Clients.items {
		if client.User != nil {
			usersID[client.User.ID] = client
		}
	}
	return usersID
}

func (h *Hub) GetUsersClient(userID int) (*Client, bool) {
	h.Clients.RLock()
	defer h.Clients.RUnlock()
	for client := range h.Clients.items {
		if client.User != nil && client.User.ID == userID {
			return client, true
		}
	}
	return nil, false
}

func (h *Hub) IsThereClient(client *Client) bool {
	h.Clients.RLock()
	defer h.Clients.RUnlock()
	_, ok := h.Clients.items[client]
	return ok
}

func (h *Hub) SendMessageToAllClients(message []byte) {
	h.messageForAll <- message
}
