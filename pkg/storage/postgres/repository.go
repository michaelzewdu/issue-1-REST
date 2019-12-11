/*
Package postgres contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package postgres

import (
	"github.com/slim-crown/Issue-1/pkg/listing"
)

// DBHandler ...
type DBHandler interface {
	Query(query string)
}

// Repository ...
type Repository struct {
	dbhandler DBHandler
}

// NewRepository ...
func NewRepository() (*Repository, error) {
	return nil, nil
}

// FindUser ...
func (repo *Repository) FindUser(username string) (*listing.User, error) {
	return nil, nil
}
