package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
)

// GET: /posts/:id ...getpost(id)
// getPost returns a handler for GET /posts/{id} requests
func getPost(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "postID",
				ErrorMessage: fmt.Sprintf("invalid postID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Post %d", id)
			rel, err := d.PostService.GetPost(id)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *rel
				d.Logger.Log("success fetching post %d", id)
			case post.ErrPostNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "postID",
					ErrorMessage: fmt.Sprintf("post of postID %d not found", id),
				}
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of post failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching post"
				statusCode = http.StatusInternalServerError

			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// POST: /posts/.....addpost
// postPost returns a handler for POST: /posts/.....{addpost}
func postPost(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		newPost := new(post.Post)
		{ // checks if requests uses forms or JSON and parses then
			newPost.PostedByUsername = r.FormValue("username")
			if newPost.PostedByUsername != "" {
				newPost.Title = r.FormValue("title")
				newPost.Description = r.FormValue("description")
				newPost.OriginChannel = r.FormValue("channelName")

			} else {
				err := json.NewDecoder(r.Body).Decode(newPost)
				if err != nil {
					response.Data = jSendFailData{
						ErrorReason: "request format",
						ErrorMessage: `bad request, use format
				{"poster":"username len 5-22 chars",
				"originChannel":"channel",
				"title":"title",
				"description":"description"
				}`,
					}
					s.Logger.Log("bad update post request")
					statusCode = http.StatusBadRequest
				}
			}
		}

		if response.Data == nil {

			if newPost.OriginChannel == "" {
				response.Data = jSendFailData{
					ErrorReason:  "username",
					ErrorMessage: "username is required",
				}
			}
			if newPost.Title == "" {
				response.Data = jSendFailData{
					ErrorReason:  "username",
					ErrorMessage: "username is required",
				}
			}
			if newPost.PostedByUsername == "" {
				response.Data = jSendFailData{
					ErrorReason:  "username",
					ErrorMessage: "username is required",
				}
			} else {
				if len(newPost.PostedByUsername) > 22 || len(newPost.PostedByUsername) < 5 {
					response.Data = jSendFailData{
						ErrorReason:  "username",
						ErrorMessage: "username length shouldn't be shorter that 5 and longer than 22 chars",
					}
				}
			}
			if response.Data == nil {
				s.Logger.Log("trying to add post %s", newPost.PostedByUsername, newPost.Title, newPost.OriginChannel, newPost.Description)
				pos, err := s.PostService.AddPost(newPost)
				switch err {
				case nil:
					response.Status = "success"
					response.Data = *pos
					s.Logger.Log("success adding user %s", pos.PostedByUsername, pos.Title, pos.OriginChannel, pos.Description)
				default:
					_ = s.PostService.DeletePost(pos.ID)
					s.Logger.Log("adding of post failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when adding post"
					statusCode = http.StatusInternalServerError
				}
			} else {
				// if required fields aren't present
				s.Logger.Log("bad adding user request")
				statusCode = http.StatusBadRequest
			}
		}
		writeResponseToWriter(response, w, statusCode)

	}

}

