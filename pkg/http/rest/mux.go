package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/slim-crown/issue-1-REST/pkg/domain/channel"
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
)

/*
// Logger ...
type Logger interface {
	Log(format string, a ...interface{})
}
*/
// Setup is used to inject dependencies and other required data used by the handlers.
type Setup struct {
	Config
	Dependencies
}

// Dependencies contains dependencies used by the handlers.
type Dependencies struct {
	StrictSanitizer *bluemonday.Policy
	MarkupSanitizer *bluemonday.Policy
	UserService     user.Service
	FeedService     feed.Service
	ChannelService  channel.Service
	ReleaseService  release.Service
	PostService     post.Service
	CommentService  comment.Service
	jwtBackend      *JWTAuthenticationBackend
	Logger          *log.Logger
}

// Config contains the different settings used to set up the handlers
type Config struct {
	ImageServingRoute, ImageStoragePath, HostAddress, Port string
	TokenAccessLifetime, TokenRefreshLifetime              time.Duration
	TokenSigningSecret                                     []byte
}

// NewMux returns a new multiplexer with all the used setup.
func NewMux(s *Setup) *mux.Router {
	mainRouter := mux.NewRouter().StrictSlash(true)
	secureRouter := mainRouter.NewRoute().Subrouter()

	// setup static file server
	mainRouter.PathPrefix(s.ImageServingRoute).Handler(
		http.StripPrefix(s.ImageServingRoute, http.FileServer(http.Dir(s.ImageStoragePath))))

	// setup security
	s.jwtBackend = NewJWTAuthenticationBackend(s)
	mainRouter.Use(ParseAuthTokenMiddleware(s))
	secureRouter.Use(CheckForAuthMiddleware(s))

	// attach routes
	attachAuthRoutesToRouters(mainRouter, secureRouter, s)
	attachUserRoutesToRouters(mainRouter, secureRouter, s)
	attachReleaseRoutesToRouters(mainRouter, secureRouter, s)
	attachChannelRoutesToRouters(mainRouter, secureRouter, s)
	attachFeedRoutesToRouters(secureRouter, s)
	attachCommentRoutesToRouters(mainRouter, secureRouter, s)
	attachPostRoutesToRouters(mainRouter, secureRouter, s)

	return mainRouter
}

func attachCommentRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{id:[0-9]+}", getComment(setup)).Methods(http.MethodGet)
	mainRouter.HandleFunc("/posts/{postID:[0-9]+}/comments", getComments(setup)).Methods(http.MethodGet)
	secureRouter.HandleFunc("/posts/{postID:[0-9]+}/comments", postComment(setup)).Methods(http.MethodPost)
	secureRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{id:[0-9]+}", patchComment(setup)).Methods(http.MethodPatch)
	secureRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{id:[0-9]+}", deleteComment(setup)).Methods(http.MethodDelete)

	mainRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{rootCommentID:[0-9]+}/replies/{id:[0-9]+}", getComment(setup)).Methods(http.MethodGet)
	mainRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{rootCommentID:[0-9]+}/replies", getCommentReplies(setup)).Methods(http.MethodGet)
	secureRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{rootCommentID:[0-9]+}/replies", postComment(setup)).Methods(http.MethodPost)
	secureRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{rootCommentID:[0-9]+}/replies/{id:[0-9]+}", patchComment(setup)).Methods(http.MethodPatch)
	secureRouter.HandleFunc("/posts/{postID:[0-9]+}/comments/{rootCommentID:[0-9]+}/replies/{id:[0-9]+}", deleteComment(setup)).Methods(http.MethodDelete)
}

func attachAuthRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/token-auth", postTokenAuth(setup)).Methods("POST")
	mainRouter.HandleFunc("/token-auth-refresh", getTokenAuthRefresh(setup)).Methods("GET")
	secureRouter.HandleFunc("/logout", getLogout(setup)).Methods("GET")
}

func attachFeedRoutesToRouters(secureRouter *mux.Router, setup *Setup) {
	//TODO secure these routes
	secureRouter.HandleFunc("/users/{username}/feed", getFeed(setup)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/posts", getFeedPosts(setup)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", getFeedChannels(setup)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/feed/channels", postFeedChannel(setup)).Methods("POST")
	secureRouter.HandleFunc("/users/{username}/feed", putFeed(setup)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/feed/channels/{channelname}", deleteFeedChannel(setup)).Methods("DELETE")

}
func attachPostRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/posts", getPosts(setup)).Methods("GET")
	secureRouter.HandleFunc("/posts", postPost(setup)).Methods("POST")
	//TODO secure these routes
	mainRouter.HandleFunc("/posts/{id}", getPost(setup)).Methods("GET")
	secureRouter.HandleFunc("/posts/{id}", putPost(setup)).Methods("PUT")
	secureRouter.HandleFunc("/posts/{id}", deletePost(setup)).Methods("DELETE")
	mainRouter.HandleFunc("/posts/{id}/releases", getPostReleases(setup)).Methods("GET")
	mainRouter.HandleFunc("/posts/{id}/comments", getPostComments(setup)).Methods("GET")
	mainRouter.HandleFunc("/posts/{id}/stars", getPostStars(setup)).Methods("GET")
	mainRouter.HandleFunc("/posts/{id}/stars/{username}", getPostStar(setup)).Methods("GET")
	secureRouter.HandleFunc("/posts/{id}/stars", putPostStar(setup)).Methods("PUT")

}

func attachUserRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/users", getUsers(setup)).Methods("GET")
	mainRouter.HandleFunc("/users", postUser(setup)).Methods("POST")
	//TODO secure these routes
	secureRouter.HandleFunc("/users/{username}", getUser(setup)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}", putUser(setup)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}", deleteUser(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/users/{username}/bookmarks", getUserBookmarks(setup)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/bookmarks/{postID}", putUserBookmarks(setup)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/bookmarks/{postID}", deleteUserBookmarks(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/users/{username}/bookmarks", postUserBookmarks(setup)).Methods("POST")
	secureRouter.HandleFunc("/users/{username}/picture", getUserPicture(setup)).Methods("GET")
	secureRouter.HandleFunc("/users/{username}/picture", putUserPicture(setup)).Methods("PUT")
	secureRouter.HandleFunc("/users/{username}/picture", deleteUserPicture(setup)).Methods("DELETE")
}

func attachReleaseRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/releases", getReleases(setup)).Methods("GET")
	mainRouter.HandleFunc("/releases", postRelease(setup)).Methods("POST")
	//TODO secure these routes
	secureRouter.HandleFunc("/releases/{id}", getRelease(setup)).Methods("GET")
	secureRouter.HandleFunc("/releases/{id}", putRelease(setup)).Methods("PUT")
	secureRouter.HandleFunc("/releases/{id}", deleteRelease(setup)).Methods("DELETE")
}
func attachChannelRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/channels", postChannel(setup)).Methods("POST")
	mainRouter.HandleFunc("/channels", getChannels(setup)).Methods("GET")
	mainRouter.HandleFunc("/channels/{channelUsername}", getChannel(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}", putChannel(setup)).Methods("PUT")
	secureRouter.HandleFunc("/channels/{channelUsername}", deleteChannel(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/channels/{channelUsername}/admins", getAdmins(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/admins/{adminUsername}", putAdmin(setup)).Methods("PUT")
	secureRouter.HandleFunc("/channels/{channelUsername}/admins/{adminUsername}", deleteAdmin(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/channels/{channelUsername}/owners", getOwner(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/owners/{ownerUsername}", putOwner(setup)).Methods("PUT")
	secureRouter.HandleFunc("/channels/{channelUsername}/Posts", getChannelPosts(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/Posts/{postID}", getChannelPost(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/catalog", getCatalog(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/catalogs/{catalogID}", deleteReleaseFromCatalog(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/channels/{channelUsername}/official/{catalogID}", deleteReleaseFromOfficialCatalog(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/channels/{channelUsername}/catalogs/{catalogID}", getReleaseFromCatalog(setup)).Methods("GET")
	mainRouter.HandleFunc("/channels/{channelUsername}/official", getOfficialCatalog(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/catalogs/{catalogID}", putReleaseInCatalog(setup)).Methods("PUT")
	secureRouter.HandleFunc("/channels/{channelUsername}/catalogs}", postReleaseInCatalog(setup)).Methods("POST")
	//secureRouter.HandleFunc("/channels/{channelUsername}/official/{catalogID}", putReleaseInOfficialCatalog(setup)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{channelUsername}/Posts/stickiedPosts", getStickiedPosts(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/Posts/{postID}", stickyPost(setup)).Methods("PUT")
	secureRouter.HandleFunc("/channels/{channelUsername}/Posts/stickiedPosts/{stickiedPostID}", deleteStickiedPost(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/channels/{channelUsername}/picture", putChannelPicture(setup)).Methods("PUT")
	mainRouter.HandleFunc("/channels/{channelUsername}/picture", getChannelPicture(setup)).Methods("GET")
	secureRouter.HandleFunc("/channels/{channelUsername}/picture", deleteChannelPicture(setup)).Methods("DELETE")

}

type jSendResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
type jSendFailData struct {
	ErrorReason  string `json:"errorReason"`
	ErrorMessage string `json:"errorMessage"`
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

var errUnacceptedType = fmt.Errorf("file mime type not accepted")
var errReadingFromImage = fmt.Errorf("err reading image file from request")

func saveImageFromRequest(r *http.Request, fileName string) (*os.File, string, error) {
	file, header, err := r.FormFile(fileName)
	if err != nil {
		return nil, "", errReadingFromImage
	}
	defer file.Close()
	err = checkIfFileIsAcceptedType(file)
	if err != nil {
		return nil, "", err
	}
	newFile, err := ioutil.TempFile("", "tempIMG*.jpg")
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(newFile, file)
	if err != nil {
		return nil, "", err
	}
	return newFile, header.Filename, nil
}

func generateFileNameForStorage(fileName, prefix string) string {
	v4uuid, _ := uuid.NewV4()
	return prefix + "." + v4uuid.String() + "." + fileName
}

func checkIfFileIsAcceptedType(file multipart.File) error { // this block checks if image is of accepted types
	acceptedTypes := map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
	}
	tempBuffer := make([]byte, 512)
	_, err := file.ReadAt(tempBuffer, 0)
	if err != nil {
		return errReadingFromImage
	}
	contentType := http.DetectContentType(tempBuffer)
	if _, ok := acceptedTypes[contentType]; !ok {
		return errUnacceptedType
	}
	return err
}

func saveTempFilePermanentlyToPath(tmpFile *os.File, path string) error {
	newFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer newFile.Close()

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = io.Copy(newFile, tmpFile)
	if err != nil {
		return err
	}

	return nil
}
