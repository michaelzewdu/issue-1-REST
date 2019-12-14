/*
Package rest ...*/
package rest

import "net/http"

// Server ...
type Server struct {
	logger   *Logger
	services *map[string]interface{}
}

// Logger ...
type Logger interface {
	Log(message string)
}

// NewServer ...
func NewServer(logger *Logger, services *map[string]interface{}) Server {
	return Server{logger, services}
}

// ListenAndServe ...
func (server *Server) ListenAndServe(address string) {
	mux := http.NewServeMux()
	// fs := http.FileServer(http.Dir("assets"))
	// mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// mux.HandleFunc("/", indexHandler)
	// mux.HandleFunc("/about", aboutHandler)
	http.ListenAndServe(address, mux)
}
