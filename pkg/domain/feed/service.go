/*
Package feed contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package feed

// Service ...
type Service interface {
	AddFeed(feed *Feed) error
}

// Repository ...
type Repository interface {
	AddFeed(feed *Feed) error
}

// NewService ...
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{allServices, repo}
}

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

func (service *service) AddFeed(feed *Feed) error {
	return nil
}
