/*
Package comment contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package comment

// Service ...
type Service interface {
	AddComment(comment *Comment) error
}

// Repository ...
type Repository interface {
	AddComment(comment *Comment) error
}

// NewService ...
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{allServices, repo}
}

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

func (service *service) AddComment(comment *Comment) error {
	return nil
}
