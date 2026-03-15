// Services/affinity_store.go
package Services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var affinityDB *sql.DB

func affinityInit() error {
	if affinityDB != nil {
		return nil
	}
	dbPath := os.Getenv("AFFINITY_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/marisa_affinity.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)

	schema := `
CREATE TABLE IF NOT EXISTS user_affinity(
  user_id INTEGER PRIMARY KEY,
  score INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS affinity_log(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  delta INTEGER NOT NULL,
  reason TEXT NOT NULL,
  user_text TEXT,
  marisa_text TEXT,
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_aff_log_user ON affinity_log(user_id, created_at DESC);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	affinityDB = db
	return nil
}

func clampAffinity(x int) int {
	if x < 0 {
		return 0
	}
	if x > 100 {
		return 100
	}
	return x
}

func ensureAffinityRow(userID int64) error {
	if err := affinityInit(); err != nil {
		return err
	}
	// insert-if-not-exists
	_, err := affinityDB.Exec(`
INSERT INTO user_affinity(user_id, score, updated_at)
SELECT ?, 0, ?
WHERE NOT EXISTS(SELECT 1 FROM user_affinity WHERE user_id=?)
`, userID, time.Now().Unix(), userID)
	return err
}

func GetAffinity(userID int64) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user_id")
	}
	if err := ensureAffinityRow(userID); err != nil {
		return 0, err
	}

	var score int
	err := affinityDB.QueryRow(`SELECT score FROM user_affinity WHERE user_id=?`, userID).Scan(&score)
	if err != nil {
		return 0, err
	}
	return score, nil
}

func ApplyAffinityDelta(userID int64, delta int, reason, userText, marisaText string) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user_id")
	}
	if err := ensureAffinityRow(userID); err != nil {
		return 0, err
	}

	// read current
	cur, err := GetAffinity(userID)
	if err != nil {
		return 0, err
	}
	next := clampAffinity(cur + delta)

	now := time.Now().Unix()
	tx, err := affinityDB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`UPDATE user_affinity SET score=?, updated_at=? WHERE user_id=?`, next, now, userID)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`INSERT INTO affinity_log(user_id, delta, reason, user_text, marisa_text, created_at) VALUES(?,?,?,?,?,?)`,
		userID, delta, reason, userText, marisaText, now,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return next, nil
}

type AffinityLogItem struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Delta     int    `json:"delta"`
	Reason    string `json:"reason"`
	UserText  string `json:"user_text"`
	MarisaText string `json:"marisa_text"`
	CreatedAt int64  `json:"created_at"`
}

func ListAffinityLogs(userID int64, limit int) ([]AffinityLogItem, error) {
	if err := affinityInit(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var rows *sql.Rows
	var err error
	if userID > 0 {
		rows, err = affinityDB.Query(`
SELECT id, user_id, delta, reason, user_text, marisa_text, created_at
FROM affinity_log
WHERE user_id=?
ORDER BY id DESC
LIMIT ?`, userID, limit)
	} else {
		rows, err = affinityDB.Query(`
SELECT id, user_id, delta, reason, user_text, marisa_text, created_at
FROM affinity_log
ORDER BY id DESC
LIMIT ?`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AffinityLogItem, 0, limit)
	for rows.Next() {
		var it AffinityLogItem
		if err := rows.Scan(&it.ID, &it.UserID, &it.Delta, &it.Reason, &it.UserText, &it.MarisaText, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func AffinityLevel(score int) string {
	switch {
	case score >= 100:
		return "max"
	case score >= 70:
		return "close"
	case score >= 40:
		return "friendly"
	default:
		return "normal"
	}
}

func CanIntimate(score int) bool {
	return score >= 100
}
