package wsmodel

import (
	"testing"
)

func TestValidate(t *testing.T) {
	UCs := []UserCredentials{
		{
			Username:  "usserAdd",
			Email:     "emailtAdd@email",
			Password:  "pass",
			DateBirth: "2002-03-03",
			Gender:    "He",
			FirstName: "John",
			LastName:  "second",
		},
		{
			Username:  "",
			Email:     "emailtAdd@email",
			Password:  "pass",
			DateBirth: "2002-03-03",
			Gender:    "He",
			FirstName: "John",
			LastName:  "second",
		},
		{
			Username:  "usserAdd",
			Email:     "",
			Password:  "pass",
			DateBirth: "2002-03-03",
			Gender:    "He",
			FirstName: "John",
			LastName:  "second",
		},
		{
			Username:  "usserAdd",
			Email:     "emailtAdd@email",
			Password:  "",
			DateBirth: "2002-03-03",
			Gender:    "He",
			FirstName: "John",
			LastName:  "second",
		},
		{
			Username:  "usserAdd",
			Email:     "emailtAdd@email",
			Password:  "pass",
			DateBirth: "",
			Gender:    "He",
			FirstName: "John",
			LastName:  "second",
		},
		{
			Username:  "usserAdd",
			Email:     "emailtAdd@email",
			Password:  "pass",
			DateBirth: "2002-03-03",
			Gender:    "",
			FirstName: "John",
			LastName:  "second",
		},
		{
			Username:  "usserAdd",
			Email:     "emailtAdd@email",
			Password:  "pass",
			DateBirth: "2002-03-03",
			Gender:    "He",
			FirstName: "",
			LastName:  "second",
		},
		{
			Username:  "usserAdd",
			Email:     "emailtAdd@email",
			Password:  "pass",
			DateBirth: "2002-03-03",
			Gender:    "He",
			FirstName: "John",
			LastName:  "",
		},
		{
			Username:  "usserAdd",
			Email:     "emailtAddemail",
			Password:  "pass",
			DateBirth: "2002-03-03",
			Gender:    "He",
			FirstName: "John",
			LastName:  "second",
		},
	}

	message := []string{
		"",
		"username missing",
		"email missing",
		"password missing",
		"dateBirth missing",
		"gender missing",
		"firstName missing",
		"lastName missing",
		"wrong email",
	}

	for i := 0; i < len(UCs); i++ {
		res := UCs[i].Validate()
		if res != message[i] {
			t.Fatalf("# %d: result is\n%s\n, but expected to be\n%s\n", i, res, message[i])
		}
	}
}
