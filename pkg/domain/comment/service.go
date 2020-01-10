/*
Package comment contains definiton and implemntation of a service that deals with User entities */
package comment

// Service specifies a method to service Release entities.
type Service interface {
	GetSortedComment(c *Comment, sort time) (*Comment, error)
	DeleteComment(id int) error
	AddComment(c *Comment) (*Comment, error)
	GetComment(postID int) (*Comment, error)
	UpdateComment()
}

// Repository specifies a repo interface to serve the Comment Service interface
type Repository interface {
	GetComment(id int) (*Comment, error)
}
