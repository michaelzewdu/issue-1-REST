/*
Package adding contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package adding

// Service ...
type Service interface {
	AddUser(user *User) error
}

// Repository ...
type Repository interface {
	AddUser(user *User) error
}

type service struct {
	repo Repository
}

func (service *service) AddUser(user *User) error {
	return nil
}
