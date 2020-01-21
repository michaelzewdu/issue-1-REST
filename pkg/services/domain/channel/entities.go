package channel

import "time"

// Channel represents a singular stream of posts that a user can subscribe to
// under adminstration by certain users.
type Channel struct {
	ChannelUsername    string    `json:"channelUsername"`
	Name               string    `json:"name,omitempty"`
	Description        string    `json:"description,omitempty"`
	PictureURL         string    `json:"pictureURL,omitempty"`
	OwnerUsername      string    `json:"ownerUsername,omitempty"`
	AdminUsernames     []string  `json:"adminUsernames,omitempty"`
	PostIDs            []uint    `json:"postIDs,omitempty"`
	StickiedPostIDs    []uint    `json:"stickiedPostIDs,omitempty "`
	ReleaseIDs         []uint    `json:"releaseIDs,omitempty"`
	OfficialReleaseIDs []uint    `json:"officialReleaseIDs,omitempty"`
	CreationTime       time.Time `json:"creationTime,omitempty"`
}
