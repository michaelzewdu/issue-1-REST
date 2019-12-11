/*
Package memory contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package memory

import (
	"github.com/slim-crown/Issue-1/pkg/listing"
)

// Repository ...
type Repository struct {
	userCache []User
}

// NewRepository ...
func NewRepository() (*Repository, error) {
	return nil, nil
}

// FindUser ...
func (repo *Repository) FindUser(username string) (*listing.User, error) {
	return nil, nil
}
