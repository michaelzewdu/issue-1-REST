package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/slim-crown/Issue-1/pkg/domain/user"

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
				var data struct {
					Data string `json:"message"`
				}
				data.Data = `bad request, use format
				{"username":"username",
				"passHash":"passHash",
				"email":"email",
				"firstName":"firstName",
				"middleName":"middleName",
				"lastName":"lastName"}`
				response.Data = data
				(*logger).Log("bad update user request")
				w.WriteHeader(http.StatusBadRequest)
				encoder.Encode(response)
				return
			}
		}
		{ // this block checks for required fields
			if user.Username == "" {
				var data struct {
					Data string `json:"username"`
				}
				data.Data = "username is required"
				response.Data = data
			}
			if user.PassHash == "" {
				var data struct {
					Data string `json:"passHash"`
				}
				data.Data = "passHash is required"
				response.Data = data
			}
			if user.FirstName == "" {
				var data struct {
					Data string `json:"firstName"`
				}
				data.Data = "firstName is required"
				response.Data = data
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
				var data struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				data.Data = "server error when adding user"
				response.Data = data
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
			var data struct {
				Data string `json:"username"`
			}
			data.Data = fmt.Sprintf("username %s not found", username)
			response.Data = data
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
		var sortBy, sortOrder string

		{ // this block reads the query strings if any
			switch sortByQuery := r.URL.Query().Get("sortBy"); sortByQuery {
			case "username":
				sortBy = user.SortByUsername
			case "first-name":
				sortBy = user.SortByFirstName
			case "last-name":
				sortBy = user.SortByLastName
			default:
				sortBy = user.SortCreationTime
				sortOrder = user.SortDescending
			}
			switch sortOrderQuery := r.URL.Query().Get("sortOrder"); sortOrderQuery {
			case "dsc":
				sortOrder = user.SortDescending
			default:
				sortOrder = user.SortAscending
			}
			if limitPageRaw := r.URL.Query().Get("limit"); limitPageRaw != "" {
				limit, err = strconv.Atoi(limitPageRaw)
				if err != nil || limit < 0 {
					(*logger).Log("bad get users request, limit")
					var data struct {
						Data string `json:"limit"`
					}
					data.Data = "bad request, limit can't be negative"
					w.WriteHeader(http.StatusBadRequest)
				}
			}
			if offsetRaw := r.URL.Query().Get("offset"); offsetRaw != "" {
				offset, err = strconv.Atoi(offsetRaw)
				if err != nil || offset < 0 {
					(*logger).Log("bad request, offset")
					var data struct {
						Data string `json:"offset"`
					}
					data.Data = "bad request, offset can't be negative"
					response.Data = data
					w.WriteHeader(http.StatusBadRequest)
				}
			}
			// pattern = r.URL.Query().Get("pattern")
		}
		// if queries are clean
		if response.Data == nil {
			users, err := (*service).SearchUser(pattern, sortBy, sortOrder, limit, offset)
			if err != nil {
				(*logger).Log("fetching of users failed because: %s", err.Error())

				var data struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				data.Data = "server error when getting users"
				response.Data = data
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
			var data struct {
				Data string `json:"message"`
			}
			data.Data = "bad request"
			response.Data = data
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
					var data struct {
						Data string `json:"message"`
					}
					response.Status = "error"
					data.Data = "server error when adding user"
					response.Data = data
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					response.Status = "success"
				}
				(*logger).Log("success adding user %s", user.Username, user.PassHash, user.Email, user.FirstName, user.MiddleName, user.LastName)
			} else {
				// else, update user

				// check data for bad request
				if user.FirstName == "" && user.Username == "" && user.Email == "" && user.LastName == "" && user.MiddleName == "" && user.PassHash == "" {
					var data struct {
						Data string `json:"message"`
					}
					data.Data = "bad request"
					response.Data = data
					w.WriteHeader(http.StatusBadRequest)
				} else {
					(*logger).Log("trying to update user %s", username, user.Username, user.PassHash, user.Email, user.FirstName, user.MiddleName, user.LastName)

					if err = (*service).UpdateUser(username, &user); err != nil {

						(*logger).Log("update of user failed because: %s", err.Error())

						var data struct {
							Data string `json:"message"`
						}
						data.Data = "server error when updating user"
						response.Status = "error"
						response.Data = data
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
			var data struct {
				Data string `json:"username"`
			}
			data.Data = fmt.Sprintf("username %s not found", username)
			response.Data = data
			w.WriteHeader(http.StatusNotFound)
		} else {
			response.Status = "success"
			(*logger).Log("success deleting user %s", username)
		}
		json.NewEncoder(w).Encode(response)
	}
}
