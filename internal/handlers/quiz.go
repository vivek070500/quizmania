package handlers

import (
	"daily-quiz/internal/db"
	"daily-quiz/internal/middleware"
	"daily-quiz/internal/models"
	"database/sql"
	"log"
	"net/http"
)

type QuizData struct {
	Question   models.Question
	Difficulty string
	Streak     int
}

type ResultData struct {
	Question      models.Question
	Selected      string
	IsCorrect     bool
	PointsEarned  int
	NewStreak     int
	TotalPoints   int
	LevelInfo     models.LevelInfo
	NewBadges     []models.Badge
	StreakBroken   bool
	PreviousLevel int
	LeveledUp     bool
}

func ShowQuiz(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	today := TodayUTC()

	// Check if already answered today
	var answeredCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM user_answers WHERE user_id = ? AND answer_date = ?",
		user.ID, today).Scan(&answeredCount)
	if answeredCount > 0 {
		http.Redirect(w, r, "/result", http.StatusSeeOther)
		return
	}

	// Refresh user data
	db.DB.QueryRow("SELECT total_points, current_streak FROM users WHERE id = ?", user.ID).
		Scan(&user.TotalPoints, &user.CurrentStreak)

	// Check if streak is still active (answered correctly yesterday)
	activeStreak := getActiveStreak(user.ID)

	// Get or assign today's question
	question, err := getOrAssignQuestion(user.ID, today, activeStreak)
	if err != nil {
		log.Printf("Error getting question: %v", err)
		data := PageData{Title: "Quiz", User: user, Error: "No questions available. Please try again later."}
		RenderTemplate(w, "quiz.html", data)
		return
	}

	difficulty := models.GetDifficultyForStreak(activeStreak)
	quizData := QuizData{
		Question:   *question,
		Difficulty: models.GetDifficultyLabel(difficulty),
		Streak:     activeStreak,
	}

	data := PageData{
		Title:          "Daily Quiz",
		User:           user,
		Data:           quizData,
		TimeUntilReset: GetTimeUntilReset(),
	}
	RenderTemplate(w, "quiz.html", data)
}

