package sqlpkg

import (
	"fmt"
	"testing"
)

func TestLikes(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f := ForumModel{db}

	fmt.Println("--insert 2,5 - true --")
	id, err := f.InsertPostLike(2, 5, true)
	if err != nil {
		t.Fatal(err)
	}
	err = f.printLikes("posts_likes")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("--update 2,5->false--")
	err = f.UpdatePostLike(id, false)
	if err != nil {
		t.Fatal(err)
	}
	err = f.printLikes("posts_likes")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("--delete 2,5--")
	err = f.DeletePostLike(id)
	if err != nil {
		t.Fatal(err)
	}
	err = f.printLikes("posts_likes")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetLikes(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f := ForumModel{db}

	fmt.Println("--get likes for post 3 userReaction for user 2--")
	likes, usReact, err := f.GetPostLikes(3, 2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("likes: %v,  user id 2 react: %d \n", likes, usReact)

	fmt.Println("--get likes for post 3 by user 2--")
	id, like, err := f.GetUsersPostLike(2, 3)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("id= %d, like=%v\n", id, like)
	/*
		fmt.Println("--get likes for post 1 by user 2 (no)--")
		id,like, err = f.GetUsersPostLike(2,1)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("id= %d, like=%v\n",id, like)
	*/
	fmt.Println("--get likes for post 1 by user 3 (no)--")
	id, like, err = f.GetUsersPostLike(3, 1)
	fmt.Printf("err=%v, type err - %T\n", err, err)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("id= %d, like=%v\n", id, like)
}

func (f *ForumModel) printLikes(table string) error {
	q := `SELECT * FROM ` + table

	rows, err := f.DB.Query(q, table)
	if err != nil {
		return err
	}
	defer rows.Close()
	fmt.Printf("--id--\t--userID--\t--messageID--\t--like--\t\n")
	for rows.Next() {
		var id, userID, postID int
		var like bool
		err := rows.Scan(&id, &userID, &postID, &like)
		if err != nil {
			return err
		}
		fmt.Printf("  %d  \t    %d   \t    %d    \t\t  %v\n", id, userID, postID, like)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
