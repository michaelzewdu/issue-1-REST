package memory

import (
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
)

//commentRepository ...
type commentRepository struct {
	cache         map[int]comment.Comment
	secondaryRepo *comment.Comment
	allRepos      *map[string]interface{}
}

// NewCommentRepository returns a struct that implements the comment.Repository using
func NewCommentRepository(secondaryRepo *comment.Repository, allRepos *map[string]interface{}) comment.Repository {
	return &comment.Repository{cache: make(map[string]feed.Feed), secondaryRepo: secondaryRepo, allRepos: allRepos}
}

// GetComment returns the feed belonging to the given username from either the
// cache if found there or the secondary repos it wraps.
func (repo *commentRepository) GetComment(username string) (*comment.Comment, error) {
	if _, ok := repo.cache[username]; ok == false {
		err := repo.cacheFeed(username)
		if err != nil {
			return nil, err
		}
	}
	user := repo.cache[username]
	return &user, nil
}
