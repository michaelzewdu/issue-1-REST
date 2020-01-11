/*
Package comment contains definiton and implemntation of a service that deals with User entities */
package comment

import "fmt"

// Service specifies a method to service Release entities.
type Service interface {
	GetComment(id int) (*Comment, error)
	AddComment(comment *Comment) error
	UpdateComment(comment *Comment, id int) error
	DeleteComment(id int) error
	GetReply(id int) (*Comment, error)
	AddReply(comment *Comment) error
	UpdateReply(comment *Comment, id int) error
	DeleteReply(id int) error
}

// Repository specifies a repo interface to serve the Comment Service interface
type Repository interface {
	GetComment(id int) (*Comment, error)
	AddComment(comment *Comment) error
	UpdateComment(comment *Comment, id int) error
	DeleteComment(id int) error
	GetReply(id int) (*Comment, error)
	AddReply(comment *Comment) error
	UpdateReply(comment *Comment, id int) error
	DeleteReply(id int) error
}

// ErrCommentNotFound is returned when the requested comment is not found
var ErrCommentNotFound = fmt.Errorf("comment not found")

type service struct {
	repo *Repository
}

//todo

// SortOrder holds enums used by SearchPost methods the order of Users are sorted with
//type SortOrder string

// SortBy  holds enums used by SearchPost methods the attribute of Users are sorted with
//type SortBy string

func (s service) GetComment(id int) (*Comment, error) {
	return (*s.repo).GetComment(id)
}

func (s service) AddComment(comment *Comment) error {
	return (*s.repo).AddComment(comment)
}

func (s service) UpdateComment(comment *Comment, id int) error {
	return (*s.repo).UpdateComment(comment, id)
}

func (s service) DeleteComment(id int) error {
	if _, err := s.GetComment(id); err != nil {
		return err
	}
	return (*s.repo).DeleteComment(id)
}

func (s service) GetReply(id int) (*Comment, error) {
	return (*s.repo).GetReply(id)
}

func (s service) AddReply(comment *Comment) error {
	return (*s.repo).AddReply(comment)
}

func (s service) UpdateReply(comment *Comment, id int) error {
	return (*s.repo).UpdateReply(comment, id)
}

func (s service) DeleteReply(id int) error {
	if _, err := s.GetReply(id); err != nil {
		return err
	}
	return (*s.repo).DeleteReply(id)
}

// NewService returns a struct that implements the release.Release interface
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}
