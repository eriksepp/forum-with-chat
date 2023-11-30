package controllers

import (
	"testing"
	"time"

	"forum/model"
	"forum/wsmodel"
)

func TestCreateUser(t *testing.T) {
	userExpected := model.User{
		Name:       "usserAdd",
		Email:      "emailtAdd@email",
		DateCreate: time.Now(),
		DateBirth:  time.Date(2002, time.March, 3, 0, 0, 0, 0, time.UTC),
		Gender:     "He",
		FirstName:  "John",
		LastName:   "second",
	}

	uc := wsmodel.UserCredentials{
		Username:  "usserAdd",
		Email:     "emailtAdd@email",
		DateBirth: "2002-03-03",
		Gender:    "He",
		FirstName: "John",
		LastName:  "second",
	}
	user, err := createUser(uc)
	if err != nil {
		t.Fatalf("err is %s\n", err)
	}
	if userExpected.Name != user.Name || userExpected.Email != user.Email || userExpected.DateBirth != user.DateBirth || userExpected.Gender != user.Gender || userExpected.FirstName != user.FirstName || userExpected.LastName != user.LastName {
		t.Fatalf("result is\n%s\n, but expected to be\n%s\n", user, &userExpected)
	}
}
