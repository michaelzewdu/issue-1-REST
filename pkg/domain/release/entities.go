package release

import "time"

// Release represents an atomic work of creativity.
type Release struct {
	ID           int    `json:"id"`
	OwnerChannel string `json:"ownerChannel"`
	Type         Type   `json:"type"`
	Content      string `json:"content"`
	Metadata     `json:"metadata,omitempty"`
	CreationTime time.Time `json:"creationTime,omitempty"`
}

type Type string

const (
	Image Type = "image"
	Text  Type = "text"
)

// Metadata is a value object holds all the metadata of releases.
// genreDefining is the genre classification that defines the release most.
// authors contains username in string form if author is an issue#1 user
// or plain names otherwise.
// description is for data like blurb.
type Metadata struct {
	Title         string   `json:"title"`
	GenreDefining string   `json:"genreDefining,omitempty"`
	Description   string   `json:"description,omitempty"`
	Authors       []string `json:"authors"`
	Genres        []string `json:"genres"`
	//Cover         string   `json:"cover"`
}
