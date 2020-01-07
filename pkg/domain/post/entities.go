package post

import "time"

// Post is an aggregrate entity of Releases along with sociall interactable
// components such as stars, posting user and comments attached to the post
type Post struct {
	ID               int            `json: "id"`
	PostedByUsername string         `json: "poster"`
	OriginChannel    string         `json:"orginChannel,omitempty"`
	Title            string         `json: "title"`
	Description      string         `json: "description"`
	ContentsID       []int          `json: "listOfContents"`
	Stars            map[string]int `json: "stars"` // map of a username to the nmber of stars (range of 0 to 5) given
	CommentsID       []int          `json: "listOfComments"`
	CreationTime     time.Time      `json: "creationTime"`
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
