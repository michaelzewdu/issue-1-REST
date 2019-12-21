/*
Package memory contains cache based implementations of the different
Repository interfaces defined in the domain packages */
package memory

import (
	"github.com/slim-crown/Issue-1-REST/pkg/domain/user"
)

// UserRepository ...
type UserRepository struct {
	cache         map[string]user.User
	secondaryRepo *user.Repository
	allRepos      *map[string]interface{}
}

// // ChannelRepository ...
// type ChannelRepository struct {
// 	Cache         map[string]channel.Channel
// 	SecondaryRepo *channel.Repository
// 	AllRepos      *map[string]interface{}
// }

// // CommentRepository ...
// type CommentRepository struct {
// 	Cache         []comment.Comment
// 	SecondaryRepo *comment.Repository
// 	AllRepos      *map[string]interface{}
// }

// // FeedRepository ...
// type FeedRepository struct {
// 	Cache         []feed.Feed
// 	SecondaryRepo *feed.Repository
// 	AllRepos      *map[string]interface{}
// }

// // ReleaseRepository ...
// type ReleaseRepository struct {
// 	Cache         []release.Release
// 	SecondaryRepo *release.Repository
// 	AllRepos      *map[string]interface{}
// }
