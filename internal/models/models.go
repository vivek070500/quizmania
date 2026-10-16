package models

import "time"

type User struct {
	ID            int64
	Email         string
	DisplayName   string
	PasswordHash  string
	TotalPoints   int
	CurrentStreak int
	LongestStreak int
	CreatedAt     time.Time
}

type Question struct {
	ID            int64
	Text          string
	OptionA       string
	OptionB       string
	OptionC       string
	OptionD       string
	CorrectOption string // "A", "B", "C", or "D"
	Difficulty    int    // 1-5
	Category      string
}

type UserAnswer struct {
	ID             int64
	UserID         int64
	QuestionID     int64
	AnswerDate     string // "YYYY-MM-DD"
	SelectedOption string
	IsCorrect      bool
	PointsEarned   int
	CreatedAt      time.Time
}

type DailyAssignment struct {
	UserID     int64
	QuizDate   string
	QuestionID int64
}

type Badge struct {
	Key         string
	Name        string
	Description string
	Icon        string
}

type UserBadge struct {
	BadgeKey string
	EarnedAt time.Time
}

type LeaderboardEntry struct {
	Rank        int
	DisplayName string
	Email       string
	TotalPoints int
	Level       int
	LevelTitle  string
	Streak      int
}

// Level info
type LevelInfo struct {
	Level         int
	Title         string
	Reward        string
	PointsNeeded  int
	NextLevel     int
	NextLevelPts  int
	Progress      float64 // 0-100 percentage
	CurrentPoints int
}

// All defined badges
var AllBadges = []Badge{
	{Key: "first_step", Name: "First Step", Description: "Answer your first question", Icon: "👣"},
	{Key: "getting_started", Name: "Getting Started", Description: "5 consecutive correct answers", Icon: "🌱"},
	{Key: "on_fire", Name: "On Fire", Description: "10 consecutive correct answers", Icon: "🔥"},
	{Key: "unstoppable", Name: "Unstoppable", Description: "25 consecutive correct answers", Icon: "⚡"},
	{Key: "legendary", Name: "Legendary Streak", Description: "50 consecutive correct answers", Icon: "👑"},
	{Key: "perfect_week", Name: "Perfect Week", Description: "7-day correct streak", Icon: "📅"},
	{Key: "perfect_month", Name: "Perfect Month", Description: "30-day correct streak", Icon: "🗓️"},
	{Key: "century", Name: "Century Club", Description: "Reach 100 total points", Icon: "💯"},
	{Key: "high_roller", Name: "High Roller", Description: "Reach 500 total points", Icon: "🎰"},
	{Key: "thousandaire", Name: "Thousandaire", Description: "Reach 1000 total points", Icon: "💰"},
	{Key: "sharp_mind", Name: "Sharp Mind", Description: "Answer a difficulty 3+ question correctly", Icon: "🧠"},
	{Key: "genius", Name: "Genius", Description: "Answer a difficulty 5 question correctly", Icon: "🎓"},
}

// Level definitions
type LevelDef struct {
	Level  int
	Points int
	Title  string
	Reward string
}

var Levels = []LevelDef{
	{Level: 1, Points: 100, Title: "Quiz Apprentice", Reward: "🏅 Bronze Badge Unlocked"},
	{Level: 2, Points: 300, Title: "Knowledge Seeker", Reward: "🎖️ Silver Badge Unlocked"},
	{Level: 3, Points: 600, Title: "Trivia Expert", Reward: "🏆 Gold Badge Unlocked"},
	{Level: 4, Points: 1000, Title: "Quiz Master", Reward: "💎 Diamond Badge Unlocked"},
	{Level: 5, Points: 1500, Title: "Grandmaster", Reward: "👑 Crown Unlocked"},
	{Level: 6, Points: 2100, Title: "Quiz Legend", Reward: "⭐ Star Badge Unlocked"},
	{Level: 7, Points: 2800, Title: "Omniscient", Reward: "🌟 Cosmic Badge Unlocked"},
	{Level: 8, Points: 3600, Title: "Quiz Deity", Reward: "✨ Celestial Badge Unlocked"},
}

func GetUserLevel(totalPoints int) LevelInfo {
	currentLevel := 0
	currentTitle := "Beginner"
	currentReward := ""
	currentThreshold := 0
	nextLevel := 1
	nextThreshold := 100

	for _, l := range Levels {
		if totalPoints >= l.Points {
			currentLevel = l.Level
			currentTitle = l.Title
			currentReward = l.Reward
			currentThreshold = l.Points
		} else {
			nextLevel = l.Level
			nextThreshold = l.Points
			break
		}
	}

	if currentLevel >= len(Levels) {
		nextLevel = currentLevel + 1
		nextThreshold = (currentLevel + 1) * (currentLevel + 2) * 50
	}

	progress := 0.0
	if nextThreshold > currentThreshold {
		progress = float64(totalPoints-currentThreshold) / float64(nextThreshold-currentThreshold) * 100
	}

	if currentLevel == 0 {
		currentReward = "Keep going!"
	}

	return LevelInfo{
		Level:         currentLevel,
		Title:         currentTitle,
		Reward:        currentReward,
		PointsNeeded:  nextThreshold - totalPoints,
		NextLevel:     nextLevel,
		NextLevelPts:  nextThreshold,
		Progress:      progress,
		CurrentPoints: totalPoints,
	}
}

func GetDifficultyForStreak(streak int) int {
	switch {
	case streak <= 5:
		return 1
	case streak <= 15:
		return 2
	case streak <= 30:
		return 3
	case streak <= 50:
		return 4
	default:
		return 5
	}
}

func GetDifficultyLabel(d int) string {
	switch d {
	case 1:
		return "Easy"
	case 2:
		return "Medium"
	case 3:
		return "Hard"
	case 4:
		return "Expert"
	case 5:
		return "Master"
	default:
		return "Unknown"
	}
}
