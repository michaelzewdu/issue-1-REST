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
	GetPostReleases(p *Post) ([]*Release, error)
	GetPostRelease(pId int, rId int) (*Release, error)
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
	GetPostReleases(p *Post) ([]*Release, error)
	GetPostRelease(pId int, rId int) (*Release, error)
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
