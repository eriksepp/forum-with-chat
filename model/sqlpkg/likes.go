package sqlpkg

import (
	"database/sql"
	"errors"

	"forum/model"
)

/****
the group of function for getting likes
****/

/* returns quantity of likes/dislikes from the given table (posts or comments) for the given id of a message*/
func (f *ForumModel) GetLikes(tableName string, messageID int, userIDForReaction int) ([]int, int8, error) {
	likes := []int{0, 0}
	q := `SELECT  count(CASE WHEN like THEN TRUE END) AS likes, count(CASE WHEN NOT like THEN TRUE END)  AS dislikes,   
	(SELECT like FROM posts_likes WHERE userID = ? AND messageID = ?) AS user_like 
	FROM ` + tableName + ` WHERE messageID=? `
	row := f.DB.QueryRow(q, userIDForReaction, messageID, messageID)

	var userLike sql.NullBool
	err := row.Scan(&likes[model.LIKE], &likes[model.DISLIKE], &userLike)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, -1, model.ErrNoRecord
		}
		return nil, -1, err
	}

	var usersReaction int8
	if userLike.Valid {
		if userLike.Bool {
			usersReaction = int8(model.LIKE)
		} else {
			usersReaction = int8(model.DISLIKE)
		}
	} else {
		usersReaction = -1
	}

	return likes, usersReaction, nil
}

func (f *ForumModel) GetPostLikes(messageID int, userIDForReaction int) ([]int,int8, error) {
	return f.GetLikes(model.POST+"s_likes", messageID,userIDForReaction)
}

func (f *ForumModel) GetCommentLikes(messageID int, userIDForReaction int) ([]int,int8, error) {
	return f.GetLikes(model.COMMENT+"s_likes", messageID,userIDForReaction)
}

/* returns quantity of likes/dislikes from the given table (posts or comments) for the given user and message*/
func (f *ForumModel) getUsersLike(tableName string, userID, messageID int) (int, bool, error) {
	var id int
	var like bool
	q := `SELECT id,like FROM ` + tableName + ` WHERE userID=? AND messageID=?`
	row := f.DB.QueryRow(q, userID, messageID)

	err := row.Scan(&id, &like)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, model.ErrNoRecord
		}
		return 0, false, err
	}

	return id, like, nil
}

func (f *ForumModel) GetUsersPostLike(userID, messageID int) (int, bool, error) {
	return f.getUsersLike(model.POST+"s_likes", userID, messageID)
}

func (f *ForumModel) GetUsersCommentLike(userID, messageID int) (int, bool, error) {
	return f.getUsersLike(model.COMMENT+"s_likes", userID, messageID)
}

/****
the group of function for changing likes (inser, update, delete)
****/
/*inserts a like/dislike to the given table.*/
func (f *ForumModel) insertLike(tableName string, userID, messageID int, like bool) (int, error) {
	q := `INSERT INTO ` + tableName + ` (userID, messageID, like) VALUES (?,?,?)`
	res, err := f.DB.Exec(q, userID, messageID, like)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

/*sets a new value of like/dislike in the given table.*/
func (f *ForumModel) updateLike(tableName string, id int, like bool) error {
	q := `UPDATE ` + tableName + ` SET like=? WHERE id=?`
	res, err := f.DB.Exec(q, like, id)
	if err != nil {
		return err
	}

	return f.checkUnique(res)
}

/*deletes a row from the given table.*/
func (f *ForumModel) deleteLike(tableName string, id int) error {
	q := `DELETE FROM ` + tableName + ` WHERE id=?`
	res, err := f.DB.Exec(q, id)
	if err != nil {
		return err
	}

	return f.checkUnique(res)
}

func (f *ForumModel) InsertPostLike(userID, messageID int, like bool) (int, error) {
	return f.insertLike(model.POST+"s_likes", userID, messageID, like)
}

func (f *ForumModel) UpdatePostLike(id int, like bool) error {
	return f.updateLike(model.POST+"s_likes", id, like)
}

func (f *ForumModel) DeletePostLike(id int) error {
	return f.deleteLike(model.POST+"s_likes", id)
}

func (f *ForumModel) InsertCommentLike(userID, messageID int, like bool) (int, error) {
	return f.insertLike(model.COMMENT+"s_likes", userID, messageID, like)
}

func (f *ForumModel) UpdateCommentLike(id int, like bool) error {
	return f.updateLike(model.COMMENT+"s_likes", id, like)
}

func (f *ForumModel) DeleteCommentLike(id int) error {
	return f.deleteLike(model.COMMENT+"s_likes", id)
}

/*deletes a row from the given table by message ID.*/
func (f *ForumModel) deleteLikeByMessageID(tableName string, id int) error {
	q := `DELETE FROM ` + tableName + ` WHERE messageID=?`
	_, err := f.DB.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}

func (f *ForumModel) DeletePostLikeByMessageID(id int) error {
	return f.deleteLikeByMessageID(model.POST+"s_likes"+"s_likes", id)
}

func (f *ForumModel) DeleteCommentLikeByMessageID(id int) error {
	return f.deleteLikeByMessageID(model.COMMENT+"s_likes", id)
}
