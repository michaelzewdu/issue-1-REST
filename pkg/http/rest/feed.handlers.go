package rest

import (
	"encoding/json"
	"fmt"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// getFeed returns a handler for GET /users/{username}/feed/ requests
func getFeed(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		(*logger).Log("trying to fetch feed %s", username)

		f, err := (*service).GetFeed(username)
		if err != nil {
			(*logger).Log("fetching of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		} else {
			response.Status = "success"
			response.Data = *f

			(*logger).Log("success fetching feed of %s", username)
		}
		json.NewEncoder(w).Encode(response)
	}
}

// getFeedPosts returns a handler for GET /users/{username}/feed/posts?sort=new requests
func getFeedPosts(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		f, err := (*service).GetFeed(username)
		if err != nil {
			(*logger).Log("fetching of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		} else {
			(*logger).Log("trying to fetch posts for feed")

			limit := 25
			offset := 0
			var sort feed.Sorting
			{ // this block reads the query strings if any
				switch sortQuery := r.URL.Query().Get("sort"); sortQuery {
				case "hot":
					sort = feed.SortHot
				case "new":
					sort = feed.SortNew
				case "top":
					sort = feed.SortTop
				default:
					sort = f.Sorting
				}
				if limitPageRaw := r.URL.Query().Get("limit"); limitPageRaw != "" {
					limit, err = strconv.Atoi(limitPageRaw)
					if err != nil || limit < 0 {
						(*logger).Log("bad get feed request, limit")
						var responseData struct {
							Data string `json:"limit"`
						}
						responseData.Data = "bad request, limit can't be negative"
						response.Data = responseData
						w.WriteHeader(http.StatusBadRequest)
					}
				}
				if offsetRaw := r.URL.Query().Get("offset"); offsetRaw != "" {
					offset, err = strconv.Atoi(offsetRaw)
					if err != nil || offset < 0 {
						(*logger).Log("bad request, offset")
						var responseData struct {
							Data string `json:"offset"`
						}
						responseData.Data = "bad request, offset can't be negative"
						response.Data = responseData
						w.WriteHeader(http.StatusBadRequest)
					}
				}
			}
			// if queries are clean
			if response.Data == nil {
				posts, err := (*service).GetPosts(f, sort, limit, offset)
				if err != nil {
					(*logger).Log("fetching of posts from feed failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when getting posts"
					response.Data = responseData
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
					response.Data = posts
					(*logger).Log("success fetching posts for feed")
				}
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}

// getFeedChannels returns a handler for GET /users/{username}/feed/channels requests
func getFeedChannels(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		f, err := (*service).GetFeed(username)
		if err != nil {
			(*logger).Log("fetching of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		} else {
			if response.Data == nil {
				channels, err := (*service).GetChannels(f)
				if err != nil {
					(*logger).Log("fetching of channels of feed failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when getting channels"
					response.Data = responseData
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
					response.Data = channels
					(*logger).Log("success fetching channels of feed")
				}
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}

// postFeedChannel returns a handler for POST /users/{username}/feed/channels requests
func postFeedChannel(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		f, err := (*service).GetFeed(username)
		if err != nil {
			(*logger).Log("subscribing feed to channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		} else {
			c := feed.Channel{}
			channelUsername := r.FormValue("username")
			if channelUsername == "" {
				err := json.NewDecoder(r.Body).Decode(&c)
				if err != nil {
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = `bad request, use format
										{"channelUsername":"username"}`
					response.Data = responseData
					(*logger).Log("bad subscribe channel request")
					w.WriteHeader(http.StatusBadRequest)
				}
			}

			// if queries are clean
			if response.Data == nil {
				err := (*service).SubscribeChannel(&c, f)
				if err != nil {
					(*logger).Log("subscribing feed to channel failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when subscribing to channel"
					response.Data = responseData
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
					(*logger).Log("success subscribing feed to channel")
				}
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}

// putFeed returns a handler for PUT /users/{username}/feed requests
func putFeed(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		var newFeed feed.Feed
		err := json.NewDecoder(r.Body).Decode(&newFeed)

		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "bad request"
			response.Data = responseData
			(*logger).Log("bad update feed request")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			// if JSON parsing doesn't fail
			if oldFeed, err := (*service).GetFeed(username); err != nil {
				// if PUT username doesn't exist, create a new user
				(*logger).Log("creating new user because username on PUT not recognized: %s", err.Error())

				newFeed.OwnerUsername = username //make sure created user has the new username

				err := (*service).AddFeed(&newFeed)
				if err != nil {
					(*logger).Log("creation of feed failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when creating Feed"
					response.Data = responseData
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
				}
				(*logger).Log("success creating feed for user %s", username)
			} else {
				// else, update user

				// check data for bad request
				if newFeed.Sorting == oldFeed.Sorting && newFeed.OwnerUsername == oldFeed.OwnerUsername {
					response.Status = "success"
				} else {
					(*logger).Log("trying to update feed of user %s", username)
					if newFeed.OwnerUsername == "" {
						newFeed.OwnerUsername = username
					}
					if err = (*service).UpdateFeed(&newFeed); err != nil {

						(*logger).Log("update of feed failed because: %s", err.Error())

						var responseData struct {
							Data string `json:"message"`
						}
						responseData.Data = "server error when updating user"
						response.Status = "error"
						response.Data = responseData
						w.WriteHeader(http.StatusNotFound)
					} else {
						(*logger).Log("success updating of user %s", username)
						response.Status = "success"
					}
				}
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}

// deleteFeedChannel returns a handler for DELETE /users/{username}/feed/channels/{channelname} requests
func deleteFeedChannel(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		f, err := (*service).GetFeed(username)
		if err != nil {
			(*logger).Log("fetching of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		} else {
			channelname := vars["channelname"]
			if response.Data == nil {
				err := (*service).UnsubscribeChannel(channelname, f)
				if err != nil {
					(*logger).Log("unsubscription of feed from failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when unsubscription from channel"
					response.Data = responseData
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
					(*logger).Log("success unsubscription feed from channel")
				}
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}
