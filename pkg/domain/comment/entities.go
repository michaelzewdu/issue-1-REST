package comment

import "time"

// Comment repetesents standard comments users can attach
// to a post or another comment.
// replyTo is either and id of another comment or -1 if
// it's a reply to original post.
type Comment struct {
	ID            uint
	rootCommentID int
	OriginPost    uint
	Commenter     string
	Content       string
	ReplyTo       uint
	CreationTime  time.Time
}

type Post struct {
	ID int
}
