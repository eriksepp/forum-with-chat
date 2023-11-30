package sqlpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"forum/logger"
	"forum/model"
)

/*
inserts a new comment into DB, returns an ID for the comment
*/
func (f *ForumModel) InsertComment(postID int, content string, images []string, authorID int, dateCreate time.Time) (int, error) {
	var strOfImages sql.NullString

	if len(images) != 0 {
		strOfImages.String = strings.Join(images, ",")
		strOfImages.Valid = true
	}
	q := `INSERT INTO comments (content, images, authorID, dateCreate, postID) VALUES (?,?,?,?,?)`
	res, err := f.DB.Exec(q, content, strOfImages, authorID, dateCreate, postID)
	if err != nil {
		return 0, err
	}
	commentID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(commentID), nil
}

/*
modify a comment with the given id
*/
func (f *ForumModel) ModifyComment(id int, content string, images []string) error {
	fields := ""
	fieldsValues := []any{}
	if content != "" {
		fields += "content=?, "
		fieldsValues = append(fieldsValues, content)
	}
	if len(images) != 0 {
		fields += "images=?, "
		fieldsValues = append(fieldsValues, strings.Join(images, ","))
	}
	fields, ok := strings.CutSuffix(fields, ", ")
	if !ok {
		panic("cant cut the , after fields list in func modufyPost")
	}
	fieldsValues = append(fieldsValues, id)

	q := fmt.Sprintf("UPDATE comments SET %s WHERE id=?", fields)
	_, err := f.DB.Exec(q, fieldsValues...)
	if err != nil {
		return err
	}

	return nil
}

/*
search in the DB a comment by the given ID returns comment and its postID
*/
func (f *ForumModel) GetCommentByID(id int) (*model.Comment, error) {
	query := `SELECT c.id, c.content, c.images, c.authorID, u.name, u.dateCreate, c.dateCreate, c.postID, 
			count(CASE WHEN cl.like THEN TRUE END), count(CASE WHEN NOT cl.like THEN TRUE END) 
	    FROM comments c
		LEFT JOIN users u ON u.id=c.authorID
	    LEFT JOIN comments_likes cl ON cl.messageID=c.id 
		WHERE c.id = ?		 
		GROUP BY c.id;
		`

	row := f.DB.QueryRow(query, id)
	// get a comment
	comment := &model.Comment{}
	comment.Message.Author = &model.User{}
	comment.Message.Likes = make([]int, model.N_LIKES)
	var images sql.NullString
	// parse the row with fields:
	// c.id, c.content, c.images, c.authorID, u.name, u.dateCreate, c.dateCreate, c.postID,
	// count(CASE WHEN cl.like THEN TRUE END), count(CASE WHEN NOT cl.like THEN TRUE END)
	err := row.Scan(&comment.ID,
		&comment.Message.Content, &images,
		&comment.Message.Author.ID, &comment.Message.Author.Name, &comment.Message.Author.DateCreate,
		&comment.Message.DateCreate, &comment.PostID,
		&comment.Message.Likes[model.LIKE], &comment.Message.Likes[model.DISLIKE],
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}

	comment.Message.Images = getImagesArray(images)

	return comment, nil
}

/*
Delete a post from the database (this is "hard" delete, we should use "soft" delete instead)
*/
func (f *ForumModel) Delete(id int) error {
	_, err := f.DB.Exec("DELETE FROM comments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("comment delete failed: func '%s' : %w", logger.GetCurrentFuncName(), err)
	}
	return nil
}
