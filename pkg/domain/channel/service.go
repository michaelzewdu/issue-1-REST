/*
Package channel contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package channel

// Service ...
type Service interface {
	AddChannel(channel *Channel) error
}

// Repository ...
type Repository interface {
	AddChannel(channel *Channel) error
}

// NewService ...
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{allServices, repo}
}

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

func (service *service) AddChannel(channel *Channel) error {
	return nil
}
