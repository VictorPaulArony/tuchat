package handlers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var DB *sql.DB

// Display posts
func RenderPostsPage(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        // First, get all categories for the dropdown
        var categories []struct {
            ID   string
            Name string
        }
        categoryRows, err := DB.Query("SELECT id, name FROM categories ORDER BY name")
        if err != nil {
            http.Error(w, "Failed to load categories", http.StatusInternalServerError)
            return
        }
        defer categoryRows.Close()

        for categoryRows.Next() {
            var cat struct {
                ID   string
                Name string
            }
            if err := categoryRows.Scan(&cat.ID, &cat.Name); err != nil {
                continue
            }
            categories = append(categories, cat)
        }

        // Get all posts with their details and username
        rows, err := DB.Query(`
            SELECT p.id, u.username, p.title, p.content, p.media, p.content_type, p.created_at 
            FROM posts p 
            JOIN users u ON p.user_id = u.id
        `)
        if err != nil {
            fmt.Println(err)
            http.Error(w, "Failed to load posts", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var posts []map[string]interface{}
        for rows.Next() {
            var id, username, title, content, contentType string
            var createdAt time.Time
            var media []byte

            if err := rows.Scan(&id, &username, &title, &content, &media, &contentType, &createdAt); err != nil {
                http.Error(w, "Failed to parse posts", http.StatusInternalServerError)
                fmt.Println(err)
                return
            }

            // Convert media to base64 if it exists
            var mediaBase64 string
            if len(media) > 0 {
                mediaBase64 = base64.StdEncoding.EncodeToString(media)
            }

            // Fetch categories for this post
            categoryRows, err := DB.Query(`
                SELECT c.id, c.name 
                FROM categories c 
                JOIN post_categories pc ON c.id = pc.category_id 
                WHERE pc.post_id = ?`, id)
            if err != nil {
                http.Error(w, "Failed to fetch post categories", http.StatusInternalServerError)
                return
            }
            defer categoryRows.Close()

            var postCategories []map[string]string
            for categoryRows.Next() {
                var catID, catName string
                if err := categoryRows.Scan(&catID, &catName); err != nil {
                    continue
                }
                postCategories = append(postCategories, map[string]string{
                    "ID":   catID,
                    "Name": catName,
                })
            }

            // Fetch likes and dislikes
            var likes, dislikes int
            err = DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND type = 'like'", id).Scan(&likes)
            if err != nil {
                http.Error(w, "Failed to fetch likes", http.StatusInternalServerError)
                return
            }

            err = DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND type = 'dislike'", id).Scan(&dislikes)
            if err != nil {
                http.Error(w, "Failed to fetch dislikes", http.StatusInternalServerError)
                return
            }

            // Fetch comments
            commentRows, err := DB.Query(`
                SELECT c.id, c.content, c.created_at,
                (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND type = 'like') as likes,
                (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND type = 'dislike') as dislikes
                FROM comments c
                WHERE c.post_id = ?
                ORDER BY c.created_at DESC`, id)
            if err != nil {
                http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
                return
            }
            defer commentRows.Close()

            var comments []map[string]interface{}
            for commentRows.Next() {
                var commentID string
                var commentContent string
                var commentCreatedAt time.Time
                var commentLikes, commentDislikes int

                err := commentRows.Scan(&commentID, &commentContent, &commentCreatedAt, &commentLikes, &commentDislikes)
                if err != nil {
                    http.Error(w, "Failed to parse comment", http.StatusInternalServerError)
                    return
                }

                comments = append(comments, map[string]interface{}{
                    "ID":        commentID,
                    "Content":   commentContent,
                    "CreatedAt": commentCreatedAt,
                    "Likes":     commentLikes,
                    "Dislikes":  commentDislikes,
                })
            }

            posts = append(posts, map[string]interface{}{
                "ID":         id,
                "Title":      title,
                "Content":    content,
                "Likes":      likes,
                "Dislikes":   dislikes,
                "Comments":   comments,
                "Username":   username,  // Now using username instead of user_id
                "Categories": postCategories,
                "Media":      mediaBase64,
                "ContentType": contentType,
                "CreatedAt":  createdAt,
            })
        }

        data := map[string]interface{}{
            "Posts":      posts,
            "Categories": categories,
        }

        // RenderTemplates(w, "posts.html", data)
		RenderTemplates(w, "base.html", data)
    }
}
// function to render the html templates pages
func RenderTemplates(w http.ResponseWriter, fileName string, data interface{}) {
	if err := templates.ExecuteTemplate(w, fileName, data); err != nil {
		ErrorHandler(w, http.StatusInternalServerError)
		log.Println("Templates failed to execute:", err)
		return
	}
}

// function to handle the get method  of the home page
func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorHandler(w, http.StatusNotFound)
	}
	if r.Method == http.MethodGet {
		RenderTemplates(w, "base.html", nil)
		return
	}

	// fmt.Println(posts)

	RenderTemplates(w, "base.html", nil)
}

// function to handle the Errors in the system
func ErrorHandler(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	temp, err := template.ParseFiles("templates/error.html")
	if err != nil {
		log.Println("Error while parsing the error page:", err)
		http.Error(w, "Page temporarily down", http.StatusInternalServerError)
		return
	}

	if err := temp.Execute(w, map[string]int{"Code": code}); err != nil {
		log.Println("Error while executing the error template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

var templates *template.Template

// Initialize templates (call this function in your main.go or init function)
// func InitTemplates(templateDir string) {
// 	templates = template.Must(template.ParseGlob(templateDir + "/*.html"))
// }

// InitTemplates initializes the templates with necessary functions
func InitTemplates(templateDir string) {
	// Create a FuncMap with our template functions
	funcMap := template.FuncMap{
		// Convert []byte to base64 string
		"base64": func(b []byte) string {
			return base64.StdEncoding.EncodeToString(b)
		},
		// Format time
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 02, 2006 at 15:04")
		},
		// Add any other helper functions you need here
	}

	// Create a new template and register the FuncMap
	var err error
	templates = template.New("")

	// Register the function map
	templates = templates.Funcs(funcMap)

	// Parse all templates in the directory
	templates, err = templates.ParseGlob(templateDir + "/*.html")
	if err != nil {
		panic(err)
	}
}
