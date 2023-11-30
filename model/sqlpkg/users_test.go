package sqlpkg

import (
	"fmt"
	"testing"
	"time"

	"forum/controllers/chat"
	"forum/model"

	"golang.org/x/crypto/bcrypt"
)

func TestAddUserSession(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("--- add a session to the user 1 ---")
	fmt.Println(f.AddUsersSession(1, "ses1", time.Now()))
	fmt.Println("--- add a session to the user 10(not existing) ---")
	fmt.Println(f.AddUsersSession(10, "ses1", time.Now()))
}

func TestDeleteUsersSession(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("--- delete the session ses1 ---")
	fmt.Println(f.DeleteUsersSession("ses1"))
	fmt.Println("--- delete the session d8ce41bc-a504-4c4d-9285-c560a4bcaa7b ---")
	fmt.Println(f.DeleteUsersSession("d8ce41bc-a504-4c4d-9285-c560a4bcaa7b"))
}

func TestInsertUser(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	user := model.User{
		Name:       "ussertest1",
		Email:      "emailt1@email",
		DateCreate: time.Now(),
		DateBirth:  time.Date(2001, time.March, 3, 12, 12, 21, 0, time.UTC),
		Gender:     "He",
		FirstName:  "John",
		LastName:   "First",
	}
	user.Password, _ = bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost) // hash the password

	fmt.Println("--- add a user with the existing name ---")
	n, err := f.InsertUser(&user)
	fmt.Printf("n= %d, err= %v\n", n, err)

	user = model.User{
		Name:       "usserAdd",
		Email:      "emailtAdd@email",
		DateCreate: time.Now(),
		DateBirth:  time.Date(2002, time.March, 3, 12, 12, 21, 0, time.UTC),
		Gender:     "He",
		FirstName:  "John",
		LastName:   "second",
	}
	user.Password, _ = bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost) // hash the password

	id, err := f.AddUser(&user)
	fmt.Printf("user= %v, err= %v\n", id, err)
}

func TestGetUserByUUID(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("------")
	u, err := f.GetUserByUUID("d8ce41bc-a504-4c4d-9285-c560a4bcaa7b")
	fmt.Printf("user= %v, \nerr= %v\n", u, err)

	fmt.Println("---no that session---")
	u, err = f.GetUserByUUID("d8ce41bc-a504-4c4d-9285-c560a4b")
	fmt.Printf("user= %v, \nerr= %v\n", u, err)
}

func TestGetUserByID(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("---id 2---")
	u, err := f.GetUserByID(2)
	fmt.Printf("user= %v, \nerr= %v\n", u, err)

	fmt.Println("---id 4---")
	u, err = f.GetUserByID(4)
	fmt.Printf("user= %v, \nerr= %v\n", u, err)
}

func TestGetUserByName(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("---name test1---")
	u, err := f.GetUserByName("test1")
	fmt.Printf("user= %v, \nerr= %v\n", u, err)

	fmt.Println("---name test4---")
	u, err = f.GetUserByName("test4")
	fmt.Printf("user= %v, \nerr= %v\n", u, err)
}

func TestGetUserByEmail(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("---Email test1---")
	u, err := f.GetUserByEmail("test1@forum")
	fmt.Printf("user= %v, \nerr= %v\n", u, err)

	fmt.Println("---Email test4---")
	u, err = f.GetUserByEmail("test@ff.f4")
	fmt.Printf("user= %v, \nerr= %v\n", u, err)
}

func TestGetFilteredUsersOrderedByMessageDate(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	users := chat.MapID{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil, 9: nil, 10: nil}
	// users :=  make(chat.MapID)

	uss, err := f.GetFilteredUsersOrderedByMessagesToGivenUser(users, 4)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	fmt.Println("---------------")
	for _, user := range uss {
		fmt.Printf("user: = %v, last message: %s\n", user, user.LastMessageDate)
	}
}

