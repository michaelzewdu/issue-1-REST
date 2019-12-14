/*
Package memory contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package memory

import (
	"github.com/slim-crown/Issue-1/pkg/domain/channel"
	"github.com/slim-crown/Issue-1/pkg/domain/comment"
	"github.com/slim-crown/Issue-1/pkg/domain/feed"
	"github.com/slim-crown/Issue-1/pkg/domain/release"
	"github.com/slim-crown/Issue-1/pkg/domain/user"
)

// UserRepository ...
type UserRepository struct {
	cache      []User
	repository *user.Repository
	allRepos   *map[string]interface{}
}

// ChannelRepository ...
type ChannelRepository struct {
	cache      []Channel
	repository *channel.Repository
	allRepos   *map[string]interface{}
}

// CommentRepository ...
type CommentRepository struct {
	cache      []Comment
	repository *comment.Repository
	allRepos   *map[string]interface{}
}

// FeedRepository ...
type FeedRepository struct {
	cache      []Feed
	repository *feed.Repository
	allRepos   *map[string]interface{}
}

// ReleaseRepository ...
type ReleaseRepository struct {
	cache      []Release
	repository *release.Repository
	allRepos   *map[string]interface{}
}

// NewUserRepository ...
func NewUserRepository(dbRepo *user.Repository, allRepos *map[string]interface{}) *UserRepository {
	return &UserRepository{make([]User, 100), dbRepo, allRepos}
}

// AddUser ...
func (repo *UserRepository) AddUser(user *user.User) error {
	return nil
}

// NewChannelRepository ...
func NewChannelRepository(dbRepo *channel.Repository, allRepos *map[string]interface{}) *ChannelRepository {
	return &ChannelRepository{make([]Channel, 100), dbRepo, allRepos}
}

// AddChannel ...
func (repo *ChannelRepository) AddChannel(channel *channel.Channel) error {
	return nil
}
