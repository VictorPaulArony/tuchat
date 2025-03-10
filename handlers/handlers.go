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
func RenderPostsPage() interface{} {
	// if r.Method == http.MethodGet {
	// First, get all categories for the dropdown
	var categories []struct {
		ID   string
		Name string
	}
	categoryRows, err := DB.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		log.Println("Failed to load categories", http.StatusInternalServerError)
		return nil
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
			ORDER BY p.created_at DESC 
        `)
	if err != nil {
		fmt.Println(err)
		log.Println("Failed to load posts", http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var id, username, title, content, contentType string
		var createdAt time.Time
		var media []byte

		if err := rows.Scan(&id, &username, &title, &content, &media, &contentType, &createdAt); err != nil {
			log.Println("Failed to parse posts", http.StatusInternalServerError)
			fmt.Println(err)
			return nil
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
			log.Println("Failed to fetch post categories", http.StatusInternalServerError)
			return nil
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
			log.Println("Failed to fetch likes", http.StatusInternalServerError)
			return nil
		}

		err = DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND type = 'dislike'", id).Scan(&dislikes)
		if err != nil {
			log.Println("Failed to fetch dislikes", http.StatusInternalServerError)
			return nil
		}

		// Fetch top-level comments with nested comments
		comments, err := fetchNestedComments(id)
		if err != nil {
			fmt.Println("err", err)
			log.Println("Failed to fetch comments", http.StatusInternalServerError)
			return nil
		}

		posts = append(posts, map[string]interface{}{
			"ID":          id,
			"Title":       title,
			"Content":     content,
			"Likes":       likes,
			"Dislikes":    dislikes,
			"Comments":    comments,
			"Username":    username,
			"Categories":  postCategories,
			"Media":       mediaBase64,
			"ContentType": contentType,
			"CreatedAt":   createdAt,
		})
	}

	data := map[string]interface{}{
		"Posts":      posts,
		"Categories": categories,
	}

	return data
	// }
}

// fetchNestedComments retrieves top-level comments and their nested replies
func fetchNestedComments(postID string) ([]map[string]interface{}, error) {
	// Recursive function to fetch nested comments with unlimited depth
	var fetchCommentReplies func(parentCommentID string) ([]map[string]interface{}, error)
	fetchCommentReplies = func(parentCommentID string) ([]map[string]interface{}, error) {
		// Fetch comments/replies with their details
		commentRows, err := DB.Query(`
            SELECT c.id, c.content, c.created_at,
            (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND type = 'like') as likes,
            (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND type = 'dislike') as dislikes,
            u.username
            FROM comments c
            JOIN users u ON c.user_id = u.id
            WHERE c.parent_id = ?
            ORDER BY c.created_at ASC`, parentCommentID)
		if err != nil {
			return nil, err
		}
		defer commentRows.Close()

		var comments []map[string]interface{}
		for commentRows.Next() {
			var commentID, commentContent, username string
			var commentCreatedAt time.Time
			var commentLikes, commentDislikes int

			err := commentRows.Scan(&commentID, &commentContent, &commentCreatedAt, &commentLikes, &commentDislikes, &username)
			if err != nil {
				return nil, err
			}

			// Recursively fetch nested replies for this comment
			nestedReplies, err := fetchCommentReplies(commentID)
			if err != nil {
				return nil, err
			}

			comments = append(comments, map[string]interface{}{
				"ID":        commentID,
				"Content":   commentContent,
				"CreatedAt": commentCreatedAt,
				"Likes":     commentLikes,
				"Dislikes":  commentDislikes,
				"Username":  username,
				"Replies":   nestedReplies,
			})
		}

		return comments, nil
	}

	// Fetch top-level comments
	commentRows, err := DB.Query(`
        SELECT c.id, c.content, c.created_at,
        (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND type = 'like') as likes,
        (SELECT COUNT(*) FROM comment_likes WHERE comment_id = c.id AND type = 'dislike') as dislikes,
        u.username
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ? AND c.parent_id IS NULL
        ORDER BY c.created_at DESC`, postID)
	if err != nil {
		return nil, err
	}
	defer commentRows.Close()

	var comments []map[string]interface{}
	for commentRows.Next() {
		var commentID, commentContent, username string
		var commentCreatedAt time.Time
		var commentLikes, commentDislikes int

		err := commentRows.Scan(&commentID, &commentContent, &commentCreatedAt, &commentLikes, &commentDislikes, &username)
		if err != nil {
			return nil, err
		}

		// Fetch nested replies for this comment using the recursive function
		nestedReplies, err := fetchCommentReplies(commentID)
		if err != nil {
			return nil, err
		}

		comments = append(comments, map[string]interface{}{
			"ID":        commentID,
			"Content":   commentContent,
			"CreatedAt": commentCreatedAt,
			"Likes":     commentLikes,
			"Dislikes":  commentDislikes,
			"Username":  username,
			"Replies":   nestedReplies,
		})
	}

	return comments, nil
}

// function to handle the filters
func FilterPost(categories []string) interface{} {
	var posts []map[string]interface{}
	for _, category := range categories {

		var id string
		query := `
	  		SELECT  pc.post_id
	  		FROM post_categories pc
	  		WHERE pc.category_id = ?
		`
		err := DB.QueryRow(query, category).Scan(&id)
		if err != nil {
			fmt.Println(err)
			log.Println("Failed to load post id", http.StatusInternalServerError)
			return nil
		}

		// Get all posts with their details and username
		post, err := DB.Query(`
            SELECT p.id, u.username, p.title, p.content, p.media, p.content_type, p.created_at 
            FROM posts p 
            JOIN users u ON p.user_id = u.id 
			WHERE p.id = ?
        `, id)
		if err != nil {
			fmt.Println(err)
			log.Println("Failed to load posts", http.StatusInternalServerError)
			return nil
		}
		defer post.Close()

		for post.Next() {
			var id, username, title, content, contentType string
			var createdAt time.Time
			var media []byte

			if err := post.Scan(&id, &username, &title, &content, &media, &contentType, &createdAt); err != nil {
				log.Println("Failed to parse posts", http.StatusInternalServerError)
				fmt.Println(err)
				return nil
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
				log.Println("Failed to fetch post categories", http.StatusInternalServerError)
				return nil
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
				log.Println("Failed to fetch likes", http.StatusInternalServerError)
				return nil
			}

			err = DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND type = 'dislike'", id).Scan(&dislikes)
			if err != nil {
				log.Println("Failed to fetch dislikes", http.StatusInternalServerError)
				return nil
			}

			// Fetch top-level comments with nested comments
			comments, err := fetchNestedComments(id)
			if err != nil {
				fmt.Println("err", err)
				log.Println("Failed to fetch comments", http.StatusInternalServerError)
				return nil
			}

			posts = append(posts, map[string]interface{}{
				"ID":          id,
				"Title":       title,
				"Content":     content,
				"Likes":       likes,
				"Dislikes":    dislikes,
				"Comments":    comments,
				"Username":    username,
				"Categories":  postCategories,
				"Media":       mediaBase64,
				"ContentType": contentType,
				"CreatedAt":   createdAt,
			})
		}
	}
	data := map[string]interface{}{
		"Posts": posts,
	}
	return data
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
	// if r.URL.Path != "/" {
	// 	ErrorHandler(w, http.StatusNotFound)
	// }
	if r.Method == http.MethodGet {
		data := RenderPostsPage()
		RenderTemplates(w, "posts.html", data)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
			return
		}

		categories := r.Form["categories"]
		data := FilterPost(categories)

		RenderTemplates(w, "posts.html", data)
	}
	// fmt.Println(posts)
}

// function to handle the Errors in the system
func ErrorHandler(w http.ResponseWriter, code int) {
	// w.WriteHeader(code)
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
	// funcMap := template.FuncMap{
	// 	// Convert []byte to base64 string
	// 	"base64": func(b []byte) string {
	// 		return base64.StdEncoding.EncodeToString(b)
	// 	},
	// 	// Format time
	// 	"formatDate": func(t time.Time) string {
	// 		return t.Format("Jan 02, 2006 at 15:04")
	// 	},
	// 	// Add any other helper functions you need here
	// }
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil
				}
				dict[key] = values[i+1]
			}
			return dict
		},
	}

	var err error
	// tmpl := template.New("posts.html").Funcs(funcMap)
	// tmpl, err = tmpl.ParseFiles("posts.html")

	// Create a new template and register the FuncMap
	templates = template.New("")

	// Register the function map
	templates = templates.Funcs(funcMap)

	// Parse all templates in the directory
	templates, err = templates.ParseGlob(templateDir + "/*.html")
	if err != nil {
		panic(err)
	}
}
