package sqlpkg

import (
	"fmt"
	"testing"
	"time"

	"forum/model"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const DBPath = "../../forumDB.db"

func TestCreateDB(t *testing.T) {
	admin := model.User{
		ID:        1,
		Name:      "admin",
		Email:     "admin@forum.com",
		Password:  []byte("admin"),
		DateBirth: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC),
		Gender:    "They",
		FirstName: "AD",
		LastName:  "MIN",
	}
	db, err := CreateDB("test.db", &admin)
	// db, err := OpenDB(DBPath,"webuser","webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f := ForumModel{db}

	uss, err := f.GetAllUsers()
	if err != nil {
		t.Fatal(err)
	}

	for _, us := range uss {
		fmt.Println(us)
	}
	fmt.Println("------------")

	user := model.User{
		Name:       "test11",
		Email:      "test1@email",
		DateCreate: time.Date(2023, time.March, 3, 12, 12, 21, 0, time.UTC),
		DateBirth:  time.Date(2003, time.March, 3, 12, 12, 21, 0, time.UTC),
		Gender:     "He",
		FirstName:  "John",
		LastName:   "Test",
	}
	user.Password, _ = bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost) // hash the password

	id, err := f.InsertUser(&user)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("--id=%d-----\n", id)
	fmt.Println("------------")
	uss, err = f.GetAllUsers()
	if err != nil {
		t.Fatal(err)
	}

	for _, us := range uss {
		fmt.Println(us)
	}
	fmt.Println("----end-----")
}

func TestAuthenDB(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser") // open as not admin
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var sqlconn *sqlite3.SQLiteConn
	err = sqlconn.AuthUserAdd("webuser1", "webuser", false)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("----end-----")
}
