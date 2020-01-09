package channel

import "time"

// Channel represents a singular stream of posts that a user can subscribe to
// under adminstration by certain users.
type Channel struct {
	Username           string    `json:"username"`
	Name               string    `json:"name,omitempty"`
	Description        string    `json:"description,omitempty"`
	OwnerUsername      string    `json:"ownerUsername,omitempty"`
	AdminUsernames     []string  `json:"adminUsernames,omitempty"`
	PostIDs            []int     `json:"postIDs,omitempty"`
	StickiedPostIDs    []int     `json:"stickiedPostIDs,omitempty "`
	ReleaseIDs         []int     `json:"releaseIDs,omitempty"`
	OfficialReleaseIDs []int     `json:"officialReleaseIDs,omitempty"`
	CreationTime       time.Time `json:"creationTime,omitempty"`
}
