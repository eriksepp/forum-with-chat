package wsmodel

import (
	"net/mail"
)

type UserCredentials struct {
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
	DateBirth string `json:"dateBirth,omitempty"`
	Gender    string `json:"gender,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

func (u *UserCredentials) Validate() string {
	if isEmpty(u.Username) {
		return "username missing"
	}
	if isEmpty(u.Email) {
		return "email missing"
	}

	// check email
	// mail.ParseAddress accepts also local domens e.g. witout .(dot)
	address, err := mail.ParseAddress(u.Email)
	if err != nil {
		return "wrong email"
	}
	u.Email = address.Address // in case of full address like "Barry Gibbs <bg@example.com>"
	// the regex allows only Internet emails, e.g. with dot-atom domain (https://www.rfc-editor.org/rfc/rfc5322.html#section-3.4)
	// if !regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`).Match([]byte(email)) {
	// 	return "wrong email"
	// }

	if isEmpty(u.Password) {
		return "password missing"
	}
	if isEmpty(u.DateBirth) { // TODO find out which field of the date was empty (year, month, day)
		return "dateBirth missing"
	}
	if isEmpty(u.Gender) {
		return "gender missing"
	}
	if isEmpty(u.FirstName) {
		return "First name missing"
	}
	if isEmpty(u.LastName) {
		return "Last name missing"
	}

	return ""
}

type UserWithMessageDate struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	LastMessageDate string `json:"lastMessageDate"`
}
