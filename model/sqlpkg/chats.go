package sqlpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"forum/model"
)

func (f *ForumModel) CreatePrivatChat(userID1, userID2 int) (*model.Chat, error) {
	q := `INSERT INTO chats (name, type) VALUES (?,?);`
	name := ""
	if userID1 < userID2 {
		name = fmt.Sprintf("%d-%d", userID1, userID2)
	} else {
		name = fmt.Sprintf("%d-%d", userID2, userID1)
	}
	res, err := f.DB.Exec(q, name, 0)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	chatID := int(id)
	err = f.AddUserToChat(chatID, userID1)
	if err != nil {
		return nil, err
	}
	err = f.AddUserToChat(chatID, userID2)
	if err != nil {
		return nil, err
	}

	return &model.Chat{ID: chatID, Name: name, Type: 0, Messages: []model.ChatMessage{}}, nil
}

func (f *ForumModel) AddUserToChat(chatID, userID int) error {
	q := `INSERT INTO chat_members (chatID, userID) VALUES (?,?);`
	_, err := f.DB.Exec(q, chatID, userID)
	if err != nil {
		q = `DELETE FROM chats WHERE id=?`
		_, errDel := f.DB.Exec(q, chatID)
		if errDel != nil {
			return errDel
		}
		return err
	}
	return nil
}

/*
inserts a new chat message into DB, returns an ID for the message
*/
func (f *ForumModel) InsertChatMessage(chatID int, authorID int, content string, images []string, dateCreate time.Time) (int, error) {
	var strOfImages sql.NullString
	chatMembersID, err := f.getChatMembersID(chatID, authorID)
	if err != nil {
		return 0, fmt.Errorf("failed in getting a chatMembersID: %w", err)
	}
	if len(images) != 0 {
		strOfImages.String = strings.Join(images, ",")
		strOfImages.Valid = true
	}
	q := `INSERT INTO chat_messages (content, images, chat_membersID, dateCreate) VALUES (?,?,?,?)`
	res, err := f.DB.Exec(q, content, strOfImages, chatMembersID, dateCreate)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (f *ForumModel) DeleteChatMessage(id int) error {
	q := `DELETE FROM chat_messages WHERE id=?`
	_, err := f.DB.Exec(q, id)
	if err != nil {
		return err
	}
	return nil
}

func (f *ForumModel) GetLastMessageDateFromUserToRecipient(userId int, recipientID int) (string, error) {
	q := `SELECT max(ms.dateCreate) FROM chat_messages ms 
	WHERE ms.chat_membersID IN (SELECT  mb.id as mbID FROM chat_members mb
		WHERE mb.userID=? AND mb.chatID IN (SELECT chatID FROM chat_members WHERE userID=?)) 
	GROUP BY ms.chat_membersID `

	//var messageDateCreate sql.NullTime
	var messageDateCreate sql.NullString
	row := f.DB.QueryRow(q, recipientID, userId)
	err := row.Scan(&messageDateCreate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return  messageDateCreate.String, model.ErrNoRecord
		}
		return messageDateCreate.String, err
	}

	return  messageDateCreate.String, nil
}

func (f *ForumModel) GetPrivateChat(usrID1, usrID2 int) (int, string, error) {
	var chatID int
	var chatName string

	q := `SELECT ch.id, ch.name 
			FROM chats ch
			LEFT JOIN  chat_members mb ON ch.id=mb.chatID 
			
			WHERE ch.id IN (SELECT chatID FROM chat_members cmb WHERE cmb.userID=?) 
			  AND ch.id IN (SELECT chatID FROM chat_members cmb WHERE cmb.userID=?) 
			  AND ch.type=0 ` // for privat chats type=0

	rows, err := f.DB.Query(q, usrID1, usrID2)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	counter := 0
	for rows.Next() {
		if counter == 2 {
			return 0, "", errors.New("too many chat members, expected 2")
		}
		// SELECT  ch.id
		err := rows.Scan(&chatID, &chatName)
		if err != nil {
			return 0, "", err
		}

		counter++
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", model.ErrNoRecord
		}
		return 0, "", err
	}

	if chatID == 0 {
		return 0, "", model.ErrNoRecord
	}
	return chatID, chatName, nil
}

