package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/channel"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
	"net/url"
	"os"

	"strconv"
	"strings"

	"net/http"
)

func sanitizeChannel(c *channel.Channel, s *Setup) {
	c.ChannelUsername = s.StrictSanitizer.Sanitize(c.ChannelUsername)
	c.Name = s.StrictSanitizer.Sanitize(c.Name)
	c.Description = s.StrictSanitizer.Sanitize(c.Description)
}

// getChannel returns a handler for GET /channels/{channelUsername} requests
func getChannel(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]

		s.Logger.Printf("trying to fetch channel %s", channelUsername)
		c, err := s.ChannelService.GetChannel(channelUsername)

		switch err {
		case nil:
			response.Status = "success"
			{
				// this block sanitizes the returned User if it's not the user herself accessing the route
				if channelUsername != r.Header.Get("authorized_username") {
					s.Logger.Printf("user %s fetched channel %s", r.Header.Get("authorized_username"), c.ChannelUsername)
					c.AdminUsernames = nil
					c.ReleaseIDs = nil
					c.OwnerUsername = ""
					c.AdminUsernames = nil

				}
			}
			if c.PictureURL != "" {
				c.PictureURL = s.HostAddress + s.ImageServingRoute + url.PathEscape(c.PictureURL)
			}
			response.Data = *c
			s.Logger.Printf("success fetching channel %s", channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf("Fetching of none existent channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channel of channelUsername %s not found", channelUsername),
			}

			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("Fetching of channel failed because %s", err)
			response.Data = jSendFailData{
				ErrorReason:  "Error",
				ErrorMessage: "server error when fetching channel",
			}

			statusCode = http.StatusInternalServerError
		}

		writeResponseToWriter(response, w, statusCode)
	}
}


