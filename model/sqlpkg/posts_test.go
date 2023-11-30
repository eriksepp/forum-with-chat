package sqlpkg

import (
	"fmt"
	"testing"
	"time"

	"forum/model"
)

func TestDeletePost(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f := ForumModel{db}

	q := `DELETE FROM posts  WHERE id=3`
	res, err := f.DB.Exec(q)
	if err != nil {
		t.Fatal(err)
	}

	postID, err := res.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("---id=%d-------\n", postID)
}

func TestInsertPost(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f := ForumModel{db}

	id, err := f.InsertPost("theme1", "it's content1", []string{}, 1, time.Date(2023, time.March, 7, 12, 12, 21, 0, time.UTC), []int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("---id=%d-------\n", id)
}

func TestGetPostsByCondition(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("------------check reaction of user id 3 --------------------")
	posts, err := f.getPostsByCondition(30, "", nil, 3)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}
	for _, post := range posts {
		if post.Message.UserReaction != -1 {
			t.Fatalf("for the user with id=3: it is expected to be no reaction, ther is a reaction to the post with id=%d", post.ID)
		}
	}

	fmt.Println("------------check reaction of user id 2 --------------------")
	posts, err = f.getPostsByCondition(20, "", nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}
	for _, post := range posts {
		if !(post.ID == 2 || post.ID == 3 || post.ID == 4) && post.Message.UserReaction != -1 {
			t.Fatalf("for the user with id=2: it is expected to be no reaction for the all posts except for id=2, 3 or 4 , ther is a reaction to the post with id=%d", post.ID)
		}
		if post.ID == 2 && post.Message.UserReaction != 1 {
			t.Fatalf("for the user with id=2: it is expected like (reaction = 1 ) for the posts  id=2, ther is a reaction %d to the post", post.Message.UserReaction)
		}
		if (post.ID == 3 || post.ID == 4) && post.Message.UserReaction != 0 {
			t.Fatalf("for the user with id=2: it is expected dislike (reaction = 0 ) for the posts  id=3 or 4, ther is a reaction %d to the post id %d", post.Message.UserReaction, post.ID)
		}
	}
}

func BenchmarkGetPostsByCondition(b *testing.B) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	posts, err := f.getPostsByCondition(20, "", nil, 3)
	if err != nil {
		b.Fatal(err)
	}

	for _, post := range posts {
		if post.Message.UserReaction != -1 {
			b.Fatalf("for the user with id=3: it is expected to be no reaction, ther is a reaction to the post with id=%d", post.ID)
		}
	}

	posts, err = f.getPostsByCondition(20, "", nil, 2)
	if err != nil {
		b.Fatal(err)
	}

	for _, post := range posts {
		if !(post.ID == 2 || post.ID == 3 || post.ID == 4) && post.Message.UserReaction != -1 {
			b.Fatalf("for the user with id=2: it is expected to be no reaction for the all posts except for id=2, 3 or 4 , ther is a reaction to the post with id=%d", post.ID)
		}
		if post.ID == 2 && post.Message.UserReaction != 0 {
			b.Fatalf("for the user with id=2: it is expected like (reaction = 0 ) for the posts  id=2, ther is a reaction %d to the post", post.Message.UserReaction)
		}
		if (post.ID == 3 || post.ID == 4) && post.Message.UserReaction != 1 {
			b.Fatalf("for the user with id=2: it is expected dislike (reaction = 1 ) for the posts  id=3 or 4, ther is a reaction %d to the post id %d", post.Message.UserReaction, post.ID)
		}
	}
}

func TestGetPostsFilters(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("--get posts--")

	filter := &model.Filter{
		AuthorID:      0,
		CategoryID:    nil,
		LikedByUserID: 0,
	}
	posts, err := f.GetPosts(-1, 20, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts author 2--")
	filter = &model.Filter{
		AuthorID:      2,
		CategoryID:    nil,
		LikedByUserID: 0,
	}
	posts, err = f.GetPosts(-1, 20, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts category 2--")
	filter = &model.Filter{
		AuthorID:      0,
		CategoryID:    []int{2},
		LikedByUserID: 0,
	}
	posts, err = f.GetPosts(-1, 20, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts liked by user 1--")
	filter = &model.Filter{
		AuthorID:      0,
		CategoryID:    nil,
		LikedByUserID: 1,
	}
	posts, err = f.GetPosts(-1, 20, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}
	fmt.Println("--get posts by author 1 category 2--")
	filter = &model.Filter{
		AuthorID:      1,
		CategoryID:    []int{2},
		LikedByUserID: 0,
	}
	posts, err = f.GetPosts(-1, 20, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts liked by user 1 category 2--")
	filter = &model.Filter{
		AuthorID:      0,
		CategoryID:    []int{2},
		LikedByUserID: 1,
	}
	posts, err = f.GetPosts(-1, 20, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts category 1,2--")
	filter = &model.Filter{
		AuthorID:      0,
		CategoryID:    []int{1, 2},
		LikedByUserID: 0,
	}
	posts, err = f.GetPosts(-1, 20, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}
}

func TestGetPostsNumberPosts(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	filter := &model.Filter{
		AuthorID:      0,
		CategoryID:    nil,
		LikedByUserID: 0,
	}

	fmt.Println("--get posts id <10 5posts--")
	posts, err := f.GetPosts(10, 5, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get postsid <3 5posts: must be 2 posts--")
	posts, err = f.GetPosts(3, 5, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get all posts (beforeId=0, postNumbers=0)--")
	posts, err = f.GetPosts(0, 0, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts beforeId 0 (from the last one) 5 posts--")
	posts, err = f.GetPosts(0, 5, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get the all posts after id 10 (postNumbers=0)--")
	posts, err = f.GetPosts(10, 0, filter, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}
}

func TestGetPostsByCategory(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("--get posts cat 2--")

	posts, err := f.GetPostsByCategory(-1, 20, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts cat 0--")

	posts, err = f.GetPostsByCategory(-1, 20, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}
}

func TestGetPostsLikedByUser(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("--get liked by user 1 --")

	posts, err := f.GetPostsLikedByUser(-1, 20, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get posts liked by user 2 --")

	posts, err = f.GetPostsLikedByUser(-1, 20, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}

	fmt.Println("--get liked by user 0 --")

	posts, err = f.GetPostsLikedByUser(-1, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, post := range posts {
		fmt.Printf("%s\n", post.String())
	}
}

func TestGetPostByDI(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	f := ForumModel{db}

	fmt.Println("--get post 1--")

	post, err := f.GetPostByID(1, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", post.String())

	fmt.Println("--get post 3--")

	post, err = f.GetPostByID(3, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", post.String())
}
