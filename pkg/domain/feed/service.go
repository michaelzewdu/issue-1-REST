/*
Package feed contains definition and implementation of a service that deals with Feeds  */
package feed

// Service specifies a method to service Feeds .
type Service interface {
	GetFeed(username string) (*Feed, error)
	GetPosts(f *Feed, sort Sorting, limit, offset int) ([]*Post, error)
	GetChannels(f *Feed) ([]*Channel, error)
	SubscribeChannel(c *Channel, f *Feed) error
	AddFeed(f *Feed) error
	UpdateFeed(f *Feed) error
	UnsubscribeChannel(channelname string, f *Feed) error
}

// Sorting constants used by SearchUser methods

type Sorting int

const (
	SortTop Sorting = iota
	SortHot Sorting = iota
	SortNew Sorting = iota
)
