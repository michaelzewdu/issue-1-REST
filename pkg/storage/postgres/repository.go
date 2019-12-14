/*
Package postgres contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package postgres

import (
	"database/sql"

	"github.com/slim-crown/Issue-1/pkg/domain/channel"
	"github.com/slim-crown/Issue-1/pkg/domain/user"
)

/*
//DBHandler ...
type DBHandler interface {
	Execute(statment string) error
	Query(query string) Row
}

// Row ...
type Row interface {
	Scan(dest ...interface{})
	Next() bool
}
*/

type repository struct {
	dbConnection *sql.DB
	allRepos     *map[string]interface{}
}

// UserRepository ...
type UserRepository repository

// ChannelRepository ...
type ChannelRepository repository

// CommentRepository ...
type CommentRepository repository

// FeedRepository ...
type FeedRepository repository

// ReleaseRepository ...
type ReleaseRepository repository

// NewUserRepository ...
func NewUserRepository(dbConnection *sql.DB, allRepos *map[string]interface{}) *UserRepository {
	return &UserRepository{dbConnection, allRepos}
}

// AddUser ...
func (repo *UserRepository) AddUser(user *user.User) error {
	return nil
}

// NewChannelRepository ...
func NewChannelRepository(dbConnection *sql.DB, allRepos *map[string]interface{}) *ChannelRepository {
	return &ChannelRepository{dbConnection, allRepos}
}

// AddChannel ...
func (repo *ChannelRepository) AddChannel(channel *channel.Channel) error {
	return nil
}
