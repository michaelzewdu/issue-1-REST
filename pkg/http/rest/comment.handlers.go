package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/slim-crown/issue-1-REST/pkg/domain/comment"
	"github.com/slim-crown/issue-1-REST/pkg/domain/feed"
	"net/http"
	"strconv"
)

//getCommentID returns a handler for GET  /posts/{postID}/comments/{commentID}
func getComment(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)

		commentID := vars["commentID"]

		{ // this block secures the route
			//TODO
			if commentID != r.Header.Get("authorized_User") {
				d.Logger.Log("unauthorized post comment request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		idC, err := strconv.Atoi(commentID)

		if err != nil {
			d.Logger.Log("fetch attempt of non invalid type of comment ID %s", idC)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", idC),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Comment %d", idC)
			rel, err := d.CommentService.GetComment(idC)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *rel
				d.Logger.Log("success fetching comment %d", idC)
			case comment.ErrCommentNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "commentID",
					ErrorMessage: fmt.Sprintf("comment of commentID %d not found", idC),
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
//TODO i don't know how to handle this query yet
func getComments(d *Setup) func(w http.ResponseWriter, r *http.Request) {
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

//postComment returns the handler for POST  /posts/{postID}/comments
func postComment(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		//{ // this block secures the route
		//	if commenter != r.Header.Get("authorized_User") {
		//		d.Logger.Log("unauthorized user comment request")
		//		w.WriteHeader(http.StatusUnauthorized)
		//		return
		//	}
		//}'
		newComment := comment.Comment{}
		{ // checks if requests uses forms or JSON and parses then
			newComment.Content = r.FormValue("content")
			if newComment.Content != "" {
				newComment.Content = r.FormValue("content")
				newComment.Commenter = r.FormValue("commenter")
				idr := r.FormValue("replyTo")
				a, errrt := strconv.Atoi(idr)
				newComment.ReplyTo = a
				if errrt != nil {
					d.Logger.Log("fetch attempt of non invalid type of replyto ID %s", idr)
					response.Data = jSendFailData{
						ErrorReason:  "replyTo ID",
						ErrorMessage: fmt.Sprintf("invalid replyto %d", idr),
					}
					statusCode = http.StatusBadRequest
				}
				ido := r.FormValue(("originPost"))
				b, errop := strconv.Atoi((ido))
				newComment.OriginPost = b
				if errop != nil {
					d.Logger.Log("fetch attempt of non invalid type of originPost ID %s", ido)
					response.Data = jSendFailData{
						ErrorReason:  "originPost ID",
						ErrorMessage: fmt.Sprintf("invalid originPost %d", ido),
					}
					statusCode = http.StatusBadRequest
				}

				//newComment.OriginPost = r.FormValue("originPost")
				//newComment.ReplyTo = r.FormValue("replyTo")
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

			if newComment.Content == "" {
				response.Data = jSendFailData{
					ErrorReason:  "content",
					ErrorMessage: "content is required",
				}
			}

		} else {
			//TODO do we limit content
			if len(newComment.Content) > 100 || len(newComment.Content) < 1 {
				response.Data = jSendFailData{
					ErrorReason:  "conent",
					ErrorMessage: "content length shouldn't be shorter that 1 and longer than 100 chars",
				}
			}
		}
		if response.Data == nil {
			d.Logger.Log("trying to add comment %s", newComment.Content)
			err := d.CommentService.AddComment(&newComment)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = newComment
				d.Logger.Log("success adding comment %s", newComment.Content)
			default:

				d.Logger.Log("adding of comment failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when adding comment"
				statusCode = http.StatusInternalServerError
			}
		} else {
			// if required fields aren't present
			d.Logger.Log("bad adding comment request")
			statusCode = http.StatusBadRequest
		}

		writeResponseToWriter(response, w, statusCode)

	}

}

//patchComment returns the handler PATCH  /posts/{postID}/comments/{commentID}
func patchComment(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		idRaw := vars["commentID"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid type of comment ID %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", idRaw),
			}
			statusCode = http.StatusBadRequest
		}

		newComment := new(comment.Comment)
		erro := json.NewDecoder(r.Body).Decode(newComment)
		if erro != nil {
			response.Data = jSendFailData{
				ErrorReason: "request format",
				ErrorMessage: `bad request, use format
			{"content":"username len 5-22 chars",
			"originPost":"channel",
			"replyId":"title",
			"commenter":"commenter"
			}`,
			}
			d.Logger.Log("bad update comment request")
			statusCode = http.StatusBadRequest
		}
		if response.Data == nil {
			// if JSON parsing doesn't fail

			if newComment.Content == "" {
				response.Data = jSendFailData{
					ErrorReason:  "request",
					ErrorMessage: "request doesn't contain updatable data",
				}
				statusCode = http.StatusBadRequest
			} else {
				erron := d.CommentService.UpdateComment(newComment, id)
				switch erron {
				case nil:
					d.Logger.Log("success patch post %s", newComment.Content)
					response.Status = "success"
					response.Data = *newComment

				default:
					d.Logger.Log("adding of comment failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when adding comment"
					statusCode = http.StatusInternalServerError
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

//deleteComment returns the handlers for DELETE  /posts/{postID}/comments/{commentID}
func deleteComment(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["commentID"]

		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("delete attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", id),
			}
			statusCode = http.StatusBadRequest
		} else {
			d.Logger.Log("trying to delete Comment %d", id)
			err := d.CommentService.DeleteComment(id)
			switch err {
			case nil:
				response.Status = "success"
				d.Logger.Log("success deleting Comment %d", id)
			case comment.ErrCommentNotFound:
				d.Logger.Log("deletion of Post failed because: %v", err)
				response.Data = jSendFailData{
					ErrorReason:  "commentID",
					ErrorMessage: fmt.Sprintf("Comment of id %d not found", id),
				}
				statusCode = http.StatusNotFound

			default:
				response.Status = "error"
				response.Message = "server error when adding Post"
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)

	}

}

//GET  getReply returns a handler for /posts/{postID}/comments/{rootCommentID}/replies/{commentID}
func getReply(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		c := new(comment.Comment)
		rootCommentIDRaw := vars["rootCommentID"]
		e, erre := strconv.Atoi(rootCommentIDRaw)
		c.ReplyTo = e
		if erre != nil {
			c.ReplyTo = -1
		}
		commentID := vars["commentID"]

		{ // this block secures the route
			//TODO
			if commentID != r.Header.Get("authorized_User") {
				d.Logger.Log("unauthorized post comment request")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		idC, err := strconv.Atoi(commentID)

		if err != nil {
			d.Logger.Log("fetch attempt of non invalid type of comment ID %s", idC)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", idC),
			}
			statusCode = http.StatusBadRequest
		}

		if response.Data == nil {
			d.Logger.Log("trying to fetch Comment %d", idC)
			rel, err := d.CommentService.GetReply(idC)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = *rel
				d.Logger.Log("success fetching comment %d", idC)
			case comment.ErrCommentNotFound:
				response.Data = jSendFailData{
					ErrorReason:  "commentID",
					ErrorMessage: fmt.Sprintf("comment of commentID %d not found", idC),
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
//TODO
//addReply returns the handler for POST  /posts/{postID}/comments/{rootCommentID}/replies
func addReply(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		//{ // this block secures the route
		//	if commenter != r.Header.Get("authorized_User") {
		//		d.Logger.Log("unauthorized user comment request")
		//		w.WriteHeader(http.StatusUnauthorized)
		//		return
		//	}
		//}'
		newComment := comment.Comment{}
		vars := mux.Vars(r)
		rootCommentIDRaw := vars["rootCommentID"]
		e, erre := strconv.Atoi(rootCommentIDRaw)
		newComment.ReplyTo = e
		if erre != nil {
			newComment.ReplyTo = -1
		}
		{ // checks if requests uses forms or JSON and parses then
			newComment.Content = r.FormValue("content")
			if newComment.Content != "" {
				newComment.Content = r.FormValue("content")
				newComment.Commenter = r.FormValue("commenter")
				idr := r.FormValue("replyTo")
				a, errrt := strconv.Atoi(idr)
				newComment.ReplyTo = a
				if errrt != nil {
					d.Logger.Log("fetch attempt of non invalid type of replyto ID %s", idr)
					response.Data = jSendFailData{
						ErrorReason:  "replyTo ID",
						ErrorMessage: fmt.Sprintf("invalid replyto %d", idr),
					}
					statusCode = http.StatusBadRequest
				}
				ido := r.FormValue("originPost")
				b, errop := strconv.Atoi(ido)
				newComment.OriginPost = b
				if errop != nil {
					d.Logger.Log("fetch attempt of non invalid type of originPost ID %s", ido)
					response.Data = jSendFailData{
						ErrorReason:  "originPost ID",
						ErrorMessage: fmt.Sprintf("invalid originPost %d", ido),
					}
					statusCode = http.StatusBadRequest
				}

				//newComment.OriginPost = r.FormValue("originPost")
				//newComment.ReplyTo = r.FormValue("replyTo")
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

			if newComment.Content == "" {
				response.Data = jSendFailData{
					ErrorReason:  "content",
					ErrorMessage: "content is required",
				}
			}

		} else {
			//TODO do we limit content
			if len(newComment.Content) > 100 || len(newComment.Content) < 1 {
				response.Data = jSendFailData{
					ErrorReason:  "content",
					ErrorMessage: "content length shouldn't be shorter that 1 and longer than 100 chars",
				}
			}
		}
		if response.Data == nil {
			d.Logger.Log("trying to add comment %s", newComment.Content)
			err := d.CommentService.AddReply(&newComment)
			switch err {
			case nil:
				response.Status = "success"
				response.Data = newComment
				d.Logger.Log("success adding comment %s", newComment.Content)
			default:

				d.Logger.Log("adding of comment failed because: %v", err)
				response.Status = "error"
				response.Message = "server error when adding comment"
				statusCode = http.StatusInternalServerError
			}
		} else {
			// if required fields aren't present
			d.Logger.Log("bad adding comment request")
			statusCode = http.StatusBadRequest
		}

		writeResponseToWriter(response, w, statusCode)

	}

}

//patchReply returns the handler for PATCH  /posts/{postID}/comments/{rootCommentID}/replies/{commentID}
func patchReply(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK
		vars := mux.Vars(r)
		c := new(comment.Comment)
		rootCommentIDRaw := vars["rootCommentID"]
		e, erre := strconv.Atoi(rootCommentIDRaw)
		c.ReplyTo = e
		if erre != nil {
			c.ReplyTo = -1
		}
		idRaw := vars["commentID"]
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("fetch attempt of non invalid type of comment ID %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", idRaw),
			}
			statusCode = http.StatusBadRequest
		}

		newComment := new(comment.Comment)
		erro := json.NewDecoder(r.Body).Decode(newComment)
		if erro != nil {
			response.Data = jSendFailData{
				ErrorReason: "request format",
				ErrorMessage: `bad request, use format
			{"content":"username len 5-22 chars",
			"originPost":"channel",
			"replyId":"title",
			"commenter":"commenter"
			}`,
			}
			d.Logger.Log("bad update comment request")
			statusCode = http.StatusBadRequest
		}
		if response.Data == nil {
			// if JSON parsing doesn't fail

			if newComment.Content == "" {
				response.Data = jSendFailData{
					ErrorReason:  "request",
					ErrorMessage: "request doesn't contain updatable data",
				}
				statusCode = http.StatusBadRequest
			} else {
				erron := d.CommentService.UpdateReply(newComment, id)
				switch erron {
				case nil:
					d.Logger.Log("success patch post %s", newComment.Content)
					response.Status = "success"
					response.Data = *newComment

				default:
					d.Logger.Log("adding of comment failed because: %v", err)
					response.Status = "error"
					response.Message = "server error when adding comment"
					statusCode = http.StatusInternalServerError
				}
			}
		}
		writeResponseToWriter(response, w, statusCode)
	}
}

//deleteReply returns the handler for DELETE  /posts/{postID}/comments/{rootCommentID}/replies/{commentID}
func deleteReply(d *Setup) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response jSendResponse
		response.Status = "fail"
		statusCode := http.StatusOK

		vars := mux.Vars(r)
		idRaw := vars["commentID"]
		c := new(comment.Comment)
		rootCommentIDRaw := vars["rootCommentID"]
		e, erre := strconv.Atoi(rootCommentIDRaw)
		c.ReplyTo = e
		if erre != nil {
			c.ReplyTo = -1
		}

		id, err := strconv.Atoi(idRaw)
		if err != nil {
			d.Logger.Log("delete attempt of non invalid post id %s", idRaw)
			response.Data = jSendFailData{
				ErrorReason:  "commentID",
				ErrorMessage: fmt.Sprintf("invalid commentID %d", id),
			}
			statusCode = http.StatusBadRequest
		} else {
			d.Logger.Log("trying to delete Comment %d", id)
			err := d.CommentService.DeleteReply(id)
			switch err {
			case nil:
				response.Status = "success"
				d.Logger.Log("success deleting Comment %d", id)
			case comment.ErrCommentNotFound:
				d.Logger.Log("deletion of Post failed because: %v", err)
				response.Data = jSendFailData{
					ErrorReason:  "commentID",
					ErrorMessage: fmt.Sprintf("Comment of id %d not found", id),
				}
				statusCode = http.StatusNotFound

			default:
				response.Status = "error"
				response.Message = "server error when adding Post"
				statusCode = http.StatusInternalServerError
			}
		}
		writeResponseToWriter(response, w, statusCode)

	}

}
