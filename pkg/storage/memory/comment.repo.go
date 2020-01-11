package memory

import (
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
)

type commentRepository struct {
	cache         map[int]comment.Comment
	secondaryRepo *comment.Repository
}

func NewRepository(secondaryRepo *comment.Repository) comment.Repository {
	return &commentRepository{make(map[int]comment.Comment), secondaryRepo}
}

func (repo *commentRepository) AddComment(c *comment.Comment) (*comment.Comment, error) {
	c, err := (*repo.secondaryRepo).AddComment(c)
	if err == nil {
		repo.cache[c.ID] = *c
	}
	return c, err
}

func (repo *commentRepository) GetComment(id int) (*comment.Comment, error) {
	if c, ok := repo.cache[id]; ok == false {
		var err error
		c, err := (*repo.secondaryRepo).GetComment(id)
		if err != nil {
			return nil, err
		}
		repo.cache[id] = *c
		return c, nil
	} else {
		return &c, nil
	}
}

func (repo *commentRepository) GetComments(postID int, by string, order string, limit, offset int) ([]*comment.Comment, error) {
	result, err := (*repo.secondaryRepo).GetComments(postID, by, order, limit, offset)
	if err == nil {
		for _, c := range result {
			repo.cache[c.ID] = *c
		}
	}
	return result, err
}

func (repo *commentRepository) GetReplies(commentID int, by string, order string, limit, offset int) ([]*comment.Comment, error) {
	result, err := (*repo.secondaryRepo).GetReplies(commentID, by, order, limit, offset)
	if err == nil {
		for _, c := range result {
			repo.cache[c.ID] = *c
		}
	}
	return result, err
}

func (repo *commentRepository) UpdateComment(c *comment.Comment) (*comment.Comment, error) {
	c, err := (*repo.secondaryRepo).UpdateComment(c)
	if err == nil {
		repo.cache[c.ID] = *c
	}
	return c, err
}

func (repo *commentRepository) DeleteComment(id int) error {
	err := (*repo.secondaryRepo).DeleteComment(id)
	if err == nil {
		delete(repo.cache, id)
	}
	return err
}
