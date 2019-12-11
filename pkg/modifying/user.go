package modifying

import "time"

// User reperesents standard user entity of issue#1.
// bookmarkedPosts map contains the postId mapped to the time it was bookmarked.
type User struct {
	username                               string
	email, firstName, middleName, lastName string
	bio                                    string
	bookmarkedPosts                        map[int]time.Time
	feed                                   Feed
	creationTime                           time.Time
}

// Feed is a value object that tracks channels that a user subbed to
// and other settings
type Feed struct {
	subscriptions []Channel
	sorting       string
	// hiddenPosts   []Post
}
