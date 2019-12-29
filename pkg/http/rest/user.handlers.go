package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/slim-crown/issue-1-REST/pkg/domain/user"

	"github.com/gorilla/mux"
)

// postUser returns a handler for POST /users requests
func postUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		u := user.User{}

		{ // checks if requests uses forms or JSON and parses then
			u.Username = r.FormValue("username")
			if u.Username != "" {
				u.Password = r.FormValue("passHash")
				u.Email = r.FormValue("email")
				u.FirstName = r.FormValue("firstName")
				u.MiddleName = r.FormValue("middleName")
				u.LastName = r.FormValue("lastName")
			} else {
				err := json.NewDecoder(r.Body).Decode(&u)
				if err != nil {
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = `bad request, use format
				{"username":"username",
				"passHash":"passHash",
				"email":"email",
				"firstName":"firstName",
				"middleName":"middleName",
				"lastName":"lastName"}`
					response.Data = responseData
					(*logger).Log("bad update user request")
					statusCode = http.StatusBadRequest
				}
			}
		}
		if response.Data == nil {
			// this block checks for required fields
			if u.Username == "" {
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "username is required"
				response.Data = responseData
			}
			if u.Password == "" {
				var responseData struct {
					Data string `json:"passHash"`
				}
				responseData.Data = "passHash is required"
				response.Data = responseData
			}
			if u.FirstName == "" {
				var responseData struct {
					Data string `json:"firstName"`
				}
				responseData.Data = "firstName is required"
				response.Data = responseData
			}
			if response.Data == nil {
				(*logger).Log("trying to add user %s", u.Username, u.Email, u.FirstName, u.MiddleName, u.LastName, u.Password)
				// TODO u name is occupied error
				// TODO email occupied
				err := (*service).AddUser(&u)
				switch err {
				case nil:
					response.Status = "success"
					(*logger).Log("success adding user %s", u.Username, u.Email, u.FirstName, u.MiddleName, u.LastName, u.Password)
				case user.ErrUserNameOccupied:
					(*logger).Log("adding of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"username"`
					}
					responseData.Data = "username is occupied"
					response.Data = responseData
					statusCode = http.StatusConflict
				case user.ErrEmailIsOccupied:
					(*logger).Log("adding of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"email"`
					}
					responseData.Data = "email is occupied"
					response.Data = responseData
					statusCode = http.StatusConflict
				default:
					(*logger).Log("adding of user failed because: %s", err.Error())
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

// getUser returns a handler for GET /users/{username} requests
func getUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
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

// getUsers returns a handler for GET /users?sort=new&limit=5&offset=0&pattern=Joe requests
func getUsers(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		pattern := ""
		limit := 25
		offset := 0
		var sortBy user.SortBy
		var sortOrder user.SortOrder

		{ // this block reads the query strings if any
			pattern = r.URL.Query().Get("pattern")

			if limitPageRaw := r.URL.Query().Get("limit"); limitPageRaw != "" {
				limit, err = strconv.Atoi(limitPageRaw)
				if err != nil || limit < 0 {
					(*logger).Log("bad get users request, limit")
					var responseData struct {
						Data string `json:"limit"`
					}
					responseData.Data = "bad request, limit can't be negative"
					response.Data = responseData
					statusCode = http.StatusBadRequest
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
					statusCode = http.StatusBadRequest
				}
			}

			sort := r.URL.Query().Get("sort")
			sortSplit := strings.Split(sort, "_")

			sortOrder = user.SortAscending
			switch sortByQuery := sortSplit[0]; sortByQuery {
			case "username":
				sortBy = user.SortByUsername
			case "firstname":
				sortBy = user.SortByFirstName
			case "lastname":
				sortBy = user.SortByLastName
			default:
				sortBy = user.SortCreationTime
				sortOrder = user.SortDescending
			}
			if len(sortSplit) > 1 {
				switch sortOrderQuery := sortSplit[1]; sortOrderQuery {
				case "dsc":
					sortOrder = user.SortDescending
				default:
					sortOrder = user.SortAscending
				}
			}

		}
		// if queries are clean
		if response.Data == nil {
			users, err := (*service).SearchUser(pattern, sortBy, sortOrder, limit, offset)
			if err != nil {
				(*logger).Log("fetching of users failed because: %s", err.Error())

				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when getting users"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			} else {
				response.Status = "success"
				response.Data = users
				(*logger).Log("success fetching users")
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// postUser returns a handler for PUT /users/{username} requests
func putUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		username := vars["username"]

		{ // this block blocks user updating of user if is not the user herself accessing the route
			if username != r.Header.Get("authorized_username") {
				if _, err := (*service).GetUser(username); err == nil {
					(*logger).Log("unauthorized update user attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		}
		var u user.User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "bad request"
			response.Data = responseData
			(*logger).Log("bad update user request")
			statusCode = http.StatusBadRequest
		} else {
			// if JSON parsing doesn't fail
			if u.FirstName == "" && u.Username == "" && u.Bio == "" && u.Email == "" && u.LastName == "" && u.MiddleName == "" && u.Password == "" {
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "bad request"
				response.Data = responseData
				statusCode = http.StatusBadRequest
			} else {
				err = (*service).UpdateUser(username, &u)
				switch err {
				case nil:
					(*logger).Log("success put user %s", username, u.Username, u.Password, u.Email, u.FirstName, u.MiddleName, u.LastName)
					response.Status = "success"
				case user.ErrUserNameOccupied:
					(*logger).Log("adding of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"username"`
					}
					responseData.Data = "username is occupied by a channel"
					response.Data = responseData
					statusCode = http.StatusConflict
				case user.ErrEmailIsOccupied:
					(*logger).Log("adding of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"email"`
					}
					responseData.Data = "email is occupied"
					response.Data = responseData
					statusCode = http.StatusConflict
				case user.ErrInvalidUserData:
					(*logger).Log("adding of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"data"`
					}
					responseData.Data = "user must have email & password to be created"
					response.Data = responseData
					statusCode = http.StatusBadRequest
				default:
					(*logger).Log("update of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = "server error when updating user"
					response.Status = "error"
					response.Data = responseData
					statusCode = http.StatusInternalServerError
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteUser returns a handler for DELETE /users/{username} requests
func deleteUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block blocks user deletion of a user if is not the user herself accessing the route
			if username != r.Header.Get("authorized_username") {
				(*logger).Log("unauthorized update user attempt")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		(*logger).Log("trying to delete user %s", username)
		err := (*service).DeleteUser(username)
		if err != nil {
			(*logger).Log("deletion of user failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("username %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		} else {
			response.Status = "success"
			(*logger).Log("success deleting user %s", username)
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getUserBookmarks returns a handler for GET /users/{username}/bookmarks/ requests
func getUserBookmarks(service *user.Service, logger *Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block blocks user deletion of a user if is not the user herself accessing the route
			if username != r.Header.Get("authorized_username") {
				(*logger).Log("unauthorized update user attempt")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		u, err := (*service).GetUser(username)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = u.BookmarkedPosts
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
			responseData.Data = "server error when fetching user bookmarks"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// postUserBookmarks returns a handler for POST /users/{username}/bookmarks/ requests
func postUserBookmarks(service *user.Service, logger *Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block blocks user deletion of a user if is not the user herself accessing the route
			if username != r.Header.Get("authorized_username") {
				(*logger).Log("unauthorized update user attempt")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		var post struct {
			PostID int `json:"postID"`
		}
		{ // this block extracts post ID from the request
			temp := r.FormValue("postID")
			if temp != "" {
				post.PostID, err = strconv.Atoi(temp)
				if err != nil {
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = `bad request, postID must be an integer`
					response.Data = responseData
					(*logger).Log("bad bookmark post request")
					statusCode = http.StatusBadRequest
				}
			} else {
				err := json.NewDecoder(r.Body).Decode(&post)
				if err != nil {
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = `bad request, use format
										{"postID":"postID"}`
					response.Data = responseData
					(*logger).Log("bad bookmark post request")
					statusCode = http.StatusBadRequest
				}
			}
		}
		// if queries are clean
		(*logger).Log(fmt.Sprintf("bookmarking post: %v", post))
		if response.Data == nil {
			err := (*service).BookmarkPost(username, post.PostID)
			switch err {
			case nil:
				(*logger).Log(fmt.Sprintf("success adding bookmark %d to user %s", post.PostID, username))
				response.Status = "success"
			case user.ErrUserNotFound:
				(*logger).Log(fmt.Sprintf("bookmarking of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "user doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			case user.ErrPostNotFound:
				(*logger).Log(fmt.Sprintf("bookmarking of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"postID"`
				}
				responseData.Data = "post doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log(fmt.Sprintf("bookmarking of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when bookmarking post"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// putUserBookmarks returns a handler for PUT /users/{username}/bookmarks/ requests
func putUserBookmarks(service *user.Service, logger *Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		username := vars["username"]
		{ // this block blocks user deletion of a user if is not the user herself accessing the route
			if username != r.Header.Get("authorized_username") {
				(*logger).Log("unauthorized update user attempt")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		postID, err := strconv.Atoi(vars["postID"])
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, postID must be an integer`
			response.Data = responseData
			(*logger).Log("bad bookmark post request")
			statusCode = http.StatusBadRequest
		}
		// if queries are clean
		if response.Data == nil {
			err := (*service).BookmarkPost(username, postID)
			switch err {
			case nil:
				(*logger).Log(fmt.Sprintf("success adding bookmark %d to user %s", postID, username))
				response.Status = "success"
			case user.ErrUserNotFound:
				(*logger).Log(fmt.Sprintf("bookmarking of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "user doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			case user.ErrPostNotFound:
				(*logger).Log(fmt.Sprintf("bookmarking of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"postID"`
				}
				responseData.Data = "post doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log(fmt.Sprintf("bookmarking of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when bookmarking post"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteUserBookmarks returns a handler for DELETE /users/{username}/bookmarks/{postID} requests
func deleteUserBookmarks(service *user.Service, logger *Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		username := vars["username"]
		postID, err := strconv.Atoi(vars["postID"])
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, postID must be an integer`
			response.Data = responseData
			(*logger).Log("bad bookmark post request")
			statusCode = http.StatusBadRequest
		}
		{ // this block blocks user deletion of a user if is not the user herself accessing the route
			if username != r.Header.Get("authorized_username") {
				(*logger).Log("unauthorized update user attempt")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		// if queries are clean
		if response.Data == nil {
			err = (*service).DeleteBookmark(username, postID)
			switch err {
			case nil:
				(*logger).Log(fmt.Sprintf("success removing bookmark %d from user %s", postID, username))
				response.Status = "success"
			case user.ErrUserNotFound:
				(*logger).Log(fmt.Sprintf("deletion of bookmark failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "user doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log(fmt.Sprintf("deletion of bookmark failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when bookmarking post"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
