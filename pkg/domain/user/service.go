/*
Package user contains definition and implementation of a service that deals with User entities */
package user

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

// Service specifies a method to service User entities.
type Service interface {
	AddUser(user *User) error
	GetUser(username string) (*User, error)
	UpdateUser(username string, u *User) error
	DeleteUser(username string) error
	SearchUser(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*User, error)
	Authenticate(u *User) (bool, error)
	BookmarkPost(username string, postID int) error
	DeleteBookmark(username string, postID int) error
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
	DeleteBookmark(username string, postID int) error
	UsernameOccupied(username string) (bool, error)
	EmailOccupied(email string) (bool, error)
}

// SortOrder holds enums used by SearchUser methods the order of Users are sorted with
type SortOrder string

// SortBy  holds enums used by SearchUser methods the attribute of Users are sorted with
type SortBy string


// Sorting constants used by SearchUser methods
const (
	SortAscending  SortOrder = "ASC"
	SortDescending SortOrder = "DESC"

	SortCreationTime SortBy = "creation_time"
	SortByUsername   SortBy = "username"
	SortByFirstName  SortBy = "first_name"
	SortByLastName   SortBy = "last_name"
)

// ErrUserNotFound is returned when the the username specified isn't recognized
var ErrUserNotFound = fmt.Errorf("user not found")
// ErrUserNameOccupied is returned when the the username specified is occupied
var ErrUserNameOccupied = fmt.Errorf("user name is occupied")
// ErrEmailIsOccupied is returned when the the email specified is occupied
var ErrEmailIsOccupied = fmt.Errorf("email is occupied")
// ErrPostNotFound is returned when the the username specified isn't recognized
var ErrPostNotFound = fmt.Errorf("post not found")
// ErrInvalidUserData is returned when the the username specified isn't recognized
var ErrInvalidUserData = fmt.Errorf("passed user data is invalid")


type service struct {
	//privateKey *rsa.PrivateKey
	//PublicKey  *rsa.PublicKey
	allServices *map[string]interface{}
	repo        *Repository
}

// NewService returns a struct that implements the Service interface
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	s := &service{allServices: allServices, repo: repo}
	return s
}

// AddUser adds a user according to the given username
func (service *service) AddUser(u *User) error {
	// TODO - check if username is occupied by a channel
	if u.FirstName == "" || u.Email == "" || u.Password == "" {
		return ErrInvalidUserData
	}
	{
		if occupied, err := (*service.repo).UsernameOccupied(u.Username); err == nil {
			if occupied {
				return ErrUserNameOccupied
			}
		} else {
			return fmt.Errorf("couldn't check if username occupied because of: %s", err.Error())
		}
		if occupied, err := (*service.repo).EmailOccupied(u.Email); err == nil {
			if occupied {
				return ErrEmailIsOccupied
			}
		} else {
			return fmt.Errorf("couldn't check if email occupied because of: %s", err.Error())
		}
	}
	var hashedPassword string
	{
		// this block hashes the password
		cat := u.Password + u.Username
		hashedPasswordArr := sha512.Sum512([]byte(cat))
		hashedPassword = hex.EncodeToString(hashedPasswordArr[:])
	}
	u.Password = hashedPassword
	return (*service.repo).AddUser(u)
}

// GetUser returns the user according to the given username
func (service *service) GetUser(username string) (*User, error) {
	return (*service.repo).GetUser(username)
}

// UpdateUser updates the user of the given username according to the User struct given
func (service *service) UpdateUser(username string, u *User) error {
	if _, err := service.GetUser(username); err != nil {
		u.Username = username
		return service.AddUser(u)
	}
	// Checks if username is trying to be changed, then if the new username is occupied
	if u.Username != "" {
		if occupied, err := (*service.repo).UsernameOccupied(u.Username); err == nil {
			if occupied {
				return ErrUserNameOccupied
			}
		} else {
			return fmt.Errorf("couldn't check if username occupied because of: %s", err.Error())
		}
	}
	return (*service.repo).UpdateUser(username, u)
}

// DeleteUser removes the user of the given username
func (service *service) DeleteUser(username string) error {
	if _, err := service.GetUser(username); err != nil {
		return err
	}
	return (*service.repo).DeleteUser(username)
}

// SearchUser returns a list of users that match against the pattern.
// If pattern is empty, it returns all users.
// Sorting and pagination can be specified.
func (service *service) SearchUser(pattern string, sortBy SortBy, sortOrder SortOrder, limit, offset int) ([]*User, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	return (*service.repo).SearchUser(pattern, string(sortBy), string(sortOrder), limit, offset)
}

// Authenticate checks the given pass hash against the pass hash found in the database for the username.
func (service *service) Authenticate(u *User) (bool, error) {
	if _, err := service.GetUser(u.Username); err != nil {
		return false, err
	}
	var hashedPassword string
	{
		// this block hashes the password
		cat := u.Password + u.Username
		hashedPasswordArr := sha512.Sum512([]byte(cat))
		hashedPassword = hex.EncodeToString(hashedPasswordArr[:])
	}
	return (*service.repo).PassHashIsCorrect(u.Username, hashedPassword), nil
}

// BookmarkPost bookmarks the given postID for the user of the given username.
func (service *service) BookmarkPost(username string, postID int) error {
	if _, err := service.GetUser(username); err != nil {
		return err
	}
	return (*service.repo).BookmarkPost(username, postID)
}

// DeleteBookmark removes the post of the given postID from the users bookmarks
func (service *service) DeleteBookmark(username string, postID int) error {
	if _, err := service.GetUser(username); err != nil {
		return err
	}
	return (*service.repo).DeleteBookmark(username, postID)
}