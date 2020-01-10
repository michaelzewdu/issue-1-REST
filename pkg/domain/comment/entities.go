package comment

import "time"

// Comment repetesents standard comments users can attach
// to a post or another comment.
// replyTo is either and id of another comment or -1 if
// it's a reply to original post.
type Comment struct {
	ID           uint      `json:"id"`
	OriginPost   uint      `json:"originpost"`
	Commenter    string    `json:"comment"`
	Content      string    `json:"content"`
	ReplyTo      uint      `json:"replyto"`
	CreationTime time.Time `json:"creationtime"`
}

type Post struct {
	ID int
}
