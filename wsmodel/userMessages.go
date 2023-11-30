package wsmodel

// TODO ?if it will not take too much work, add Data field to Post and Comment structs to get it from frontend
// TODO ?if it will not take too much work, rename fields Text to Content or fields Content in model package to Text
import (
	"time"

	"forum/model"
)

type Post struct {
	Theme        string    `json:"theme"`
	Content      string    `json:"content"`
	CategoriesID []int     `json:"categoriesID"`
	Date         time.Time `json:"date"`
}

func (p *Post) Validate() string {
	if isEmpty(p.Theme) {
		return "Post's theme missing"
	}
	if isEmpty(p.Content) {
		return "Post's text missing"
	}

	if len(p.CategoriesID) == 0 {
		return "Choose at least one category"
	}

	if p.Date.Before(time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)) {
		return "Date is too old"
	}

	return ""
}

type Comment struct {
	PostID  int       `json:"post_id"`
	Content string    `json:"content"`
	Date    time.Time `json:"date"`
}

func (c *Comment) Validate() string {
	if isEmpty(c.Content) {
		return "Comment's text missing"
	}
	if c.PostID <= 0 {
		return "invalide post's ID"
	}

	if c.Date.Before(time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)) {
		return "Date is too old"
	}
	return ""
}

type ChatMessage struct {
	MessageContent string      `json:"messageContent"`
	Author         *model.User `json:"author,omitempty"`
	Date           time.Time   `json:"date"`
}

func (m *ChatMessage) Validate() string {
	if m.Date.Before(time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC)) {
		return "Date is too old"
	}
	if isEmpty(m.MessageContent) {
		return "text is missing"
	}

	return ""
}

type PrivatChat struct {
	ID            int                 `json:"id"`
	Name          string              `json:"name"`
	CurrentUser   *model.User         `json:"currentUser"`
	RecipientUser *model.User         `json:"recipientUser"`
	Messages      []model.ChatMessage `json:"message"`
}
