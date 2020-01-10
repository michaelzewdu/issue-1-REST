package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
	"golang.org/x/tools/go/analysis/passes/cgocall/testdata/src/c"
	"net/http"
	"strconv"
)
//getCommentID returns a handler for GET  /posts/{postID}/comments/{commentID}
func getCommentID(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		commenter := vars["commenter"]
		//postID := vars["postID"]
		id, err := strconv.Atoi(idRaw)
		{ // this block secures the route
			if commenter != r.Header.Get("authorized_User") {
				d.Logger.Log("unauthorized post comment request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid comment id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Comment %d", id)
			rel, err := d.CommentService.GetComment(id)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *rel
				d.Logger.Log("success fetching comment %d", id)
			case comment.ErrCommentNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "commentID",
					ErrorMessage: fmt.Sprintf("comment of commentID %d not found", id),
				}
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of comment failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching comment"
				statusCode = http.StatusInternalServerError

			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
//getComment returns a handler forGET  /posts/{postID}/comments?sort=time
func getComment(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		statusCode := http.StatusOK

		response.Status = "fail"

		vars := mux.Vars(r)
		postID := vars["postID"]
		commenter := vars["commenter"]
		{ // this block secures the route
			if commenter != r.Header.Get("authorized_User") {
				d.Logger.Log("unauthorized post comment request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		d.Logger.Log("trying to fetch comment %s", postID)
		f, err := d.FeedService.GetFeed(postID)
		switch err {
		case nil:
			response.Status = "success"
			response.Data = *f
			d.Logger.Log("success fetching comment of %s", postID)
		case feed.ErrFeedNotFound:
			d.Logger.Log("fetching of comment failed because: %v", err)
			response.Data = jSendFailData{
				ErrorReason:  "postID",
				ErrorMessage: fmt.Sprintf("comment of postID %s not found", postID),
			}
			statusCode = http.StatusNotFound
		default:
			d.Logger.Log("getting feed failed because: %v", err)
			response.Status = "error"
			response.Message = "server error when getting feed"
			statusCode = http.StatusNotFound
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

//postComment return the handler for POST  /posts/{postID}/comments
func postComment(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		commenter := vars["commenter"]
		//postID := vars["postID"]
		{ // this block secures the route
			if commenter != r.Header.Get("authorized_User") {
				d.Logger.Log("unauthorized user comment request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		newComment := new(comment.Comment)
		{ // checks if requests uses forms or JSON and parses then
			newComment.Commenter= r.FormValue("username")
			if newComment.Commenter!= "" {
				newComment.Content = r.FormValue("content")
				//i don't know if i should add more
				newComment.OriginPost = r.FormValue("originPost")
				newComment.ReplyTo = r.FormValue("replyTo")
			} else {
				err := json.NewDecoder(r.Body).Decode(newComment)
				if err != nil {
					response.Data = jSendFailData{
						ErrorReason: "request format",
						ErrorMessage: `bad request, use format
				{"originPost":"originPost",
				"commenter":"commenter",
				"content":"content",
				"replyTo":"replyTo"
				}`,
					}
					d.Logger.Log("bad update post request")
					statusCode = http.StatusBadRequest
				}
			}
		}

		if response.Data == nil {

			if newComment.Commenter == "" {
				response.Data = jSendFailData{
					ErrorReason:  "username",
					ErrorMessage: "username is required",
				}
			}
			if newComment.OriginPost == "" {
				response.Data = jSendFailData{
					ErrorReason:  "username",
					ErrorMessage: "username is required",
				}
			}
			} else {
				if len(newComment.Commenter) > 22 || len(newComment.Commenter) < 5 {
					response.Data = jSendFailData{
						ErrorReason:  "username",
						ErrorMessage: "username length shouldn't be shorter that 5 and longer than 22 chars",
					}
				}
			}
			if response.Data == nil {
				d.Logger.Log("trying to add comment %s", newComment.Commenter,newComment.Content, newComment.OriginPost)
				com, err := d.CommentService.AddComment(newComment)
				switch err {
				case nil:
					response.Status = "success"
					response.Data = *com
					d.Logger.Log("success adding user %s", com.Commenter, com.Content, com.OriginPost)
				default:
					_ = d.CommentService.DeletePost(com.ID)
					d.Logger.Log("adding of comment failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when adding comment"
					statusCode = http.StatusInternalServerError
				}
			} else {
				// if required fields aren't present
				d.Logger.Log("bad adding user request")
				statusCode = http.StatusBadRequest
			}
		}
		writeResponseToWriter(response, w, statusCode)

	}

}


//PATCH  /posts/{postID}/comments/{commentID}
//DELETE  /posts/{postID}/comments/{commentID}
//GET  getReplyCommentID returns a handler for /posts/{postID}/comments/{rootCommentID}/replies/{commentID}
func getReplyCommentID(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["id"]
		rootCommentIDRaw :=vars["rootCommentID"]
		c.ReplyTo, err = strconv.Atoi(rootCommentIDRaw)

		if err != nil{
			c.ReplyTo =-1
		}
		commenter := vars["commenter"]
		//postID := vars["postID"]
		id, err := strconv.Atoi(idRaw)
		{ // this block secures the route
			if commenter != r.Header.Get("authorized_User") {
				d.Logger.Log("unauthorized post comment request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid comment id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", id),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Comment %d", id)
			rel, err := d.CommentService.GetComment(id)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *rel
				d.Logger.Log("success fetching comment %d", id)
			case comment.ErrCommentNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "commentID",
					ErrorMessage: fmt.Sprintf("comment of commentID %d not found", id),
				}
				statusCode = http.StatusNotFound
			default:
				d.Logger.Log("fetching of comment failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when fetching comment"
				statusCode = http.StatusInternalServerError

			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}
//GET  /posts/{postID}/comments/{rootCommentID}/replies/?sort=time
//POST  /posts/{postID}/comments/{rootCommentID}/replies
//PATCH  /posts/{postID}/comments/{rootCommentID}/replies/{commentID}
//DELETE  /posts/{postID}/comments/{rootCommentID}/replies/{commentID}

