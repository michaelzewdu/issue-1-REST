/*
Package user contains definiton and implemntation of a service that deals with User entities */
package user

import (
	"fmt"
)

// Service specifies a method to service User entities.
type Service interface {
	AddUser(user *User) error
	GetUser(username string) (*User, error)
	UpdateUser(username string, u *User) error
	DeleteUser(username string) error
	SearchUser(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*User, error)
	PassHashIsCorrect(username, passHash string) (bool, error)
	BookmarkPost(username string, postID int) error
}

// Repository specifies a repo interface to serve the Service interface
type Repository interface {
	AddUser(user *User) error
	GetUser(username string) (*User, error)
	UpdateUser(username string, u *User) error
	DeleteUser(username string) error
	SearchUser(pattern, sortBy, sortOrder string, limit, offset int) ([]*User, error)
	PassHashIsCorrect(username, passHash string) bool
	BookmarkPost(username string, postID int) error
}

type SortOrder string
type SortBy string

// Sorting constants used by SearchUser methods
const (
	SortAscending SortOrder  = "ASC"
	SortDescending SortOrder = "DESC"

	SortCreationTime SortBy = "creation_time"
	SortByUsername SortBy   = "username"
	SortByFirstName SortBy  = "first_name"
	SortByLastName SortBy   = "last_name"
)

type service struct {
	allServices *map[string]interface{}
	repo        *Repository
}

// NewService returns a struct that implements the Service interface
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{allServices, repo}
}

// AddUser adds a user according to the given username
func (service *service) AddUser(user *User) error {
	// TODO - check if username is occupied by a channel
	if u, _ := service.GetUser(user.Username); u != nil {
		return fmt.Errorf("the username %s is occupied", (*user).Username)
	}
	return (*service.repo).AddUser(user)
}

// GetUser returns the user according to the given username
func (service *service) GetUser(username string) (*User, error) {
	return (*service.repo).GetUser(username)
}

// UpdateUser updates the user of the given username according to the User struct given
func (service *service) UpdateUser(username string, u *User) error {
	if _, err := service.GetUser(username); err != nil {
		return fmt.Errorf("the user %s could not be found because of: %s", username, err.Error())
	}
	// Checks if username is trying to be changed and if the new username is occupied
	if u.Username != "" {
		if u, _ := service.GetUser(u.Username); u != nil {
			return fmt.Errorf("the new username %s is occupied", u.Username)
		}
	}
	return (*service.repo).UpdateUser(username, u)
}

// DeleteUser removes the user of the given username
func (service *service) DeleteUser(username string) error {
	if _, err := service.GetUser(username); err != nil {
		return fmt.Errorf("the user %s could not be found because of: %s", username, err.Error())
	}
	return (*service.repo).DeleteUser(username)
}

// SearchUser returns a list of users that match against the pattern.
// If pattern is empty, it returns all users.
// Sorting and pagination can be specified.
func (service *service) SearchUser(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*User, error) {
	// TODO - check for valid sort method
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	return (*service.repo).SearchUser(pattern, string(sortBy), string(sortOrder), limit, offset)
}

// PassHashIsCorrect checks the given pass hash agains the pass hash found in the database for the username.
func (service *service) PassHashIsCorrect(username, passHash string) (bool, error) {
	if _, err := service.GetUser(username); err != nil {
		return false, fmt.Errorf("the user %s could not be found because of: %s", username, err.Error())
	}
	return (*service.repo).PassHashIsCorrect(username, passHash), nil
}

// BookmarkPost bookmarks the given postID for the user of the given username.
func (service *service) BookmarkPost(username string, postID int) error {
	if _, err := service.GetUser(username); err != nil {
		return fmt.Errorf("the user %s could not be found because of: %s", username, err.Error())
	}
	return (*service.repo).BookmarkPost(username, postID)
}
