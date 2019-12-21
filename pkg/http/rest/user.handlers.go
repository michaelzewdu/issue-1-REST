package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/slim-crown/Issue-1-REST/pkg/domain/user"

	"github.com/gorilla/mux"
)

// postUser returns a handler for POST /users requests
func postUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		encoder := json.NewEncoder(w)

		user := user.User{}

		// checks if requests uses forms or JSON
		user.Username = r.FormValue("username")
		if user.Username != "" {
			user.PassHash = r.FormValue("passHash")
			user.Email = r.FormValue("email")
			user.FirstName = r.FormValue("firstName")
			user.MiddleName = r.FormValue("middleName")
			user.LastName = r.FormValue("lastName")
		} else {
			err := json.NewDecoder(r.Body).Decode(&user)
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
				w.WriteHeader(http.StatusBadRequest)
				encoder.Encode(response)
				return
			}
		}
		{ // this block checks for required fields
			if user.Username == "" {
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "username is required"
				response.Data = responseData
			}
			if user.PassHash == "" {
				var responseData struct {
					Data string `json:"passHash"`
				}
				responseData.Data = "passHash is required"
				response.Data = responseData
			}
			if user.FirstName == "" {
				var responseData struct {
					Data string `json:"firstName"`
				}
				responseData.Data = "firstName is required"
				response.Data = responseData
			}
		}
		if response.Data != nil {
			// if required fields aren't present
			(*logger).Log("bad adding user request")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			(*logger).Log("trying to add user %s", user.Username, user.PassHash, user.Email, user.FirstName, user.MiddleName, user.LastName)

			if err := (*service).AddUser(&user); err != nil {
				(*logger).Log("adding of user failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when adding user"
				response.Data = responseData
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				response.Status = "success"
				(*logger).Log("success adding user %s", user.Username, user.PassHash, user.Email, user.FirstName, user.MiddleName, user.LastName)
			}
		}
		encoder.Encode(response)
	}
}

// getUser returns a handler for GET /users/{username} requests
func getUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		(*logger).Log("trying to fetch user %s", username)

		user, err := (*service).GetUser(username)
		if err != nil {
			(*logger).Log("fetching of user failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		} else {
			response.Status = "success"
			response.Data = *user

			(*logger).Log("success fetching user %s", username)
		}
		json.NewEncoder(w).Encode(response)

	}
}

// getUsers returns a handler for GET /users requests
func getUsers(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		var response jSendResponse
		response.Status = "fail"

		(*logger).Log("trying to fetch all user")

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

			sort := r.URL.Query().Get("pattern")
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
		(*logger).Log("testing: %s", string(sortBy))
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
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				response.Status = "success"
				response.Data = users
				(*logger).Log("success fetching users")
			}
		}
		json.NewEncoder(w).Encode(response)

	}
}

// postUser returns a handler for PUT /users/{username} requests
func putUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		vars := mux.Vars(r)
		username := vars["username"]

		// TODO token authorization
		var user user.User
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "bad request"
			response.Data = responseData
			(*logger).Log("bad update user request")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			// if JSON parsing doesn't fail
			if _, err = (*service).GetUser(username); err != nil {
				// if PUT username doesn't exist, create a new user

				(*logger).Log("creating new user because username on PUT not recognized: %s", err.Error())

				user.Username = username //make sure created user has the new username

				err := (*service).AddUser(&user)
				if err != nil {
					(*logger).Log("adding of user failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					responseData.Data = "server error when adding user"
					response.Data = responseData
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
				}
				(*logger).Log("success adding user %s", user.Username, user.PassHash, user.Email, user.FirstName, user.MiddleName, user.LastName)
			} else {
				// else, update user

				// check data for bad request
				if user.FirstName == "" && user.Username == "" && user.Email == "" && user.LastName == "" && user.MiddleName == "" && user.PassHash == "" {
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = "bad request"
					response.Data = responseData
					w.WriteHeader(http.StatusBadRequest)
				} else {
					(*logger).Log("trying to update user %s", username, user.Username, user.PassHash, user.Email, user.FirstName, user.MiddleName, user.LastName)

					if err = (*service).UpdateUser(username, &user); err != nil {

						(*logger).Log("update of user failed because: %s", err.Error())

						var responseData struct {
							Data string `json:"message"`
						}
						responseData.Data = "server error when updating user"
						response.Status = "error"
						response.Data = responseData
						w.WriteHeader(http.StatusNotFound)
					} else {
						(*logger).Log("success update user %s", username, user.Username, user.PassHash, user.Email, user.FirstName, user.MiddleName, user.LastName)
						response.Status = "success"
					}
				}
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}

// deleteUser returns a handler for DELETE /users/{username} requests
func deleteUser(service *user.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"

		// TODO token authorization

		vars := mux.Vars(r)
		username := vars["username"]

		(*logger).Log("trying to delete user %s", username)

		err := (*service).DeleteUser(username)
		if err != nil {
			(*logger).Log("deletion of user failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("username %s not found", username)
			response.Data = responseData
			w.WriteHeader(http.StatusNotFound)
		} else {
			response.Status = "success"
			(*logger).Log("success deleting user %s", username)
		}
		json.NewEncoder(w).Encode(response)
	}
}
