/*
Package post contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package post

// Service ...
type Service interface {
	AddPost(post *Post) error
}

// Repository ...
type Repository interface {
	AddPost(post *Post) error
}

// NewService ...
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{allServices, repo}
}

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

func (service *service) AddPost(post *Post) error {
	return nil
}