func HandleAnswer(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	today := TodayUTC()

	// Prevent double submission
	var answeredCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM user_answers WHERE user_id = ? AND answer_date = ?",
		user.ID, today).Scan(&answeredCount)
	if answeredCount > 0 {
		http.Redirect(w, r, "/result", http.StatusSeeOther)
		return
	}

	selectedOption := r.FormValue("answer")
	if selectedOption == "" {
		http.Redirect(w, r, "/quiz", http.StatusSeeOther)
		return
	}

	// Get today's assigned question
	var questionID int64
	err := db.DB.QueryRow("SELECT question_id FROM daily_assignments WHERE user_id = ? AND quiz_date = ?",
		user.ID, today).Scan(&questionID)
	if err != nil {
		http.Redirect(w, r, "/quiz", http.StatusSeeOther)
		return
	}

	// Load the question
	var q models.Question
	err = db.DB.QueryRow(`SELECT id, text, option_a, option_b, option_c, option_d, correct_option, difficulty, category
		FROM questions WHERE id = ?`, questionID).
		Scan(&q.ID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectOption, &q.Difficulty, &q.Category)
	if err != nil {
		http.Redirect(w, r, "/quiz", http.StatusSeeOther)
		return
	}

	isCorrect := selectedOption == q.CorrectOption
	activeStreak := getActiveStreak(user.ID)

	var pointsEarned int
	var newStreak int
	streakBroken := false

	if isCorrect {
		newStreak = activeStreak + 1
		pointsEarned = newStreak
	} else {
		newStreak = 0
		pointsEarned = 0
		if activeStreak > 0 {
			streakBroken = true
		}
	}

	// Refresh user data
	db.DB.QueryRow("SELECT total_points, current_streak, longest_streak FROM users WHERE id = ?", user.ID).
		Scan(&user.TotalPoints, &user.CurrentStreak, &user.LongestStreak)

	previousLevel := models.GetUserLevel(user.TotalPoints).Level
	newTotalPoints := user.TotalPoints + pointsEarned

	longestStreak := user.LongestStreak
	if newStreak > longestStreak {
		longestStreak = newStreak
	}

	// Save the answer
	_, err = db.DB.Exec(`INSERT INTO user_answers (user_id, question_id, answer_date, selected_option, is_correct, points_earned)
		VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, questionID, today, selectedOption, isCorrect, pointsEarned)
	if err != nil {
		log.Printf("Error saving answer: %v", err)
		http.Redirect(w, r, "/quiz", http.StatusSeeOther)
		return
	}

	// Update user stats
	_, err = db.DB.Exec(`UPDATE users SET total_points = ?, current_streak = ?, longest_streak = ? WHERE id = ?`,
		newTotalPoints, newStreak, longestStreak, user.ID)
	if err != nil {
		log.Printf("Error updating user stats: %v", err)
	}

	// Check and award badges
	checkAndAwardBadges(user.ID, newStreak, newTotalPoints, q.Difficulty, isCorrect)

	_ = streakBroken
	_ = previousLevel

	http.Redirect(w, r, "/result", http.StatusSeeOther)
}

func ShowResult(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	today := TodayUTC()

	var answer struct {
		questionID     int64
		selectedOption string
		isCorrect      bool
		pointsEarned   int
	}

	err := db.DB.QueryRow(`SELECT question_id, selected_option, is_correct, points_earned 
		FROM user_answers WHERE user_id = ? AND answer_date = ?`,
		user.ID, today).Scan(&answer.questionID, &answer.selectedOption, &answer.isCorrect, &answer.pointsEarned)
	if err != nil {
		http.Redirect(w, r, "/quiz", http.StatusSeeOther)
		return
	}

	var q models.Question
	db.DB.QueryRow(`SELECT id, text, option_a, option_b, option_c, option_d, correct_option, difficulty, category
		FROM questions WHERE id = ?`, answer.questionID).
		Scan(&q.ID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectOption, &q.Difficulty, &q.Category)

	// Refresh user data
	db.DB.QueryRow("SELECT total_points, current_streak, longest_streak FROM users WHERE id = ?", user.ID).
		Scan(&user.TotalPoints, &user.CurrentStreak, &user.LongestStreak)

	levelInfo := models.GetUserLevel(user.TotalPoints)

	resultData := ResultData{
		Question:     q,
		Selected:     answer.selectedOption,
		IsCorrect:    answer.isCorrect,
		PointsEarned: answer.pointsEarned,
		NewStreak:    user.CurrentStreak,
		TotalPoints:  user.TotalPoints,
		LevelInfo:    levelInfo,
	}

	data := PageData{
		Title:          "Result",
		User:           user,
		Data:           resultData,
		LevelInfo:      levelInfo,
		TimeUntilReset: GetTimeUntilReset(),
	}
	RenderTemplate(w, "result.html", data)
}

func getActiveStreak(userID int64) int {
	yesterday := YesterdayUTC()

	// Check if user answered correctly yesterday
	var yesterdayCorrect bool
	err := db.DB.QueryRow(`SELECT is_correct FROM user_answers WHERE user_id = ? AND answer_date = ?`,
		userID, yesterday).Scan(&yesterdayCorrect)

	if err == sql.ErrNoRows {
		// Didn't answer yesterday — check if this is a brand new user or streak is broken
		// Also check: maybe user hasn't answered before at all (new user, streak = 0)
		var currentStreak int
		db.DB.QueryRow("SELECT current_streak FROM users WHERE id = ?", userID).Scan(&currentStreak)

		// If they had a streak but didn't answer yesterday, streak is broken
		if currentStreak > 0 {
			// Reset streak in DB
			db.DB.Exec("UPDATE users SET current_streak = 0 WHERE id = ?", userID)
		}
		return 0
	}

	if err != nil || !yesterdayCorrect {
		return 0
	}

	// Yesterday was correct, return current streak from DB
	var streak int
	db.DB.QueryRow("SELECT current_streak FROM users WHERE id = ?", userID).Scan(&streak)
	return streak
}

func getOrAssignQuestion(userID int64, date string, streak int) (*models.Question, error) {
	// Check if question already assigned
	var questionID int64
	err := db.DB.QueryRow("SELECT question_id FROM daily_assignments WHERE user_id = ? AND quiz_date = ?",
		userID, date).Scan(&questionID)

	if err == nil {
		// Already assigned, load and return it
		var q models.Question
		err = db.DB.QueryRow(`SELECT id, text, option_a, option_b, option_c, option_d, correct_option, difficulty, category
			FROM questions WHERE id = ?`, questionID).
			Scan(&q.ID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectOption, &q.Difficulty, &q.Category)
		if err != nil {
			return nil, err
		}
		return &q, nil
	}

	// Need to assign a new question
	difficulty := models.GetDifficultyForStreak(streak)

	// Pick a random question from the appropriate difficulty that the user hasn't seen
	var q models.Question
	err = db.DB.QueryRow(`SELECT id, text, option_a, option_b, option_c, option_d, correct_option, difficulty, category
		FROM questions
		WHERE difficulty = ? AND id NOT IN (
			SELECT question_id FROM daily_assignments WHERE user_id = ?
		)
		ORDER BY RANDOM()
		LIMIT 1`, difficulty, userID).
		Scan(&q.ID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectOption, &q.Difficulty, &q.Category)

	if err != nil {
		// If no unseen questions at this difficulty, try any difficulty
		err = db.DB.QueryRow(`SELECT id, text, option_a, option_b, option_c, option_d, correct_option, difficulty, category
			FROM questions
			WHERE id NOT IN (
				SELECT question_id FROM daily_assignments WHERE user_id = ?
			)
			ORDER BY ABS(difficulty - ?) ASC, RANDOM()
			LIMIT 1`, userID, difficulty).
			Scan(&q.ID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectOption, &q.Difficulty, &q.Category)
		if err != nil {
			// All questions exhausted, recycle: pick any random question
			err = db.DB.QueryRow(`SELECT id, text, option_a, option_b, option_c, option_d, correct_option, difficulty, category
				FROM questions ORDER BY RANDOM() LIMIT 1`).
				Scan(&q.ID, &q.Text, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectOption, &q.Difficulty, &q.Category)
			if err != nil {
				return nil, err
			}
		}
	}

	// Save the assignment
	_, err = db.DB.Exec("INSERT INTO daily_assignments (user_id, quiz_date, question_id) VALUES (?, ?, ?)",
		userID, date, q.ID)
	if err != nil {
		return nil, err
	}

	return &q, nil
}

func checkAndAwardBadges(userID int64, streak int, totalPoints int, questionDifficulty int, isCorrect bool) []models.Badge {
	var newBadges []models.Badge

	// Helper to award badge if not already earned
	award := func(key string) {
		var exists int
		db.DB.QueryRow("SELECT COUNT(*) FROM user_badges WHERE user_id = ? AND badge_key = ?",
			userID, key).Scan(&exists)
		if exists == 0 {
			db.DB.Exec("INSERT INTO user_badges (user_id, badge_key) VALUES (?, ?)", userID, key)
			for _, b := range models.AllBadges {
				if b.Key == key {
					newBadges = append(newBadges, b)
					break
				}
			}
		}
	}

	// First answer badge
	award("first_step")

	if isCorrect {
		if streak >= 5 {
			award("getting_started")
		}
		if streak >= 10 {
			award("on_fire")
		}
		if streak >= 25 {
			award("unstoppable")
		}
		if streak >= 50 {
			award("legendary")
		}
		if streak >= 7 {
			award("perfect_week")
		}
		if streak >= 30 {
			award("perfect_month")
		}
		if questionDifficulty >= 3 {
			award("sharp_mind")
		}
		if questionDifficulty >= 5 {
			award("genius")
		}
	}

	if totalPoints >= 100 {
		award("century")
	}
	if totalPoints >= 500 {
		award("high_roller")
	}
	if totalPoints >= 1000 {
		award("thousandaire")
	}

	return newBadges
}
