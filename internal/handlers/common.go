package handlers

import (
	"daily-quiz/internal/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type PageData struct {
	Title          string
	User           *models.User
	Error          string
	Success        string
	Data           interface{}
	LevelInfo      models.LevelInfo
	TimeUntilReset string
}

var templates map[string]*template.Template

func InitTemplates(templatesDir string) error {
	templates = make(map[string]*template.Template)

	funcMap := template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"sub":      func(a, b int) int { return a - b },
		"mul":      func(a, b int) int { return a * b },
		"mod":      func(a, b int) int { return a % b },
		"divFloat": func(a, b int) float64 { return float64(a) / float64(b) * 100 },
		"printf":   fmt.Sprintf,
	}

	pages := []string{
		"home.html",
		"login.html",
		"register.html",
		"quiz.html",
		"result.html",
		"profile.html",
		"leaderboard.html",
	}

	layoutPath := filepath.Join(templatesDir, "layout.html")

	for _, page := range pages {
		pagePath := filepath.Join(templatesDir, page)
		tmpl, err := template.New("").Funcs(funcMap).ParseFiles(layoutPath, pagePath)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", page, err)
		}
		templates[page] = tmpl
	}

	return nil
}

func RenderTemplate(w http.ResponseWriter, name string, data PageData) {
	tmpl, ok := templates[name]
	if !ok {
		log.Printf("Template not found: %s", name)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("Template execution error: %s: %v", name, err)
	}
}

func GetTimeUntilReset() string {
	now := time.Now().UTC()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	diff := tomorrow.Sub(now)

	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func TodayUTC() string {
	return time.Now().UTC().Format("2006-01-02")
}

func YesterdayUTC() string {
	return time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
}
