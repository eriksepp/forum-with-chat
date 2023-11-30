package controllers

import (
	"errors"
	"fmt"

	"forum/application"
	"forum/model"
	"forum/wsmodel"
	"forum/wsmodel/parse"
)

func replyOpenChat(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any, error) {
	userID, err := parse.PayloadToInt(message.Payload)
	if err != nil || userID <= 0 {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Can't open a chat: invalid userID '%s'", message.Payload), err)
	}

	// check if the user is still online
	userClient, ok := app.Hub.GetUsersClient(userID)
	if !ok {
		return nil, badRequestHelper(app, currConnection, message, fmt.Sprintf("user with id %d is offline", userID))
	}

	chat, err := getChatHistory(app, currConnection, userClient.User.ID, message)
	if err != nil {
		return nil, err
	}

	currConnection.Client.OpenedChatWith.ChatID = chat.ID
	currConnection.Client.OpenedChatWith.ChatName = chat.Name
	currConnection.Client.OpenedChatWith.UserClient = userClient

	app.InfoLog.Printf("Chat with '%s' is opened.", currConnection.Client.OpenedChatWith.UserClient.User.Name)

	return createPrivateChatForReply(chat, currConnection.Client.User, currConnection.Client.OpenedChatWith.UserClient.User), nil
}

func replyCloseChat(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any, error) {
	err := checkLoggedStatus(app, currConnection, message)
	if err != nil {
		return nil, err
	}

	app.InfoLog.Printf("Chat with '%s' is closed.", currConnection.Client.OpenedChatWith.UserClient.User.Name)
	currConnection.Client.OpenedChatWith.ChatID = 0
	currConnection.Client.OpenedChatWith.UserClient = nil

	return nil, nil
}

func replySendMessageToOpendChat(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any, error) {
	
	chatMessage, err := parse.PayloadToChatMessage(message.Payload)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Invalid payload for a chat message: '%s'", message.Payload), err)
	}

	errmessage := chatMessage.Validate()
	if errmessage != "" {
		return nil, badRequestHelper(app, currConnection, message, errmessage)
	}

	id, err := app.ForumData.InsertChatMessage(currConnection.Client.OpenedChatWith.ChatID, currConnection.session.User.ID, chatMessage.MessageContent, nil, chatMessage.Date)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("save chat message in DB failed: '%s'", chatMessage.MessageContent), err)
	}

	// check if the user is still online
	if !app.Hub.IsThereClient(currConnection.Client.OpenedChatWith.UserClient) {
		var ok bool
		currConnection.Client.OpenedChatWith.UserClient, ok = app.Hub.GetUsersClient(currConnection.Client.OpenedChatWith.UserClient.User.ID)
		if !ok {
			errDel := app.ForumData.DeleteChatMessage(id)
			if errDel != nil {
				app.ErrLog.Printf("DeleteChatMessage failed: %v\n", err)
			}
			return nil, badRequestHelper(app, currConnection, message, fmt.Sprintf("user with id %d is offline", currConnection.Client.OpenedChatWith.UserClient.User.ID))
		}
	}

	err = sendMessageToRecipient(app, currConnection, chatMessage)
	if err != nil {
		errDel := app.ForumData.DeleteChatMessage(id)
		if errDel != nil {
			app.ErrLog.Printf("DeleteChatMessage failed: %v\n", err)
		}
		return nil, err
	}

	return "delivered", nil
}

func replyChatPortion(app *application.Application, currConnection *usersConnection, message wsmodel.WSMessage) (any, error) {
	
	beforeID, err := parse.PayloadToInt(message.Payload)
	if err != nil {
		return nil, errHelper(app, currConnection, fmt.Sprintf("Invalid payload for the portion of messages '%s'", message.Payload), err)
	}

	chat, err := app.ForumData.GetPrivateChatMessagesByChatId(currConnection.Client.OpenedChatWith.ChatID, beforeID, CHAT_MESSAGES_PORTION)
	if err != nil {
		return nil, errHelper(app, currConnection, "get the next portion of chat messages from DB failed", err)
	}

	chat.ID = currConnection.Client.OpenedChatWith.ChatID
	chat.Name = currConnection.Client.OpenedChatWith.ChatName

	app.InfoLog.Printf("Send the next portion of messages to chat '%s-%s'.", currConnection.Client.OpenedChatWith.UserClient.User.Name, currConnection.Client.User.Name)
	
	return createPrivateChatForReply(chat, currConnection.Client.User, currConnection.Client.OpenedChatWith.UserClient.User),nil
}

/*
search for chat between 2 users in DB. If the chat is not found, creates and saves to DB a new chat.
Returns the chat with maximum CHAT_MESSAGES_PORTION (10) last messages from DB.
Used in replyOpenChat function
*/
func getChatHistory(app *application.Application, currConnection *usersConnection, userIDChatWith int, message wsmodel.WSMessage) (*model.Chat, error) {
	chatID, chatName, err := app.ForumData.GetPrivateChat(currConnection.Client.User.ID, userIDChatWith)
	if errors.Is(err, model.ErrNoRecord) {
		return createChat(app, currConnection, userIDChatWith)
	}
	if err != nil {
		return nil, errHelper(app, currConnection, "get the chat id from DB failed", err)
	}
	if chatID == 0 {
		return createChat(app, currConnection, userIDChatWith)
	}

	chat, err := app.ForumData.GetPrivateChatMessagesByChatId(chatID, 0, CHAT_MESSAGES_PORTION)
	if err != nil {
		return nil, errHelper(app, currConnection, "get the chat messages from DB failed", err)
	}

	chat.ID = chatID
	chat.Name = chatName

	return chat, nil
}

/*
creates and saves in DB a new chat.
Used in getChatHistory function
*/
func createChat(app *application.Application, currConnection *usersConnection, userIDChatWith int) (*model.Chat, error) {
	chat, err := app.ForumData.CreatePrivatChat(currConnection.Client.User.ID, userIDChatWith)
	if err != nil {
		return nil, errHelper(app, currConnection, "creating a chat failed", err)
	}
	app.InfoLog.Printf("chat created with id %d", chat.ID)
	return chat, nil
}

/*
sends a chat message to the user stored in OpenedChatWith
used in replySendMessageToOpendChat
*/
func sendMessageToRecipient(app *application.Application, currConnection *usersConnection, chatMessage wsmodel.ChatMessage) error {
	recipientConnection := &usersConnection{Client: currConnection.Client.OpenedChatWith.UserClient}
	chatMessage.Author = currConnection.Client.User
	return sendSuccessMessage(app, recipientConnection, wsmodel.InputChatMessage, chatMessage)
}

func createPrivateChatForReply(chat *model.Chat, currentUser, recipientUser *model.User) wsmodel.PrivatChat {
	privateChat := wsmodel.PrivatChat{
		ID:            chat.ID,
		Name:          chat.Name,
		CurrentUser:   currentUser,
		RecipientUser: recipientUser,
		Messages:      chat.Messages,
	}
	return privateChat
}
