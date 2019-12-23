package rest

import (
	"encoding/json"
	"fmt"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"net/http"
	"strconv"
	"strings"

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
		switch err {
		case nil:
			response.Status = "success"
			response.Data = *f
			(*logger).Log("success fetching feed of %s", username)
		case feed.ErrFeedNotFound:
			(*logger).Log("fetching of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		default:
			(*logger).Log("getting feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when getting feed"
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

// getFeedPosts returns a handler for GET /users/{username}/feed/posts?sort=new&limit=5&offset=0 requests
func getFeedPosts(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		(*logger).Log("trying to fetch posts for feed")

		f := feed.Feed{OwnerUsername: username}

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
				sort = feed.NotSet
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
			posts, err := (*service).GetPosts(&f, sort, limit, offset)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = posts
				(*logger).Log("success fetching posts for feed")
				// TODO deliver actual posts from post service
			case feed.ErrFeedNotFound:
				(*logger).Log("fetching of feed failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = fmt.Sprintf("feed of username %s not found", username)
				response.Data = responseData
				w.WriteHeader(http.StatusNotFound)
			default:
				(*logger).Log("fetching of posts from feed failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when getting posts"
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

// getFeedChannels returns a handler for GET /users/{username}/feed/channels?sort=sub-time_dsc requests
func getFeedChannels(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO test function
		var err error
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		var sortBy feed.SortBy
		var sortOrder feed.SortOrder
		{ // this block reads the query strings if any

			sort := r.URL.Query().Get("sort")
			sortSplit := strings.Split(sort, "_")

			switch sortByQuery := sortSplit[0]; sortByQuery {
			case "username":
				sortBy = feed.SortByUsername
			case "name":
				sortBy = feed.SortByName
			case "sub-time":
				fallthrough
			default:
				sortBy = feed.SortBySubscriptionTime
				sortOrder = feed.SortDescending
			}
			sortOrder = feed.SortAscending
			if len(sortSplit) > 1 {
				switch sortOrderQuery := sortSplit[1]; sortOrderQuery {
				case "dsc":
					sortOrder = feed.SortDescending
				case "asc":
					fallthrough
				default:
					sortOrder = feed.SortAscending
				}
			}
		}

		channels, err := (*service).GetChannels(&feed.Feed{OwnerUsername: username}, sortBy, sortOrder)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = channels
			(*logger).Log("success fetching channels of feed")
		case feed.ErrFeedNotFound:
			(*logger).Log("fetching of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		default:
			(*logger).Log("fetching of channels of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when getting channels"
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

// postFeedChannel returns a handler for POST /users/{username}/feed/channels requests
func postFeedChannel(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		vars := mux.Vars(r)
		username := vars["username"]

		c := feed.Channel{}
		c.Username = r.FormValue("username")
		if c.Username == "" {
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
			err := (*service).Subscribe(&feed.Feed{OwnerUsername: username}, c.Username)
			switch err {
			case nil:
				response.Status = "success"
				(*logger).Log("success subscribing feed to channel")
			case feed.ErrFeedNotFound:
				(*logger).Log("fetching of feed failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = fmt.Sprintf("feed of username %s not found", username)
				response.Data = responseData
				w.WriteHeader(http.StatusNotFound)
			case feed.ErrChannelDoesNotExist:
				(*logger).Log("fetching of feed failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"channel"`
				}
				responseData.Data = fmt.Sprintf("channel of username %s not found", c.Username)
				response.Data = responseData
				w.WriteHeader(http.StatusNotFound)
			default:
				(*logger).Log("subscribing feed to channel failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when subscribing to channel"
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

// putFeed returns a handler for PUT /users/{username}/feed requests
func putFeed(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]
		// TODO check auth
		var requestData struct {
			Sorting string `json:"defaultSorting"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, use format {"defaultSorting":"hot"}`
			response.Data = responseData
			(*logger).Log("bad update feed request")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			newFeed := feed.Feed{}
			{
				switch requestData.Sorting {
				case "hot":
					newFeed.Sorting = feed.SortHot
				case "new":
					newFeed.Sorting = feed.SortNew
				case "top":
					newFeed.Sorting = feed.SortTop
				default:
					newFeed.Sorting = feed.NotSet
				}
			}
			switch oldFeed, err := (*service).GetFeed(username); err {
			case nil:
				// if oldFeed exists, update
				if newFeed.Sorting == oldFeed.Sorting && newFeed.OwnerUsername == oldFeed.OwnerUsername {
					response.Status = "success"
				} else {
					(*logger).Log("trying to update feed of user %s", username)
					if newFeed.OwnerUsername == "" {
						newFeed.OwnerUsername = username
					}
					if err = (*service).UpdateFeed(oldFeed.ID, &newFeed); err != nil {

						(*logger).Log("update of feed failed because: %s", err.Error())

						var responseData struct {
							Data string `json:"message"`
						}
						responseData.Data = "server error when updating feed"
						response.Status = "error"
						response.Data = responseData
						w.WriteHeader(http.StatusNotFound)
					} else {
						(*logger).Log("success updating of feed %s", username)
						response.Status = "success"
					}
				}
			case feed.ErrFeedNotFound:
				// if oldFeed not found, create
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
			default:
				(*logger).Log("updating of feed failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when updating Feed"
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

// deleteFeedChannel returns a handler for DELETE /users/{username}/feed/channels/{channelname} requests
func deleteFeedChannel(service *feed.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		channelname := vars["channelname"]
		err := (*service).Unsubscribe(&feed.Feed{OwnerUsername: username}, channelname)
		switch err {
		case nil:
			response.Status = "success"
			(*logger).Log("success unsubscription feed from channel")
		case feed.ErrChannelDoesNotExist:
			(*logger).Log("fetching of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"channel"`
			}
			responseData.Data = fmt.Sprintf("channel of username %s not found", channelname)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		case feed.ErrFeedNotFound:
			(*logger).Log("deletion of feed failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("feed of username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		default:
			(*logger).Log("unsubscription of feed from failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when unsubscription from channel"
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
