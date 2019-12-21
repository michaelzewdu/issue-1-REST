package channel

import "time"

// Channel represents a singular stream of posts that a user can subscribe to
// under adminstration by certain users.
type Channel struct {
	username, name, description string
	ownerUsername               string
	adminUsernames              []string
	postIDs                     []int
	stickiedPostIDs             [2]int
	catalogIDs                  []int
	creationTime                time.Time
}
