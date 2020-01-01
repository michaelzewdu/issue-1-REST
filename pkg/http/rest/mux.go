package rest

import (
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
	"net/http"
)

// Logger ...
type Logger interface {
	Log(format string, a ...interface{})
}

type Enviroment struct {
	Config
	Dependencies
}

type Dependencies struct {
	UserService    user.Service
	FeedService    feed.Service
	ReleaseService release.Service
	Logger         Logger
}
type Config struct {
	ImageServingRoute, ImageStoragePath, HostAddress, Port string
}

// NewMux ...
func NewMux(setup *Enviroment) *mux.Router {

	mainRouter := mux.NewRouter().StrictSlash(true)
	secureRouter := mainRouter.NewRoute().Subrouter()

	mainRouter.PathPrefix(setup.ImageServingRoute).Handler(
		http.StripPrefix(setup.ImageServingRoute, http.FileServer(http.Dir(setup.ImageStoragePath))),
	)

	attachUserRoutesToRouters(mainRouter, secureRouter, setup, &setup.Logger)
	attachReleaseRoutesToRouters(mainRouter, secureRouter, setup)

	feedService := &setup.FeedService
	//TODO secure these routes
	secureRouter.HandleFunc("/users/{username}/feed", getFeed(feedService, &setup.Logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/posts", getFeedPosts(feedService, &setup.Logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", getFeedChannels(feedService, &setup.Logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", postFeedChannel(feedService, &setup.Logger)).Methods("POST")
	secureRouter.HandleFunc("/users/{username}/feed", putFeed(feedService, &setup.Logger)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/feed/channels/{channelname}", deleteFeedChannel(feedService, &setup.Logger)).Methods("DELETE")

	jwtBackend := NewJWTAuthenticationBackend(&setup.UserService, []byte("secret"), 3600)
	secureRouter.Use(CheckForAuthenticationMiddleware(jwtBackend, &setup.Logger))

	mainRouter.HandleFunc("/token-auth", postTokenAuth(jwtBackend, &setup.Logger)).Methods("POST")
	secureRouter.Handle("/token-auth-refresh", negroni.New(negroni.HandlerFunc(getTokenAuthRefresh(jwtBackend, &setup.Logger)))).Methods("GET")
	secureRouter.Handle("/logout", negroni.New(negroni.HandlerFunc(getLogout(jwtBackend, &setup.Logger)))).Methods("GET")
	return mainRouter
}

func attachUserRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Enviroment, logger *Logger) {
	mainRouter.HandleFunc("/users", getUsers(&setup.UserService, logger)).Methods("GET")
	mainRouter.HandleFunc("/users", postUser(&setup.UserService, logger)).Methods("POST")
	//TODO secure these routes
	secureRouter.HandleFunc("/users/{username}", getUser(&setup.UserService, logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}", putUser(&setup.UserService, logger)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}", deleteUser(&setup.UserService, logger)).Methods("DELETE")
	secureRouter.HandleFunc("/users/{username}/bookmarks", getUserBookmarks(&setup.UserService, logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/bookmarks/{postID}", putUserBookmarks(&setup.UserService, logger)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/bookmarks/{postID}", deleteUserBookmarks(&setup.UserService, logger)).Methods("DELETE")
	secureRouter.HandleFunc("/users/{username}/bookmarks", postUserBookmarks(&setup.UserService, logger)).Methods("POST")
}

func attachReleaseRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Enviroment) {
	mainRouter.HandleFunc("/releases", getReleases(setup)).Methods("GET")
	mainRouter.HandleFunc("/releases", postReleases(setup)).Methods("POST")
	//TODO secure these routes
	secureRouter.HandleFunc("/releases/{id}", getRelease(setup)).Methods("GET")
	secureRouter.HandleFunc("/releases/{id}", putRelease(setup)).Methods("PUT")
	secureRouter.HandleFunc("/releases/{id}", deleteRelease(setup)).Methods("DELETE")
}

type jSendResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
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
