package post

import "time"

// Post is an aggregrate entity of Releases along with sociall interactable
// components such as stars, posting user and comments attached to the post
type Post struct {
	ID               int            `json: "Id"`
	PostedByUsername string         `json:"PostedByUsername,omitempty"`
	OriginChannel    string         `json:"orginChannel,omitempty"`
	Title            string         `json: "Title"`
	Description      string         `json: "Description"`
	ContentsID       []int          `json: "ContentsID"`
	Stars            map[string]int `json: "Stars"` // map of a username to the nmber of stars (range of 0 to 5) given
	CommentsID       []int          `json: "CommentsID"`
	CreationTime     time.Time      `json: "CreationTime"`
}

//Star is a key value pair of username and numberofStars
type Star struct {
	Username   string `json:"username,omitempty"`
	NumOfStars int    `json:"num_of_stars,omitempty"`
}

//
type Release struct {
	id int `json:"id,omitempty"`
}
