package rest

import (
	"github.com/gorilla/mux"
	"github.com/slim-crown/Issue-1-REST/pkg/domain/user"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
)

// Logger ...
type Logger interface {
	Log(format string, a ...interface{})
}

type jSendResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// NewMux ...
func NewMux(logger *Logger, services *map[string]interface{}) *mux.Router {
	mux := mux.NewRouter()
	userService, _ := (*services)["User"].(*user.Service)
	mux.HandleFunc("/users", postUser(userService, logger)).Methods("POST")
	mux.HandleFunc("/users", getUsers(userService, logger)).Methods("GET")
	mux.HandleFunc("/users/{username}", getUser(userService, logger)).Methods("GET")
	mux.HandleFunc("/users/{username}", putUser(userService, logger)).Methods("PUT")
	mux.HandleFunc("/users/{username}", deleteUser(userService, logger)).Methods("DELETE")

	feedService, _ := (*services)["Feed"].(*feed.Service)
	mux.HandleFunc("/users/{username}/feed/", getFeed(feedService, logger)).Methods("GET")
	mux.HandleFunc("/users/{username}/feed/posts", getFeedPosts(feedService, logger)).Methods("GET")
	mux.HandleFunc("/users/{username}/feed/channels", getFeedChannels(feedService, logger)).Methods("GET")
	mux.HandleFunc("/users/{username}/feed/channels", postFeedChannel(feedService, logger)).Methods("POST")
	mux.HandleFunc("/users/{username}/feed", putFeed(feedService, logger)).Methods("PUT")
	mux.HandleFunc("/users/{username}/feed/channels/{channelname}", deleteFeedChannel(feedService, logger)).Methods("DELETE")

	return mux
}
