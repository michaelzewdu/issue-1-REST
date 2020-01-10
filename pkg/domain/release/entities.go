package release

import "time"

// Type signifies the content type of the release. Either Image or Text.
type Type string

const (
	// Image type releases include webcomics, art, memes...etc
	Image Type = "image"
	// Text type releases include web-series, essays, blogs, anecdote...etc
	Text Type = "text"
)

// Release represents an atomic work of creativity.
type Release struct {
	ID           int    `json:"id"`
	OwnerChannel string `json:"ownerChannel"`
	Type         Type   `json:"type"`
	Content      string `json:"content"`
	Metadata     `json:"metadata,omitempty"`
	CreationTime time.Time `json:"creationTime,omitempty"`
}

// Metadata is a value object holds all the metadata of releases.
// genreDefining is the genre classification that defines the release most.
// authors contains username in string form if author is an issue#1 user
// or plain names otherwise.
// description is for data like blurb.
type Metadata struct {
	Title         string    `json:"title,omitempty"`
	ReleaseDate   time.Time `json:"releaseDate,omitempty"`
	GenreDefining string    `json:"genreDefining,omitempty"`
	Description   string    `json:"description,omitempty"`
	Other         `json:"other,omitempty"`
	//Cover         string   `json:"cover"`
}

// Other is a struct used to contain metadata not necessarily present in all releases
type Other struct {
	Authors []string `json:"authors,omitempty"`
	Genres  []string `json:"genres,omitempty"`
}
