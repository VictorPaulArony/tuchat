package handlers

import (
    "fmt"
    "net/http"
    "github.com/google/uuid"
)

// LikeHandler handles likes/dislikes for both posts and comments
func LikeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        ErrorHandler(w, http.StatusMethodNotAllowed)
        return
    }

    // Check user authentication
    cookies, err := r.Cookie("Token")
    if err != nil {
        http.Error(w, "User not logged in", http.StatusUnauthorized)
        return
    }
    userID := cookies.Value

    // Parse form data
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form data", http.StatusBadRequest)
        return
    }

    // Get form values
    itemID := r.FormValue("id")          // This could be either post_id or comment_id
    itemType := r.FormValue("item_type") // "post" or "comment"
    likeType := r.FormValue("type")      // "like" or "dislike"

    // Validate inputs
    if itemID == "" {
        http.Error(w, "ID is required", http.StatusBadRequest)
        return
    }

    if itemType != "post" && itemType != "comment" {
        http.Error(w, "Invalid item type", http.StatusBadRequest)
        return
    }

    if likeType != "like" && likeType != "dislike" {
        http.Error(w, "Invalid like type", http.StatusBadRequest)
        return
    }

    // Process the like/dislike based on item type
    err = processLike(itemType, itemID, userID, likeType)
    if err != nil {
        http.Error(w, "Failed to process like/dislike", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("liked/disliked created successfully"))
}

// processLike handles the database operations for likes
func processLike(itemType, itemID, userID, likeType string) error {
    var tableName, idColumn string

    // Set table and column names based on item type
    if itemType == "post" {
        tableName = "post_likes"
        idColumn = "post_id"
    } else {
        tableName = "comment_likes"
        idColumn = "comment_id"
    }

    // First verify the item exists
    exists, err := checkItemExists(itemType, itemID)
    if err != nil {
        return fmt.Errorf("error checking item existence: %v", err)
    }
    if !exists {
        return fmt.Errorf("item not found")
    }

    // Check for existing like
    var existingType string
    query := fmt.Sprintf("SELECT type FROM %s WHERE user_id = ? AND %s = ?", tableName, idColumn)
    err = DB.QueryRow(query, userID, itemID).Scan(&existingType)

    if err == nil {
        // Update existing like
        updateQuery := fmt.Sprintf("UPDATE %s SET type = ? WHERE user_id = ? AND %s = ?", tableName, idColumn)
        _, err = DB.Exec(updateQuery, likeType, userID, itemID)
    } else {
        // Insert new like
        insertQuery := fmt.Sprintf("INSERT INTO %s (id, user_id, %s, type) VALUES (?, ?, ?, ?)", tableName, idColumn)
        _, err = DB.Exec(insertQuery, uuid.New().String(), userID, itemID, likeType)
    }

    return err
}

// checkItemExists verifies if the post or comment exists
func checkItemExists(itemType, itemID string) (bool, error) {
    var exists bool
    var query string

    if itemType == "post" {
        query = "SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)"
    } else {
        query = "SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)"
    }

    err := DB.QueryRow(query, itemID).Scan(&exists)
    return exists, err
}