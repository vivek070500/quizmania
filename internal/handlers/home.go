package handlers

import (
	"daily-quiz/internal/db"
	"daily-quiz/internal/middleware"
	"daily-quiz/internal/models"
	"net/http"
)

type HomeData struct {
	AnsweredToday  bool
	TodayCorrect   bool
	TodayPoints    int
	ActiveStreak   int
	LevelInfo      models.LevelInfo
	TotalUsers     int
	TotalAnswers   int
	TimeUntilReset string
}

func ShowHome(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	if user == nil {
		// Landing page for unauthenticated users
		var totalUsers, totalAnswers int
		db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
		db.DB.QueryRow("SELECT COUNT(*) FROM user_answers").Scan(&totalAnswers)

		homeData := HomeData{
			TotalUsers:     totalUsers,
			TotalAnswers:   totalAnswers,
			TimeUntilReset: GetTimeUntilReset(),
		}

		data := PageData{
			Title: "QuizMania — Daily Quiz Challenge",
			Data:  homeData,
		}
		RenderTemplate(w, "home.html", data)
		return
	}

	// Dashboard for authenticated users
	today := TodayUTC()

	// Refresh user data
	db.DB.QueryRow("SELECT total_points, current_streak, longest_streak FROM users WHERE id = ?", user.ID).
		Scan(&user.TotalPoints, &user.CurrentStreak, &user.LongestStreak)

	var answeredToday bool
	var todayCorrect bool
	var todayPoints int

	var isCorrect bool
	var points int
	err := db.DB.QueryRow("SELECT is_correct, points_earned FROM user_answers WHERE user_id = ? AND answer_date = ?",
		user.ID, today).Scan(&isCorrect, &points)
	if err == nil {
		answeredToday = true
		todayCorrect = isCorrect
		todayPoints = points
	}

	levelInfo := models.GetUserLevel(user.TotalPoints)

	// Active streak check
	activeStreak := user.CurrentStreak
	if !answeredToday {
		yesterday := YesterdayUTC()
		var lastCorrectDate string
		err := db.DB.QueryRow(`SELECT answer_date FROM user_answers WHERE user_id = ? AND is_correct = 1 
			ORDER BY answer_date DESC LIMIT 1`, user.ID).Scan(&lastCorrectDate)
		if err != nil || lastCorrectDate != yesterday {
			activeStreak = 0
		}
	}

	homeData := HomeData{
		AnsweredToday:  answeredToday,
		TodayCorrect:   todayCorrect,
		TodayPoints:    todayPoints,
		ActiveStreak:   activeStreak,
		LevelInfo:      levelInfo,
		TimeUntilReset: GetTimeUntilReset(),
	}

	data := PageData{
		Title:     "Dashboard",
		User:      user,
		Data:      homeData,
		LevelInfo: levelInfo,
	}
	RenderTemplate(w, "home.html", data)
}
