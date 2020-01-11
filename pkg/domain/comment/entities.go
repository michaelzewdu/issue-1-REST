package comment

import "time"

// Comment repetesents standard comments users can attach
// to a post or another comment.
// replyTo is either and id of another comment or -1 if
// it's a reply to original post.
type Comment struct {
	ID           int       `json:"id"`
	OriginPost   int       `json:"originpost"`
	Commenter    string    `json:"comment"`
	Content      string    `json:"content"`
	ReplyTo      int       `json:"replyto"`
	CreationTime time.Time `json:"creationtime"`
}

type Post struct {
	ID int `json:"id"`
}
