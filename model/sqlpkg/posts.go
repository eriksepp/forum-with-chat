package sqlpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"forum/model"
)

/*
inserts a new post into DB, returns an ID for the post
*/
func (f *ForumModel) InsertPost(theme, content string, images []string, authorID int, dateCreate time.Time, categoriesID []int) (int, error) {
	var strOfImages sql.NullString
	if len(images) != 0 {
		strOfImages.String = strings.Join(images, ",")
		strOfImages.Valid = true
	}
	q := `INSERT INTO posts (theme, content, images, authorID, dateCreate) VALUES (?,?,?,?,?)`
	res, err := f.DB.Exec(q, theme, content, strOfImages, authorID, dateCreate)
	if err != nil {
		return 0, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	q = `INSERT INTO post_categories (categoryID, postID) VALUES (?,?)`
	for i := 1; i < len(categoriesID); i++ {
		q += `,(?,?)`
	}
	insertData := make([]any, 2*len(categoriesID))
	for i := 0; i < len(categoriesID); i++ {
		insertData[2*i] = categoriesID[i]
		insertData[2*i+1] = int(postID)
	}
	res, err = f.DB.Exec(q, insertData...)
	if err != nil {
		return 0, err
	}

	_, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(postID), nil
}

/*
search in the DB a post by the given ID
*/
func (f *ForumModel) GetPostByID(id int, userID int) (*model.Post, error) {
	query := `SELECT p.id, p.theme, p.content, p.images, p.authorID, u.name, u.dateCreate, c.id, c.name,  p.dateCreate, 
				 count(CASE WHEN pl.like THEN TRUE END), count(CASE WHEN NOT pl.like THEN TRUE END), 
				 (CASE WHEN p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=true)  THEN 1
				      WHEN p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=false) THEN 0
					  ELSE -1 END)
			  FROM posts p
 			  LEFT JOIN users u ON u.id=p.authorID
			  LEFT JOIN post_categories pc ON pc.postID=p.id
			  LEFT JOIN categories c ON c.id=pc.categoryID
			  LEFT JOIN posts_likes pl ON pl.messageID=p.id 
			  WHERE p.id = ?		 
			  GROUP BY c.id;
		`

	// exequting the query
	var rows *sql.Rows
	var err error
	rows, err = f.DB.Query(query, userID, userID, id)
	if err != nil {
		return nil, err
	}

	// parsing the query's result
	var post *model.Post
	var category *model.Category

	// add the first post without condition
	if rows.Next() {
		post, category, err = rowScanForPostByID(rows)
		if err != nil {
			return nil, err
		}

		post.Categories = append(post.Categories, category)
	} else {
		return nil, model.ErrNoRecord
	}

	for rows.Next() {
		// add categories only
		postTmp, categoryTmp, err := rowScanForPostByID(rows)
		if err != nil {
			return nil, err
		}

		if postTmp.ID != post.ID {
			return nil, fmt.Errorf("select failed: two different posts by one ID: '%d', '%d'", post.ID, postTmp.ID)
		}
		post.Categories = append(post.Categories, categoryTmp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	query = `-- select comments.
		SELECT c.id, c.content, c.images, c.authorID, u.name, u.dateCreate, c.dateCreate, 
			count(CASE WHEN cl.like THEN TRUE END), count(CASE WHEN NOT cl.like THEN TRUE END), 
			(CASE WHEN c.id IN (SELECT messageID FROM comments_likes cl  WHERE cl.userID = ? AND cl.like=true)  THEN 1
				      WHEN c.id IN (SELECT messageID FROM comments_likes cl  WHERE cl.userID = ? AND cl.like=false) THEN 0
					  ELSE -1 END)
	    FROM comments c
		LEFT JOIN users u ON u.id=c.authorID
	    LEFT JOIN comments_likes cl ON cl.messageID=c.id 
		WHERE c.postID = ?		 
		GROUP BY c.id;
		`
	rows, err = f.DB.Query(query, userID, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}
	defer rows.Close()

	// get comments
	for rows.Next() {
		comment := &model.Comment{}
		comment.Message.Author = &model.User{}
		comment.Message.Likes = make([]int, model.N_LIKES)
		var images sql.NullString

		// parse the row with fields:
		// c.id, c.content, c.images, c.authorID, u.name, u.dateCreate, c.dateCreate,
		// count(CASE WHEN cl.like THEN TRUE END), count(CASE WHEN NOT cl.like THEN TRUE END)
		// (CASE WHEN p.id IN (SELECT messageID FROM comments_likes cl  WHERE cl.userID = ? AND cl.like=true)  THEN 1
		// 		      WHEN p.id IN (SELECT messageID FROM comments_likes cl  WHERE cl.userID = ? AND cl.like=false) THEN 0
		// 			  ELSE -1 END)
		err := rows.Scan(&comment.ID,
			&comment.Message.Content, &images,
			&comment.Message.Author.ID, &comment.Message.Author.Name, &comment.Message.Author.DateCreate,
			&comment.Message.DateCreate,
			&comment.Message.Likes[model.LIKE], &comment.Message.Likes[model.DISLIKE],
			&comment.Message.UserReaction,
		)
		if err != nil {
			return nil, err
		}
		comment.Message.Images = getImagesArray(images)
		post.Comments = append(post.Comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	post.CommentsQuantity = len(post.Comments)

	return post, nil
}

/*
convert string containing a list of images file names to an array
*/
func getImagesArray(imagesStr sql.NullString) []string {
	if imagesStr.Valid {
		imagesNames := strings.Split(imagesStr.String, ",")
		for i := 0; i < len(imagesNames)-1; i++ {
			if imagesNames[i] == "" {
				imagesNames = append(imagesNames[:i], imagesNames[i+1])
			}
		}
		if imagesNames[len(imagesNames)-1] == "" {
			imagesNames = imagesNames[:len(imagesNames)-1]
		}
		return imagesNames
	}
	return nil
}

/*
scan and prefilles an item of modelPost for getPostByID
*/
func rowScanForPostByID(rows *sql.Rows) (*model.Post, *model.Category, error) {
	post := &model.Post{}
	post.Message.Likes = make([]int, model.N_LIKES)
	post.Message.Author = &model.User{}
	category := &model.Category{}
	var images sql.NullString

	// parse the row with fields:
	// p.id, p.theme, p.content, p.images, p.authorID, u.name, u.dateCreate, c.id, c.name,  p.dateCreate,
	// count(CASE WHEN pl.like THEN TRUE END), count(CASE WHEN NOT pl.like THEN TRUE END)
	// (CASE WHEN p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=true)  THEN 1
	// 			      WHEN p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=false) THEN 0
	// 				  ELSE -1 END)
	err := rows.Scan(&post.ID, &post.Theme,
		&post.Message.Content, &images,
		&post.Message.Author.ID, &post.Message.Author.Name, &post.Message.Author.DateCreate,
		&category.ID, &category.Name,
		&post.Message.DateCreate,
		&post.Message.Likes[model.LIKE], &post.Message.Likes[model.DISLIKE],
		&post.Message.UserReaction,
	)

	post.Message.Images = getImagesArray(images)

	return post, category, err
}

/*
returns 'postNumbers' posts with ids less than 'beforeId' and matching the filter.
The posts are sorted by date of creation in descending order.
The parametr userID is used to mark the user's reaction to a post.
If beforeId and/or postNumbers is less or equal to 0 it will ignore the conditions.
*/
func (f *ForumModel) GetPosts(beforeId int, postNumbers int, filter *model.Filter, userIDForReaction int) ([]*model.Post, error) {
	condition := ""
	var arguments []any
	v := reflect.ValueOf(*filter)
	for _, field := range reflect.VisibleFields(reflect.TypeOf(*filter)) {
		// if either of the fields !=0 add conditions to the query
		if !v.FieldByIndex(field.Index).IsZero() {
			condition = ` WHERE `
			arguments = []any{}
			if filter.AuthorID != 0 {
				condition += ` p.authorID= ? AND `
				arguments = append(arguments, filter.AuthorID)
			}

			if len(filter.CategoryID) != 0 {

				condition += ` p.id IN (SELECT postID FROM post_categories pc  WHERE `
				for _, c := range filter.CategoryID {
					condition += ` pc.categoryID = ? OR `
					arguments = append(arguments, c)
				}
				condition = strings.TrimSuffix(condition, `OR `)
				condition += `GROUP BY pc.postID) AND `
			}

			if filter.LikedByUserID != 0 {
				condition += ` p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=true) AND `
				arguments = append(arguments, filter.LikedByUserID)
			}

			if filter.DisLikedByUserID != 0 {
				condition += ` p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=false) AND `
				arguments = append(arguments, filter.DisLikedByUserID)
			}

			condition = strings.TrimSuffix(condition, `AND `)
			break
		}
	}

	if beforeId > 0 {
		if condition == "" {
			condition = `WHERE `
		} else {
			condition += " AND "
		}
		condition += `p.id < ?`
		arguments = append(arguments, beforeId)
	}

	return f.getPostsByCondition(postNumbers, condition, arguments, userIDForReaction)
}

/*
returns posts that have the given category
*/
func (f *ForumModel) GetPostsByCategory(beforeId int, postNumbers int, category int, userIDForReaction int) ([]*model.Post, error) {
	condition := ` WHERE p.id IN (SELECT postID FROM post_categories pc  WHERE pc.categoryID = ?) `
	arguments := []any{category}

	if beforeId > 0 {
		condition += ` AND p.id < ?`
		arguments = append(arguments, beforeId)
	}

	return f.getPostsByCondition(postNumbers, condition, arguments, userIDForReaction)
}

/*
returns posts that have got the given category
*/
func (f *ForumModel) GetPostsLikedByUser(beforeId int, postNumbers int, userID int) ([]*model.Post, error) {
	condition := ` WHERE p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=true) `
	arguments := []any{userID}

	if beforeId > 0 {
		condition += ` AND p.id < ?`
		arguments = append(arguments, beforeId)
	}

	return f.getPostsByCondition(postNumbers, condition, arguments, userID)
}

/*
addes the condition to a query and run it. Returnes found posts
*/
func (f *ForumModel) getPostsByCondition(postNumbers int, condition string, argumentsForCondition []any, userID int) ([]*model.Post, error) {
	query := `SELECT p.id, p.theme, p.content, p.images, p.authorID, u.name, u.dateCreate, c.id, c.name,  p.dateCreate, 
				(SELECT count(id) FROM comments cm WHERE cm.postID=p.id),
				count(CASE WHEN pl.like THEN TRUE END), count(CASE WHEN NOT pl.like THEN TRUE END),
				(CASE WHEN ul.like is NULL THEN -1 WHEN ul.like THEN 1 WHEN NOT  ul.like THEN 0 END)
				
		  FROM posts p
		  LEFT JOIN users u ON u.id=p.authorID
		  LEFT JOIN post_categories pc ON pc.postID=p.id
		  LEFT JOIN categories c ON c.id=pc.categoryID
		  LEFT JOIN posts_likes pl ON pl.messageID=p.id 
		  LEFT JOIN (SELECT messageID, like FROM posts_likes pl  WHERE pl.userID = ?) ul ON ul.messageID=p.id 
		` + condition +
		` GROUP BY p.id, c.id 
		ORDER BY p.dateCreate DESC, p.id, c.id
		`
	// exequting the query
	var rows *sql.Rows
	var err error
	rows, err = f.DB.Query(query, append([]any{userID}, argumentsForCondition...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// parsing the query's result
	var posts []*model.Post
	authors := make(map[int]*model.User)
	postCounter := 0 // the number of the last added post

	// add the first post without condition
	if rows.Next() {
		post, category, author, err := scanRowForPosts(rows)
		if err != nil {
			return nil, err
		}
		addNewPostStruct(&posts, post, category, author, authors)
	}

	for rows.Next() {
		post, category, author, err := scanRowForPosts(rows)
		if err != nil {
			return nil, err
		}

		// found out do we need to add a new post or to add a category to the previouse post
		// if the next row contains the same postID not create new post, just add a category to the u post
		if post.ID == posts[postCounter].ID {
			posts[postCounter].Categories = append(posts[postCounter].Categories, category)
		} else {
			postCounter++
			if postCounter == postNumbers {
				break
			}
			addNewPostStruct(&posts, post, category, author, authors)
		}
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoRecord
		}
		return nil, err
	}

	return posts, nil
}

/*
scans and prefilles an item of modelPost for getPosts
*/
func scanRowForPosts(rows *sql.Rows) (*model.Post, *model.Category, *model.User, error) {
	post := &model.Post{}
	post.Message.Likes = make([]int, model.N_LIKES)
	author := &model.User{}
	category := &model.Category{}
	var images sql.NullString

	// parse the row with fields:
	// p.id, p.theme, p.content,  p.images, p.authorID, u.name, u.dateCreate, c.id, c.name,  p.dateCreate, count (cm.id),
	// count(CASE WHEN pl.like THEN TRUE END), count(CASE WHEN NOT pl.like THEN TRUE END)
	// (CASE WHEN p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=true)  THEN 1
	// 	     WHEN p.id IN (SELECT messageID FROM posts_likes pl  WHERE pl.userID = ? AND pl.like=false) THEN 0
	// 		 ELSE -1 END)
	err := rows.Scan(&post.ID, &post.Theme, &post.Message.Content, &images,
		&author.ID, &author.Name, &author.DateCreate,
		&category.ID, &category.Name,
		&post.Message.DateCreate,
		&post.CommentsQuantity,
		&post.Message.Likes[model.LIKE], &post.Message.Likes[model.DISLIKE],
		&post.Message.UserReaction,
	)
	post.Message.Images = getImagesArray(images)

	return post, category, author, err
}

/*
creates an item of modelPost type and addes it to the slice. Used in the getPostsByCondition function
*/
func addNewPostStruct(posts *[]*model.Post, post *model.Post, category *model.Category, author *model.User, authors map[int]*model.User) {
	post.Categories = append(post.Categories, category)

	// find out if the author in the current row is found before, if yes, keep that previouse one
	if existingAuthor, ok := authors[author.ID]; ok {
		post.Message.Author = existingAuthor
	} else {
		post.Message.Author = author
		authors[author.ID] = author
	}
	*posts = append(*posts, post)
}

/*
modify a post with the given id
*/
func (f *ForumModel) ModifyPost(id int, theme, content string, images []string) error {
	fields := ""
	fieldsValues := []any{}
	if theme != "" {
		fields += "theme=?, "
		fieldsValues = append(fieldsValues, theme)
	}
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

	q := fmt.Sprintf("UPDATE posts SET %s WHERE id=?", fields)
	_, err := f.DB.Exec(q, fieldsValues...)
	if err != nil {
		return err
	}

	return nil
}
