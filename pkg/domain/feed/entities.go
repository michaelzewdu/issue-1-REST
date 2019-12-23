package feed

import "time"

type (
	// Feed is a value object that tracks channels that a user subbed to
	// and other settings
	Feed struct {
		ID            uint      `json:"id,omitempty"`
		OwnerUsername string    `json:"ownerUsername"`
		Sorting       Sorting   `json:"defaultSorting"`
		//Subscriptions []*Channel `json:"subscriptions"`
		// hiddenPosts   []Post
	}
	// Channel represents a singular stream of posts that a user can subscribe to
	// under administration by certain users.
	Channel struct {
		Username         string    `json:"username"`
		Name             string    `json:"name"`
		SubscriptionTime time.Time `json:"subscriptionTime"`
	}
	// Post is an aggregate entity of Releases along with socially interactive
	// components such as stars, posting user and comments attached to Releases
	Post struct {
		ID int `json:"id"`
		//OwnerChannel     string    `json:"ownerChannel"`
		//PosterUsername   string    `json:"posterUsername"`
		//Title            string    `json:"title"`
		//ContentIDs       []int     `json:"contentIDs,omitempty"`
		//StarCount           int       `json:"starCount"`
		//StarsByFeedOwner int       `json:"starsByFeedOwner"`
		//CommentCount     int       `json:"commentCount"`
		//CreationTime     time.Time `json:"creationTime"`
	}
)
