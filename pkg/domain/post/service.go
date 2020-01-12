/*
Package post contains definiton and implemntation of a service that deals with User entities */
package post

import (
	"fmt"
)

// Service specifies a method to service Release entities.
type Service interface {
	GetPost(id int) (*Post, error)
	DeletePost(id int) error
	AddPost(p *Post) (*Post, error)
	UpdatePost(pos *Post, id int) (*Post, error)
	SearchPost(pattern string, by SortBy, order SortOrder, limit int, offset int) ([]*Post, error)
	GetPostStar(id int, username string) (*Star, error)
	DeletePostStar(id int, username string) error
	AddPostStar(id int, star *Star) (*Star, error)
	UpdatePostStar(id int, star *Star) (*Star, error)
}

// Repository specifies a repo interface to serve the Post Service interface
type Repository interface {
	GetPost(id int) (*Post, error)
	DeletePost(id int) error
	AddPost(p *Post) (*Post, error)
	UpdatePost(pos *Post, id int) (*Post, error)
	SearchPost(pattern string, by SortBy, order SortOrder, limit int, offset int) ([]*Post, error)
	GetPostStar(id int, username string) (*Star, error)
	DeletePostStar(id int, username string) error
	AddPostStar(id int, star *Star) (*Star, error)
	UpdatePostStar(id int, star *Star) (*Star, error)
}

// SortOrder holds enums used by SearchPost methods the order of Users are sorted with
type SortOrder string

// SortBy  holds enums used by SearchPost methods the attribute of Users are sorted with
type SortBy string

// Sorting constants used by SearchPost methods
const (
	SortAscending  SortOrder = "ASC"
	SortDescending SortOrder = "DESC"

	SortByCreationTime SortBy = "creation_time"
	SortByChannel      SortBy = "channel_from"
	SortByPoster       SortBy = "posted_by"
	SortByTitle        SortBy = "title"
)

//ErrPostNotFound is returned when requested post is not found
var ErrPostNotFound = fmt.Errorf("Post not found")

//ErrUserNotFound is returned when requested User is not found
var ErrUserNotFound = fmt.Errorf("User not found")

//ErrReleaseNotFound is returned when requested Release is not found
var ErrReleaseNotFound = fmt.Errorf("Release not found")

//ErrStarNotFound is returned when requested Star is not found
var ErrStarNotFound = fmt.Errorf("Star not found")

//ErrSomePostDataNotPersisted is returned when data aren't properly added to post database
var ErrSomePostDataNotPersisted = fmt.Errorf("Data not properly added")

type service struct {
	repo *Repository
}

// NewService returns a struct that implements the Service interface
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

// GetPost gets the Post stored under the given id.
func (s service) GetPost(id int) (*Post, error) {
	p, err := (*s.repo).GetPost(id)
	if err != nil {
		return nil, ErrPostNotFound
	}

	return p, err
}

// DeletePost Deletes the Post stored under the given id.
func (s service) DeletePost(id int) error {
	err := (*s.repo).DeletePost(id)
	if err != nil {
		return ErrPostNotFound
	}
	return err

}

// AddPost Adds the Post stored under the given id.
func (s service) AddPost(p *Post) (*Post, error) {

	return (*s.repo).AddPost(p)
}

//UpdatePost updates the post with given id and post struct
func (s service) UpdatePost(pos *Post, id int) (*Post, error) {
	return (*s.repo).UpdatePost(pos, id)
}


func (s service) SearchPost(pattern string, by SortBy, order SortOrder, limit int, offset int) ([]*Post, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	return (*s.repo).SearchPost(pattern, by, order, limit, offset)

}
func (s service) GetPostStar(id int, username string) (*Star, error) {
	return (*s.repo).GetPostStar(id, username)
}
func (s service) DeletePostStar(id int, username string) error {

	return (*s.repo).DeletePostStar(id, username)
}
func (s service) AddPostStar(id int, star *Star) (*Star, error) {
	return (*s.repo).AddPostStar(id, star)
}
func (s service) UpdatePostStar(id int, star *Star) (*Star, error) {
	return (*s.repo).UpdatePostStar(id, star)
}
