package model

import (
	"errors"
	"time"
)

type Time time.Time

// Tables' names for likes
const (
	POST    = "post"
	COMMENT = "comment"
)
const N_LIKES = 2

const (
	DISLIKE UserReactions = iota
	LIKE
)

var (
	ErrNoRecord        = errors.New("there is no record in the DB")
	ErrTooManyRecords  = errors.New("there are more than one record")
	ErrUnique          = errors.New("unique constraint failed")
	ErrUniqueUserName  = errors.New("user with the given name already exists")
	ErrUniqueUserEmail = errors.New("user with the given email already exists")
)

type UserReactions int

type User struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Password        []byte    `json:"-"`
	Email           string    `json:"email,omitempty"`
	DateCreate      time.Time `json:"dateCreate,omitempty"`
	DateBirth       time.Time `json:"dateBirth,omitempty"`
	Gender          string    `json:"gender,omitempty"`
	FirstName       string    `json:"firstName,omitempty"`
	LastName        string    `json:"lastName,omitempty"`
	Uuid            string    `json:"uuid,omitempty"`
	ExpirySession   time.Time `json:"expirySession,omitempty"`
	LastMessageDate string    `json:"lastMessageDate"`
}

type message struct {
	Author       *User         `json:"author,omitempty"`
	Content      string        `json:"content"`
	DateCreate   time.Time     `json:"dateCreate,omitempty"`
	Likes        []int         `json:"likes,omitempty"` // index 0 keeps number of dislikes, index 1 keeps number of likes
	Images       []string      `json:"-"`
	UserReaction UserReactions `json:"userReaction,omitempty"` //-1 => no reaction
}

type Post struct {
	ID               int         `json:"id,omitempty"`
	Theme            string      `json:"theme"`
	Message          message     `json:"message"`
	Categories       []*Category `json:"categories"`
	Comments         []*Comment  `json:"comments,omitempty"`
	CommentsQuantity int         `json:"commentsQuantity,omitempty"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

type Comment struct {
	ID      int     `json:"id,omitempty"`
	PostID  int     `json:"postID"`
	Message message `json:"message"`
}

type Chat struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	Type     int           `json:"type"`
	Messages []ChatMessage `json:"message"`
}

type ChatMessage struct {
	ID         int       `json:"id"`
	Author     *User     `json:"author,omitempty"`
	Content    string    `json:"content"`
	DateCreate time.Time `json:"dateCreate,omitempty"`
	Images     []string  `json:"-"`
}

type Filter struct {
	CategoryID       []int `json:"categoryID"`
	AuthorID         int   `json:"authorID"`
	LikedByUserID    int   `json:"likedByUserID"`
	DisLikedByUserID int   `json:"disLikedByUserID"`
}

/*
checkID must return true if a user with the given ID has to be added to the result of selection from DB
*/
type IdChecker interface {
	CheckID(id int) bool
}
