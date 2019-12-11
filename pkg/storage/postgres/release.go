package postgres

import "time"

// Release represents an atomic work of creativity.
type Release struct {
	id           int
	owner        User
	metadata     Metadata
	contentType  string
	content      ContentGetter
	creationTime time.Time
}

// ContentGetter is an interface that allows us to get the content as a byte slice.
// Allows to abstract away the difference storage methods for text (db)
// and image (file) releases.
type ContentGetter interface {
	getContent() []byte
}

// Metadata is a value object holds all the metadata of releases.
// genreDefining is the genre classification that defines the release most.
// authors contains username in string form if author is an issue#1 user
// or plain names otherwise.
// description is for data like blurb.
type Metadata struct {
	genreDefining string
	authors       []string
	description   string
}
