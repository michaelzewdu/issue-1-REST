package rest

import (
	"encoding/json"
	"fmt"
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/slim-crown/issue-1-REST/pkg/domain/user"

	"github.com/gorilla/mux"
)

// postComments returns a handler for POST /comments requests
func postComment(service *comment.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		c := comment.Comment{}

		{ // checks if requests uses forms or JSON and parses then
			c.Commenter = r.FormValue("commenter")
			if c.Commenter != "" { //do i need to check for origin post too?
				c.OriginPost = r.FormValue("originPost")
				c.Content = r.FormValue("content")
				// todo
				c.ReplyTo, _ = strconv.Atoi(r.FormValue("replyTo"))
			} else {
				err := json.NewDecoder(r.Body).Decode(&c)
				if err != nil {
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = `bad request, use format
				{"originPost":"originPost",
				"commenter":"commenter",
				"content":"content",
				"replyTo":"replyTo"}`
					response.Data = responseData
					(*logger).Log("bad update user request")
					statusCode = http.StatusBadRequest
				}
			}
		}
		if response.Data == nil {
			// this block checks for required fields
			if c.Commenter == "" {
				var responseData struct {
					Data string `json:"commenter"`
				}
				responseData.Data = "Commenter Username is required"
				response.Data = responseData
			}
			if c.OriginPost == "" {
				var responseData struct {
					Data string `json:"originPost"`
				}
				responseData.Data = "Origin Post ID is required"
				response.Data = responseData
			}
			if response.Data == nil {
				(*logger).Log("trying to add a comment %s", c.Commenter, c.Content, c.OriginPost, c.ReplyTo)
				err := (*service).AddComment(&c)
				switch err {
				case nil:
					response.Status = "success"
					(*logger).Log("success adding Comment %s", c.Commenter, c.Content, c.OriginPost, c.ReplyTo)

				//TODO with the updated code
				case comment.ErrUserNameOccupied:
					(*logger).Log("adding of comment failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"commenter"`
					}
					responseData.Data = "username is occupied"
					response.Data = responseData
					statusCode = http.StatusConflict
				case user.ErrOriginPostDoesntExist: //highly doubtfull about this one
					(*logger).Log("adding of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"originPost"`
					}
					responseData.Data = "Origin Post doesn't exist"
					response.Data = responseData
					statusCode = http.StatusConflict
				default:
					(*logger).Log("adding of comment failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when adding user"
					response.Data = responseData
					statusCode = http.StatusInternalServerError
				}
			} else {
				// if required fields aren't present
				(*logger).Log("bad adding user request")
				statusCode = http.StatusBadRequest
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

func getUser(service *comment.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]

		(*logger).Log("trying to fetch user %s", username)

		u, err := (*service).GetUser(username)
		switch err {
		case nil:
			response.Status = "success"
			{ // this block sanitizes the returned User if it's not the user herself accessing the route
				if username != r.Header.Get("authorized_username") {
					(*logger).Log(fmt.Sprintf("user %s fetched user %s", r.Header.Get("authorized_username"), u.Username))
					u.Email = ""
					u.BookmarkedPosts = make(map[int]time.Time)
				}
			}
			response.Data = *u
			(*logger).Log("success fetching user %s", username)
		case user.ErrUserNotFound:
			(*logger).Log("fetch attempt of non existing user %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("user of username %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of user failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when fetching user"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getComment returns a handler for GET /Comment requests
func getComment(service *comment.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		commenter := vars["commenter"]

		(*logger).Log("trying to fetch comment %s", commenter)

		u, err := (*service).GetComment(commenter)
		switch err {
		case nil:
			response.Status = "success"
			{ // this block sanitizes the returned comment if it's not the commenter herself accessing the route
				if commenter != r.Header.Get("authorized_commenter") {
					//i don't get this
					//u.Email = ""
					//u.BookmarkedPosts = make(map[int]time.Time)
				}
			}
			response.Data = *c
			(*logger).Log("success fetching commenter %s", commenter)
		case user.ErrUserNotFound:
			(*logger).Log("fetch attempt of non existing commenter %s", commenter)
			var responseData struct {
				Data string `json:"commenter"`
			}
			responseData.Data = fmt.Sprintf("commenter username %s not found", commenter)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of commenter failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when fetching commenter"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getCommentID returns a handler for GET /comments/:id
func getCommentID(service *comment.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		id := vars["id"]

		(*logger).Log("trying to fetch Comment %d", id)

		f, err := (*service).GetCommentID(id)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = *f
			(*logger).Log("success fetching comment id of %d", id)
		case post.ErrPostNotFound:
			(*logger).Log("fetching of post failed because: %s", err.Error())
			var responseData struct {
				Data int `json:"id"`
			}
			responseData.Data = fmt.Sprintf("comment of id %d not found", id)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		default:
			(*logger).Log("getting comment failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when getting comment"
			response.Data = responseData
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t\t")
		err = encoder.Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

// putComment returns a handler for PUT /comments/:id
func putComment(service *comment.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		id := vars["id"]

		newComment := comment.Comment{}
		errs := json.NewDecoder(r.Body).Decode(&newcomment)
		commenter := r.FormValue("commenter")
		if errs != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "bad request"
			response.Data = responseData
			(*logger).Log("bad update comment request")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			// if JSON parsing doesn't fail
			switch oldComment, err := (*service).GetComment(id); err != nil {
			case nil:
				if commentPost.Content == oldComment.Content && newComment.commenter == oldPost.commenter { // i don't really know what i did here
					response.Status = "success"
				} else {
					(*logger).Log("trying to update comment of user %s", oldComment.commenter)
					if newComment.commenter == "" {
						newComment.Commenter = oldComment.Commenter
					}
					if err = (*service).UpdateComment(oldComment.ID, &newComment); err != nil {

						(*logger).Log("update of Comment failed because: %s", err.Error())

						var responseData struct {
							Data string `json:"message"`
						}
						responseData.Data = "server error when updating Comment"
						response.Status = "error"
						response.Data = responseData
						w.WriteHeader(http.StatusNotFound)
					} else {
						(*logger).Log("success updating of comment %s", oldComment.commenter)
						response.Status = "success"
					}
				}
			case post.ErrPostNotFound:
				newComment.commenter = username //make sure created user has the new username
				err := (*service).AddComment(&newComment)
				if err != nil {
					(*logger).Log("creation of comment failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when creating comment"
					response.Data = responseData
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
				}
				(*logger).Log("success creating comment for user %s", username)
			default:
				(*logger).Log("updating of comment failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when updating comment"
				response.Data = responseData
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t\t")
		err = encoder.Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

	}
}

// deleteComment returns a handler for DELETE /comments/:id
func deleteComment(service *comment.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		commenter = r.FormValue("commenter")

		vars := mux.Vars(r)
		id := vars["id"]

		(*logger).Log("trying to delete comment %d", id)

		err := (*service).DeleteComment(id)
		switch err {
		case nil:
			response.Status = "success"
			(*logger).Log("success deleting comment %d", id)

		case post.ErrPostNotFound:
			(*logger).Log("deletion of comment failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"commenter"`
			}
			responseData.Data = fmt.Sprintf("feed of commenter %s not found", commenter)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		default:
			(*logger).Log("deletion of from failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when deletion of comment"
			response.Data = responseData
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t\t")
		err = encoder.Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

	}
}

// getCommentReplies returns a handler for DELETE /comments/:id/replies
func getPostReplies(service *comment.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		id := vars["id"]

		(*logger).Log("trying to reply to a comment %d", id)

		p, err := (*service).GetComment(id)
		switch err {
		case nil:
			r, er := (*service).getCommentReplies(&c)
			switch er {
			case nil:
				response.Status = "success"
				response.Data = *r
				(*logger).Log("success fetching replies of a comment %d", id)
			case post.ErrReleaseNotFound:
				(*logger).Log("fetching of reply failed because: %s", err.Error())
				var responseData struct {
					Data int `json:"id"`
				}
				responseData.Data = fmt.Sprintf("replies of comment %d not found", id)
				response.Data = responseData
				w.WriteHeader(http.StatusNotFound)
			default:
				(*logger).Log("getting comment failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when getting post"
				response.Data = responseData
				w.WriteHeader(http.StatusInternalServerError)
			}

		case post.ErrFeedNotFound:
			(*logger).Log("fetching of comment failed because: %s", err.Error())
			var responseData struct {
				Data int `json:"id"`
			}
			responseData.Data = fmt.Sprintf("commment  %d not found", id)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		default:
			(*logger).Log("getting comment failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when getting comment"
			response.Data = responseData
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t\t")
		err = encoder.Encode(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

// POST /comments/:id/replies
