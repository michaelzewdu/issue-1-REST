package rest

import (
	"encoding/json"
	"fmt"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// postReleases returns a handler for POST /releases requests
func postReleases(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
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
					d.Logger.Log("bad add release request")
					statusCode = http.StatusBadRequest
				} else {
					{
						// TODO check if given channel exists and auth
					}
				}
			}
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
							d.Logger.Log(fmt.Sprintf("temp file saved: %s", tmpFile.Name()))
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
					d.Logger.Log("trying to add release")

					newRelease, err := d.ReleaseService.AddRelease(newRelease)
					switch err {
					case nil:
						if newRelease.Type == release.Image {
							err := saveTempFilePermanentlyToPath(tmpFile, d.ImageStoragePath+newRelease.Content)
							if err != nil {
								d.Logger.Log("adding of release failed because: %v", err)
								response.Status = "error"
								response.Message = "server error when adding release"
								statusCode = http.StatusInternalServerError
								_ = d.ReleaseService.DeleteRelease(newRelease.ID)
							}
						}
						if response.Message == "" {
							{ // TODO add to channel catalog

							}
							response.Status = "success"
							newRelease.Content = d.HostAddress + d.ImageServingRoute + url.PathEscape(newRelease.Content)
							response.Data = *newRelease
							d.Logger.Log("success adding release %d to channel %s", newRelease.ID, newRelease.OwnerChannel)
						}
					case release.ErrSomeReleaseDataNotPersisted:
						fallthrough
					default:
						_ = d.ReleaseService.DeleteRelease(newRelease.ID)
						d.Logger.Log("adding of release failed because: %v", err)
						response.Status = "error"
						response.Message = "server error when adding release"
						statusCode = http.StatusInternalServerError
					}
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// getRelease returns a handler for GET /releases/{id} requests
func getRelease(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid release id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "releaseID",
				ErrorMessage: fmt.Sprintf("invalid releaseID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch release %d", id)
			rel, err := d.ReleaseService.GetRelease(id)
			switch err {
			case nil:
				response.Status = "success"
				if rel.Type == release.Image {
					rel.Content = d.HostAddress + d.ImageServingRoute + url.PathEscape(rel.Content)
				}
				// TODO secure route
				//{ // this block sanitizes the returned User if it'd not the user herself accessing the route
				//	if releaseID != r.Header.Get("authorized_username") {
				//		(*logger).Log(fmt.Sprintf("user %d fetched user %d", r.Header.Get("authorized_username"), rel.Username))
				//		rel.Email = ""
				//		rel.BookmarkedPosts = make(map[int]time.Time)
				//	}
				//}
				response.Data = *rel
				d.Logger.Log("success fetching release %d", id)
			case release.ErrReleaseNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "releaseID",
					ErrorMessage: fmt.Sprintf("release of releaseID %d not found", id),
				}
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of release failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching release"
				statusCode = http.StatusInternalServerError
			}
		}

		writeResponseToWriter(response, w, statusCode)
	}
}

