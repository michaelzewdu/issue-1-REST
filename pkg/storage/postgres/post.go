package postgres

import "time"

// Post is an aggregrate entity of Releases along with sociall interactable
// components such as stars, posting user and comments attached to the post
type Post struct {
	id                 int
	postedBy           User
	title, description string
	content            []Release
	stars              map[string]int // map of a username to the nmber of stars (range of 0 to 5) given
	comments           []Comment
	creationTime       time.Time
}
