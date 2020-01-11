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

var ErrCommentNotFound = fmt.Errorf("comment not found")
