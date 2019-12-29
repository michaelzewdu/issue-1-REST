/*
Package postgres contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package postgres

import (
	"database/sql"
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
	db       *sql.DB
	allRepos *map[string]interface{}
}

// ChannelRepository ...
type ChannelRepository repository

// CommentRepository ...
type CommentRepository repository

// ReleaseRepository ...
type ReleaseRepository repository
