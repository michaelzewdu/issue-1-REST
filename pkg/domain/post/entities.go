package post

import "time"

// Post is an aggregate entity of Releases along with socially interactive
// components such as stars, posting user and comments attached to the post
type Post struct {
	ID               int            `json:"id"`
	PostedByUsername string         `json:"PostedByUsername,omitempty"`
	OriginChannel    string         `json:"originChannel,omitempty"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	ContentsID       []int          `json:"contentsID"`
	Stars            map[string]int `json:"stars"`
	CommentsID       []int          `json:"commentsID"`
	CreationTime     time.Time      `json:"creationTime"`
}

//Star is a key value pair of username and number of stars
type Star struct {
	Username   string `json:"username,omitempty"`
	NumOfStars int    `json:"stars,omitempty"`
}

//
type Release struct {
	ID int `json:"id,omitempty"`
}
