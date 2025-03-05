package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// function that allow the logged in user to comment onthe posts
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure only POST method is allowed
	if r.Method != http.MethodPost {
		ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}

	// Ensure the user is logged in
	cookies, err := r.Cookie("Token")
	if err != nil {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	userID := cookies.Value

	// Parse form data
	if err := r.ParseForm(); err != nil {
		ErrorHandler(w, http.StatusBadRequest)
		return
	}

	// Extract comment details
	postID := r.FormValue("post_id")
	content := r.FormValue("content")
	parentID := r.FormValue("comment_id")

	// Generate a new comment ID
	commentID := uuid.New().String()

	// Validate comment content
	if content == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	// Prepare the query based on whether it's a reply or top-level comment
	var query string
	var args []interface{}

	if parentID != "" {
		query = `
		INSERT INTO comments (id, user_id, parent_id, content) 
		VALUES (?, ?, ?, ?)`
		args = []interface{}{commentID, userID, parentID, content}
		

	} else if postID != "" {
		// Prepare query for top-level comment
		query = `
		INSERT INTO comments (id, post_id, user_id, content) 
		VALUES (?, ?, ?, ?)`
		args = []interface{}{commentID, postID, userID, content}
	} else {
		// No post ID provided
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Execute the insert
	_, err = DB.Exec(query, args...)
	if err != nil {
		fmt.Println(err)
		// Log the actual error for server-side debugging
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Comment created successfully"))
}
