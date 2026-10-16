package handlers

import (
	"daily-quiz/internal/db"
	"daily-quiz/internal/middleware"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func ShowLogin(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := PageData{
		Title: "Login",
	}
	RenderTemplate(w, "login.html", data)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	if email == "" || password == "" {
		data := PageData{Title: "Login", Error: "Please fill in all fields"}
		RenderTemplate(w, "login.html", data)
		return
	}

	var userID int64
	var passwordHash string
	err := db.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", email).
		Scan(&userID, &passwordHash)
	if err != nil {
		data := PageData{Title: "Login", Error: "Invalid email or password"}
		RenderTemplate(w, "login.html", data)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		data := PageData{Title: "Login", Error: "Invalid email or password"}
		RenderTemplate(w, "login.html", data)
		return
	}

	token, err := middleware.CreateSession(userID)
	if err != nil {
		data := PageData{Title: "Login", Error: "Something went wrong. Please try again."}
		RenderTemplate(w, "login.html", data)
		return
	}

	middleware.SetSessionCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ShowRegister(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := PageData{Title: "Register"}
	RenderTemplate(w, "register.html", data)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if email == "" || displayName == "" || password == "" {
		data := PageData{Title: "Register", Error: "Please fill in all fields"}
		RenderTemplate(w, "register.html", data)
		return
	}

	if len(password) < 6 {
		data := PageData{Title: "Register", Error: "Password must be at least 6 characters"}
		RenderTemplate(w, "register.html", data)
		return
	}

	if password != confirmPassword {
		data := PageData{Title: "Register", Error: "Passwords do not match"}
		RenderTemplate(w, "register.html", data)
		return
	}

	if len(displayName) > 30 {
		data := PageData{Title: "Register", Error: "Display name must be 30 characters or less"}
		RenderTemplate(w, "register.html", data)
		return
	}

	var exists int
	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if exists > 0 {
		data := PageData{Title: "Register", Error: "An account with this email already exists"}
		RenderTemplate(w, "register.html", data)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		data := PageData{Title: "Register", Error: "Something went wrong. Please try again."}
		RenderTemplate(w, "register.html", data)
		return
	}

	result, err := db.DB.Exec("INSERT INTO users (email, display_name, password_hash) VALUES (?, ?, ?)",
		email, displayName, string(hash))
	if err != nil {
		data := PageData{Title: "Register", Error: "Something went wrong. Please try again."}
		RenderTemplate(w, "register.html", data)
		return
	}

	userID, _ := result.LastInsertId()
	token, err := middleware.CreateSession(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	middleware.SetSessionCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		middleware.DeleteSession(cookie.Value)
	}
	middleware.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
