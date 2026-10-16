package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			display_name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			total_points INTEGER DEFAULT 0,
			current_streak INTEGER DEFAULT 0,
			longest_streak INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			option_a TEXT NOT NULL,
			option_b TEXT NOT NULL,
			option_c TEXT NOT NULL,
			option_d TEXT NOT NULL,
			correct_option TEXT NOT NULL,
			difficulty INTEGER NOT NULL,
			category TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS daily_assignments (
			user_id INTEGER NOT NULL,
			quiz_date TEXT NOT NULL,
			question_id INTEGER NOT NULL,
			PRIMARY KEY (user_id, quiz_date),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (question_id) REFERENCES questions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_answers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			question_id INTEGER NOT NULL,
			answer_date TEXT NOT NULL,
			selected_option TEXT NOT NULL,
			is_correct BOOLEAN NOT NULL,
			points_earned INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, answer_date),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (question_id) REFERENCES questions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_badges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			badge_key TEXT NOT NULL,
			earned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, badge_key),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_user_answers_user_date ON user_answers(user_id, answer_date)`,
		`CREATE INDEX IF NOT EXISTS idx_questions_difficulty ON questions(difficulty)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_users_total_points ON users(total_points DESC)`,
	}

	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %s: %w", m[:60], err)
		}
	}

	return nil
}

type JSONQuestion struct {
	Text     string   `json:"text"`
	Options  []string `json:"options"`
	Correct  int      `json:"correct"`
	Diff     int      `json:"difficulty"`
	Category string   `json:"category"`
}

func SeedQuestions(filePath string) error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM questions").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Printf("Questions already seeded (%d questions found)", count)
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read questions file: %w", err)
	}

	var questions []JSONQuestion
	if err := json.Unmarshal(data, &questions); err != nil {
		return fmt.Errorf("failed to parse questions: %w", err)
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO questions (text, option_a, option_b, option_c, option_d, correct_option, difficulty, category)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	optionLetters := []string{"A", "B", "C", "D"}

	for _, q := range questions {
		if len(q.Options) != 4 {
			log.Printf("Skipping question with %d options: %s", len(q.Options), q.Text[:50])
			continue
		}
		correctLetter := optionLetters[q.Correct]
		_, err := stmt.Exec(q.Text, q.Options[0], q.Options[1], q.Options[2], q.Options[3], correctLetter, q.Diff, q.Category)
		if err != nil {
			return fmt.Errorf("failed to insert question: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Seeded %d questions successfully", len(questions))
	return nil
}
