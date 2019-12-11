/*
Package listing contains implementations of business level objects
all contained in a single package to help with circular dependecies */
package listing

// Service ...
type Service interface {
	FindUser(username string) error
}

// Repository ...
type Repository interface {
	FindUser(user *User) error
}

type service struct {
	repo Repository
}

func (service *service) FindUser(username string) (*User, error) {
	return nil, nil
}
