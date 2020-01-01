package rest

import (
	"encoding/json"
	"fmt"
	"github.com/slim-crown/issue-1-REST/pkg/domain/release"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// postReleases returns a handler for POST /releases requests
func postReleases(d *Enviroment) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		newRelease := new(release.Release)
		var tmpFile *os.File
		{ // this block parses the JSON part of the request
			err := json.Unmarshal([]byte(r.PostFormValue("JSON")), newRelease)
			if err != nil {
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "use multipart for for posting Releases. A part named 'JSON' for Release data \nand a file called 'image' if release is of image type JPG/PNG."
				statusCode = http.StatusBadRequest
				response.Data = responseData
			}
		}
		if response.Data == nil {
			{
				// this block checks for required fields
				if newRelease.OwnerChannel == "" {
					var responseData struct {
						Data string `json:"OwnerChannel"`
					}
					responseData.Data = "OwnerChannel is required"
					response.Data = responseData
					// if required fields aren't present
					d.Logger.Log("bad add release request")
					statusCode = http.StatusBadRequest
				} else {
					{
						// TODO check if given channel exists
					}
				}
			}
			if response.Data == nil {
				{ // this block extracts the image file if necessary
					switch newRelease.Type {
					case release.Image:
						var responseData struct {
							Data string `json:"image"`
						}
						tmpFile, fileName, err := saveImageFromRequest(r, "image")
						switch err {
						case nil:
							{ //TODO this block generates the name of the file
								d.Logger.Log(fmt.Sprintf("saving files: %s", fileName))
								newRelease.Content = fileName
							}
							defer os.Remove(tmpFile.Name())
						case errUnacceptedType:
							responseData.Data = "only types image/jpeg & image/png are accepted"
							response.Data = responseData
							statusCode = http.StatusBadRequest
						case errReadingFromImage:
							responseData.Data = "unable to read image file"
							response.Data = responseData
							statusCode = http.StatusBadRequest
						default:
							response.Status = "error"
							responseData.Data = "server error when adding release"
							response.Data = responseData
							statusCode = http.StatusInternalServerError
						}
					case release.Text:
					default:
						var responseData struct {
							Data string `json:"type"`
						}
						responseData.Data = "type can only be 'text' or 'image'"
						statusCode = http.StatusBadRequest
						response.Data = responseData
					}
				}
				if response.Data == nil {
					d.Logger.Log("trying to add release")

					newRelease, err := d.ReleaseService.AddRelease(newRelease)
					switch err {
					case nil:
						err := saveTempFilePermanentlyToPath(tmpFile, d.ImageStoragePath+newRelease.Content)
						if err != nil {
							d.Logger.Log("adding of release failed because: %v", err)
							var responseData struct {
								Data string `json:"message"`
							}
							response.Status = "error"
							responseData.Data = "server error when adding release"
							response.Data = responseData
							statusCode = http.StatusInternalServerError
						} else {
							{ // TODO add to channel catalog

							}
							response.Status = "success"
							response.Data = *newRelease
							d.Logger.Log("success adding release %d to channel %s", newRelease.ID, newRelease.OwnerChannel)
						}
					default:
						d.Logger.Log("adding of release failed because: %v", err)
						var responseData struct {
							Data string `json:"message"`
						}
						response.Status = "error"
						responseData.Data = "server error when adding release"
						response.Data = responseData
						statusCode = http.StatusInternalServerError
					}
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

var errUnacceptedType = fmt.Errorf("file mime type not accepted")
var errReadingFromImage = fmt.Errorf("file mime type not accepted")

func saveImageFromRequest(r *http.Request, fileName string) (*os.File, string, error) {
	file, header, err := r.FormFile(fileName)
	if err != nil {
		return nil, "", errReadingFromImage
	}
	defer file.Close()
	err = checkIfFileIsAcceptedType(file)
	if err != nil {
		return nil, "", err
	}
	newFile, err := ioutil.TempFile("", "tempIMG")
	if err != nil {
		return nil, "", err
	}
	defer newFile.Close()
	_, err = io.Copy(newFile, file)
	if err != nil {
		return nil, "", err
	}
	return newFile, header.Filename, nil
}

func checkIfFileIsAcceptedType(file multipart.File) error { // this block checks if image is of accepted types
	acceptedTypes := map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
	}
	tempBuffer := make([]byte, 512)
	_, err := file.ReadAt(tempBuffer, 0)
	if err != nil {
		return errReadingFromImage
	}
	contentType := http.DetectContentType(tempBuffer)
	if _, ok := acceptedTypes[contentType]; !ok {
		return errUnacceptedType
	}
	return err
}

func saveTempFilePermanentlyToPath(tmpFile *os.File, path string) error {
	newFile, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = io.Copy(newFile, tmpFile)
	if err != nil {
		return err
	}
	return nil
}

// getRelease returns a handler for GET /releases/{id} requests
func getRelease(d *Enviroment) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid release id %s", idRaw)
			var responseData struct {
				Data string `json:"releaseID"`
			}
			responseData.Data = fmt.Sprintf("invalid releaseID %d", id)
			response.Data = responseData
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch release %d", id)
			rel, err := d.ReleaseService.GetRelease(id)
			switch err {
			case nil:
				response.Status = "success"
				if rel.Type == release.Image {
					rel.Content = d.HostAddress + d.ImageServingRoute + rel.Content
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
				d.Logger.Log("fetch attempt of non existing release %d", id)
				var responseData struct {
					Data string `json:"releaseID"`
				}
				responseData.Data = fmt.Sprintf("release of releaseID %d not found", id)
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of release failed because: %d", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when fetching release"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}

		writeResponseToWriter(response, w, statusCode)
	}
}

// getReleases returns a handler for GET /releases?sort=new&limit=5&offset=0&pattern=Joe requests
func getReleases(d *Enviroment) func(w http.ResponseWriter, r *http.Request) {
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
					d.Logger.Log("bad request, offset")
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
				d.Logger.Log("fetching of releases failed because: %d", err.Error())

				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = "server error when getting releases"
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			} else {
				response.Status = "success"
				response.Data = releases
				d.Logger.Log("success fetching releases")
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// putRelease returns a handler for PUT /releases/{id} requests
func putRelease(d *Enviroment) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("put attempt of non invalid release id %s", idRaw)
			var responseData struct {
				Data string `json:"releaseID"`
			}
			responseData.Data = fmt.Sprintf("invalid releaseID %d", id)
			response.Data = responseData
			statusCode = http.StatusBadRequest
		}

		{ // TODO secure block
			//if username != r.Header.Get("authorized_username") {
			//	if _, err := (*service).GetUser(username); err == nil {
			//		(*logger).Log("unauthorized update user attempt")
			//		w.WriteHeader(http.StatusUnauthorized)
			//		return
			//	}
			//}
		}
		rel := new(release.Release)
		var tmpFile *os.File
		{ // this block parses the request
			err := json.Unmarshal([]byte(r.PostFormValue("JSON")), rel)
			if err != nil {
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "use multipart for for posting Releases. A part named 'JSON' for Release data \nand a file called 'image' if release is of image type JPG/PNG."
				statusCode = http.StatusBadRequest
				response.Data = responseData
			}
		}
		if response.Data == nil {
			// if JSON parsing doesn't fail
			if rel.Title == "" && rel.GenreDefining == "" && rel.Description == "" && len(rel.Genres) == 0 && len(rel.Authors) == 0 {
				var responseData struct {
					Data string `json:"message"`
				}
				responseData.Data = "bad request, data sent doesn't contain updatable data"
				response.Data = responseData
				statusCode = http.StatusBadRequest
			} else {
				{ // this block extracts the image file if necessary
					switch rel.Type {
					case release.Image:
						var responseData struct {
							Data string `json:"image"`
						}
						tmpFile, fileName, err := saveImageFromRequest(r, "image")
						switch err {
						case nil:
							{ //TODO this block generates the name of the file
								d.Logger.Log(fmt.Sprintf("saving files: %s", fileName))
								rel.Content = fileName
							}
							defer os.Remove(tmpFile.Name())
						case errUnacceptedType:
							responseData.Data = "only types image/jpeg & image/png are accepted"
							response.Data = responseData
							statusCode = http.StatusBadRequest
						case errReadingFromImage:
							responseData.Data = "unable to read image file"
							response.Data = responseData
							statusCode = http.StatusBadRequest
						default:
							response.Status = "error"
							responseData.Data = "server error when adding release"
							response.Data = responseData
							statusCode = http.StatusInternalServerError
						}
					case release.Text:
					default:
						var responseData struct {
							Data string `json:"type"`
						}
						responseData.Data = "type can only be 'text' or 'image'"
						statusCode = http.StatusBadRequest
						response.Data = responseData
					}
				}
				if response.Data == nil {
					rel.ID = id
					rel, err = d.ReleaseService.UpdateRelease(rel)
					switch err {
					case nil:
						err := saveTempFilePermanentlyToPath(tmpFile, d.ImageStoragePath+rel.Content)
						if err != nil {
							d.Logger.Log("adding of release failed because: %v", err)
							var responseData struct {
								Data string `json:"message"`
							}
							response.Status = "error"
							responseData.Data = "server error when adding release"
							response.Data = responseData
							statusCode = http.StatusInternalServerError
						} else {
							d.Logger.Log("success updating release %d", id)
							response.Status = "success"
							response.Data = *rel
						}
					case release.ErrReleaseNotFound:
						d.Logger.Log("update attempt of non existing release %d", id)
						var responseData struct {
							Data string `json:"releaseID"`
						}
						responseData.Data = fmt.Sprintf("release of id %d not found", id)
						response.Data = responseData
						statusCode = http.StatusNotFound
					default:
						d.Logger.Log("update of release failed because: %d", err.Error())
						var responseData struct {
							Data string `json:"message"`
						}
						response.Status = "error"
						responseData.Data = "server error when updating release"
						response.Data = responseData
						statusCode = http.StatusInternalServerError
					}
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// deleteRelease returns a handler for DELETE /releases/{id} requests
func deleteRelease(s *Enviroment) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]

		id, err := strconv.Atoi(idRaw)
		if err != nil {
			s.Logger.Log("bad delete release request")
			var responseData struct {
				Data string `json:"id"`
			}
			responseData.Data = "invalid id"
			response.Data = responseData
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
			s.Logger.Log("trying to delete release %d", id)
			err := s.ReleaseService.DeleteRelease(id)
			switch err {
			case nil:
				response.Status = "success"
				s.Logger.Log("success deleting release %d", id)
			case release.ErrReleaseNotFound:
				s.Logger.Log("deletion of release failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"id"`
				}
				responseData.Data = fmt.Sprintf("release %d not found", id)
				response.Data = responseData
				statusCode = http.StatusNotFound
			default:
				s.Logger.Log("deletion of release failed because: %s", err.Error())
				var responseData struct {
					Data string `json:"message"`
				}
				response.Status = "error"
				responseData.Data = fmt.Sprintf("server error when deleting feed")
				response.Data = responseData
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
