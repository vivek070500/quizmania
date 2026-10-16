package main

import (
	"daily-quiz/internal/db"
	"daily-quiz/internal/handlers"
	"daily-quiz/internal/middleware"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "quizmania.db"
	}

	// Initialize database
	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("✅ Database initialized successfully")

	// Seed questions
	if err := db.SeedQuestions("data/questions.json"); err != nil {
		log.Printf("Warning: Failed to seed questions: %v", err)
	}

	// Initialize templates
	if err := handlers.InitTemplates("templates"); err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}
	log.Println("✅ Templates loaded successfully")

	// Start session cleanup goroutine
	middleware.CleanupExpiredSessions()

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public routes
	mux.HandleFunc("GET /", handlers.ShowHome)
	mux.HandleFunc("GET /login", handlers.ShowLogin)
	mux.HandleFunc("POST /login", handlers.HandleLogin)
	mux.HandleFunc("GET /register", handlers.ShowRegister)
	mux.HandleFunc("POST /register", handlers.HandleRegister)
	mux.HandleFunc("GET /logout", handlers.HandleLogout)
	mux.HandleFunc("GET /leaderboard", handlers.ShowLeaderboard)

	// Protected routes
	mux.HandleFunc("GET /quiz", middleware.RequireAuth(handlers.ShowQuiz))
	mux.HandleFunc("POST /quiz/answer", middleware.RequireAuth(handlers.HandleAnswer))
	mux.HandleFunc("GET /result", middleware.RequireAuth(handlers.ShowResult))
	mux.HandleFunc("GET /profile", middleware.RequireAuth(handlers.ShowProfile))

	// Wrap all routes with auth middleware (loads user into context)
	handler := middleware.AuthMiddleware(mux)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("🎯 QuizMania server starting on http://localhost%s", addr)
	log.Println("✅ Ready! Open your browser and enjoy 🚀")
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
