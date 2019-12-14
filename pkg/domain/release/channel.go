package release

import "time"

// Channel represents a singular stream of posts that a user can subscribe to
// under adminstration by certain users.
type Channel struct {
	username, name, description string
	owner                       User
	admins                      []User
	posts                       []Post
	stickiedPosts               [2]Post
	catalog                     []Release
	creationTime                time.Time
}
