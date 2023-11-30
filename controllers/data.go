package controllers

import (
	"errors"
	"net/http"

	"forum/application"
	"forum/controllers/chat"
	"forum/session"
	"forum/wsmodel"
)

// form fields
const (
	F_NAME         = "name"
	F_PASSWORD     = "password"
	F_EMAIL        = "email"
	F_CONTENT      = "text"
	F_IMAGES       = "images"
	F_AUTHORID     = "authorID"
	F_THEME        = "title"
	F_CATEGORIESID = "categoriesID"
	F_LIKEBY       = "likedby"
	F_DISLIKEBY    = "dislikedby"
)

const POST_PREVIEW_LENGTH = 450

const (
	POSTS_ON_POSTSVIEW    = 10
	CHAT_MESSAGES_PORTION = 10
)

const USER_IMAGES_DIR = "./images"

const (
	MaxFileUploadSize = 20 << 20               // 20MB
	MaxUploadSize     = 10 * MaxFileUploadSize // 10 files by 20MB
)

type usersConnection struct {
	session *session.Session
	Client  *chat.Client
}

func (uc *usersConnection) renewClientForUser(app *application.Application, session *session.Session) error {
	oldClient := uc.Client
	// (online changes)uc.Client = chat.NewClient(app.Hub, session.User, uc.Client.Conn, uc.Client.ReceivedMessages, uc.Client.ClientRegistered)
	err := uc.createNewClientAndSendUserOnline(app, uc.Client.Conn, uc.Client.ReceivedMessages, uc.Client.ClientRegistered)
	if err != nil {
		return err
	}
	app.InfoLog.Printf("registered new client: %s", uc.Client)
	
	// (online changes)app.Hub.UnRegisterFromHub(oldClient)
	err=uc.deleteClientAndSendUserOffline(app,oldClient)
	if err != nil {
		return err
	}
	app.InfoLog.Printf("send client %s to unregister", oldClient)
	return nil
}

type (
	replierAuthenticator func(*application.Application, http.ResponseWriter, *usersConnection, wsmodel.WSMessage) error
	replier              func(*application.Application, *usersConnection, wsmodel.WSMessage) error
)

var (
	replierAuthenticators = map[string]replierAuthenticator{
		wsmodel.RegisterRequest: replyRegister,
		wsmodel.LoginRequest:    replyLogin,
		wsmodel.LogoutRequest:   replyLogout,
	}
	repliers = map[string]replier{
		wsmodel.PostsPortionRequest:           sendReplyForLoggedUser(replyPosts),
		wsmodel.FullPostAndCommentsRequest:    sendReplyForLoggedUser(replyFullPostAndComments),
		wsmodel.NewPostRequest:                sendReplyForLoggedUser(replyNewPost),
		wsmodel.NewCommentRequest:             sendReplyForLoggedUser(replyNewComment),
		wsmodel.OpenChatRequest:               sendReplyForLoggedUser(replyOpenChat),
		wsmodel.SendMessageToOpendChatRequest: sendReplyForLoggedUser(replySendMessageToOpendChat),
		wsmodel.CloseChatRequest:              sendReplyForLoggedUser(replyCloseChat),
		wsmodel.ChatPortionRequest:            sendReplyForLoggedUser(replyChatPortion),
	}
)

var ErrNoPost = errors.New("could not find post") // TODO Naming:  perhaps ErrMissedPost would be better

type replyDataCreator func(*application.Application, *usersConnection, wsmodel.WSMessage) (any, error)

func sendReplyForLoggedUser(createReplyData replyDataCreator) replier {
	return func(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) error {
		err := checkLoggedStatus(app, currConnection, message)
		if err != nil {
			return err
		}

		replyData, err := createReplyData(app, currConnection, message)
		if err != nil {
			return err
		}

		return sendReply(app, currConnection, message, replyData)
	}
}