/*
returns 'messageNumbers' messages with ids less than 'beforeId' for the chat with the given 'id'.
The messages are sorted by date of creation.
If beforeId and/or messageNumbers is less or equal to 0 it will ignore the conditions.
*/
func (f *ForumModel) GetPrivateChatMessagesByChatId(id int, beforeId int, messageNumbers int) (*model.Chat, error) {
	condition, arguments := createConditionSelectMessagesByChatId(id, beforeId)
	query := createQuerySelectMessages(
		" chat_messages.id, chat_messages.content, chat_messages.images, chat_messages.dateCreate, chat_members.userID, users.name ",
		condition,
	)
	chat, err := f.execQuerySelectMessages(query, arguments, messageNumbers, scanRowToChatMessageAndUser)
	if err != nil {
		return nil, err
	}
	chat.ID = id
	return chat, nil
}

/*
creates a condition for the query to select chat messages.
Used in the GetChatMessagesByChatId function.
*/
func createConditionSelectMessagesByChatId(chatID int, beforeId int) (string, []any) {
	condition := ` WHERE chats.id = ? `
	arguments := []any{chatID}

	if beforeId > 0 {
		condition += ` AND chat_messages.id < ? `
		arguments = append(arguments, beforeId)
	}
	return condition, arguments
}

/*
creates the query to select chat messages.
Used in the GetChatMessagesByChatId and GetChatMessagesByUsersId function.
*/
func createQuerySelectMessages(fields, condition string) (query string) {
	query = `
		SELECT  ` + fields + ` 
	    FROM chat_messages
	    LEFT JOIN chat_members ON chat_messages.chat_membersID=chat_members.id  
		LEFT JOIN chats ON chat_members.chatID=chats.id 
		LEFT JOIN users ON chat_members.userID=users.id 
		` + condition + ` 
		ORDER BY chat_messages.dateCreate DESC;
		`
	return
}

/*
executes a query to select chat messages.
Used in the GetChatMessage function.
*/
func (f *ForumModel) execQuerySelectMessages(query string, arguments []any, messageNumbers int, scanRowToChat func(*sql.Rows, *model.Chat) error) (*model.Chat, error) {
	rows, err := f.DB.Query(query, arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chat := &model.Chat{}
	authors := make(map[int]*model.User) // to create the user's struct for the author of a messages just once and use in each message the same pointer to the same authors
	messageCounter := 0                  // the number of the last added message

	for rows.Next() {
		err := scanRowToChat(rows, chat)
		if err != nil {
			return nil, err
		}

		determinMessageAuthor(chat, messageCounter, authors)

		messageCounter++
		if messageCounter == messageNumbers {
			break
		}
	}

	if err = rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}

	return chat, nil
}

/*
scans a row from a query to the the item of model.Chat.
Used in the execQuerySelectMessages function.
*/
func scanRowToChatMessageAndUser(rows *sql.Rows, chat *model.Chat) error {
	author := model.User{}
	message := model.ChatMessage{}
	var images sql.NullString

	// parse the row with fields:
	// SELECT  chat_messages.id, chat_messages.content, chat_messages.images, chat_messages.dateCreate, chat_members.userID, users.name
	err := rows.Scan(&message.ID, &message.Content, &images, &message.DateCreate, &author.ID, &author.Name)
	message.Images = getImagesArray(images)
	message.Author = &author
	chat.Messages = append(chat.Messages, message)
	return err
}

/*
checks if the author of the message was found before (in a previouse row),
if it's true replace a pointer to the author to the exintent one.
Used in the execQuerySelectMessages function.
*/
func determinMessageAuthor(chat *model.Chat, lastMessageIndex int, authors map[int]*model.User) {
	// find out if the author in the current row is found before, if yes, keep that previouse one
	author := chat.Messages[lastMessageIndex].Author

	if existingAuthor, ok := authors[author.ID]; ok {
		chat.Messages[lastMessageIndex].Author = existingAuthor
	} else {
		authors[author.ID] = author
	}
}

/*
returns id of the row in chat_members table for a given chat and user
used in InsertChatMessage
*/
func (f *ForumModel) getChatMembersID(chatID int, authorID int) (int, error) {
	var id int
	q := `SELECT id FROM chat_members WHERE chatID=? AND userID=? `
	row := f.DB.QueryRow(q, chatID, authorID)

	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrNoRecord
		}
		return 0, err
	}

	return id, nil
}
