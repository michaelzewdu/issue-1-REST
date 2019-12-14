package release

import "time"

// User reperesents standard user entity of issue#1.
// bookmarkedPosts map contains the postId mapped to the time it was bookmarked.
type User struct {
	username                               string
	email, firstName, middleName, lastName string
	creationTime                           time.Time
}
