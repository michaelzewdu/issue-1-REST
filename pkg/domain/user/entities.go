package user

import (
	"time"
)

// User represents standard user entity of issue#1.
// bookmarkedPosts map contains the postId mapped to the time it was bookmarked.
type User struct {
	Username        string            `json:"username"`
	Email           string            `json:"email"`
	FirstName       string            `json:"firstName"`
	MiddleName      string            `json:"middleName"`
	LastName        string            `json:"lastName"`
	CreationTime    time.Time         `json:"creationTime"`
	Bio             string            `json:"bio"`
	BookmarkedPosts map[int]time.Time `json:"bookmarkedPosts"`
	Password        string            `json:"password,omitempty"`
	PictureURL      string            `json:"pictureURL"`
}
