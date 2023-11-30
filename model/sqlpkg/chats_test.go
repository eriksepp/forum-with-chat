package sqlpkg

import (
	"fmt"
	"testing"

	"forum/model"
)

func TestGetChatMessagesByUsersIds(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var chat *model.Chat

	f := ForumModel{db}
	fmt.Println("--------------------------------")
	id, name, err := f.GetPrivateChat(2, 3)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("id=", id)
	if id != 0 {
		chat, err = f.GetPrivateChatMessagesByChatId(id, 0, 3)
		if err != nil {
			t.Fatal(err)
		}
	}
	chat.ID = id
	chat.Name = name
	fmt.Printf("%s\n", chat.String())

	fmt.Println("--------------------------------")
	id, name, err = f.GetPrivateChat(4, 5)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("id=", id)
	fmt.Println("-----get last 3 messages")
	if id != 0 {
		chat, err = f.GetPrivateChatMessagesByChatId(id, 0, 3)
		if err != nil {
			t.Fatal(err)
		}
	}
	chat.ID = id
	chat.Name = name
	fmt.Printf("%s\n", chat.String())

	beforeID:=5
	fmt.Println("-----get 3 messages beforeID ", beforeID)
	if id != 0 {
		chat, err = f.GetPrivateChatMessagesByChatId(id, beforeID, 3)
		if err != nil {
			t.Fatal(err)
		}
	}
	chat.ID = id
	chat.Name = name
	fmt.Printf("%s\n", chat.String())
	
	beforeID=4
	fmt.Println("-----get 3 messages beforeID ", beforeID)
	if id != 0 {
		chat, err = f.GetPrivateChatMessagesByChatId(id, beforeID, 3)
		if err != nil {
			t.Fatal(err)
		}
	}
	chat.ID = id
	chat.Name = name
	fmt.Printf("%s\n", chat.String())

	fmt.Println("--------------------------------")
	id, name, err = f.GetPrivateChat(2, 5)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("id=", id)
	if id != 0 {
		chat, err = f.GetPrivateChatMessagesByChatId(id, 0, 3)
		if err != nil {
			t.Fatal(err)
		}
	}
	chat.ID = id
	chat.Name = name
	fmt.Printf("%s\n", chat.String())
}


func TestGetLastMessageDateFromUserToRecipientt(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	date, err := f.GetLastMessageDateFromUserToRecipient(4, 8)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	fmt.Println("---------------")

	fmt.Printf("from user %d to user %d mes date: %v\n", 4,10, date)
}