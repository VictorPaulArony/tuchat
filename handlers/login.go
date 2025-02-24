package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// function to handle the login of the users in the system
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// handle the get request to display the login page to the user
	if r.Method == http.MethodGet {
		RenderTemplates(w, "login.html", nil)
		return
	}

	// handle the login page for the post request from the user
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			ErrorHandler(w, http.StatusBadRequest)
			return
		}

		// get the user input from the login form
		userName := r.FormValue("userName")
		password := r.FormValue("password")

		var hashedPassword string
		var userId string

		// retrieving the user data from the database for confirmation
		err := DB.QueryRow("SELECT id, password FROM users WHERE username = ?", userName).Scan(&userId, &hashedPassword)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, "Error retrieving user", http.StatusInternalServerError)
			return
		}

		// check if the password entered is correct
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid user password", http.StatusUnauthorized)
			return
		}

		// generate a session token for the logged in user
		// session := uuid.New().String()
		session := userId

		http.SetCookie(w, &http.Cookie{
			Name:     "Token",
			Value:    session,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		w.Write([]byte("Login successful"))
		// http.Redirect(w, r, "/posts/create", http.StatusSeeOther)
	}
}
