package sqlpkg

import (
	"database/sql"
	"errors"
	"time"

	"forum/model"
)

const ConstFields = ` u.id, u.name, u.email, u.dateCreate, u.dateBirth, u.gender, u.firstName, u.lastName `

/*
returns list of all users in DB
*/
func (f *ForumModel) GetAllUsers() ([]*model.User, error) {
	return f.getUsersByCondition("")
}

/*
returns list of users who filtered by userIDs.CheckID().
*/
func (f *ForumModel) GetFilteredUsers(userIDs model.IdChecker) ([]*model.User, error) {
	q := `SELECT ` + ConstFields + ` FROM users u `
	rows, err := f.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.DateCreate, &user.DateBirth, &user.Gender, &user.FirstName, &user.LastName)
		if err != nil {
			return nil, err
		}
		if userIDs.CheckID(user.ID) {
			users = append(users, user)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

/*
returns list of users who filtered by userIDs.CheckID().
The list is ordered in descending order by date of chat messages sent from the user in the list to the user 'forUserID',
if there is no message from a user, shuch users will sort by their names.
*/
func (f *ForumModel) GetFilteredUsersOrderedByMessagesToGivenUser(userIDs model.IdChecker, forUserID int) ([]*model.User, error) {
	q := `SELECT u.id, u.name , max(ms.dateCreate) FROM users u
	LEFT JOIN (SELECT  mb.id as mbID,  mb.userID as UserID FROM chat_members mb
		WHERE mb.userID!=? AND mb.chatID IN (SELECT chatID FROM chat_members WHERE userID=?)) 
		userMb ON u.id=userMb.UserID 
	LEFT JOIN chat_messages ms ON userMb.mbID=ms.chat_membersID
	WHERE u.id!=?
	GROUP BY u.id  ORDER BY ms.dateCreate desc, lower(name)`

	rows, err := f.DB.Query(q, forUserID, forUserID, forUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		var dateCreate sql.NullString
		err := rows.Scan(&user.ID, &user.Name, &dateCreate)
		if err != nil {
			return nil, err
		}
		if userIDs.CheckID(user.ID) {
			user.LastMessageDate=dateCreate.String
			users = append(users, user)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (f *ForumModel) getUsersByCondition(condition string) ([]*model.User, error) {
	q := `SELECT ` + ConstFields + ` FROM users u ` + condition

	rows, err := f.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.DateCreate, &user.DateBirth, &user.Gender, &user.FirstName, &user.LastName)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

/*
returns a user from DB by ID
*/
func (f *ForumModel) GetUserByID(id int) (*model.User, error) {
	q := `SELECT ` + ConstFields + `, u.password, s.uuid, s.expirySession 
	      FROM users u LEFT JOIN usersessions s ON u.id=s.userID WHERE u.id=?`

	user := &model.User{}
	var uuidInDB sql.NullString
	var expirySessionInDB sql.NullTime
	row := f.DB.QueryRow(q, id)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.DateCreate, &user.DateBirth, &user.Gender, &user.FirstName, &user.LastName, &user.Password, &uuidInDB, &expirySessionInDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}

	user.Uuid = uuidInDB.String
	user.ExpirySession = expirySessionInDB.Time
	return user, nil
}

/*
returns a user from DB by the name
*/
func (f *ForumModel) GetUserByName(name string) (*model.User, error) {
	q := `SELECT ` + ConstFields + `, u.password, s.uuid, s.expirySession  
	      FROM users u LEFT JOIN usersessions s ON u.id=s.userID
		  WHERE u.name=?`

	user := &model.User{}
	var uuidInDB sql.NullString
	var expirySessionInDB sql.NullTime
	row := f.DB.QueryRow(q, name)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.DateCreate, &user.DateBirth, &user.Gender, &user.FirstName, &user.LastName, &user.Password, &uuidInDB, &expirySessionInDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}

		return nil, err
	}

	user.Uuid = uuidInDB.String
	user.ExpirySession = expirySessionInDB.Time

	return user, nil
}

/*
returns a user from DB by the email
*/
func (f *ForumModel) GetUserByEmail(email string) (*model.User, error) {
	q := `SELECT ` + ConstFields + `, u.password, s.uuid, s.expirySession 
	      FROM users u LEFT JOIN usersessions s ON u.id=s.userID 
		  WHERE u.email=?`

	user := &model.User{}
	var uuidInDB sql.NullString
	var expirySessionInDB sql.NullTime
	row := f.DB.QueryRow(q, email)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.DateCreate, &user.DateBirth, &user.Gender, &user.FirstName, &user.LastName, &user.Password, &uuidInDB, &expirySessionInDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}

	user.Uuid = uuidInDB.String
	user.ExpirySession = expirySessionInDB.Time

	return user, nil
}

