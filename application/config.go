package application

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"forum/controllers/chat"
	"forum/logger"
	"forum/model"
	"forum/model/sqlpkg"
	"forum/view"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var (
	ADMIN = model.User{
		ID:        1,
		Name:      "admin",
		Email:     "admin@forum.com",
		Password:  []byte("admin"),
		DateBirth: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC),
		Gender:    "They",
		FirstName: "AD",
		LastName:  "MIN",
	}
	ADM_PASS = "admin"
)

type Application struct {
	ErrLog    *log.Logger
	InfoLog   *log.Logger
	View      *view.View
	Hub       *chat.Hub
	ForumData *sqlpkg.ForumModel
	Upgrader  websocket.Upgrader
	Server    *http.Server
}

func New() (*Application, error) {
	var application Application
	var err error
	application.ErrLog, application.InfoLog = logger.CreateLoggers()

	application.View, err = view.New("index.html")
	if err != nil {
		return &application, err
	}

	application.Hub = chat.NewHub()
	go application.Hub.Run()

	application.InfoLog.Println("The chat Hub is created")

	application.Upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return &application, nil
}

func (app *Application) CreateDB(fileName string) error {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(ADMIN.Password), 8)
	if err != nil {
		return fmt.Errorf("password crypting failed: %w", err)
	}

	ADMIN.Password = hashPassword
	db, err := sqlpkg.CreateDB(fileName, &ADMIN)
	if err != nil {
		return fmt.Errorf("creating DB faild: %w", err)
	}
	app.ForumData = &sqlpkg.ForumModel{DB: db}
	app.InfoLog.Printf("DB has been created")
	return nil
}

func (app *Application) FillTestDB(path string) error {
	hashPassword1, err := bcrypt.GenerateFromPassword([]byte("test1"), 8)
	if err != nil {
		return fmt.Errorf("password crypting failed: %w", err)
	}
	hashPassword2, err := bcrypt.GenerateFromPassword([]byte("test2"), 8)
	if err != nil {
		return fmt.Errorf("password crypting failed: %w", err)
	}
	app.InfoLog.Println("DB has been filled by examles of data")
	return app.ForumData.FillInDB(path, string(hashPassword1), string(hashPassword2))
}
