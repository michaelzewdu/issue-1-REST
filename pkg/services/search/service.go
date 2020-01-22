/*
Package search contains definition and implementation of a service that deals searching.*/
package search

import "fmt"

// Service specifies a method to provide searching functionality.
type Service interface {
	SearchComments(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*Comment, error)
}

// Repository specifies a repo interface to serve the search.Service interface
type Repository interface {
	SearchComments(pattern string, sortBy string, sortOrder string, limit, offset int) ([]*Comment, error)
}

// SortOrder holds  that specify how comments are sorted
type SortOrder string

// SortBy holds enums that specify by what attribute comments are sorted
type SortBy string

// Sorting constants used by SearchUser methods
const (
	SortAscending      SortOrder = "ASC"
	SortDescending     SortOrder = "DESC"
	SortByCreationTime SortBy    = "creation_time"
	SortByRank         SortBy    = "rank"
)

type service struct {
	repo *Repository
}

// NewService returns a struct that implements the search.Service interface
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

// SearchComments returns a list of comments based on the given pattern and pagination parameters.
func (s service) SearchComments(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*Comment, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	return (*s.repo).SearchComments(pattern, string(sortBy), string(sortOrder), limit, offset)
}
