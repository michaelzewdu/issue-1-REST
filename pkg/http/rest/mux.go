package rest

import (
	"encoding/json"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"github.com/slim-crown/issue-1-REST/pkg/domain/user"
)

// Logger ...
type Logger interface {
	Log(format string, a ...interface{})
}

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
	ReleaseService  release.Service
	jwtBackend      *JWTAuthenticationBackend
	Logger          Logger
}

// Config contains the different settings used to set up the handlers
type Config struct {
	ImageServingRoute, ImageStoragePath, HostAddress, Port string
}

// NewMux returns a new multiplexer with all the used setup.
func NewMux(s *Setup) *mux.Router {

	mainRouter := mux.NewRouter().StrictSlash(true)
	secureRouter := mainRouter.NewRoute().Subrouter()

	mainRouter.PathPrefix(s.ImageServingRoute).Handler(
		http.StripPrefix(s.ImageServingRoute, http.FileServer(http.Dir(s.ImageStoragePath))))

	attachUserRoutesToRouters(mainRouter, secureRouter, s)
	attachReleaseRoutesToRouters(mainRouter, secureRouter, s)
	attachFeedRoutesToRouters(secureRouter, s)

	s.jwtBackend = NewJWTAuthenticationBackend(&s.UserService, []byte("secret"), 3600)
	secureRouter.Use(CheckForAuthenticationMiddleware(s))

	attachAuthRoutesToRouters(mainRouter, secureRouter, s)

	return mainRouter
}
func attachAuthRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/token-auth", postTokenAuth(setup)).Methods("POST")
	secureRouter.Handle("/token-auth-refresh", negroni.New(negroni.HandlerFunc(getTokenAuthRefresh(setup)))).Methods("GET")
	secureRouter.Handle("/logout", negroni.New(negroni.HandlerFunc(getLogout(setup)))).Methods("GET")
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
func attachCommentRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/posts/{postID}/comments/{commentID}", getComment(setup)).Methods("GET")
	mainRouter.HandleFunc("/posts/{postID}/comments?sort=time", getComments(setup)).Methods("GET")
	//TODO secure these routes
	secureRouter.HandleFunc("/posts/{postID}/comments", postComment(setup)).Methods("POST")
	secureRouter.HandleFunc("/posts/{postID}/comments/{commentID}", updateComment(setup)).Methods("PATCH")
	secureRouter.HandleFunc("/posts/{postID}/comments/{commentID}", deleteComment(setup)).Methods("DELETE")
	secureRouter.HandleFunc("/posts/{postID}/comments/{rootCommentID}/replies/{commentID}", getReply(setup)).Methods("GET")
	secureRouter.HandleFunc("/posts/{postID}/comments/{rootCommentID}/replies/?sort=time", getReplys(setup)).Methods("GET")
	secureRouter.HandleFunc("/posts/{postID}/comments/{rootCommentID}/replies", postReply(setup)).Methods("POST")
	secureRouter.HandleFunc("/posts/{postID}/comments/{rootCommentID}/replies/{commentID}", updateReply(setup)).Methods("PATCH")
	secureRouter.HandleFunc("/posts/{postID}/comments/{rootCommentID}/replies/{commentID}", deleteReply(setup)).Methods("DELETE")
}

func attachReleaseRoutesToRouters(mainRouter, secureRouter *mux.Router, setup *Setup) {
	mainRouter.HandleFunc("/releases", getReleases(setup)).Methods("GET")
	mainRouter.HandleFunc("/releases", postReleases(setup)).Methods("POST")
	//TODO secure these routes
	secureRouter.HandleFunc("/releases/{id}", getRelease(setup)).Methods("GET")
	secureRouter.HandleFunc("/releases/{id}", putRelease(setup)).Methods("PUT")
	secureRouter.HandleFunc("/releases/{id}", deleteRelease(setup)).Methods("DELETE")
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