// getReleases returns a handler for GET /releases?sort=new&limit=5&offset=0&pattern=Joe requests
func getReleases(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		pattern := ""
		limit := 25
		offset := 0
		var sortBy release.SortBy
		var sortOrder release.SortOrder

		{ // this block reads the query strings if any
			pattern = r.URL.Query().Get("pattern")

			if limitPageRaw := r.URL.Query().Get("limit"); limitPageRaw != "" {
				limit, err = strconv.Atoi(limitPageRaw)
				if err != nil || limit < 0 {
					d.Logger.Log("bad get releases request, limit")
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
					d.Logger.Log("bad request, offset")
					response.Data = jSendFailData{
						ErrorReason:  "offset",
						ErrorMessage: "bad request, offset can't be negative",
					}
					statusCode = http.StatusBadRequest
				}
			}

			sort := r.URL.Query().Get("sort")
			sortSplit := strings.Split(sort, "_")

			sortOrder = release.SortAscending
			switch sortByQuery := sortSplit[0]; sortByQuery {
			case "type":
				sortBy = release.SortByType
			case "channel":
				sortBy = release.SortByOwner
			default:
				sortBy = release.SortCreationTime
				sortOrder = release.SortDescending
			}
			if len(sortSplit) > 1 {
				switch sortOrderQuery := sortSplit[1]; sortOrderQuery {
				case "dsc":
					sortOrder = release.SortDescending
				default:
					sortOrder = release.SortAscending
				}
			}

		}
		// if queries are clean
		if response.Data == nil {
			releases, err := d.ReleaseService.SearchRelease(pattern, sortBy, sortOrder, limit, offset)
			if err != nil {
				d.Logger.Log("fetching of releases failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when getting releases"
				statusCode = http.StatusInternalServerError
			} else {
				response.Status = "success"
				for _, rel := range releases {
					if rel.Type == release.Image {
						rel.Content = d.HostAddress + d.ImageServingRoute + url.PathEscape(rel.Content)
					}
				}
				response.Data = releases
				d.Logger.Log("success fetching releases")
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// putRelease returns a handler for PUT /releases/{id} requests
func putRelease(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("put attempt of non invalid release id %s", idRaw)
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
							d.Logger.Log("image found on put request")
							defer os.Remove(tmpFile.Name())
							defer tmpFile.Close()
							d.Logger.Log(fmt.Sprintf("temp file saved: %s", tmpFile.Name()))
							rel.Content = generateFileNameForStorage(fileName, "release")
							rel.Type = release.Image
						case errUnacceptedType:
							response.Data = jSendFailData{
								ErrorMessage: "image",
								ErrorReason:  "only types image/jpeg & image/png are accepted",
							}
							statusCode = http.StatusBadRequest
						case errReadingFromImage:
							d.Logger.Log("image not found on put request")
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
			{ // TODO secure block
				// TODO check if release in an official catalog
				//if username != r.Header.Get("authorized_username") {
				//	if _, err := (*service).GetUser(username); err == nil {
				//		(*logger).Log("unauthorized update user attempt")
				//		w.WriteHeader(http.StatusUnauthorized)
				//		return
				//	}
				//}
			}
			if response.Data == nil {
				if response.Data == nil {
					rel.ID = id
					rel, err = d.ReleaseService.UpdateRelease(rel)
					switch err {
					case nil:
						if rel.Type == release.Image {
							err := saveTempFilePermanentlyToPath(tmpFile, d.ImageStoragePath+rel.Content)
							if err != nil {
								d.Logger.Log("updating of release failed because: %v", err)
								response.Status = "error"
								response.Message = "server error when updating release"
								statusCode = http.StatusInternalServerError
								_ = d.ReleaseService.DeleteRelease(rel.ID)
							}
						}
						if response.Message == "" {
							d.Logger.Log("success updating release %d", id)
							response.Status = "success"
							rel.Content = d.HostAddress + d.ImageServingRoute + url.PathEscape(rel.Content)
							response.Data = *rel
							// TODO delete old image if image updated
						}
					case release.ErrAttemptToChangeReleaseType:
						d.Logger.Log("update attempt of release type for release %d", id)
						response.Data = jSendFailData{
							ErrorReason:  "type",
							ErrorMessage: "release type cannot be changed",
						}
						statusCode = http.StatusNotFound
					case release.ErrReleaseNotFound:
						d.Logger.Log("update attempt of non existing release %d", id)
						response.Data = jSendFailData{
							ErrorReason:  "releaseID",
							ErrorMessage: fmt.Sprintf("release of id %d not found", id),
						}
						statusCode = http.StatusNotFound
					case release.ErrSomeReleaseDataNotPersisted:
						_ = d.ReleaseService.DeleteRelease(rel.ID)
						fallthrough
					default:
						d.Logger.Log("update of release failed because: %v", err)
						response.Status = "error"
						response.Message = "server error when adding release"
						statusCode = http.StatusInternalServerError
					}
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteRelease returns a handler for DELETE /releases/{id} requests
func deleteRelease(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]

		id, err := strconv.Atoi(idRaw)
		if err != nil {
			s.Logger.Log("delete attempt of non invalid release id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "releaseID",
				ErrorMessage: fmt.Sprintf("invalid releaseID %d", id),
			}
			statusCode = http.StatusBadRequest
		} else {
			// TODO secure route
			//{ // this block sanitizes the returned User if it's not the user herself accessing the route
			//	if releaseID != r.Header.Get("authorized_username") {
			//		(*logger).Log(fmt.Sprintf("user %s fetched user %s", r.Header.Get("authorized_username"), rel.Username))
			//		rel.Email = ""
			//		rel.BookmarkedPosts = make(map[int]time.Time)
			//	}
			//}
			// TODO delete image if image type
			s.Logger.Log("trying to delete release %d", id)
			err := s.ReleaseService.DeleteRelease(id)
			switch err {
			case nil:
				response.Status = "success"
				s.Logger.Log("success deleting release %d", id)
			case release.ErrReleaseNotFound:
				s.Logger.Log("deletion of release failed because: %v", err)
				response.Data = jSendFailData{
					ErrorReason:  "releaseID",
					ErrorMessage: fmt.Sprintf("release of id %d not found", id),
				}
				statusCode = http.StatusNotFound
				statusCode = http.StatusNotFound
			default:
				s.Logger.Log("deletion of release failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when adding release"
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
