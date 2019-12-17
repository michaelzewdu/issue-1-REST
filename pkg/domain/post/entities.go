package post

import "time"

// Post is an aggregrate entity of Releases along with sociall interactable
// components such as stars, posting user and comments attached to the post
type Post struct {
	ID                 int
	PostedByUsername   string
	Title, Description string
	ContentsID         []int
	Stars              map[string]int // map of a username to the nmber of stars (range of 0 to 5) given
	CommentsID         []int
	CreationTime       time.Time
}
