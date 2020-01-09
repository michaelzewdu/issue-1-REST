package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/channel"

	"strconv"
	"strings"

	"net/http"
)

func getChannel(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]

		(*logger).Log("trying to fetch channel %s", username)
		c, err := (*service).GetChannel(username)

		switch err {
		case nil:
			response.Status = "success"
			{
				//TODO
				//AUTHORIZATION
			}
			response.Data = *c
			(*logger).Log("success fetching channel %s", username)
		case channel.ErrChannelNotFound:
			(*logger).Log("Fetching of none existent channel %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("channel of username %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("Fetching of channel failed because %s", err)
			var responseData struct {
				Data string `json:"message"`
			}
			response.Status = "error"
			responseData.Data = "server error when fetching channel"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}

		writeResponseToWriter(response, w, statusCode)
	}
}
func postChannel(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		c := channel.Channel{}
		{
			c.Username = r.FormValue("username")
			if c.Username != "" {
				c.Name = r.FormValue("description")
				c.Description = r.FormValue("description")
			} else {
				err := json.NewDecoder(r.Body).Decode(&c)
				if err != nil {
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = `bad request, use format
				{"username":"username",
		        "name":"name",
				"description":"description"}`
					response.Data = responseData
					(*logger).Log("bad post channel request")
					statusCode = http.StatusBadRequest
				}
			}
		}
		if response.Data == nil {
			if c.Username == "" {
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "username is required"
				response.Data = responseData
			}
			if c.Name == "" {
				var responseData struct {
					Data string `json:"name"`
				}
				responseData.Data = "name is required"
				response.Data = responseData
			}

			if response.Data == nil {
				(*logger).Log("trying to add channel %s", c.Username, c.Name, c.Description)
				if &c != nil {
					err := (*service).AddChannel(&c)
					switch err {
					case nil:
						response.Status = "success"
						(*logger).Log("success adding channel %s", c.Username, c.Name, c.Description)
					case channel.ErrInvalidChannelData:
						(*logger).Log("creating of channel failed because: %s", err.Error())
						var responseData struct {
							Data string `json:"data"`
						}
						responseData.Data = "channel must have name & username to be created"
						response.Data = responseData
						statusCode = http.StatusBadRequest

					case channel.ErrUserNameOccupied:
						(*logger).Log("adding of channel failed because: %s", err.Error())
						var responseData struct {
							Data string `json:"username"`
						}
						responseData.Data = "username is occupied"
						response.Data = responseData
						statusCode = http.StatusConflict

					default:
						(*logger).Log("adding of channel failed because: %s", err.Error())
						var responseData struct {
							Data string `json:"message"`
						}
						response.Status = "error"
						responseData.Data = "server error when adding channel"
						response.Data = responseData
						statusCode = http.StatusInternalServerError
					}
				}

			} else {
				// if required fields aren't present
				(*logger).Log("bad adding channel request")
				statusCode = http.StatusBadRequest
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func putChannel(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		var c channel.Channel
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "bad request"
			response.Data = responseData
			(*logger).Log("bad update channel request")
			statusCode = http.StatusBadRequest
		} else {
			if c.Name == "" && c.Description == "" && c.Username == "" {
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "bad request"
				response.Data = responseData
				statusCode = http.StatusBadRequest
			} else {
				err := (*service).UpdateChannel(username, &c)
				switch err {
				case nil:
					(*logger).Log("success put channel %s", username)
					response.Status = "success"
				case channel.ErrUserNameOccupied:
					(*logger).Log("adding of channel failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"username"`
					}
					responseData.Data = "username is occupied by a channel"
					response.Data = responseData
					statusCode = http.StatusConflict
				case channel.ErrInvalidChannelData:
					(*logger).Log("updating of channel failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"data"`
					}
					responseData.Data = "channel must have name & username to be created"
					response.Data = responseData
					statusCode = http.StatusBadRequest
				default:
					(*logger).Log("update of channel failed because: %s", err.Error())
					var responseData struct {
						Data string `json:"message"`
					}
					responseData.Data = "server error when updating channel"
					response.Status = "error"
					response.Data = responseData
					statusCode = http.StatusInternalServerError
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func getChannels(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		pattern := ""
		limit := 25
		offset := 0
		var sortBy channel.SortBy
		var sortOrder channel.SortOrder
		{
			pattern = r.URL.Query().Get("pattern")
			if limitPageRaw := r.URL.Query().Get("limit"); limitPageRaw != "" {
				limit, err = strconv.Atoi(limitPageRaw)
				if err != nil || limit < 0 {
					(*logger).Log("bad get channels request, limit")
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
			sortOrder = channel.SortAscending
			switch sortByQuery := sortSplit[0]; sortByQuery {
			case "username":
				sortBy = channel.SortByUsername
			case "name":
				sortBy = channel.SortByName
			default:
				sortBy = channel.SortCreationTime
				sortOrder = channel.SortDescending
			}
			if len(sortSplit) > 1 {
				switch sortOrderQuery := sortSplit[1]; sortOrderQuery {
				case "dsc":
					sortOrder = channel.SortDescending
				default:
					sortOrder = channel.SortAscending
				}
			}
		}
		if response.Data == nil {
			channels, err := (*service).SearchChannels(pattern, sortBy, sortOrder, limit, offset)
			if err != nil {
				(*logger).Log("fetching of channels failed because: %s", err.Error())

				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when getting channels"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			} else {
				response.Status = "success"
				response.Data = channels
				(*logger).Log("success fetching channels")
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func deleteChannel(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		var err error
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		(*logger).Log("trying to delete channel %s", username)
		err = (*service).DeleteChannel(username)
		if err != nil {
			(*logger).Log("deletion of channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("username %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		} else {
			response.Status = "success"
			(*logger).Log("success deleting channel %s", username)
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func getAdmins(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		c, err := (*service).GetChannel(username)
		switch err {
		case nil:

			response.Status = "success"
			response.Data = c.AdminUsernames
			(*logger).Log("success fetching admins of channel %s", username)
		case channel.ErrChannelNotFound:
			(*logger).Log("fetch attempt of non existing channel %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("user of channel %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of admins of channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when fetching admins of channel"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func putAdmin(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]

		{
			//TODO
			//AUTHORIZATION
		}
		adminUsername := vars["adminUsername"]
		err := (*service).AddAdmin(username, adminUsername)
		switch err {
		case nil:
			response.Status = "success"
			(*logger).Log("success adding admin  %s in to channel %s", adminUsername, username)
		case channel.ErrChannelNotFound:
			(*logger).Log(fmt.Sprintf("Adding of Admin failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = "channel doesn't exits"
			response.Data = responseData
			statusCode = http.StatusNotFound
		case channel.ErrAdminNotFound:
			(*logger).Log(fmt.Sprintf("Adding of Admin failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"adminUsername"`
			}
			responseData.Data = "Admin user doesn't exits"
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log(fmt.Sprintf("Adding of Admin failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when Adding of Admin"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func deleteAdmin(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		adminUsername := vars["adminUsername"]
		err := (*service).DeleteAdmin(username, adminUsername)
		switch err {
		case nil:
			response.Status = "success"
			(*logger).Log("success deleting admin  %s in to channel %s", adminUsername, username)
		case channel.ErrChannelNotFound:
			(*logger).Log(fmt.Sprintf("Deleting of Admin failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = "channel doesn't exits"
			response.Data = responseData
			statusCode = http.StatusNotFound
		case channel.ErrAdminNotFound:
			(*logger).Log(fmt.Sprintf("Deleting of Admin failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"adminUsername"`
			}
			responseData.Data = "Admin user doesn't exits"
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log(fmt.Sprintf("Deleting of Admin failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when Deleting of Admin"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func getOwner(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		c, err := (*service).GetChannel(username)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = c.OwnerUsername
			(*logger).Log("success fetching owner of channel %s", username)
		case channel.ErrChannelNotFound:
			(*logger).Log("fetch attempt of non existing channel %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("channel of %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of owner of channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when fetching owner of channel"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}

		writeResponseToWriter(response, w, statusCode)
	}
}
func putOwner(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		ownerUsername := vars["ownerUsername"]
		err := (*service).ChangeOwner(username, ownerUsername)
		switch err {
		case nil:
			response.Status = "success"
			(*logger).Log("success updating owner of  %s  channel to %s", username, ownerUsername)
		case channel.ErrChannelNotFound:
			(*logger).Log(fmt.Sprintf("Update of owner failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = "channel doesn't exits"
			response.Data = responseData
			statusCode = http.StatusNotFound
		case channel.ErrOwnerNotFound:
			(*logger).Log(fmt.Sprintf("Update of owner failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"adminUsername"`
			}
			responseData.Data = "Owner user doesn't exits"
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log(fmt.Sprintf("Update of owner failed because: %s", err.Error()))
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when Update of owner"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

func getCatalog(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		c, err := (*service).GetChannel(username)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = c.ReleaseIDs
			(*logger).Log("success fetching catalog of channel %s", username)
		case channel.ErrChannelNotFound:
			(*logger).Log("fetch attempt of catalog from non existent channel %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("channel of %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of catalog of channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when fetching catalog of channel"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func getOfficialCatalog(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		c, err := (*service).GetChannel(username)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = c.OfficialReleaseIDs
			(*logger).Log("success fetching official catalog of channel %s", username)
		case channel.ErrChannelNotFound:
			(*logger).Log("fetch attempt of official catalog from non existent channel %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("channel of %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of official catalog of channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when fetching official catalog of channel"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func deleteReleaseFromCatalog(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		ReleaseID, err := strconv.Atoi(vars["catalogID"])
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, ReleaseID must be an integer`
			response.Data = responseData
			(*logger).Log("bad delete Release request")
			statusCode = http.StatusBadRequest

		} else {
			errC := (*service).DeleteReleaseFromCatalog(username, ReleaseID)
			switch errC {
			case nil:
				response.Status = "success"
				(*logger).Log("success deleting release  %s from channel %s's Catalog", ReleaseID, username)
			case channel.ErrChannelNotFound:
				(*logger).Log(fmt.Sprintf("Deleting of release failed because: %s", errC.Error()))
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "channel doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			case channel.ErrReleaseNotFound:
				(*logger).Log(fmt.Sprintf("Deleting of Admin failed because: %s", errC.Error()))
				var responseData struct {
					Data string `json:"releaseID"`
				}
				responseData.Data = "Release doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log(fmt.Sprintf("Deleting of Release failed because: %s", errC.Error()))
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when Deleting of Release"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}

		}

		writeResponseToWriter(response, w, statusCode)
	}
}
func getReleaseFromCatalog(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}

		ReleaseID, errC := strconv.Atoi(vars["catalogID"])
		if errC != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, ReleaseID must be an integer`
			response.Data = responseData
			(*logger).Log("bad delete Release request")
			statusCode = http.StatusBadRequest

		} else {
			c, err := (*service).GetChannel(username)
			(*logger).Log("geyying")
			switch err {
			case nil:
				for i := 0; i < len(c.ReleaseIDs); i++ {
					if c.ReleaseIDs[i] == ReleaseID {
						response.Status = "success"
						response.Data = c.ReleaseIDs
						(*logger).Log("success fetching release of  catalog of channel %s", username)
					} else {

						var responseData struct {
							Data string `json:"releaseID"`
						}
						responseData.Data = "release doesn't exits"
						response.Data = responseData
						statusCode = http.StatusNotFound

					}
				}
			case channel.ErrChannelNotFound:
				(*logger).Log("fetch attempt of catalog from non existent channel %s", username)
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = fmt.Sprintf("channel of %s not found", username)
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log("fetching of catalog of channel failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when fetching catalog of channel"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}

		writeResponseToWriter(response, w, statusCode)
	}
}
func putReleaseInCatalog(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//feature-release-service Put RELEASE handler
	}
}
func postReleaseInCatalog(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//feature-release-service Post RELEASE Handler
	}
}
func putReleaseInOfficialCatalog(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}

		releaseID, err := strconv.Atoi(vars["catalogID"])
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, releaseID must be an integer`
			response.Data = responseData
			(*logger).Log("bad bookmark post request")
			statusCode = http.StatusBadRequest
		}
		if response.Data == nil {
			c, errC := (*service).GetChannel(username)
			switch errC {
			case nil:
				for i := 0; i < len(c.ReleaseIDs); i++ {
					if c.ReleaseIDs[i] == releaseID {
						response.Status = "success"
						err := (*service).AddReleaseToOfficialCatalog(username, releaseID)
						switch err {
						case nil:
							(*logger).Log(fmt.Sprintf("success adding Release %d to channels %s's Catalog", releaseID, username))
							response.Status = "success"
						case channel.ErrChannelNotFound:
							(*logger).Log(fmt.Sprintf("Adding Release to Offical Catalog failed because: %s", err.Error()))
							var responseData struct {
								Data string `json:"username"`
							}
							responseData.Data = "channel doesn't exits"
							response.Data = responseData
							statusCode = http.StatusNotFound
						case channel.ErrReleaseNotFound:
							(*logger).Log(fmt.Sprintf("Adding Release to Offical Catalog failed because: %s", err.Error()))
							var responseData struct {
								Data string `json:"releaseID"`
							}
							responseData.Data = "release doesn't exits"
							response.Data = responseData
							statusCode = http.StatusNotFound
						default:
							(*logger).Log(fmt.Sprintf("Adding Release to Offical Catalog failed because: %s", err.Error()))
							var responseData struct {
								Data string `json:"message"`
							}
							responseData.Data = "server error when Adding Release to Offical Catalog "
							response.Status = "error"
							response.Data = responseData
							statusCode = http.StatusInternalServerError
						}
						break

					} else {

						var responseData struct {
							Data string `json:"releaseID"`
						}
						responseData.Data = "release doesn't exits"
						response.Data = responseData
						statusCode = http.StatusNotFound

					}
				}
			case channel.ErrChannelNotFound:
				(*logger).Log("fetch attempt of catalog from non existent channel %s", username)
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = fmt.Sprintf("channel of %s not found", username)
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:

				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when fetching catalog of channel"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}

		writeResponseToWriter(response, w, statusCode)
	}
}

func getPost(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		c, err := (*service).GetChannel(username)
		postID, errC := strconv.Atoi(vars["postID"])
		if errC != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, postID must be an integer`
			response.Data = responseData
			(*logger).Log("bad get post request")
			statusCode = http.StatusBadRequest

		} else {
			switch err {
			case nil:
				for i := 0; i < len(c.PostIDs); i++ {
					if c.PostIDs[i] == postID {
						response.Status = "success"
						response.Data = c.PostIDs[i]
						(*logger).Log("success fetching post of channel %s", username)
						break
					} else {

						var responseData struct {
							Data string `json:"postID"`
						}
						responseData.Data = "post doesn't exits"
						response.Data = responseData
						statusCode = http.StatusNotFound

					}
				}
			case channel.ErrChannelNotFound:
				(*logger).Log("fetch attempt of post from non existent channel %s", username)
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = fmt.Sprintf("channel of %s not found", username)
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log("fetching of post of channel failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when fetching post of channel"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func getPosts(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		c, err := (*service).GetChannel(username)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = c.PostIDs
			(*logger).Log("success fetching posts of channel %s", username)
		case channel.ErrChannelNotFound:
			(*logger).Log("fetch attempt of posts from non existent channel %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("channel of %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of posts of channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when fetching catalog of channel"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func getStickiedPosts(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		c, err := (*service).GetChannel(username)

		switch err {
		case nil:

			response.Status = "success"
			response.Data = c.StickiedPostIDs
			(*logger).Log("success fetching post of channel %s", username)

		case channel.ErrChannelNotFound:
			(*logger).Log("fetch attempt of post from non existent channel %s", username)
			var responseData struct {
				Data string `json:"username"`
			}
			responseData.Data = fmt.Sprintf("channel of %s not found", username)
			response.Data = responseData
			statusCode = http.StatusNotFound
		default:
			(*logger).Log("fetching of post of channel failed because: %s", err.Error())
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = "server error when fetching post of channel"
			response.Status = "error"
			response.Data = responseData
			statusCode = http.StatusInternalServerError
		}

		writeResponseToWriter(response, w, statusCode)
	}
}
func deleteStickiedPost(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}
		stickiedPostID, err := strconv.Atoi(vars["stickiedPostID"])
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, StickiedPostID must be an integer`
			response.Data = responseData
			(*logger).Log("bad delete StickiedPostID request")
			statusCode = http.StatusBadRequest

		} else {
			errC := (*service).DeleteStickiedPost(username, stickiedPostID)
			switch errC {
			case nil:
				response.Status = "success"
				(*logger).Log("success deleting stickied Post  %s from channel %s's Catalog", stickiedPostID, username)
			case channel.ErrChannelNotFound:
				(*logger).Log(fmt.Sprintf("Deleting of stickied Post failed because: %s", errC.Error()))
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "channel doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			case channel.ErrStickiedPostNotFound:
				(*logger).Log(fmt.Sprintf("Deleting of stickied post failed because: %s", errC.Error()))
				var responseData struct {
					Data string `json:"stickiedPostID"`
				}
				responseData.Data = "Stickied Post doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log(fmt.Sprintf("Deleting of Stickied Post failed because: %s", errC.Error()))
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when Deleting of Stickied Post"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
func stickyPost(service *channel.Service, logger *Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		username := vars["username"]
		{
			//TODO
			//AUTHORIZATION
		}

		stickyPost, err := strconv.Atoi(vars["postID"])
		if err != nil {
			var responseData struct {
				Data string `json:"message"`
			}
			responseData.Data = `bad request, PostID must be an integer`
			response.Data = responseData
			(*logger).Log("bad delete PostID request")
			statusCode = http.StatusBadRequest

		} else {
			err := (*service).StickyPost(username, stickyPost)
			switch err {
			case nil:
				response.Status = "success"
				(*logger).Log("success of stickying post  %s to channel  %s", stickyPost, username)
			case channel.ErrChannelNotFound:
				(*logger).Log(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"username"`
				}
				responseData.Data = "channel doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			case channel.ErrPostNotFound:
				(*logger).Log(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"postID"`
				}
				responseData.Data = "post doesn't exits"
				response.Data = responseData
				statusCode = http.StatusNotFound
			case channel.ErrStickiedPostFull:
				(*logger).Log(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"postID"`
				}
				responseData.Data = "stickied post full"
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				(*logger).Log(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "server error when stickying a post"
				response.Status = "error"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
