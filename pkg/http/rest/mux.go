package rest

import (
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/channel"
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
	ChannelService channel.Service
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
	//attachChannelRoutesToRouters(mainRouter, secureRouter, setup, &setup.Logger)
	attachReleaseRoutesToRouters(mainRouter, secureRouter, setup)

	feedService := &setup.FeedService
	//TODO secure these routes
	secureRouter.HandleFunc("/users/{username}/feed", getFeed(feedService, &setup.Logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/posts", getFeedPosts(feedService, &setup.Logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", getFeedChannels(feedService, &setup.Logger)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", postFeedChannel(feedService, &setup.Logger)).Methods("POST")
	secureRouter.HandleFunc("/users/{username}/feed", putFeed(feedService, &setup.Logger)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/feed/channels/{channelname}", deleteFeedChannel(feedService, &setup.Logger)).Methods("DELETE")

	mainRouter.HandleFunc("/channels", postChannel(&setup.ChannelService, &setup.Logger)).Methods("POST")
	mainRouter.HandleFunc("/channels", getChannels(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}", getChannel(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}", putChannel(&setup.ChannelService, &setup.Logger)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{username}", deleteChannel(&setup.ChannelService, &setup.Logger)).Methods("DELETE")
	mainRouter.HandleFunc("/channels/{username}/admins", getAdmins(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/admins/{adminUsername}", putAdmin(&setup.ChannelService, &setup.Logger)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{username}/admins/{adminUsername}", deleteAdmin(&setup.ChannelService, &setup.Logger)).Methods("DELETE")
	mainRouter.HandleFunc("/channels/{username}/owners", getOwner(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/owners/{ownerUsername}", putOwner(&setup.ChannelService, &setup.Logger)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{username}/Posts", getPosts(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/Posts/{postID}", getPost(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/catalog", getCatalog(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/catalogs/{catalogID}", deleteReleaseFromCatalog(&setup.ChannelService, &setup.Logger)).Methods("DELETE")
	mainRouter.HandleFunc("/channels/{username}/catalogs/{catalogID}", getReleaseFromCatalog(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/catalogs", getOfficialCatalog(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/catalogs/{catalogID}", putReleaseInCatalog(&setup.ChannelService, &setup.Logger)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{username}/catalogs}", postReleaseInCatalog(&setup.ChannelService, &setup.Logger)).Methods("POST")
	mainRouter.HandleFunc("/channels/{username}/catalogs/official/{catalogID}", putReleaseInOfficialCatalog(&setup.ChannelService, &setup.Logger)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{username}/Posts/stickiedPosts", getStickiedPosts(&setup.ChannelService, &setup.Logger)).Methods("GET")
	mainRouter.HandleFunc("/channels/{username}/Posts/{postID}", stickyPost(&setup.ChannelService, &setup.Logger)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{username}/Posts/stickiedPosts/{stickiedPostID}", deleteStickiedPost(&setup.ChannelService, &setup.Logger)).Methods("DELETE")

	jwtBackend := NewJWTAuthenticationBackend(&setup.UserService, []byte("secret"), 3600)
	secureRouter.Use(CheckForAuthenticationMiddleware(jwtBackend, &setup.Logger))

	mainRouter.HandleFunc("/token-auth", postTokenAuth(jwtBackend, &setup.Logger)).Methods("POST")
	secureRouter.Handle("/token-auth-refresh", negroni.New(negroni.HandlerFunc(getTokenAuthRefresh(jwtBackend, &setup.Logger)))).Methods("GET")
	secureRouter.Handle("/logout", negroni.New(negroni.HandlerFunc(getLogout(jwtBackend, &setup.Logger)))).Methods("GET")
	return mainRouter
}

//func attachChannelRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Enviroment, logger *Logger) {
//	mainRouter.HandleFunc("/channels",postChannel(&setup.ChannelService,logger)).Methods("POST")
//	mainRouter.HandleFunc("/channels",getChannels(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}",getChannel(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}",putChannel(&setup.ChannelService,logger)).Methods("PUT")
//	secureRouter.HandleFunc("/channels/{username}",deleteChannel(&setup.ChannelService,logger)).Methods("DELETE")
//	secureRouter.HandleFunc("/channels/{username}/admins",getAdmins(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/admins/{adminUsername}",putAdmin(&setup.ChannelService,logger)).Methods("PUT")
//	secureRouter.HandleFunc("/channels/{username}/admins/{adminUsername}",deleteAdmin(&setup.ChannelService,logger)).Methods("DELETE")
//	secureRouter.HandleFunc("/channels/{username}/owners",getOwner(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/owners/{ownerUsername}",putOwner(&setup.ChannelService,logger)).Methods("PUT")
//	secureRouter.HandleFunc("/channels/{username}/Posts",getPosts(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/Posts/{postIDs}",getPost(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/admins/catalog",getCatalog(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/catalogs/{catalogID}",deleteReleaseFromCatalog(&setup.ChannelService,logger)).Methods("DELETE")
//	secureRouter.HandleFunc("/channels/{username}/catalogs/{catalogID}",getReleaseFromCatalog(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/catalogs/official",getOfficialCatalog(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/catalogs/{catalogID}",putReleaseInCatalog(&setup.ChannelService,logger)).Methods("PUT")
//	secureRouter.HandleFunc("/channels/{username}/catalogs}",postReleaseInCatalog(&setup.ChannelService,logger)).Methods("POST")
//	secureRouter.HandleFunc("/channels/{username}/catalogs/{catalogID}",putReleaseInOfficialCatalog(&setup.ChannelService,logger)).Methods("PUT")
//	secureRouter.HandleFunc("/channels/{username}/stickiedPosts",getStickiedPosts(&setup.ChannelService,logger)).Methods("GET")
//	secureRouter.HandleFunc("/channels/{username}/Posts/{postID}",stickyPost(&setup.ChannelService,logger)).Methods("PUT")
//	secureRouter.HandleFunc("/channels/{username}/stickiedPosts{stickiedPostID}",deleteStickiedPost(&setup.ChannelService,logger)).Methods("GET")
//
//
//
//
//}

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
