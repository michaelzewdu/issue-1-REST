package rest

import (
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
	"net/http"
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

	mainRouter := mux.NewRouter().StrictSlash(true)
	secureRouter := mainRouter.NewRoute().Subrouter()

	userService, _ := (*services)["User"].(*user.Service)
	attachUserRoutesToRouters(mainRouter, secureRouter, userService, logger)

	feedService, _ := (*services)["Feed"].(*feed.Service)
	//TODO secure these routes
	secureRouter.HandleFunc("/users/{username}/feed", getFeed(feedService, logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/posts", getFeedPosts(feedService, logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", getFeedChannels(feedService, logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", postFeedChannel(feedService, logger)).Methods("POST")
	secureRouter.HandleFunc("/users/{username}/feed", putFeed(feedService, logger)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/feed/channels/{channelname}", deleteFeedChannel(feedService, logger)).Methods("DELETE")

	jwtBackend := NewJWTAuthenticationBackend(userService, []byte("secret"), 3600)
	secureRouter.Use(CheckForAuthenticationMiddleware(jwtBackend, logger))

	mainRouter.HandleFunc("/token-auth", postTokenAuth(jwtBackend, logger)).Methods("POST")
	secureRouter.Handle("/token-auth-refresh", negroni.New(negroni.HandlerFunc(getTokenAuthRefresh(jwtBackend, logger)))).Methods("GET")
	secureRouter.Handle("/logout", negroni.New(negroni.HandlerFunc(getLogout(jwtBackend, logger)))).Methods("GET")
	return mainRouter
}

func attachUserRoutesToRouters(mainRouter, secureRouter *mux.Router, userService *user.Service, logger *Logger) {
	mainRouter.HandleFunc("/users", getUsers(userService, logger)).Methods("GET")
	mainRouter.HandleFunc("/users", postUser(userService, logger)).Methods("POST")
	//TODO secure these routes
	secureRouter.HandleFunc("/users/{username}", getUser(userService, logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}", putUser(userService, logger)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}", deleteUser(userService, logger)).Methods("DELETE")
	secureRouter.HandleFunc("/users/{username}/bookmarks", getUserBookmarks(userService, logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/bookmarks/{postID}", putUserBookmarks(userService, logger)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/bookmarks/{postID}", deleteUserBookmarks(userService, logger)).Methods("DELETE")
	secureRouter.HandleFunc("/users/{username}/bookmarks", postUserBookmarks(userService, logger)).Methods("POST")
}

// writeResponseToWriter is a helper function.
func writeResponseToWriter(response jSendResponse, w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t\t")
	err := encoder.Encode(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