/*
returns a user from DB by the email
*/
func (f *ForumModel) GetUserByUUID(uuid string) (*model.User, error) {
	q := `SELECT ` + ConstFields + `, s.uuid, s.expirySession 
	FROM  users u INNER JOIN usersessions s ON u.id=s.userID WHERE s.uuid=?`

	user := &model.User{}
	var uuidInDB sql.NullString
	var expirySessionInDB sql.NullTime
	row := f.DB.QueryRow(q, uuid)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.DateCreate, &user.DateBirth, &user.Gender, &user.FirstName, &user.LastName, &uuidInDB, &expirySessionInDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}

	user.Uuid = uuidInDB.String
	user.ExpirySession = expirySessionInDB.Time

	return user, nil
}

/*
inserts the new user into DB. It doesn't do any check of unique data. But if DB have some restricts, it will return an error
*/
func (f *ForumModel) InsertUser(user *model.User) (int, error) {
	q := `INSERT INTO users  (name, email, password, dateCreate, dateBirth, gender, firstName, lastName) VALUES (?,?,?,?,?,?,?,?)`
	res, err := f.DB.Exec(q, user.Name, user.Email, user.Password, user.DateCreate, user.DateBirth, user.Gender, user.FirstName, user.LastName)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

/*
adds a session uuid to the user with the given ID
*/
func (f *ForumModel) AddUsersSession(id int, uuid string, expirySession time.Time) error {
	q := `INSERT INTO usersessions (userID, uuid,expirySession,agent) VALUES (?,?,?,?)`
	_, err := f.DB.Exec(q, id, uuid, expirySession, "")
	if err != nil {
		return err
	}

	return nil
}

/*
deletes the user's session uuid
*/
func (f *ForumModel) DeleteUsersSession(uuid string) error {
	q := `DELETE FROM usersessions WHERE uuid=?`
	res, err := f.DB.Exec(q, uuid)
	if err != nil {
		return err
	}

	return f.checkUnique(res)
}

/*
check if a user with the given name exists,  returns nil only if there is exactly one user
*/
func (f *ForumModel) CheckUserByName(name string) error {
	err := f.checkExisting("users", "name", name)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrNoRecord
	}
	return err
}

/*
check if a user with the given email exists, returns nil only if there is exactly one user
*/
func (f *ForumModel) CheckUserByEmail(email string) error {
	err := f.checkExisting("users", "email", email)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrNoRecord
	}
	return err
}

/*
adds the user to DB
*/
func (f *ForumModel) AddUser(user *model.User) (int, error) {
	id, err := f.InsertUser(user)
	if err != nil {
		errUnique := f.CheckUserByName(user.Name)
		if errUnique == nil {
			return 0, model.ErrUniqueUserName
		}
		errUnique = f.CheckUserByEmail(user.Email)
		if errUnique == nil {
			return 0, model.ErrUniqueUserEmail
		}
	}

	return id, nil
}

/*
changes an email of the user with the given id
*/
func (f *ForumModel) ChangeUsersEmail(id int, email string) error {
	err := f.changeUsersField(id, "email", email)
	if err != nil {
		errUnique := f.CheckUserByEmail(email)
		if errUnique == nil {
			return model.ErrUniqueUserEmail
		}
	}
	return err
}

/*
changes a password of the user with the given id
*/
func (f *ForumModel) ChangeUsersPassword(id int, password string) error {
	return f.changeUsersField(id, "password", password)
}

/*
changes a field in the users table for the user with the given id
*/
func (f *ForumModel) changeUsersField(id int, field, value string) error {
	q := `UPDATE users SET ` + field + `=? WHERE id=?`
	res, err := f.DB.Exec(q, value, id)
	if err != nil {
		return err
	}

	return f.checkUnique(res)
}
