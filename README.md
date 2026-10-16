# 🎯 QuizMania

**One question. Every day. No excuses.**

QuizMania is a daily trivia quiz app. You get **one question per day** — answer it correctly to earn points and build a streak. The longer your streak, the more points you earn!

---

## How It Works

1. **Sign up** with an email and password
2. **Answer one trivia question** each day (resets at midnight UTC)
3. **Build streaks** — Day 1 correct = 1 pt, Day 2 = 2 pts, Day 3 = 3 pts, and so on
4. **Level up** as you hit point milestones (100 → 300 → 600 → 1000…)
5. **Earn badges** like "On Fire 🔥" (10-day streak) or "Century Club 💯" (100 points)
6. **Climb the leaderboard** — top 10 players are displayed for everyone to see

Miss a day or answer wrong? Your streak resets to zero. Brutal, but fair 😈

### Difficulty Scaling

As your streak grows, questions get harder:

| Streak | Difficulty |
|--------|-----------|
| 1–5 | Easy |
| 6–15 | Medium |
| 16–30 | Hard |
| 31–50 | Expert |
| 51+ | Master |

---

## Tech Stack

- **Backend:** Go (standard library — no frameworks)
- **Database:** SQLite (zero setup, single file)
- **Frontend:** Go HTML templates + Tailwind CSS
- **Auth:** Email + password with bcrypt hashing

---

## Run It Yourself

### What You Need

- **Go 1.22 or newer** — [Download here](https://go.dev/dl/)

That's literally it. No Docker, no database server, no Node.js.

### Steps

```bash
# 1. Clone the repo
git clone https://github.com/YOUR_USERNAME/daily-quiz.git
cd daily-quiz

# 2. Download dependencies
go mod download

# 3. Run the app
go run main.go
```

Open **http://localhost:8080** in your browser. Done! 🎉

The database (`quizmania.db`) is created automatically the first time you run the app. Questions are seeded from `data/questions.json`.

### Optional: Build a Binary

```bash
go build -o quizmania .
./quizmania
```

### Environment Variables (all optional)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Port to run the server on |
| `DB_PATH` | `quizmania.db` | Path to the SQLite database file |

---

## Project Structure

```
daily-quiz/
├── main.go                  # Entry point & routes
├── data/
│   └── questions.json       # 200 trivia questions (add more here!)
├── internal/
│   ├── db/db.go             # Database setup & migrations
│   ├── models/models.go     # Data types, levels, badges
│   ├── handlers/            # HTTP handlers (quiz, auth, etc.)
│   └── middleware/           # Auth & session management
└── templates/               # HTML page templates
```

---

## Want to Add More Questions?

Edit `data/questions.json`. Each question looks like this:

```json
{
  "text": "What is the capital of France?",
  "options": ["Paris", "London", "Berlin", "Madrid"],
  "correct": 0,
  "difficulty": 1,
  "category": "Geography"
}
```

- `correct` is the index of the right answer (0 = first option, 1 = second, etc.)
- `difficulty` goes from 1 (easy) to 5 (master)
- Delete `quizmania.db` and restart the app to re-seed

---

Built with ☕ and Go.
