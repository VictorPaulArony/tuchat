package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxFileSize = 20 * 1024 * 1024 // 20MB to allow for some buffer
	ChunkSize   = 4096             // Read/write in 4KB chunks
)

type Post struct {
	ID          string
	UserID      string
	Title       string
	Content     string
	Media       []byte
	ContentType string
	Categories  []string
	CreatedAt   time.Time
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	// if r.Method == http.MethodGet {
	// RenderTemplates(w, "posts.html", nil)
	// return

	if r.Method == http.MethodGet {
		// Fetch categories for the form
		rows, err := DB.Query("SELECT id, name FROM categories ORDER BY name")
		if err != nil {

			http.Error(w, "Failed to load categories", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var categories []struct {
			ID   string
			Name string
		}
		for rows.Next() {
			var cat struct {
				ID   string
				Name string
			}
			if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
				continue
			}
			categories = append(categories, cat)
		}
		// fmt.Println(categories)
		RenderTemplates(w, "posts.html", map[string]interface{}{
			"Categories": categories,
		})
		return
	}
	if r.Method != http.MethodPost {
		ErrorHandler(w, http.StatusMethodNotAllowed)
		return
	}

	// Increase the maximum memory allocated for form parsing
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Get user ID from cookie
	cookie, err := r.Cookie("Token")
	if err != nil {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	userID := cookie.Value

	// Extract form data
	post := Post{
		ID:         uuid.New().String(),
		UserID:     userID,
		Title:      r.FormValue("title"),
		Content:    r.FormValue("content"),
		Categories: r.Form["categories"],
	}

	// Handle file upload
	file, header, err := r.FormFile("media")
	if err == nil {
		defer file.Close()

		// Validate file size
		if header.Size > MaxFileSize {
			http.Error(w, "File size exceeds maximum limit", http.StatusBadRequest)
			return
		}

		// Validate file type
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if !isValidFileType(ext) {
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// Read file in chunks
		buffer := make([]byte, 0, header.Size)
		tempBuffer := make([]byte, ChunkSize)
		for {
			n, err := file.Read(tempBuffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}
			buffer = append(buffer, tempBuffer[:n]...)
		}

		post.Media = buffer
		post.ContentType = getContentType(ext)
	}

	// Insert post
	_, err = DB.Exec(`
        INSERT INTO posts (id, user_id, title, content, media, content_type)
        VALUES (?, ?, ?, ?, ?, ?)`,
		post.ID, post.UserID, post.Title, post.Content, post.Media, post.ContentType,
	)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Insert categories
	for _, categoryID := range post.Categories {
		_, err = DB.Exec(`
            INSERT INTO post_categories (post_id, category_id)
            VALUES (?, ?)`,
			post.ID, categoryID,
		)
		if err != nil {
			http.Error(w, "Failed to add category", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Post created successfully"))
}

func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	var post Post
	err := DB.QueryRow(`
        SELECT id, user_id, title, content, media, content_type
        FROM posts WHERE id = ?`,
		postID,
	).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Media, &post.ContentType)

	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If this is a media request, serve the media file
	if r.URL.Query().Get("media") == "true" && post.Media != nil {
		w.Header().Set("Content-Type", post.ContentType)
		w.Write(post.Media)
		return
	}

	// Otherwise, return the post data as JSON
	// Add your JSON response handling here
}

func isValidFileType(ext string) bool {
	validTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".mp4":  true,
		".mov":  true,
		".webm": true,
	}
	return validTypes[ext]
}

func getContentType(ext string) string {
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".mov":  "video/quicktime",
		".webm": "video/webm",
	}
	return contentTypes[ext]
}