// postChannel returns a handler for POST /channels requests
func postChannel(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		c := new(channel.Channel)
		{
			c.ChannelUsername = r.FormValue("channelUsername")
			if c.ChannelUsername != "" {
				c.Name = r.FormValue("name")
				c.Description = r.FormValue("description")
			} else {
				err := json.NewDecoder(r.Body).Decode(&c)
				if err != nil {
					response.Data = jSendFailData{
						ErrorReason: "Bad Request",
						ErrorMessage: `use format
				{"channelUsername":"channelUsername",
		        "name":"name",
				"description":"description"}`,
					}

					statusCode = http.StatusBadRequest
				}
			}
		}
		sanitizeChannel(c, s)
		if response.Data == nil {
			if c.ChannelUsername == "" {
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: "channelUsername is required",
				}
			}
			if c.Name == "" {
				response.Data = jSendFailData{
					ErrorReason:  "name",
					ErrorMessage: "name is required",
				}

			} else {
				if len(c.ChannelUsername) > 24 || len(c.ChannelUsername) < 5 {
					response.Data = jSendFailData{
						ErrorReason:  "channelUsername",
						ErrorMessage: "channelUsername length shouldn't be shorter that 5 and longer than 22 chars",
					}
				}
			}

			if response.Data == nil {
				s.Logger.Printf("trying to add channel %s %s %s ", c.ChannelUsername, c.Name, c.Description)
				if &c != nil {
					err := s.ChannelService.AddChannel(c)
					switch err {
					case nil:
						response.Status = "success"
						s.Logger.Printf("success adding channel %s %s %s", c.ChannelUsername, c.Name, c.Description)
					case channel.ErrInvalidChannelData:
						s.Logger.Printf("creating of channel failed because: %s", err.Error())
						response.Data = jSendFailData{
							ErrorReason:  "needed values missing",
							ErrorMessage: "channel must have name & channelUsername to be created",
						}

						statusCode = http.StatusBadRequest

					case channel.ErrUserNameOccupied:
						s.Logger.Printf("adding of channel failed because: %s", err.Error())
						response.Data = jSendFailData{
							ErrorReason:  "channelUsername",
							ErrorMessage: "channelUsername is occupied",
						}

						statusCode = http.StatusConflict

					default:
						_ = s.ChannelService.DeleteChannel(c.ChannelUsername)
						s.Logger.Printf("adding of channel failed because: %s", err.Error())
						response.Data = jSendFailData{
							ErrorReason:  "Server Error",
							ErrorMessage: "server error when adding channel",
						}

						statusCode = http.StatusInternalServerError
					}
				}

			} else {
				// if required fields aren't present
				s.Logger.Printf(c.ChannelUsername)
				s.Logger.Printf("bad adding channel request")
				statusCode = http.StatusBadRequest
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// putChannel returns a handler for PUT /channels/{channelUsername} requests
func putChannel(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users updating of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized update channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		var c channel.Channel
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			response.Data = jSendFailData{
				ErrorReason:  "bad request",
				ErrorMessage: "bad request",
			}

			s.Logger.Printf("bad update channel request")
			statusCode = http.StatusBadRequest
		} else {
			if c.Name == "" && c.Description == "" && c.ChannelUsername == "" {
				response.Data = jSendFailData{
					ErrorReason:  "bad request",
					ErrorMessage: "bad request",
				}

				statusCode = http.StatusBadRequest
			} else {
				err := s.ChannelService.UpdateChannel(channelUsername, &c)
				switch err {
				case nil:
					s.Logger.Printf("success put channel %s", channelUsername)
					response.Status = "success"
				case channel.ErrUserNameOccupied:
					s.Logger.Printf("adding of channel failed because: %s", err.Error())
					response.Data = jSendFailData{
						ErrorReason:  "channelUsername",
						ErrorMessage: "channelUsername is occupied by channel",
					}

					statusCode = http.StatusConflict
				case channel.ErrInvalidChannelData:
					s.Logger.Printf("updating of channel failed because: %s", err.Error())
					response.Data = jSendFailData{
						ErrorReason:  "Needed values missing",
						ErrorMessage: "channel must have name & channelUsername to be created",
					}

					statusCode = http.StatusBadRequest
				default:
					s.Logger.Printf("update of channel failed because: %s", err.Error())
					response.Data = jSendFailData{
						ErrorReason:  "error",
						ErrorMessage: "server error when updating channel",
					}

					statusCode = http.StatusInternalServerError
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getChannels returns a handler for GET /channels requests
func getChannels(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		pattern := ""
		limit := 25
		offset := 0
		var sortBy = channel.SortCreationTime

		var sortOrder = channel.SortDescending
		{
			pattern = r.URL.Query().Get("pattern")
			if limitPageRaw := r.URL.Query().Get("limit"); limitPageRaw != "" {
				limit, err = strconv.Atoi(limitPageRaw)
				if err != nil || limit < 0 {
					s.Logger.Printf("bad get channels request, limit")
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
					s.Logger.Printf("bad request, offset")
					response.Data = jSendFailData{
						ErrorReason:  "offset",
						ErrorMessage: "bad request, offset can't be negative",
					}
					statusCode = http.StatusBadRequest
				}
			}
			sort := r.URL.Query().Get("sort")
			sortSplit := strings.Split(sort, "_")
			sortOrder = channel.SortAscending
			switch sortByQuery := sortSplit[0]; sortByQuery {
			case "channelUsername":
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
			channels, err := s.ChannelService.SearchChannels(pattern, sortBy, sortOrder, limit, offset)
			if err != nil {
				s.Logger.Printf("fetching of channels failed because: %s", err.Error())
				response.Data = jSendFailData{
					ErrorReason:  "error",
					ErrorMessage: "server error when getting channels",
				}
				statusCode = http.StatusInternalServerError
			} else {
				response.Status = "success"
				for _, c := range channels {
					c.AdminUsernames = nil
					c.ReleaseIDs = nil
					c.OwnerUsername = ""
					if c.PictureURL != "" {
						c.PictureURL = s.HostAddress + s.ImageServingRoute + url.PathEscape(c.PictureURL)
					}
				}
				response.Data = channels
				s.Logger.Printf("success fetching channels")
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteChannel returns a handler for DELETE /channels/{channelUsername} requests
func deleteChannel(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		var err error
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users deleting of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if !one {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized delete channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		s.Logger.Printf("trying to delete channel %s", channelUsername)
		err = s.ChannelService.DeleteChannel(channelUsername)
		if err != nil {
			s.Logger.Printf("deletion of channel failed because: %s", err.Error())
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channelUsername %s not found", channelUsername),
			}

			statusCode = http.StatusNotFound
		} else {
			response.Status = "success"
			s.Logger.Printf("success deleting channel %s", channelUsername)
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getAdmins returns a handler for GET /channels/{channelUsername}/admins requests
func getAdmins(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users getting admins of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized get admins of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		c, err := s.ChannelService.GetChannel(channelUsername)
		switch err {
		case nil:

			response.Status = "success"
			adminUsernames := c.AdminUsernames
			admins := make([]interface{}, 0)

			for _, uID := range adminUsernames {
				if temp, err := s.UserService.GetUser(uID); err == nil {
					admins = append(admins, temp)
				} else {
					admins = append(admins, uID)
				}
			}
			response.Data = admins

			s.Logger.Printf("success fetching admins of channel %s", channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf("fetch attempt of non existing channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("user of channel %s not found", channelUsername),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("fetching of admins of channel failed because: %s", err.Error())
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when fetching admins of channel",
			}
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// putAdmin returns a handler for PUT /channels/{channelUsername}/admins/{adminUsername}
func putAdmin(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]

		{
			// this block blocks users updating of admins of  channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			s.Logger.Printf(r.Header.Get("authorized_username"))
			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {

					one = true

				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized update of channel admins attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		adminUsername := vars["adminUsername"]
		err := s.ChannelService.AddAdmin(channelUsername, adminUsername)
		switch err {
		case nil:
			response.Status = "success"
			s.Logger.Printf("success adding admin  %s in to channel %s", adminUsername, channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf(fmt.Sprintf("Adding of Admin failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: "channel doesn't exits",
			}
			statusCode = http.StatusNotFound
		case channel.ErrAdminNotFound:
			s.Logger.Printf(fmt.Sprintf("Adding of Admin failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "adminUsername",
				ErrorMessage: "Admin user doesn't exits",
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf(fmt.Sprintf("Adding of Admin failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when Adding of Admin",
			}
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteAdmin returns a handler for DELETE /channels/{channelUsername}/admins/{adminUsername}
func deleteAdmin(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{

			//// this block blocks users deleting of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.OwnerUsername

			if adminUsername != r.Header.Get("authorized_username") {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized delete admins of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

			}
		}
		adminUsername := vars["adminUsername"]
		err := s.ChannelService.DeleteAdmin(channelUsername, adminUsername)
		switch err {
		case nil:
			response.Status = "success"
			s.Logger.Printf("success deleting admin  %s in to channel %s", adminUsername, channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf(fmt.Sprintf("Deleting of Admin failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: "channel doesn't exits",
			}
			statusCode = http.StatusNotFound
		case channel.ErrAdminNotFound:
			s.Logger.Printf(fmt.Sprintf("Deleting of Admin failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "adminUsername",
				ErrorMessage: "Admin user doesn't exits",
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf(fmt.Sprintf("Deleting of Admin failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when Deleting of Admin",
			}
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getOwner returns a handler for GET /channels/{channelUsername}/owners
func getOwner(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users getting owners of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized get owner of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		c, err := s.ChannelService.GetChannel(channelUsername)
		switch err {
		case nil:
			response.Status = "success"
			owner := c.OwnerUsername
			ownerUser, err := s.UserService.GetUser(owner)
			if err != nil {
				response.Data = owner
			} else {
				response.Data = ownerUser
			}

			s.Logger.Printf("success fetching owner of channel %s", channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf("fetch attempt of non existing channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("fetching of owner of channel failed because: %s", err.Error())
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when fetching owner of channel",
			}
			statusCode = http.StatusInternalServerError
		}

		writeResponseToWriter(response, w, statusCode)
	}
}

// putOwner returns a handler for PUT /channels/{channelUsername}/owners/{ownerUsername}
func putOwner(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users updating owner of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			ownerUsername := c.OwnerUsername
			if ownerUsername == r.Header.Get("authorized_username") {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized update owner of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		}
		ownerUsername := vars["ownerUsername"]
		err := s.ChannelService.ChangeOwner(channelUsername, ownerUsername)
		switch err {
		case nil:
			response.Status = "success"
			s.Logger.Printf("success updating owner of  %s  channel to %s", channelUsername, ownerUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf(fmt.Sprintf("Update of owner failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: "channel doesn't exits",
			}
			statusCode = http.StatusNotFound
		case channel.ErrOwnerNotFound:
			s.Logger.Printf(fmt.Sprintf("Update of owner failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "adminUsername",
				ErrorMessage: "Owner user doesn't exits",
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf(fmt.Sprintf("Update of owner failed because: %s", err.Error()))
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when Update of owner",
			}
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getCatalog returns a handler for GET /channels/{channelUsername}/catalog
func getCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users getting Catalog of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized get catalog of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		c, err := s.ChannelService.GetChannel(channelUsername)
		switch err {
		case nil:
			response.Status = "success"
			catalog := c.ReleaseIDs
			releases := make([]interface{}, 0)

			for _, uID := range catalog {
				if temp, err := s.ReleaseService.GetRelease(uID); err == nil {
					releases = append(releases, temp)
				} else {
					releases = append(releases, uID)
				}
			}
			response.Data = releases
			s.Logger.Printf("success fetching catalog of channel %s", channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf("fetch attempt of catalog from non existent channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("fetching of catalog of channel failed because: %s", err.Error())
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when fetching catalog of channel",
			}
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getOfficialCatalog returns a handler for GET /channels/{channelUsername}/official
func getOfficialCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		c, err := s.ChannelService.GetChannel(channelUsername)
		switch err {
		case nil:
			response.Status = "success"
			officialCatalog := c.OfficialReleaseIDs
			releases := make([]interface{}, 0)

			for _, uID := range officialCatalog {
				if temp, err := s.ReleaseService.GetRelease(uID); err == nil {
					releases = append(releases, temp)
				} else {
					releases = append(releases, uID)
				}
			}
			response.Data = releases
			s.Logger.Printf("success fetching official catalog of channel %s", channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf("fetch attempt of official catalog from non existent channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("fetching of official catalog of channel failed because: %s", err.Error())
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when fetching official catalog of channel",
			}
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteReleaseFromCatalog returns a handler for DELETE /channels/{channelUsername}/catalogs/{catalogID}
func deleteReleaseFromCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users deleting release from catalog of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized delete release of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		ReleaseID, err := strconv.Atoi(vars["catalogID"])
		if err != nil {
			response.Data = jSendFailData{
				ErrorReason:  "bad request",
				ErrorMessage: "bad request, ReleaseID must be an integer",
			}
			statusCode = http.StatusBadRequest

		} else {
			errC := s.ChannelService.DeleteReleaseFromCatalog(channelUsername, ReleaseID)
			switch errC {
			case nil:
				response.Status = "success"
				s.Logger.Printf("success deleting release  %d from channel %s's Catalog", ReleaseID, channelUsername)
			case channel.ErrChannelNotFound:
				s.Logger.Printf(fmt.Sprintf("Deleting of release failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: "channel doesn't exits",
				}
				statusCode = http.StatusNotFound
			case channel.ErrReleaseNotFound:
				s.Logger.Printf(fmt.Sprintf("Deleting of Admin failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "releaseID",
					ErrorMessage: "Release doesn't exits",
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf(fmt.Sprintf("Deleting of Release failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "error",
					ErrorMessage: "server error when Deleting of Release",
				}
				statusCode = http.StatusInternalServerError
			}

		}

		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteReleaseFromOfficialCatalog returns a handler for DELETE /channels/{channelUsername}/official/{catalogID}
func deleteReleaseFromOfficialCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users deleting release from catalog of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized delete release of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		ReleaseID, err := strconv.Atoi(vars["catalogID"])
		if err != nil {
			response.Data = jSendFailData{
				ErrorReason:  "bad request",
				ErrorMessage: "bad request, ReleaseID must be an integer",
			}
			statusCode = http.StatusBadRequest

		} else {
			errC := s.ChannelService.DeleteReleaseFromOfficialCatalog(channelUsername, ReleaseID)
			switch errC {
			case nil:
				response.Status = "success"
				s.Logger.Printf("success deleting release  %d from channel %s's Catalog", ReleaseID, channelUsername)
			case channel.ErrChannelNotFound:
				s.Logger.Printf(fmt.Sprintf("Deleting of release failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: "channel doesn't exits",
				}
				statusCode = http.StatusNotFound
			case channel.ErrReleaseNotFound:
				s.Logger.Printf(fmt.Sprintf("Deleting of release from official catalog failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "releaseID",
					ErrorMessage: "Release doesn't exits",
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf(fmt.Sprintf("Deleting of Release from official catalog failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "error",
					ErrorMessage: "server error when Deleting of Release",
				}
				statusCode = http.StatusInternalServerError
			}

		}

		writeResponseToWriter(response, w, statusCode)
	}
}

// getReleaseFromCatalog returns a handler for GET /channels/{channelUsername}/catalogs/{catalogID}
func getReleaseFromCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users get release of catalog of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized get release of catalog of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}

		ReleaseID, errC := strconv.Atoi(vars["catalogID"])
		if errC != nil {
			response.Data = jSendFailData{
				ErrorReason:  "bad request",
				ErrorMessage: "bad request, ReleaseID must be an integer",
			}
			statusCode = http.StatusBadRequest

		} else {
			c, err := s.ChannelService.GetChannel(channelUsername)

			switch err {
			case nil:
				for i := 0; i < len(c.ReleaseIDs); i++ {
					if c.ReleaseIDs[i] == ReleaseID {
						response.Status = "success"
						catalog := ReleaseID
						releases := make([]interface{}, 0)
						temp, err := s.ReleaseService.GetRelease(catalog)
						if err == nil {
							releases = append(releases, temp)
						} else {
							releases = append(releases, catalog)
							s.Logger.Printf(err.Error())
						}

						response.Data = releases
						s.Logger.Printf("success fetching release of  catalog of channel %s", channelUsername)
					} else {
						response.Data = jSendFailData{
							ErrorReason:  "releaseID",
							ErrorMessage: "release doesn't exits",
						}
						statusCode = http.StatusNotFound

					}
				}
			case channel.ErrChannelNotFound:
				s.Logger.Printf("fetch attempt of catalog from non existent channel %s", channelUsername)
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: fmt.Sprintf("channel of is %s not found", channelUsername),
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf("fetching of catalog of channel failed because: %s", err.Error())
				response.Data = jSendFailData{
					ErrorReason:  "error",
					ErrorMessage: "server error when fetching catalog of channel",
				}
				statusCode = http.StatusInternalServerError
			}
		}

		writeResponseToWriter(response, w, statusCode)
	}
}

//TODO
//// putReleaseInCatalog returns a handler for PUT /channels/{channelUsername}/catalogs/{catalogID}
func putReleaseInCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		k := false
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		idRaw := vars["catalogID"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			s.Logger.Printf("put attempt of non invalid release id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "releaseID",
				ErrorMessage: fmt.Sprintf("invalid releaseID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		rel := new(release.Release)
		var tmpFile *os.File
		{ // this block parses the JSON part of the request
			err := json.Unmarshal([]byte(r.PostFormValue("JSON")), rel)
			if err != nil {
				err = json.NewDecoder(r.Body).Decode(rel)
				if err != nil {
					// TODO send back format
					response.Data = jSendFailData{
						ErrorReason:  "message",
						ErrorMessage: "use multipart for for posting Image Releases. A part named 'JSON' for Release data \r\nand a file called 'image' if release is of image type JPG/PNG.",
					}
					statusCode = http.StatusBadRequest
				}
			} else {
				{ // this block extracts the image file if necessary
					switch rel.Type {
					case release.Text:
					case release.Image:
						fallthrough
					default:
						var fileName string
						var err error
						tmpFile, fileName, err = saveImageFromRequest(r, "image")
						switch err {
						case nil:
							s.Logger.Printf("image found on put request")
							defer os.Remove(tmpFile.Name())
							defer tmpFile.Close()
							s.Logger.Printf(fmt.Sprintf("temp file saved: %s", tmpFile.Name()))
							rel.Content = generateFileNameForStorage(fileName, "release")
							rel.Type = release.Image
						case errUnacceptedType:
							response.Data = jSendFailData{
								ErrorMessage: "image",
								ErrorReason:  "only types image/jpeg & image/png are accepted",
							}
							statusCode = http.StatusBadRequest
						case errReadingFromImage:
							s.Logger.Printf("image not found on put request")
							if rel.Type == release.Image {
								response.Data = jSendFailData{
									ErrorReason:  "image",
									ErrorMessage: "unable to read image file\nuse multipart-form for for posting Image Releases. A part named 'JSON' for Release data \nand a file called 'image' of image type JPG/PNG.",
								}
								statusCode = http.StatusBadRequest
							}
						default:
							response.Status = "error"
							response.Message = "server error when adding release"
							statusCode = http.StatusInternalServerError
						}
					}
				}
			}
		}
		if response.Data == nil {
			// if JSON parsing doesn't fail
			if rel.Content == "" && rel.Title == "" && rel.GenreDefining == "" && rel.Description == "" && len(rel.Genres) == 0 && len(rel.Authors) == 0 && rel.OwnerChannel == "" {
				response.Data = jSendFailData{
					ErrorReason:  "request",
					ErrorMessage: "bad request, data sent doesn't contain update able data",
				}
				statusCode = http.StatusBadRequest
			}

			{
				// this block blocks users get release of catalog of channel if is not the admin of the channel herself accessing the route
				c, _ := s.ChannelService.GetChannel(channelUsername)
				adminUsername := c.AdminUsernames

				one := false
				for i := 0; i < len(adminUsername); i++ {
					if adminUsername[i] == r.Header.Get("authorized_username") {
						one = true
					}
				}
				if one == false {
					if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
						s.Logger.Printf("unauthorized get release of catalog of channel attempt")
						w.WriteHeader(http.StatusUnauthorized)
						return

					}
				}
			}
			c, err := s.ChannelService.GetChannel(channelUsername)
			switch err {
			case nil:

				k = true
				s.Logger.Printf("success fetching official catalog of channel %s", channelUsername)
			case channel.ErrChannelNotFound:
				s.Logger.Printf("fetch attempt of official catalog from non existent channel %s", channelUsername)
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf("fetching of official catalog of channel failed because: %s", err.Error())
				response.Data = jSendFailData{
					ErrorReason:  "error",
					ErrorMessage: "server error when fetching official catalog of channel",
				}
				statusCode = http.StatusInternalServerError
			}
			if k {
				if response.Data == nil {
					if response.Data == nil {
						rel.ID = id
						for i := 0; i < len(c.ReleaseIDs); i++ {
							if c.ReleaseIDs[i] != rel.ID {
								s.Logger.Printf("update attempt of release %d not in catalog of channel %s", id, c.ChannelUsername)
								response.Data = jSendFailData{
									ErrorReason:  "releaseID",
									ErrorMessage: fmt.Sprintf("release of id %d not found in catalog", id),
								}
							}
						}

						rel, err = s.ReleaseService.UpdateRelease(rel)
						switch err {
						case nil:
							if rel.Type == release.Image {
								err := saveTempFilePermanentlyToPath(tmpFile, s.ImageStoragePath+rel.Content)
								if err != nil {
									s.Logger.Printf("updating of release failed because: %v", err)
									response.Status = "error"
									response.Message = "server error when updating release"
									statusCode = http.StatusInternalServerError
									_ = s.ReleaseService.DeleteRelease(rel.ID)
								}
							}
							if response.Message == "" {
								s.Logger.Printf("success updating release %d", id)
								response.Status = "success"
								rel.Content = s.HostAddress + s.ImageServingRoute + url.PathEscape(rel.Content)
								response.Data = *rel
								// TODO delete old image if image updated
							}
						case release.ErrAttemptToChangeReleaseType:
							s.Logger.Printf("update attempt of release type for release %d", id)
							response.Data = jSendFailData{
								ErrorReason:  "type",
								ErrorMessage: "release type cannot be changed",
							}
							statusCode = http.StatusNotFound
						case release.ErrReleaseNotFound:
							s.Logger.Printf("update attempt of non existing release %d", id)
							response.Data = jSendFailData{
								ErrorReason:  "releaseID",
								ErrorMessage: fmt.Sprintf("release of id %d not found", id),
							}
							statusCode = http.StatusNotFound
						case release.ErrSomeReleaseDataNotPersisted:
							fallthrough
						default:
							s.Logger.Printf("update of release failed because: %v", err)
							response.Status = "error"
							response.Message = "server error when adding release"
							statusCode = http.StatusInternalServerError
						}
					}
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

//TODO
//// postReleaseInCatalog returns a handler for POST /channels/{channelUsername}/catalogs/{catalogID}
func postReleaseInCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		k := false
		var response jSendResponse
		response.Status = "fail"
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		statusCode := http.StatusCreated

		newRelease := new(release.Release)
		var tmpFile *os.File
		{ // this block parses the JSON part of the request
			err := json.Unmarshal([]byte(r.PostFormValue("JSON")), newRelease)
			if err != nil {
				err = json.NewDecoder(r.Body).Decode(newRelease)
				if err != nil {
					response.Data = jSendFailData{
						ErrorReason:  "request format",
						ErrorMessage: "use multipart for for posting Image Releases. A part named 'JSON' with format\r\n{\n  \"ownerChannel\": \"ownerChannel\",\n  \"type\": \"image or text\",\n  \"content\": \"content if type is text\",\n  \"metadata\": {\n    \"title\": \"title\",\n    \"releaseDate\": \"unix timestamp\",\n    \"genreDefining\": \"genreDefining\",\n    \"description\": \"description\",\n    \"Other\": { \"authors\": [], \"genres\": [] }\n  }\n}\nfor Release data and a file called 'image' if release is of image type. We accept JPG/PNG formats.",
					}
					statusCode = http.StatusBadRequest
				}
			}
		}
		if response.Data == nil {
			{
				// this block checks for required fields
				if newRelease.OwnerChannel == "" {
					response.Data = jSendFailData{
						ErrorReason:  "OwnerChannel",
						ErrorMessage: "OwnerChannel is required",
					}
					// if required fields aren't present
					s.Logger.Printf("bad add release request")
					statusCode = http.StatusBadRequest
				} else {
					{
						// this block blocks users get release of catalog of channel if is not the admin of the channel herself accessing the route
						c, _ := s.ChannelService.GetChannel(channelUsername)
						adminUsername := c.AdminUsernames

						one := false
						for i := 0; i < len(adminUsername); i++ {
							if adminUsername[i] == r.Header.Get("authorized_username") {
								one = true
							}
						}
						if one == false {
							if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
								s.Logger.Printf("unauthorized get release of catalog of channel attempt")
								w.WriteHeader(http.StatusUnauthorized)
								return

							}
						}
					}
					_, err := s.ChannelService.GetChannel(channelUsername)
					switch err {
					case nil:

						k = true
						s.Logger.Printf("success fetching official catalog of channel %s", channelUsername)
					case channel.ErrChannelNotFound:
						s.Logger.Printf("fetch attempt of official catalog from non existent channel %s", channelUsername)
						response.Data = jSendFailData{
							ErrorReason:  "channelUsername",
							ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
						}
						statusCode = http.StatusNotFound
					default:
						s.Logger.Printf("fetching of official catalog of channel failed because: %s", err.Error())
						response.Data = jSendFailData{
							ErrorReason:  "error",
							ErrorMessage: "server error when fetching official catalog of channel",
						}
						statusCode = http.StatusInternalServerError
					}
				}
			}
			if k {
				if response.Data == nil {
					{ // this block extracts the image file if necessary
						switch newRelease.Type {
						case release.Image:
							var fileName string
							var err error
							tmpFile, fileName, err = saveImageFromRequest(r, "image")
							switch err {
							case nil:
								defer tmpFile.Close()
								defer os.Remove(tmpFile.Name())
								s.Logger.Printf(fmt.Sprintf("temp file saved: %s", tmpFile.Name()))
								newRelease.Content = generateFileNameForStorage(fileName, "release")
							case errUnacceptedType:
								response.Data = jSendFailData{
									ErrorMessage: "image",
									ErrorReason:  "only types image/jpeg & image/png are accepted",
								}
								statusCode = http.StatusBadRequest
							case errReadingFromImage:
								response.Data = jSendFailData{
									ErrorReason:  "image",
									ErrorMessage: "unable to read image file\nuse multipart-form for for posting Image Releases. A part named 'JSON' for Release data \nand a file called 'image' of image type JPG/PNG.",
								}
								statusCode = http.StatusBadRequest
							default:
								response.Status = "error"
								response.Message = "server error when adding release"
								statusCode = http.StatusInternalServerError
							}
						case release.Text:
						default:
							statusCode = http.StatusBadRequest
							response.Data = jSendFailData{
								ErrorMessage: "type can only be 'text' or 'image'",
								ErrorReason:  "type",
							}
						}
					}
					if response.Data == nil {
						s.Logger.Printf("trying to add release")

						newRelease, err := s.ReleaseService.AddRelease(newRelease)
						switch err {
						case nil:
							if newRelease.Type == release.Image {
								err := saveTempFilePermanentlyToPath(tmpFile, s.ImageStoragePath+newRelease.Content)
								if err != nil {
									s.Logger.Printf("adding of release failed because: %v", err)
									response.Status = "error"
									response.Message = "server error when adding release"
									statusCode = http.StatusInternalServerError
									_ = s.ReleaseService.DeleteRelease(newRelease.ID)
								}
							}
							if response.Message == "" {

								response.Status = "success"
								newRelease.OwnerChannel = channelUsername
								if newRelease.Type == release.Image {
									newRelease.Content = s.HostAddress + s.ImageServingRoute + url.PathEscape(newRelease.Content)
								}

								response.Data = *newRelease
								s.Logger.Printf("success adding release %d to channel %s", newRelease.ID, newRelease.OwnerChannel)
							}
						case release.ErrSomeReleaseDataNotPersisted:
							fallthrough
						default:
							_ = s.ReleaseService.DeleteRelease(newRelease.ID)
							s.Logger.Printf("adding of release failed because: %v", err)
							response.Status = "error"
							response.Message = "server error when adding release"
							statusCode = http.StatusInternalServerError
						}
					}
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}

}

//TODO
//// postReleaseInOfficialCatalog returns a handler for POST /channels/{channelUsername}/official
func putReleaseInOfficialCatalog(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		k := false
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users updating content of catalog of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized update catalog of  channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		var releasePost struct {
			PostID    int `json:"postID"`
			ReleaseID int `json:"releaseID"`
		}

		errC := json.NewDecoder(r.Body).Decode(&releasePost)
		if errC != nil {
			response.Data = jSendFailData{
				ErrorReason: "request format",
				ErrorMessage: `bad request, use format
			{"postID":"postID",
			"releaseID":"releaseID"
			}`,
			}

			s.Logger.Printf("bad get post info request")
			statusCode = http.StatusBadRequest
		} else {
			if releasePost.PostID == 0 && releasePost.ReleaseID == 0 {
				response.Data = jSendFailData{
					ErrorReason:  "bad request",
					ErrorMessage: "bad request",
				}

				statusCode = http.StatusBadRequest
			} else {
				pos, err := s.PostService.GetPost(releasePost.PostID)
				switch err {
				case nil:
					for i := 0; i < len(pos.ContentsID); i++ {
						if pos.ContentsID[i] == releasePost.ReleaseID {
							c, errC := s.ChannelService.GetChannel(channelUsername)
							switch errC {
							case nil:
								for i := 0; i < len(c.ReleaseIDs); i++ {
									if c.ReleaseIDs[i] == releasePost.ReleaseID {
										k = true

									}

								}
								if k {
									err := s.ChannelService.AddReleaseToOfficialCatalog(channelUsername, releasePost.ReleaseID)
									switch err {
									case nil:
										s.Logger.Printf(fmt.Sprintf("success adding Release %d to channels %s's Catalog", releasePost.ReleaseID, channelUsername))
										response.Status = "success"
									case channel.ErrChannelNotFound:
										s.Logger.Printf(fmt.Sprintf("Adding Release to Offical Catalog failed because: %s", err.Error()))
										response.Data = jSendFailData{
											ErrorReason:  "channelUsername",
											ErrorMessage: "channel doesn't exits",
										}
										statusCode = http.StatusNotFound
									case channel.ErrReleaseNotFound:
										s.Logger.Printf(fmt.Sprintf("Adding Release to Offical Catalog failed because: %s", err.Error()))
										response.Data = jSendFailData{
											ErrorReason:  "releaseID",
											ErrorMessage: "release doesn't exits",
										}
										statusCode = http.StatusNotFound
									default:
										s.Logger.Printf(fmt.Sprintf("Adding Release to Offical Catalog failed because: %s", err.Error()))
										response.Data = jSendFailData{
											ErrorReason:  "error",
											ErrorMessage: "server error when Adding Release to Official Catalog ",
										}
										statusCode = http.StatusInternalServerError
									}
								} else {
									s.Logger.Printf(fmt.Sprintf("Adding Release to Offical Catalog failed because: %s", err.Error()))
									response.Data = jSendFailData{
										ErrorReason:  "releaseID",
										ErrorMessage: "release doesn't exits in post",
									}
									statusCode = http.StatusNotFound
								}
							}
						}
					}
				case post.ErrPostNotFound:
					response.Data = jSendFailData{
						ErrorReason:  "postID",
						ErrorMessage: fmt.Sprintf("post of postID %d not found", releasePost.ReleaseID),
					}
					statusCode = http.StatusNotFound
				default:
					s.Logger.Printf("fetching of post failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when fetching post"
					statusCode = http.StatusInternalServerError

				}
			}
		}

		writeResponseToWriter(response, w, statusCode)
	}
}

// getChannelPost returns a handler for GET /channels/{channelUsername}/Posts/{postIDs}
func getChannelPost(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		c, err := s.ChannelService.GetChannel(channelUsername)
		postID, errC := strconv.Atoi(vars["postID"])
		if errC != nil {
			response.Data = jSendFailData{
				ErrorReason:  "bad request",
				ErrorMessage: "bad request, postID must be an integer",
			}
			statusCode = http.StatusBadRequest

		} else {
			switch err {
			case nil:
				for i := 0; i < len(c.PostIDs); i++ {
					if c.PostIDs[i] == postID {
						response.Status = "success"
						postid := postID
						posts := make([]interface{}, 0)
						if temp, err := s.ReleaseService.GetRelease(postid); err == nil {
							posts = append(posts, temp)
						} else {
							posts = append(posts, postid)
						}

						response.Data = posts
						s.Logger.Printf("success fetching post of channel %s", channelUsername)
						break
					} else {
						response.Data = jSendFailData{
							ErrorReason:  "postID",
							ErrorMessage: "post doesn't exits",
						}
						statusCode = http.StatusNotFound

					}
				}
			case channel.ErrChannelNotFound:
				s.Logger.Printf("fetch attempt of post from non existent channel %s", channelUsername)
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf("fetching of post of channel failed because: %s", err.Error())
				response.Data = jSendFailData{
					ErrorReason:  "error",
					ErrorMessage: "server error when fetching post of channel",
				}
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getChannelPosts returns a handler for GET /channels/{channelUsername}/Posts
func getChannelPosts(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]

		c, err := s.ChannelService.GetChannel(channelUsername)
		switch err {
		case nil:
			response.Status = "success"
			postid := c.PostIDs
			posts := make([]interface{}, 0)

			for _, pID := range postid {
				if temp, err := s.ReleaseService.GetRelease(pID); err == nil {
					posts = append(posts, temp)
				} else {
					posts = append(posts, pID)
				}
			}
			response.Data = posts
			s.Logger.Printf("success fetching posts of channel %s", channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf("fetch attempt of posts from non existent channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("fetching of posts of channel failed because: %s", err.Error())
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when fetching catalog of channel",
			}
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getStickiedPosts returns a handler for GET /channels/{channelUsername}/stickiedPosts
func getStickiedPosts(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		c, err := s.ChannelService.GetChannel(channelUsername)

		switch err {
		case nil:

			response.Status = "success"
			postid := c.StickiedPostIDs
			posts := make([]interface{}, 0)

			for _, pID := range postid {
				if temp, err := s.ReleaseService.GetRelease(pID); err == nil {
					posts = append(posts, temp)
				} else {
					posts = append(posts, pID)
				}
			}
			response.Data = posts
			s.Logger.Printf("success fetching post of channel %s", channelUsername)

		case channel.ErrChannelNotFound:
			s.Logger.Printf("fetch attempt of post from non existent channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channel of %s not found", channelUsername),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("fetching of post of channel failed because: %s", err.Error())
			response.Data = jSendFailData{
				ErrorReason:  "error",
				ErrorMessage: "server error when fetching post of channel",
			}
			statusCode = http.StatusInternalServerError
		}

		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteStickiedPost returns a handler for DELETE /channels/{channelUsername}/stickiedPosts{stickiedPostID}
func deleteStickiedPost(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users deleting stickied post of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized delete stickied post of channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		stickiedPostID, err := strconv.Atoi(vars["stickiedPostID"])
		if err != nil {
			response.Data = jSendFailData{
				ErrorReason:  "Bad Request",
				ErrorMessage: "bad request, StickiedPostID must be an integer",
			}
			statusCode = http.StatusBadRequest

		} else {
			errC := s.ChannelService.DeleteStickiedPost(channelUsername, stickiedPostID)
			switch errC {
			case nil:
				response.Status = "success"
				s.Logger.Printf("success deleting stickied Post  %d from channel %s's Catalog", stickiedPostID, channelUsername)
			case channel.ErrChannelNotFound:
				s.Logger.Printf(fmt.Sprintf("Deleting of stickied Post failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: "channel doesn't exits",
				}
				statusCode = http.StatusNotFound
			case channel.ErrStickiedPostNotFound:
				s.Logger.Printf(fmt.Sprintf("Deleting of stickied post failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "stickiedPostID",
					ErrorMessage: "Stickied Post doesn't exits",
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf(fmt.Sprintf("Deleting of Stickied Post failed because: %s", errC.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "Error",
					ErrorMessage: "server error when Deleting of Stickied Post",
				}
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// stickyPost returns a handler for PUT /channels/{channelUsername}/Posts/{postID}
func stickyPost(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users sticking a post of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized sticky a post in channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}

		stickyPost, err := strconv.Atoi(vars["postID"])
		if err != nil {
			response.Data = jSendFailData{
				ErrorReason:  "bad request",
				ErrorMessage: "bad request, PostID must be an integer",
			}
			statusCode = http.StatusBadRequest

		} else {
			err := s.ChannelService.StickyPost(channelUsername, stickyPost)
			switch err {
			case nil:
				response.Status = "success"
				s.Logger.Printf("success of stickying post  %d to channel  %s", stickyPost, channelUsername)
			case channel.ErrChannelNotFound:
				s.Logger.Printf(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: "channel doesn't exits",
				}
				statusCode = http.StatusNotFound
			case channel.ErrPostNotFound:
				s.Logger.Printf(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "postID",
					ErrorMessage: "post doesn't exits",
				}

				statusCode = http.StatusNotFound
			case channel.ErrStickiedPostFull:
				s.Logger.Printf(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "Stickied postID",
					ErrorMessage: "stickied post full",
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf(fmt.Sprintf("Stickying of post failed because: %s", err.Error()))
				response.Data = jSendFailData{
					ErrorReason:  "Server Error",
					ErrorMessage: "server error when stickying a post",
				}
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getChannelPicture returns a handler for GET /channels/{channelUsername}/picture requests
func getChannelPicture(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"

		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]

		c, err := s.ChannelService.GetChannel(channelUsername)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = s.HostAddress + s.ImageServingRoute + url.PathEscape(c.PictureURL)
			s.Logger.Printf("success fetching channel %s picture URL", channelUsername)
		case channel.ErrChannelNotFound:
			s.Logger.Printf("fetch picture URL attempt of non existing channel %s", channelUsername)
			response.Data = jSendFailData{
				ErrorReason:  "channelUsername",
				ErrorMessage: fmt.Sprintf("channel of channelUsername %s not found", channelUsername),
			}
			statusCode = http.StatusNotFound
		default:
			s.Logger.Printf("fetching of channel picture URL failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when fetching channel picture URL"
			statusCode = http.StatusInternalServerError
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// putChannelPicture returns a handler for PUT /channels/{channelUsername}/picture requests
func putChannelPicture(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users sticking a post of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized sticky a post in channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		var tmpFile *os.File
		var fileName string
		{ // this block extracts the image
			tmpFile, fileName, err = saveImageFromRequest(r, "image")
			switch err {
			case nil:
				s.Logger.Printf("image found on put channel picture request")
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()
				s.Logger.Printf("temp file saved: %s", tmpFile.Name())
				fileName = generateFileNameForStorage(fileName, "user")
			case errUnacceptedType:
				response.Data = jSendFailData{
					ErrorMessage: "image",
					ErrorReason:  "only types image/jpeg & image/png are accepted",
				}
				statusCode = http.StatusBadRequest
			case errReadingFromImage:
				s.Logger.Printf("image not found on put request")
				response.Data = jSendFailData{
					ErrorReason:  "image",
					ErrorMessage: "unable to read image file\nuse multipart-form for for posting channel pictures. A form that contains the file under the key 'image', of image type JPG/PNG.",
				}
				statusCode = http.StatusBadRequest
			default:
				response.Status = "error"
				response.Message = "server error when adding channel picture"
				statusCode = http.StatusInternalServerError
			}
		}
		// if queries are clean
		if response.Data == nil {
			err := s.ChannelService.AddPicture(channelUsername, fileName)
			s.Logger.Printf(channelUsername)
			switch err {
			case nil:
				err := saveTempFilePermanentlyToPath(tmpFile, s.ImageStoragePath+fileName)
				if err != nil {
					s.Logger.Printf("adding of picture failed  case nil because: %v", err)
					response.Status = "error"
					response.Message = "server error when setting channel picture"
					statusCode = http.StatusInternalServerError
					_ = s.ChannelService.RemovePicture(channelUsername)
				} else {
					s.Logger.Printf("success adding picture %s to channel %s", fileName, channelUsername)
					response.Status = "success"
					response.Data = s.HostAddress + s.ImageServingRoute + url.PathEscape(fileName)
				}
			case channel.ErrChannelNotFound:
				s.Logger.Printf("adding of channel picture failed because: %v", err)
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: fmt.Sprintf("user of channelUsername %s not found", channelUsername),
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf("Setting of picture of channel failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when setting channel picture"
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteChannelPicture returns a handler for DELETE /channels/{channelUsername}/picture requests
func deleteChannelPicture(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"
		vars := mux.Vars(r)
		channelUsername := vars["channelUsername"]
		{
			// this block blocks users sticking a post of channel if is not the admin of the channel herself accessing the route
			c, _ := s.ChannelService.GetChannel(channelUsername)
			adminUsername := c.AdminUsernames

			one := false
			for i := 0; i < len(adminUsername); i++ {
				if adminUsername[i] == r.Header.Get("authorized_username") {
					one = true
				}
			}
			if one == false {
				if _, err := s.ChannelService.GetChannel(channelUsername); err == nil {
					s.Logger.Printf("unauthorized sticky a post in channel attempt")
					w.WriteHeader(http.StatusUnauthorized)
					return

				}
			}
		}
		// if queries are clean
		if response.Data == nil {
			err = s.ChannelService.RemovePicture(channelUsername)
			switch err {
			case nil:
				// TODO delete picture from fs
				s.Logger.Printf("success removing piture from channel %s", channelUsername)
				response.Status = "success"
			case channel.ErrChannelNotFound:
				s.Logger.Printf("deletion of channel picture failed because: %v", err)
				response.Data = jSendFailData{
					ErrorReason:  "channelUsername",
					ErrorMessage: fmt.Sprintf("channel of channelUsername %s not found", channelUsername),
				}
				statusCode = http.StatusNotFound
			default:
				s.Logger.Printf("deletion of user pictre failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when removing user picture"
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
