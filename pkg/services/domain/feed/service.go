/*
Package feed contains definition and implementation of a service that deals with Feeds  */
package feed

import (
	"fmt"
)

// Service specifies a method to service Feeds .
type Service interface {
	AddFeed(f *Feed) error
	GetFeed(username string) (*Feed, error)
	GetPosts(f *Feed, sort Sorting, limit, offset int) ([]*Post, error)
	GetChannels(f *Feed, sortBy SortBy, sortOrder SortOrder) ([]*Channel, error)
	UpdateFeed(username string, f *Feed) error
	Subscribe(f *Feed, channelname string) error
	Unsubscribe(f *Feed, channelname string) error
}

// Repository specifies a repo interface to serve the Service interface
type Repository interface {
	AddFeed(f *Feed) error
	GetFeed(username string) (*Feed, error)
	GetPosts(f *Feed, sort Sorting, limit, offset int) ([]*Post, error)
	GetChannels(f *Feed, sortBy string, sortOrder string) ([]*Channel, error)
	UpdateFeed(id uint, f *Feed) error
	Subscribe(f *Feed, channelname string) error
	Unsubscribe(f *Feed, channelname string) error
}

// Sorting enums used by GetPosts methods signifying how retrieved posts ares sorted
type Sorting string

const (
	// SortTop sorts the posts according to their star count
	SortTop Sorting = "top"
	// SortHot sorts the posts according to their comment count
	SortHot Sorting = "hot"
	// SortNew sorts the posts according to their creation time
	SortNew Sorting = "new"
	// NotSet signifies sort hasn't been set
	NotSet Sorting = ""
)

// SortOrder enums used by GetChannel methods the order channels are sorted with
type SortOrder string

// SortBy enums used by GetChannel methods the attribute channels are sorted with
type SortBy string

const (
	// SortAscending orders in an ascending manner
	SortAscending SortOrder = "ASC"
	// SortDescending orders in a descending manner
	SortDescending SortOrder = "DESC"

	// SortBySubscriptionTime orders channels according to the time the feed subscribed to them
	SortBySubscriptionTime SortBy = "subscription_time"
	// SortByUsername orders channels according to their username
	SortByUsername SortBy = "username"
	// SortByName orders channels according to their name
	SortByName SortBy = "name"
)

// ErrFeedNotFound is returned when the requested feed is not found
var ErrFeedNotFound = fmt.Errorf("feed not found")

// ErrChannelNotFound is returned when the specified channel does not exist
var ErrChannelNotFound = fmt.Errorf("channel does not exist found")

//var ErrUserDoesNotExist = fmt.Errorf("user does not exist found")

type service struct {
	allServices *map[string]interface{}
	//userService *user.Service
	repo *Repository
}

// NewService returns a struct that implements the Service interface
func NewService(repo *Repository, allServices *map[string]interface{}) Service {
	return &service{
		allServices: allServices,
		repo:        repo,
	}
}

// GetUser returns the feed belonging to the given username
func (s service) GetFeed(username string) (*Feed, error) {
	/*if s.userService == nil {
		temp, ok := (*s.allServices)["User"]
		if !ok {
			return  nil, fmt.Errorf("user service not available")
		}
		s.userService = temp.(*user.Service)
	}
	if _, err := (*s.userService).GetUser(username); err != nil {
		return nil, UserDoesNotExist
	}
	*/
	return (*s.repo).GetFeed(username)
}

// GetPosts returns a list of posts collected from the channels
// the given feed has subscribed to sorted according to the given
// method.
// Pagination can be specified.
func (s service) GetPosts(f *Feed, sort Sorting, limit, offset int) ([]*Post, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination")
	}
	if f, err := s.GetFeed(f.OwnerUsername); err != nil {
		return nil, err
	} else {
		if sort == NotSet {
			sort = f.Sorting
		}
		return (*s.repo).GetPosts(f, sort, limit, offset)
	}
}

// GetChannels returns all the channels the given feed has subscribed to
func (s service) GetChannels(f *Feed, sortBy SortBy, sortOrder SortOrder) ([]*Channel, error) {
	if f, err := s.GetFeed(f.OwnerUsername); err != nil {
		return nil, err
	} else {
		return (*s.repo).GetChannels(f, string(sortBy), string(sortOrder))
	}
}

// Subscribe adds the channel to the list of channels that the feed
// collects posts from.
func (s service) Subscribe(f *Feed, channelname string) error {
	if f, err := s.GetFeed(f.OwnerUsername); err != nil {
		return err
	} else {
		return (*s.repo).Subscribe(f, channelname)
	}
}

// Unsubscribe removes the channel to the list of channels that the
// feed collects posts from.
func (s service) Unsubscribe(f *Feed, channelname string) error {
	if f, err := s.GetFeed(f.OwnerUsername); err != nil {
		return err
	} else {
		return (*s.repo).Unsubscribe(f, channelname)
	}
}

// AddFeed adds a feed to the username contained in the feed struct.
func (s service) AddFeed(f *Feed) error {
	if _, err := s.GetFeed(f.OwnerUsername); err == nil {
		return fmt.Errorf("feed already exists")
	}
	return (*s.repo).AddFeed(f)

}

// UpdateFeed updates the feed at the given id according to the given struct.
func (s service) UpdateFeed(username string, f *Feed) error {
	if oldFeed, err := s.GetFeed(username); err != nil {
		return err
	} else {
		f.OwnerUsername = username
		return (*s.repo).UpdateFeed(oldFeed.ID, f)
	}

}
