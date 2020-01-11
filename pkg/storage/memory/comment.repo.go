package memory

import (
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
	"strconv"
)

//commentRepository ...
type commentRepository struct {
	cache         map[int]comment.Comment
	secondaryRepo *comment.Comment
	allRepos      *map[string]interface{}
}
func (repo *commentRepository) cacheFeed(id string) error {
	u, err := (*repo.secondaryRepo).GetComment(id)
	if err != nil {
		return err
	}
	idi, _ :=strconv.Atoi(id)
	repo.cache[idi] = *u
	return nil
}

func (repo *commentRepository) GetComment(id int) (*comment.Comment, error) {
}

func (repo *commentRepository) AddComment(comment *comment.Comment) error {
	panic("implement me")
}

func (repo *commentRepository) UpdateComment(comment *comment.Comment, id int) error {
	panic("implement me")
}

func (repo *commentRepository) DeleteComment(id int) error {
	panic("implement me")
}

func (repo *commentRepository) GetReply(id int) (*comment.Comment, error) {
	panic("implement me")
}

func (repo *commentRepository) AddReply(comment *comment.Comment) error {
	panic("implement me")
}

func (repo *commentRepository) UpdateReply(comment *comment.Comment, id int) error {
	panic("implement me")
}

func (repo *commentRepository) DeleteReply(id int) error {
	panic("implement me")
}

// NewCommentRepository returns a struct that implements the comment.Repository using
// a cached based implementation.
// A database implementation of the same interface needs to be passed so that it can be
// consulted when the caches aren't enough.
func NewCommentRepository(secondaryRepo *comment.Repository, allRepos *map[string]interface{}) comment.Repository {
	return &commentRepository{cache: make(map[string]comment.Comment), secondaryRepo: secondaryRepo, allRepos: allRepos}
}

/