// PUT: /posts/:id ---
//putPost returns a handler for PUT /posts/{id} requests--{updatePost(p,id)}
func putPost(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)

		newPost := new(post.Post)
		erro := json.NewDecoder(r.Body).Decode(newPost)
		if erro != nil {
			response.Data = jSendFailData{
				ErrorReason: "request format",
				ErrorMessage: `bad request, use format
			{"poster":"username len 5-22 chars",
			"originChannel":"channel",
			"title":"title",
			"description":"description"
			}`,
			}
			s.Logger.Log("bad update post request")
			statusCode = http.StatusBadRequest
		}
		if response.Data == nil {
			// if JSON parsing doesn't fail

			if newPost.PostedByUsername == "" && newPost.OriginChannel == "" && newPost.Title == "" && newPost.Description == "" {
				response.Data = jSendFailData{
					ErrorReason:  "request",
					ErrorMessage: "request doesn't contain updatable data",
				}
				statusCode = http.StatusBadRequest
			} else {
				pos, erron := s.PostService.UpdatePost(newPost, id)
				switch erron {
				case nil:
					s.Logger.Log("success put post %s", idRaw, pos.PostedByUsername, pos.OriginChannel, pos.Title, pos.Description)
					response.Status = "success"
					response.Data = *pos

				default:
					_ = s.PostService.DeletePost(pos.ID)
					s.Logger.Log("adding of post failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when adding post"
					statusCode = http.StatusInternalServerError
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// DELETE: /posts/:id
// postPost returns a handler for DELETE: /posts/:id ---{DeletePost(id)}
func deletePost(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]

		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("delete attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "postID",
				ErrorMessage: fmt.Sprintf("invalid postID %d", id),
			}
			statusCode = http.StatusBadRequest
		} else {
			d.Logger.Log("trying to delete Post %d", id)
			err := d.PostService.DeletePost(id)
			switch err {
			case nil:
				response.Status = "success"
				d.Logger.Log("success deleting Post %d", id)
			case post.ErrPostNotFound:
				d.Logger.Log("deletion of Post failed because: %v", err)
				response.Data = jSendFailData{
					ErrorReason:  "PostID",
					ErrorMessage: fmt.Sprintf("Post of id %d not found", id),
				}
				statusCode = http.StatusNotFound

			default:
				d.Logger.Log("deletion of Post failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when adding Post"
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)

	}

}

// GET: /posts/:id/releases/
// getPostReleases returns a handler for GET: /posts/:id/releases requests ---{getPost,getPostReleases}
func getPostReleases(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "postID",
				ErrorMessage: fmt.Sprintf("invalid postID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Post %d", id)
			pos, err := d.PostService.GetPost(id)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *pos
				d.Logger.Log("success fetching post %d", id)
				rel, erro := d.PostService.GetPostReleases(pos)
				switch erro {
				case nil:
					response.Status = "success"
					response.Data = rel
					d.Logger.Log("success fetching release %d", id)
				default:
					d.Logger.Log("fetching of releases failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when fetching releases of post"
					statusCode = http.StatusInternalServerError
				}

			case post.ErrPostNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "postID",
					ErrorMessage: fmt.Sprintf("post of postID %d not found", id),
				}
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of post failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching post"
				statusCode = http.StatusInternalServerError

			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// GET: /posts/:id/releases/:release_id
// getPostRelease returns a handler for GET: /posts/:id/releases/:rId requests ---{getPost,getPostRelease}
func getPostRelease(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		rIdRaw := vars["rId"]

		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "postID",
				ErrorMessage: fmt.Sprintf("invalid postID %d", id),
			}
			statusCode = http.StatusBadRequest
		}
		rId, erro := strconv.Atoi(idRaw)
		if erro != nil {
			d.Logger.Log("fetch attempt of non invalid release of post id %s", rIdRaw)
			response.Data = jSendFailData{
				ErrorReason:  "releaseeID",
				ErrorMessage: fmt.Sprintf("invalid releaseID %d", rId),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Post %d", id)
			pos, err := d.PostService.GetPost(id)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *pos
				d.Logger.Log("success fetching post %d", id)
				rel, erro := d.PostService.GetPostRelease(pos.ID, rId)
				switch erro {
				case nil:
					response.Status = "success"
					response.Data = *rel
					d.Logger.Log("success fetching release %d", rId)
				case post.ErrReleaseNotFound:
					response.Data = jSendFailData{
						ErrorReason:  "releaseID",
						ErrorMessage: fmt.Sprintf("release of releaseID %d not found", rId),
					}
					statusCode = http.StatusNotFound
				default:
					d.Logger.Log("fetching of releases failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when fetching releases of post"
					statusCode = http.StatusInternalServerError
				}

			case post.ErrPostNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "postID",
					ErrorMessage: fmt.Sprintf("post of postID %d not found", id),
				}
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of post failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching post"
				statusCode = http.StatusInternalServerError

			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

// PUT: /posts/:id/stars

// GET:/posts?sortBy=creation-time&sortOrder=asc&limit=25&offset=0&pattern=John
// getPosts returns a handler for GET /posts?sort=new&limit=5&offset=0&pattern=Joe requests---{searchPost}
func getPosts(s *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		pattern := ""
		limit := 25
		offset := 0
		var sortBy post.SortBy
		var sortOrder post.SortOrder

		{ // this block reads the query strings if any
			pattern = r.URL.Query().Get("pattern")

			if limitPageRaw := r.URL.Query().Get("limit"); limitPageRaw != "" {
				limit, err = strconv.Atoi(limitPageRaw)
				if err != nil || limit < 0 {
					s.Logger.Log("bad get useposts request, limit")
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

			sort := r.URL.Query().Get("sort")
			sortSplit := strings.Split(sort, "_")

			sortOrder = post.SortAscending
			switch sortByQuery := sortSplit[0]; sortByQuery {
			case "username":
				sortBy = post.SortByPoster
			case "channelname":
				sortBy = post.SortByChannel
			case "title":
				sortBy = post.SortByTitle
			default:
				sortBy = post.SortByCreationTime
				sortOrder = post.SortDescending
			}
			if len(sortSplit) > 1 {
				switch sortOrderQuery := sortSplit[1]; sortOrderQuery {
				case "dsc":
					sortOrder = post.SortDescending
				default:
					sortOrder = post.SortAscending
				}
			}

		}
		// if queries are clean
		if response.Data == nil {
			posts, err := s.PostService.SearchPost(pattern, sortBy, sortOrder, limit, offset)
			if err != nil {
				s.Logger.Log("fetching of post failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when getting users"
				statusCode = http.StatusInternalServerError
			} else {
				response.Status = "success"
				// for _, u := range users {
				// 	u.Email = ""
				// 	u.BookmarkedPosts = make(map[int]time.Time)
				// 	if u.PictureURL != "" {
				// 		u.PictureURL = s.HostAddress + s.ImageServingRoute + url.PathEscape(u.PictureURL)
				// 	}
				// }
				response.Data = posts
				s.Logger.Log("success fetching posts")
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

//GET: /posts/:id/stars
// getPostStars returns a handler for GET: /posts/:id/stars requests
func getPostStars(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "postID",
				ErrorMessage: fmt.Sprintf("invalid postID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Post %d", id)
			pos, err := d.PostService.GetPost(id)
			switch err {
			case nil:

				response.Status = "success"
				response.Data = (*pos).Stars
				d.Logger.Log("success fetching post %d stars", id)

			case post.ErrPostNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "postID",
					ErrorMessage: fmt.Sprintf("post of postID %d not found", id),
				}
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of post failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching post"
				statusCode = http.StatusInternalServerError

			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

//GET: /posts/:id/stars/username
// getPostStar returns a handler for GET: /posts/:id/stars/username requests----{getPostStar}
func getPostStar(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		id, err := strconv.Atoi(idRaw)
		username := r.FormValue("username")

		if err != nil {
			d.Logger.Log("fetch attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "postID",
				ErrorMessage: fmt.Sprintf("invalid postID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Star of Post %d and username %s", id, username)
			st, err := d.PostService.GetPostStar(id, username)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *st
				d.Logger.Log("success fetch Star of Post %d and username %s", id, username)

			case post.ErrStarNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "postID/username",
					ErrorMessage: fmt.Sprintf("star of postID %d and username %s not found", id, username),
				}
				statusCode = http.StatusNotFound

			default:
				d.Logger.Log("fetching of post failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching post"
				statusCode = http.StatusInternalServerError

			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

//PUT: /posts/:id/stars
// putPostStar returns a handler for PUT: /posts/:id/stars ...AddPostStar,UpdatePostStar(),DeletePostStar
func putPostStar(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var response jSendResponse
		statusCode := http.StatusOK
		response.Status = "fail"

		vars := mux.Vars(r)
		idRaw := vars["id"]

		id, err := strconv.Atoi(idRaw)

		st := new(post.Star)
		erro := json.NewDecoder(r.Body).Decode(st)

		if erro != nil {
			response.Data = jSendFailData{
				ErrorReason: "request format",
				ErrorMessage: `bad request, use format
			{"username":"username len 5-22 chars",
			"num_of_stars":"num_of_stars(0-5)"}`,
			}
			s.Logger.Log("bad update star request")
			statusCode = http.StatusBadRequest
		} else {
			username := st.Username
			if st.NumOfStars == 0 {
				errs := s.PostService.DeletePostStar(id, username)
				switch errs {
				case nil:
					response.Status = "success"
					s.Logger.Log("success deleting Star of Post %d and username %s", id, username)

				case post.ErrStarNotFound:
					s.Logger.Log("deletion of Star failed because: %v", err)
					response.Data = jSendFailData{
						ErrorReason:  "postID/username",
						ErrorMessage: fmt.Sprintf("star of postID %d and username %s not found", id, username),
					}
					statusCode = http.StatusNotFound

				default:
					s.Logger.Log("deletion of Star failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when adding Star"
					statusCode = http.StatusInternalServerError
				}
			}
			if st.NumOfStars > 5 || st.NumOfStars < 0 {
				response.Data = jSendFailData{
					ErrorReason: "request format",
					ErrorMessage: `bad request, use format
				{"username":"username len 5-22 chars",
				"num_of_stars":"num_of_stars(0-5)"}`,
				}
				s.Logger.Log("bad update star request")
				statusCode = http.StatusBadRequest
			}
			if response.Data == nil {

				_, errr := s.PostService.GetPostStar(id, username)
				switch errr {
				case nil:
					newStar, w := s.PostService.UpdatePostStar(id, st)
					switch w {
					case nil:
						response.Status = "success"
						response.Data = *newStar
						s.Logger.Log("successful in updating Star of Post %d and username %s", id, username)

					default:
						_ = s.PostService.DeletePostStar(id, username)
						s.Logger.Log("adding of post star failed because: %v", err)
						response.Status = "error"
						response.Message = "server error when adding post star"
						statusCode = http.StatusInternalServerError
					}
				case post.ErrStarNotFound:
					newStar, _ := s.PostService.AddPostStar(id, st)
					response.Status = "success"
					response.Data = *newStar
					s.Logger.Log("successful in adding Star of Post %d and username %s", id, username)

				default:
					s.Logger.Log("fetching of post star failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when fetching post star"
					statusCode = http.StatusInternalServerError

				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
