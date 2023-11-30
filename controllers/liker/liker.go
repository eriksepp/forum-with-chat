package liker

import (
	"errors"
	"fmt"

	"forum/model"
	"forum/model/sqlpkg"
	"forum/wsmodel"
)

type LikePost struct {
	ID        int  `json:"id,omitempty"`
	User    *model.User  `json:"userID"`
	MessageID int  `json:"messageID"`
	Reaction  bool `json:"reaction"`
}

type LikeComment struct {
	ID        int  `json:"id,omitempty"`
	User    *model.User  `json:"userID"`
	MessageID int  `json:"messageID"`
	Reaction  bool `json:"reaction"`
}

type Liker interface {
	GetLike(*sqlpkg.ForumModel) error
	InsertLike(*sqlpkg.ForumModel, bool) error
	UpdateLike(*sqlpkg.ForumModel, bool) error
	DeleteLike(*sqlpkg.ForumModel) error
	CompareLike(bool) bool
	GetLikesNumbers(*sqlpkg.ForumModel) (LikesNumbers, error)
}

type LikesNumbers struct{
	Likes int `json:"likes"`
	Dislikes int `json:"dislikes"`
	UserWithReaction *model.User `json:"userWithReaction"`
	UserReaction int8 `json:"userReactions"`
}

func NewLikeComment(user *model.User, reactData wsmodel.Reaction) *LikeComment {
	var lc LikeComment
	lc.User = &model.User{ID: user.ID, Name: user.Name} 
	lc.MessageID = reactData.MessageID
	lc.Reaction = reactData.Reaction
	return &lc
}

func NewLikePost(user *model.User, reactData wsmodel.Reaction) *LikePost {
	var pc LikePost
	pc.User = &model.User{ID: user.ID, Name: user.Name} 
	pc.MessageID = reactData.MessageID
	pc.Reaction = reactData.Reaction
	return &pc
}

func (pl *LikePost) GetLike(db *sqlpkg.ForumModel) error {
	var err error
	pl.ID, pl.Reaction, err = db.GetUsersPostLike(pl.User.ID, pl.MessageID)
	return err
}

func (pl *LikePost) GetLikesNumbers(db *sqlpkg.ForumModel) (LikesNumbers, error) {
	var likesNum LikesNumbers
	likes, userReaction,err:= db.GetPostLikes( pl.MessageID, pl.User.ID)
	if err != nil {
		return likesNum, err
	}
	likesNum.Dislikes = likes[model.DISLIKE]
	likesNum.Likes = likes[model.LIKE]
	likesNum.UserReaction = userReaction
	likesNum.UserWithReaction = pl.User
	return likesNum, nil
}

func (pl *LikePost) InsertLike(db *sqlpkg.ForumModel, like bool) error {
	var err error
	pl.Reaction = like
	pl.ID, err = db.InsertPostLike(pl.User.ID, pl.MessageID, pl.Reaction)
	return err
}

func (pl *LikePost) UpdateLike(db *sqlpkg.ForumModel, like bool) error {
	pl.Reaction = like
	return db.UpdatePostLike(pl.ID, pl.Reaction)
}

func (pl *LikePost) DeleteLike(db *sqlpkg.ForumModel) error {
	return db.DeletePostLike(pl.ID)
}

func (pl *LikePost) CompareLike(like bool) bool {
	return pl.Reaction == like
}

func (cl *LikeComment) GetLike(db *sqlpkg.ForumModel) error {
	var err error
	cl.ID, cl.Reaction, err = db.GetUsersCommentLike(cl.User.ID, cl.MessageID)
	return err
}

func (cl *LikeComment) GetLikesNumbers(db *sqlpkg.ForumModel) (LikesNumbers, error) {
	var likesNum LikesNumbers
	likes, userReaction,err:= db.GetCommentLikes( cl.MessageID, cl.User.ID)
	if err != nil {
		return likesNum, err
	}
	likesNum.Dislikes = likes[model.DISLIKE]
	likesNum.Likes = likes[model.LIKE]
	likesNum.UserReaction = userReaction
	likesNum.UserWithReaction = cl.User
	return likesNum, nil
}

func (cl *LikeComment) InsertLike(db *sqlpkg.ForumModel, like bool) error {
	var err error
	cl.Reaction = like
	cl.ID, err = db.InsertCommentLike(cl.User.ID, cl.MessageID, cl.Reaction)
	return err
}

func (cl *LikeComment) UpdateLike(db *sqlpkg.ForumModel, like bool) error {
	cl.Reaction = like
	return db.UpdateCommentLike(cl.ID, cl.Reaction)
}

func (cl *LikeComment) DeleteLike(db *sqlpkg.ForumModel) error {
	return db.DeleteCommentLike(cl.ID)
}

func (cl *LikeComment) CompareLike(like bool) bool {
	return cl.Reaction == like
}

func SetLike(db *sqlpkg.ForumModel, liker Liker, newLike bool) error {
	err := liker.GetLike(db)
	if err != nil {
		// if there is no like/dislike made by the user, add a new one
		if errors.Is(err, model.ErrNoRecord) {
			err := liker.InsertLike(db, newLike)
			if err != nil {
				return fmt.Errorf("insert data to DB failed: %s", err)
			}
		} else {
			return fmt.Errorf("getting data from DB failed: %s", err)
		}
	} else {
		if liker.CompareLike(newLike) { // if it is the same like, delete it
			err := liker.DeleteLike(db)
			if err != nil {
				return fmt.Errorf("deleting data from DB failed: %s", err)
			}
		} else {
			err := liker.UpdateLike(db, newLike)
			if err != nil {
				return fmt.Errorf("updating data in DB failed: %s", err)
			}
		}
	}
	return nil
}
