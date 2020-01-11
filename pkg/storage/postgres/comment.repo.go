package postgres

import (
	"database/sql"
	"fmt"
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
)

mport (
"database/sql"

"github.com/slim-crown/issue-1-REST/pkg/domain/post"
)

type CommentRepository repository

func NewCommentRepository(DB *sql.DB, allRepos *map[string]interface{}) *CommentRepository {
	return &CommentRepository{DB, allRepos}
}

// GetComment gets the Comment stored under the given id.
func (repo *CommentRepository) GetComment(id int) (*comment.Comment, error) {
	var err error
	var c = new(comment.Comment)
	/*
	ID           int      `json:"id"`
		OriginPost   int      `json:"originpost"`
		Commenter    string    `json:"comment"`
		Content      string    `json:"content"`
		ReplyTo      int      `json:"replyto"`
		CreationTime time.Time `json:"creationtime"`*/

	err = repo.db.QueryRow(`
								SELECT COALESCE(post_id, ''), COALESCE(reply_to, ''), COALESCE(commented_by, ''),creation_time
								FROM comments
								WHERE id = $1`, id).Scan(&c.OriginPost, &c.ReplyTo, &c.Commenter, &c.CreationTime)
	if err != nil {
		return nil, comment.ErrCommentNotFound
	}
	return c,nil

}

// DeletePost Deletes the Post stored under the given id.
func (repo *CommentRepository) DeleteComment(id int) error {
	_,err := repo.db.Exec(`DELETE FROM "issue#1".comments WHERE id = $1`,id)
	if err != nil {
		fmt.Errorf("deletion from comment failed because: %v",err)
	}
	return nil
}

// AddPost Adds the Post stored under the given id.
func (repo *comment.Repository) AddPost(p *post.Post) (*post.Post, error) {
	p, err := (*repo.secondaryRepo).AddPost(p)
	if err == nil {
		repo.cache[p.ID] = *p
	}
	return p, err
}

//UpdatePost updates the post with given id and post struct
func (repo *postRepository) UpdatePost(pos *post.Post, id int) (*post.Post, error) {
	p, err := (*repo.secondaryRepo).UpdatePost(pos, id)
	if err == nil {
		repo.cache[p.ID] = *p
	}
	return p, err
}

// GetPostReleases(p *Post) ([]*Release, error)
// GetPostRelease(pId int, rId int) (*Release, error)

func (repo *postRepository) SearchPost(pattern string, by post.SortBy, order post.SortOrder, limit int, offset int) ([]*post.Post, error) {
	pos, err := (*repo.secondaryRepo).SearchPost(pattern, by, order, limit, offset)
	if err == nil {
		for _, p := range pos {
			repo.cache[p.ID] = *p

		}
	}
	return pos, err
}
func (repo *postRepository) GetPostStar(id int, username string) (*post.Star, error) {
	return (*repo.secondaryRepo).GetPostStar(id, username)
}
func (repo *postRepository) DeletePostStar(id int, username string) error {
	err := (*repo.secondaryRepo).DeletePostStar(id, username)
	if err == nil {
		errs := repo.cachePost(id)
		if errs != nil {
			return errs
		}

	}
	return err
}
func (repo *postRepository) AddPostStar(id int, star *post.Star) (*post.Star, error) {
	s, err := (*repo.secondaryRepo).AddPostStar(id, star)
	if err == nil {
		errs := repo.cachePost(id)
		if errs != nil {
			return s, errs
		}
	}
	return s, err
}
func (repo *postRepository) UpdatePostStar(id int, star *post.Star) (*post.Star, error) {
	s, err := (*repo.secondaryRepo).UpdatePostStar(id, star)
	if err == nil {
		errs := repo.cachePost(id)
		if errs != nil {
			return s, errs
		}
	}
	return s, err
}
