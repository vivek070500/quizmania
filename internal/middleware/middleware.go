package middleware

import (
	"context"
	"crypto/rand"
	"daily-quiz/internal/db"
	"daily-quiz/internal/models"
	"encoding/hex"
	"net/http"
	"time"
)

type contextKey string

const UserKey contextKey = "user"

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateSession(userID int64) (string, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	_, err = db.DB.Exec("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func DeleteSession(token string) error {
	_, err := db.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func GetUserFromSession(token string) (*models.User, error) {
	var user models.User
	err := db.DB.QueryRow(`
		SELECT u.id, u.email, u.display_name, u.password_hash, u.total_points, u.current_streak, u.longest_streak, u.created_at
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > CURRENT_TIMESTAMP
	`, token).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash,
		&user.TotalPoints, &user.CurrentStreak, &user.LongestStreak, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// AuthMiddleware loads user into context if session exists
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie.Value != "" {
			user, err := GetUserFromSession(cookie.Value)
			if err == nil {
				ctx := context.WithValue(r.Context(), UserKey, user)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth redirects to login if not authenticated
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func GetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(UserKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// CleanupExpiredSessions removes old sessions periodically
func CleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			db.DB.Exec("DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP")
		}
	}()
}
