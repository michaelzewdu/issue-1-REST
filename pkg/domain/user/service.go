/*
Package user contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package user

// Service ...
type Service interface {
	AddUser(user *User) error
}

// Repository ...
type Repository interface {
	AddUser(user *User) error
}

// NewService ...
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{allServices, repo}
}

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

func (service *service) AddUser(user *User) error {
	return nil
}
