package handlers

import (
	"net/http"

	"github.com/google/uuid"
)

// function that allow the logged in user to comment onthe posts
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}

	// Ensure the user is logged in
	cookies, err := r.Cookie("Token")
	if err != nil {
		http.Redirect(w, r, "user not logged in", http.StatusUnauthorized)
		return
	}
	userID := cookies.Value

	// Parse form data
	if err := r.ParseForm(); err != nil {
		ErrorHandler(w, http.StatusBadRequest)
		return
	}

	postID := r.FormValue("post_id")
	content := r.FormValue("content")
	parentID := r.FormValue("parent_id")

	commentID := uuid.New().String()
	if content == "" {
		http.Error(w, "Comment can not be empty ", http.StatusInternalServerError)
		return
	} else {
		// Prepare the query based on whether it's a reply or top-level comment
		var query string
		var args []interface{}

		if parentID != "" {
			// Verify parent comment exists and belongs to the same post
			var parentPostID string
			err = DB.QueryRow("SELECT post_id FROM comments WHERE id = ?", parentID).Scan(&parentPostID)
			if err != nil {
				http.Error(w, "Parent comment not found", http.StatusBadRequest)
				return
			}
			if parentPostID != postID {
				http.Error(w, "Parent comment does not belong to the specified post", http.StatusBadRequest)
				return
			}

			query = `INSERT INTO comments (id, post_id, user_id, parent_id, content) VALUES (?, ?, ?, ?, ?)`
			args = []interface{}{commentID, postID, userID, parentID, content}
		} else {
			query = `INSERT INTO comments (id, post_id, user_id, content) VALUES (?, ?, ?, ?)`
			args = []interface{}{commentID, postID, userID, content}
		}

		// Execute the insert
		_, err = DB.Exec(query, args...)
		if err != nil {
			http.Error(w, "Failed to create comment", http.StatusInternalServerError)
			return
		}

		// Insert comment into the database
		// _, err = DB.Exec("INSERT INTO comments (id, post_id, user_id, content) VALUES (?, ?, ?, ?)", commentID, postID, userID, content)
		// if err != nil {
		// 	http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		// 	return
		// }
	}

	// http.Redirect(w, r, "/posts", http.StatusOK)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Comment created successfully"))
}
