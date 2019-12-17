package comment

import "time"

// Comment repetesents standard comments users can attach
// to a post or another comment.
// replyTo is either and id of another comment or -1 if
// it's a reply to original post.
type Comment struct {
	ID                int
	CommenterUsername string
	Content           string
	ReplyTo           int
	CreationTime      time.Time
}
