/*
Package release contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package release

// Service ...
type Service interface {
	AddRelease(release *Release) error
}

// Repository ...
type Repository interface {
	AddRelease(release *Release) error
}

// NewService ...
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{allServices, repo}
}

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

func (service *service) AddRelease(release *Release) error {
	return nil
}
