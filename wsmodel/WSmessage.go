package wsmodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	CurrentSession                = "currentSession"
	ERROR                         = "ERROR"
	RegisterRequest               = "registerRequest"
	RegisterReply                 = "registerReply"
	LoginRequest                  = "loginRequest"
	LoginReply                    = "loginReply"
	LogoutRequest                 = "logoutRequest"
	LogoutReply                   = "logoutReply"
	PostsPortionRequest           = "postsPortionRequest"
	PostsPortionReply             = "postsPortionReply"
	FullPostAndCommentsRequest    = "fullPostAndCommentsRequest"
	FullPostAndCommentsReply      = "fullPostAndCommentsReply"
	NewPostRequest                = "newPostRequest"
	NewPostReply                  = "newPostReply"
	NewCommentRequest             = "newCommentRequest"
	NewCommentReply               = "newCommentReply"
	OnlineUsers                   = "onlineUsers"
	OpenChatRequest               = "openChatRequest"
	OpenChatReply                 = "openChatReply"
	SendMessageToOpendChatRequest = "sendMessageToOpendChatRequest"
	SendMessageToOpendChatReply   = "sendMessageToOpendChatReply"
	InputChatMessage              = "inputChatMessage"
	CloseChatRequest              = "closeChatRequest"
	CloseChatReply                = "closeChatReply"
	ChatPortionRequest            = "chatPortionRequest"
	ChatPortionReply              = "chatPortionReply"
	NewOnlineUser                 = "newOnlineUser"
	OfflineUser                   = "offlineUser"
)

var ErrWarning = errors.New("Warning")

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (m *WSMessage) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("Type: %s | Payload: %s\n", m.Type, m.Payload)
}

func (m *WSMessage) IsAuthentification() bool {
	return strings.HasPrefix(m.Type, "login") || strings.HasPrefix(m.Type, "logout") || strings.HasPrefix(m.Type, "register")
}

func (m *WSMessage) CreateMessageReply(result string, data any) (WSMessage, error) { // TODO rename it to CreateReplyToRequestMessage
	messageType, ok := strings.CutSuffix(m.Type, "Request")
	if !ok {
		return WSMessage{}, fmt.Errorf("bad request message type: %s", m.Type)
	}
	messageType += "Reply"

	return CreateMessage(messageType, result, data)
}

func CreateMessage(messageType string, result string, data any) (WSMessage, error) {
	payload, err := json.Marshal(Payload{Result: result, Data: data})
	if err != nil {
		return WSMessage{}, err
	}
	message := WSMessage{
		Type:    messageType,
		Payload: payload,
	}
	return message, nil
}

type Payload struct {
	Result string `json:"result"`
	Data   any    `json:"data,omitempty"` // TODO ? would be better to get 'null' in JS
}

func isEmpty(field string) bool {
	return strings.TrimSpace(field) == "" || field == "undefined"
}
