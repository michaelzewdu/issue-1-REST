package feed

// Feed is a value object that tracks channels that a user subbed to
// and other settings
type Feed struct {
	ID                 int
	OwnerUsername      string
	SubscribedChannels []string
	Sorting            string
	// hiddenPosts   []Post
}
