/*
Package modifying contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package modifying

// Service ...
type Service interface {
	ChangeEmail(username string) error
}

// Repository ...
type Repository interface {
	ChangeEmail(user *User) error
}

type service struct {
	repo Repository
}

func (service *service) ChangeEmail(user *User, email string) error {
	return nil
}
