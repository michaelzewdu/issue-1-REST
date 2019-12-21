package feed

import "time"

// Feed is a value object that tracks channels that a user subbed to
// and other settings
type Feed struct {
	ID            int     `json:"id"`
	OwnerUsername string  `json:"ownerUsername"`
	Sorting       Sorting `json:"defaultSorting"`
	// hiddenPosts   []Post
}

// Channel represents a singular stream of posts that a user can subscribe to
// under administration by certain users.
type Channel struct {
	username string `json:"username"`
	name     string `json:"name"`
}

// Post is an aggregate entity of Releases along with socially interactive
// components such as stars, posting user and comments attached to Releases
type Post struct {
	ID               int       `json:"id"`
	ownerChannel     string    `json:"ownerChannel"`
	posterUsername   string    `json:"posterUsername"`
	Title            string    `json:"title"`
	ContentIDs       []int     `json:"contentIDs"`
	Stars            int       `json:"stars"`
	StarsByFeedOwner int       `json:"starsByFeedOwner"`
	CommentCount     int       `json:"commentCount"`
	CreationTime     time.Time `json:"creationTime"`
}
