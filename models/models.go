// models/user.go
package models

type User struct {
	ID        string `json:"id"`
	UserEmail string `json:"useremail"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

// models/post.go
type Post struct {
    ID          string `json:"id"`
    UserID      string `json:"user_id"`
    Username    string `json:"username"`
    Title       string `json:"title"`
    Content     string `json:"content"`
    CreatedAt   string `json:"created_at"`
    LikeCount   int    `json:"like_count"`
    DislikeCount int   `json:"dislike_count"`
}

// models/comment.go
type Comment struct {
	ID        string `json:"id"`
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// models/category.go
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// models/like_dislike.go
type LikeDislike struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	PostID    string `json:"post_id"`
	CommentID string `json:"comment_id"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}
