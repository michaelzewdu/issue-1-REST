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
func getFeed(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		statusCode := http.StatusOK

		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block secures the route
			if username != r.Header.Get("authorized_username") {
				s.Logger.Log("unauthorized user feed request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		s.Logger.Log("trying to fetch feed %s", username)
		f, err := s.FeedService.GetFeed(username)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = *f
			s.Logger.Log("success fetching feed of %s", username)
		case feed.ErrFeedNotFound:
			s.Logger.Log("fetching of feed failed because: %v", err)
			response.Data = jSendFailData{
				ErrorReason:  "username",
				ErrorMessage: fmt.Sprintf("feed of username %s not found", username),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Log("getting feed failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when getting feed"
			statusCode = http.StatusNotFound
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getFeedPosts returns a handler for GET /users/{username}/feed/posts?sort=new&limit=5&offset=0 requests
func getFeedPosts(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK

		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block secures the route
			if username != r.Header.Get("authorized_username") {
				s.Logger.Log("unauthorized user feed request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		s.Logger.Log("trying to fetch posts for feed")

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
					s.Logger.Log("bad get feed request, limit")
					response.Data = jSendFailData{
						ErrorReason:  "limit",
						ErrorMessage: "bad request, limit can't be negative",
					}
					statusCode = http.StatusBadRequest
				}
			}
			if offsetRaw := r.URL.Query().Get("offset"); offsetRaw != "" {
				offset, err = strconv.Atoi(offsetRaw)
				if err != nil || offset < 0 {
					s.Logger.Log("bad request, offset")
					response.Data = jSendFailData{
						ErrorReason:  "offset",
						ErrorMessage: "bad request, offset can't be negative",
					}
					statusCode = http.StatusBadRequest
				}
			}
		}
		// if queries are clean
		if response.Data == nil {
			posts, err := s.FeedService.GetPosts(&f, sort, limit, offset)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = posts
				s.Logger.Log("success fetching posts for feed")
				// TODO deliver actual posts from post service
			case feed.ErrFeedNotFound:
				s.Logger.Log("fetching of feed failed because: %v", err)
				response.Data = jSendFailData{
					ErrorReason:  "username",
					ErrorMessage: fmt.Sprintf("feed of username %s not found", username),
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Log("fetching of posts from feed failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when getting posts"
				statusCode = http.StatusNotFound
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getFeedChannels returns a handler for GET /users/{username}/feed/channels?sort=sub-time_dsc requests
func getFeedChannels(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO test function
		var err error
		var response jSendResponse
		statusCode := http.StatusOK

		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block secures the route
			if username != r.Header.Get("authorized_username") {
				s.Logger.Log("unauthorized user feed request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

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

		channels, err := s.FeedService.GetChannels(&feed.Feed{OwnerUsername: username}, sortBy, sortOrder)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = channels
			s.Logger.Log("success fetching channels of feed")
		case feed.ErrFeedNotFound:
			s.Logger.Log("fetching of feed failed because: %v", err)
			response.Data = jSendFailData{
				ErrorReason:  "username",
				ErrorMessage: fmt.Sprintf("feed of username %s not found", username),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Log("fetching of channels of feed failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when getting channels"
			statusCode = http.StatusNotFound
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// postFeedChannel returns a handler for POST /users/{username}/feed/channels requests
func postFeedChannel(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		statusCode := http.StatusCreated

		response.Status = "fail"
		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block secures the route
			if username != r.Header.Get("authorized_username") {
				s.Logger.Log("unauthorized user feed request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		c := feed.Channel{}
		c.Channelname = r.FormValue("channelname")
		if c.Channelname == "" {
			err := json.NewDecoder(r.Body).Decode(&c)
			if err != nil {
				response.Data = jSendFailData{
					ErrorReason: "request format",
					ErrorMessage: `bad request, use format
										{"channelname":"username"}`,
				}
				s.Logger.Log("bad subscribe channel request")
				statusCode = http.StatusBadRequest
			}
		}
		// if queries are clean
		if response.Data == nil {
			if c.Channelname == "" {
				response.Data = jSendFailData{
					ErrorReason:  "channelname",
					ErrorMessage: `channelname is required`,
				}
			}
			if response.Data == nil {
				err := s.FeedService.Subscribe(&feed.Feed{OwnerUsername: username}, c.Channelname)
				switch err {
				case nil:
					response.Status = "success"
					s.Logger.Log("success subscribing feed to channel")
				case feed.ErrFeedNotFound:
					s.Logger.Log("fetching of feed failed because: %v", err)
					response.Data = jSendFailData{
						ErrorReason:  "username",
						ErrorMessage: fmt.Sprintf("feed of username %s not found", username),
					}
					statusCode = http.StatusNotFound
				case feed.ErrChannelDoesNotExist:
					s.Logger.Log("fetching of feed failed because: %v", err)
					response.Data = jSendFailData{
						ErrorReason:  "channelname",
						ErrorMessage: fmt.Sprintf("channel of channelname %s not found", c.Channelname),
					}
					statusCode = http.StatusNotFound
				default:
					s.Logger.Log("subscribing feed to channel failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when subscribing to channel"
					statusCode = http.StatusNotFound
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// putFeed returns a handler for PUT /users/{username}/feed requests
func putFeed(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block secures the route
			if username != r.Header.Get("authorized_username") {
				s.Logger.Log("unauthorized user feed request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		var requestData struct {
			Sorting string `json:"defaultSorting"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			response.Data = jSendFailData{
				ErrorReason:  "request format",
				ErrorMessage: `bad request, use format {"defaultSorting":"hot"}`,
			}
			s.Logger.Log("bad update feed request")
			statusCode = http.StatusBadRequest
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
			switch oldFeed, err := s.FeedService.GetFeed(username); err {
			case nil:
				// if oldFeed exists, update
				if newFeed.Sorting == oldFeed.Sorting && newFeed.OwnerUsername == oldFeed.OwnerUsername {
					response.Status = "success"
				} else {
					s.Logger.Log("trying to update feed of user %s", username)
					if newFeed.OwnerUsername == "" {
						newFeed.OwnerUsername = username
					}
					if err = s.FeedService.UpdateFeed(oldFeed.ID, &newFeed); err != nil {

						s.Logger.Log("update of feed failed because: %v", err)
						response.Status = "error"
						response.Message = "server error when updating feed"
						statusCode = http.StatusNotFound
					} else {
						s.Logger.Log("success updating of feed %s", username)
						response.Status = "success"
					}
				}
			case feed.ErrFeedNotFound:
				// if oldFeed not found, create
				s.Logger.Log("creating new user because username on PUT not recognized: %v", err)

				newFeed.OwnerUsername = username //make sure created user has the new username
				err := s.FeedService.AddFeed(&newFeed)
				if err != nil {
					s.Logger.Log("creation of feed failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when creating Feed"
					statusCode = http.StatusNotFound
				} else {
					response.Status = "success"
				}
				s.Logger.Log("success creating feed for user %s", username)
			default:
				s.Logger.Log("updating of feed failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when updating Feed"
				statusCode = http.StatusNotFound
			}

		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteFeedChannel returns a handler for DELETE /users/{username}/feed/channels/{channelname} requests
func deleteFeedChannel(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block secures the route
			if username != r.Header.Get("authorized_username") {
				s.Logger.Log("unauthorized user feed request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		channelname := vars["channelname"]
		err := s.FeedService.Unsubscribe(&feed.Feed{OwnerUsername: username}, channelname)
		switch err {
		case nil:
			response.Status = "success"
			s.Logger.Log("success unsubscription feed from channel")
		case feed.ErrFeedNotFound:
			s.Logger.Log("deletion of feed failed because: %v", err)
			response.Data = jSendFailData{
				ErrorReason:  "username",
				ErrorMessage: fmt.Sprintf("feed of username %s not found", username),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Log("unsubscription of feed from failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when unsubscription from channel"
			statusCode = http.StatusNotFound
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
