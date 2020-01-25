package post

import "time"

// Post is an aggregate entity of Releases along with socially interactive
// components such as stars, posting user and comments attached to the post
type Post struct {
	ID               int            `json:"id"`
	PostedByUsername string         `json:"PostedByUsername,omitempty"`
	OriginChannel    string         `json:"originChannel,omitempty"`
	Title            string         `json:"title,omitempty"`
	Description      string         `json:"description,omitempty"`
	ContentsID       []int          `json:"contentsID,omitempty"`
	Stars            map[string]int `json:"stars,omitempty"`
	CommentsID       []int          `json:"commentsID,omitempty"`
	CreationTime     time.Time      `json:"creationTime"`
}

//Star is a key value pair of username and number of stars
type Star struct {
	Username   string `json:"username,omitempty"`
	NumOfStars int    `json:"stars,omitempty"`
}
