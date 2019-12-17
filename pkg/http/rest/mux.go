package rest

import (
	"github.com/gorilla/mux"
	"github.com/slim-crown/Issue-1/pkg/domain/user"
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

	return mux
}
