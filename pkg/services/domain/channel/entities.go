package channel

import "time"

// Channel represents a singular stream of posts that a user can subscribe to
// under adminstration by certain users.
type Channel struct {
	ChannelUsername    string    `json:"channelUsername"`
	Name               string    `json:"name,omitempty"`
	Description        string    `json:"description,omitempty"`
	PictureURL         string    `json:"pictureURL"`
	OwnerUsername      string    `json:"ownerUsername,omitempty"`
	AdminUsernames     []string  `json:"adminUsernames,omitempty"`
	PostIDs            []int     `json:"postIDs,omitempty"`
	StickiedPostIDs    []int     `json:"stickiedPostIDs,omitempty "`
	ReleaseIDs         []int     `json:"releaseIDs,omitempty"`
	OfficialReleaseIDs []int     `json:"officialReleaseIDs,omitempty"`
	CreationTime       time.Time `json:"creationTime,omitempty"`
}
