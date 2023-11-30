package sqlpkg

import (
	"fmt"
	"testing"
	"time"
)

func TestInsertComment(t *testing.T) {
	db, err := OpenDB(DBPath, "webuser", "webuser")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f := ForumModel{db}

	id, err := f.InsertComment(2, "comment 2 tto post 2", []string{}, 1, time.Date(2023, time.March, 8, 12, 12, 21, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("---id=%d-------\n", id)
}
