package handlers

import (
	"daily-quiz/internal/db"
	"daily-quiz/internal/middleware"
	"daily-quiz/internal/models"
	"net/http"
)

type LeaderboardData struct {
	Entries  []models.LeaderboardEntry
	UserRank int
}

func ShowLeaderboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	rows, err := db.DB.Query(`
		SELECT display_name, email, total_points, current_streak
		FROM users
		ORDER BY total_points DESC
		LIMIT 10
	`)
	if err != nil {
		data := PageData{Title: "Leaderboard", User: user, Error: "Failed to load leaderboard"}
		RenderTemplate(w, "leaderboard.html", data)
		return
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry models.LeaderboardEntry
		err := rows.Scan(&entry.DisplayName, &entry.Email, &entry.TotalPoints, &entry.Streak)
		if err != nil {
			continue
		}
		entry.Rank = rank
		levelInfo := models.GetUserLevel(entry.TotalPoints)
		entry.Level = levelInfo.Level
		entry.LevelTitle = levelInfo.Title
		entries = append(entries, entry)
		rank++
	}

	// Get user's rank if logged in
	userRank := 0
	if user != nil {
		db.DB.QueryRow(`
			SELECT COUNT(*) + 1 FROM users WHERE total_points > (
				SELECT total_points FROM users WHERE id = ?
			)
		`, user.ID).Scan(&userRank)
	}

	leaderboardData := LeaderboardData{
		Entries:  entries,
		UserRank: userRank,
	}

	data := PageData{
		Title: "Leaderboard",
		User:  user,
		Data:  leaderboardData,
	}
	RenderTemplate(w, "leaderboard.html", data)
}
