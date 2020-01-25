/*
Package comment contains definition and implementation of a service that deals with User entities */
package comment

import "fmt"

// Service specifies a method to service Comment entities.
type Service interface {
	AddComment(c *Comment) (*Comment, error)
	GetComment(id int) (*Comment, error)
	GetComments(postID int, by SortBy, order SortOrder, limit, offset int) ([]*Comment, error)
	GetReplies(commentID int, by SortBy, order SortOrder, limit, offset int) ([]*Comment, error)
	UpdateComment(c *Comment) (*Comment, error)
	DeleteComment(id int) error
}

// Repository specifies a repo interface to serve the Comment Service interface
type Repository interface {
	AddComment(c *Comment) (*Comment, error)
	GetComment(id int) (*Comment, error)
	GetComments(postID int, sortBy string, sortOrder string, limit, offset int) ([]*Comment, error)
	GetReplies(commentID int, by string, order string, limit, offset int) ([]*Comment, error)
	UpdateComment(c *Comment) (*Comment, error)
	DeleteComment(id int) error
}

// SortOrder holds  that specify how comments are sorted
type SortOrder string

// SortBy holds enums that specify by what attribute comments are sorted
type SortBy string

// Sorting constants used by SearchUser methods
const (
	SortAscending      SortOrder = "ASC"
	SortDescending     SortOrder = "DESC"
	SortByCreationTime SortBy    = "creation_time"
)

// ErrPostNotFound is returned when the the post ID specified has no post under it
var ErrPostNotFound = fmt.Errorf("post not found")

// ErrUserNotFound is returned when the the username specified isn't recognized
var ErrUserNotFound = fmt.Errorf("user not found")

// ErrSomeCommentDataNotPersisted is returned when the the username specified isn't recognized
var ErrSomeCommentDataNotPersisted = fmt.Errorf("was not able to persist some user data")

// ErrCommentNotFound is returned when the requested comment is not found
var ErrCommentNotFound = fmt.Errorf("comment not found")

type service struct {
	repo *Repository
}

// NewService returns a struct that implements the comment.Service interface
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

// AddComment adds an new comment based on the passed in struct
func (s service) AddComment(c *Comment) (*Comment, error) {
	if c.ReplyTo != -1 {
		if temp, err := s.GetComment(c.ReplyTo); err != nil {
			return nil, err
		} else {
			if temp.OriginPost != c.OriginPost {
				return nil, ErrCommentNotFound
			}
		}
	}
	return (*s.repo).AddComment(c)
}

// GetComment gets the comment stored under the given id.
func (s service) GetComment(id int) (*Comment, error) {
	return (*s.repo).GetComment(id)
}

// GetComment get's all the comments found under a single post.
// This includes replies to comments.
func (s service) GetComments(postID int, by SortBy, order SortOrder, limit, offset int) ([]*Comment, error) {
	return (*s.repo).GetComments(postID, string(by), string(order), limit, offset)
}

// GetReplies returns all the comments that are replies to the comment
// under the given id.
func (s service) GetReplies(commentID int, by SortBy, order SortOrder, limit, offset int) ([]*Comment, error) {
	return (*s.repo).GetReplies(commentID, string(by), string(order), limit, offset)
}

// UpdateComment updates a comment entity based on the given struct.
func (s service) UpdateComment(c *Comment) (*Comment, error) {
	if _, err := (*s.repo).GetComment(c.ID); err != nil {
		return nil, err
	}
	return (*s.repo).UpdateComment(c)
}

// DeleteComment removes the comment under the given id.
func (s service) DeleteComment(id int) error {
	return (*s.repo).DeleteComment(id)
}
