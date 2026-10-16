package handlers

import (
	"daily-quiz/internal/db"
	"daily-quiz/internal/middleware"
	"daily-quiz/internal/models"
	"net/http"
)

type ProfileData struct {
	TotalAnswered  int
	TotalCorrect   int
	Accuracy       float64
	CurrentStreak  int
	LongestStreak  int
	TotalPoints    int
	LevelInfo      models.LevelInfo
	EarnedBadges   []EarnedBadge
	AllBadges      []BadgeDisplay
	RecentAnswers  []RecentAnswer
	DaysActive     int
}

type EarnedBadge struct {
	models.Badge
	EarnedAt string
}

type BadgeDisplay struct {
	models.Badge
	Earned   bool
	EarnedAt string
}

type RecentAnswer struct {
	Date         string
	QuestionText string
	IsCorrect    bool
	PointsEarned int
	Category     string
}

func ShowProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	// Refresh user data
	db.DB.QueryRow("SELECT total_points, current_streak, longest_streak FROM users WHERE id = ?", user.ID).
		Scan(&user.TotalPoints, &user.CurrentStreak, &user.LongestStreak)

	// Get stats
	var totalAnswered, totalCorrect int
	db.DB.QueryRow("SELECT COUNT(*) FROM user_answers WHERE user_id = ?", user.ID).Scan(&totalAnswered)
	db.DB.QueryRow("SELECT COUNT(*) FROM user_answers WHERE user_id = ? AND is_correct = 1", user.ID).Scan(&totalCorrect)

	accuracy := 0.0
	if totalAnswered > 0 {
		accuracy = float64(totalCorrect) / float64(totalAnswered) * 100
	}

	// Get days active
	var daysActive int
	db.DB.QueryRow("SELECT COUNT(DISTINCT answer_date) FROM user_answers WHERE user_id = ?", user.ID).Scan(&daysActive)

	// Get earned badges
	earnedMap := make(map[string]string) // key -> earned_at
	rows, _ := db.DB.Query("SELECT badge_key, earned_at FROM user_badges WHERE user_id = ?", user.ID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var key, earnedAt string
			rows.Scan(&key, &earnedAt)
			earnedMap[key] = earnedAt
		}
	}

	var allBadges []BadgeDisplay
	for _, b := range models.AllBadges {
		bd := BadgeDisplay{Badge: b}
		if ea, ok := earnedMap[b.Key]; ok {
			bd.Earned = true
			bd.EarnedAt = ea[:10] // just the date part
		}
		allBadges = append(allBadges, bd)
	}

	// Get recent answers (last 10)
	var recentAnswers []RecentAnswer
	answerRows, _ := db.DB.Query(`
		SELECT ua.answer_date, q.text, ua.is_correct, ua.points_earned, q.category
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		WHERE ua.user_id = ?
		ORDER BY ua.answer_date DESC
		LIMIT 10
	`, user.ID)
	if answerRows != nil {
		defer answerRows.Close()
		for answerRows.Next() {
			var ra RecentAnswer
			answerRows.Scan(&ra.Date, &ra.QuestionText, &ra.IsCorrect, &ra.PointsEarned, &ra.Category)
			recentAnswers = append(recentAnswers, ra)
		}
	}

	levelInfo := models.GetUserLevel(user.TotalPoints)

	// Check active streak
	activeStreak := user.CurrentStreak
	// Verify streak is still valid
	yesterday := YesterdayUTC()
	today := TodayUTC()
	var lastCorrectDate string
	err := db.DB.QueryRow(`SELECT answer_date FROM user_answers WHERE user_id = ? AND is_correct = 1 
		ORDER BY answer_date DESC LIMIT 1`, user.ID).Scan(&lastCorrectDate)
	if err != nil || (lastCorrectDate != yesterday && lastCorrectDate != today) {
		activeStreak = 0
	}

	profileData := ProfileData{
		TotalAnswered: totalAnswered,
		TotalCorrect:  totalCorrect,
		Accuracy:      accuracy,
		CurrentStreak: activeStreak,
		LongestStreak: user.LongestStreak,
		TotalPoints:   user.TotalPoints,
		LevelInfo:     levelInfo,
		AllBadges:     allBadges,
		RecentAnswers: recentAnswers,
		DaysActive:    daysActive,
	}

	data := PageData{
		Title:     "Profile",
		User:      user,
		Data:      profileData,
		LevelInfo: levelInfo,
	}
	RenderTemplate(w, "profile.html", data)
}
