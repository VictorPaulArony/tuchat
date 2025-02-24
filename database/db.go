package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// function to intialize the database queries for the social media system
func InitializeDataBase(databaseName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", databaseName)
	if err != nil {
		return nil, err
	}
	// Create database table
	query := `
	CREATE TABLE IF NOT EXISTS users(
		id TEXT PRIMARY KEY,
		useremail TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS posts(
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL, --discription of the post
		media BLOB, --binary data - video, image and GIFs
		content_type TEXT, --content type tracking(text, image, gif)
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS post_categories(
		post_id TEXT NOT NULL,
		category_id TEXT NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (category_id) REFERENCES categories(id),
		PRIMARY KEY (post_id, category_id)
	);

	CREATE TABLE IF NOT EXISTS categories (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS comments(
		id TEXT PRIMARY KEY,
		post_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		parent_id TEXT,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (parent_id) REFERENCES comments(id)
	);

	CREATE TABLE IF NOT EXISTS post_likes(
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		post_id TEXT,
		type TEXT CHECK(type IN ('like', 'dislike')),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (post_id) REFERENCES posts(id)
	);

	CREATE TABLE IF NOT EXISTS comment_likes(
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		comment_id TEXT NOT NULL,
		type TEXT CHECK(type IN('like', 'dislike')),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (comment_id) REFERENCES comments(id)
	);
	
	INSERT OR IGNORE INTO categories (id, name) VALUES
    ('tech','tech'),
    ('lifestyle', 'lifestyle'),
    ('gaming', 'gaming'),
    ('food', 'food'),
    ('travel', 'travel'),
    ('other', 'other');
	`

	if _, err := db.Exec(query); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
