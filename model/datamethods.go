package model

import (
	"fmt"
)

func (f *Filter) IsCheckedCategory(id int) bool {
	for _, c := range f.CategoryID {
		if id == c {
			return true
		}
	}
	return false
}

func (u *User) String() string {
	if u == nil {
		return "nil"
	}
	return fmt.Sprintf(`name: %s (id: %d)`, u.Name, u.ID)
}

func (u *User) StringFull() string {
	if u == nil {
		return "nil"
	}
	return fmt.Sprintf(`user: {
		    id:           %d
		    name:         %s
		    email:        %s
		    password:     %s
		    DataCreate:   %s
		    session: uuid %s
			till:         %v
		    DateBirth:    %s
		    Gender:       %s
		    First Name:   %s
		    Last Name:    %s
			}`,
		u.ID, u.Name, u.Email, "****", u.DateCreate.String(), u.Uuid, u.ExpirySession, u.DateBirth, u.Gender, u.FirstName, u.LastName)
}

func (p *Post) String() string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("id: %d | Theme: %s\nMessage: \n%s\nCategories: \n%v\nComments: \n%v\n",
		p.ID, p.Theme, p.Message.String(), p.Categories, p.Comments)
}

func (m *message) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("   Author: (%p)\n   %v\n   Content: %s\n   DataCreate: %s | Likes: %#v | UserReaction: %#v\n",
		m.Author, m.Author, m.Content, m.DateCreate.String(), m.Likes, m.UserReaction)
}

func (c *Category) String() string {
	if c == nil {
		return "nil"
	}
	return fmt.Sprintf("categ id: %d | name: %s\n", c.ID, c.Name)
}

func (c *Chat) String() string {
	if c == nil {
		return "nil"
	}
	// members := ""
	// for _, m := range c.Members {
	// 	members += fmt.Sprintf("- %s(id%d) -", m.Name, m.ID)
	// }
	messages := ""
	for _, m := range c.Messages {
		messages += fmt.Sprintf("%s\n--------\n", m)
	}
	return fmt.Sprintf("chat id: %d | name: %s | type: %d  \n   Messages:\n%s\n\n", c.ID, c.Name, c.Type, messages)
}

func (m ChatMessage) String() string {
	return fmt.Sprintf("    Author:    %s\n    Content: %s\n    DataCreate: %s \n",
		m.Author.Name, m.Content, m.DateCreate.String())
}

func (c *Comment) String() string {
	if c == nil {
		return "nil"
	}
	return fmt.Sprintf("comment id: %d | Comment Message: \n%s\n", c.ID, c.Message.String())
}
