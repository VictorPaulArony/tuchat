package main

import (
	"log"
	"net/http"

	"social-media/database"
	handlers "social-media/handlers"
)

func main() {
	db, err := database.InitializeDataBase("database.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	handlers.DB = db

	handlers.InitTemplates("templates")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", handlers.HomePageHandler)
	http.HandleFunc("/register", handlers.RegisterUserHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/create", handlers.CreatePostHandler)
	http.HandleFunc("/comments", handlers.CreateCommentHandler)
	http.HandleFunc("/likes", handlers.LikeHandler)
	http.HandleFunc("/posts", handlers.RenderPostsPage)

	log.Println("Server started at port: http://localhost:1234")
	log.Fatal(http.ListenAndServe(":1234", nil))
}
