//app_store.go

package Services

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// 复用 ragDB（来自 rag_store.go），确保 appInit 会先 ragInit
func appInit() error {
	if err := ragInit(); err != nil {
		return err
	}

	schema := `
CREATE TABLE IF NOT EXISTS app_stats(
  key TEXT PRIMARY KEY,
  val INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS app_users(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  pass_hash TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS app_messages(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  username TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS app_persona_chatlog(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  persona_id INTEGER NOT NULL,
  role TEXT NOT NULL,         -- "user" | "assistant"
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_pchat_user_time ON app_persona_chatlog(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_pchat_persona_time ON app_persona_chatlog(persona_id, created_at DESC);


CREATE TABLE IF NOT EXISTS app_personas(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  prompt TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_personas_user ON app_personas(user_id);

CREATE TABLE IF NOT EXISTS app_persona_teach(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  persona_id INTEGER NOT NULL,
  q TEXT NOT NULL,
  a TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_teach_persona ON app_persona_teach(persona_id);


INSERT OR IGNORE INTO app_stats(key, val) VALUES ('total_requests', 0);
INSERT OR IGNORE INTO app_stats(key, val) VALUES ('total_replies', 0);
INSERT OR IGNORE INTO app_stats(key, val) VALUES ('total_sessions', 0);
`
	_, err := ragDB.Exec(schema)
	return err
}

func incStat(key string, delta int64) error {
	if err := appInit(); err != nil {
		return err
	}
	_, err := ragDB.Exec(`UPDATE app_stats SET val = val + ? WHERE key = ?`, delta, key)
	return err
}

func getStat(key string) (int64, error) {
	if err := appInit(); err != nil {
		return 0, err
	}
	var v int64
	err := ragDB.QueryRow(`SELECT val FROM app_stats WHERE key=?`, key).Scan(&v)
	return v, err
}

func GetStats() (map[string]int64, error) {
	if err := appInit(); err != nil {
		return nil, err
	}
	totalReq, _ := getStat("total_requests")
	totalRep, _ := getStat("total_replies")
	totalSes, _ := getStat("total_sessions")
	return map[string]int64{
		"total_requests": totalReq,
		"total_replies":  totalRep,
		"total_sessions": totalSes,
	}, nil
}

func IncRequest() { _ = incStat("total_requests", 1) }
func IncReply()   { _ = incStat("total_replies", 1) }

// session 去重（内存级，重启归零——够用）
var (
	seenMu       sync.Mutex
	seenSessions = map[string]bool{}
)

func IncSessionOnce(sid string) {
	if sid == "" {
		return
	}
	seenMu.Lock()
	defer seenMu.Unlock()
	if seenSessions[sid] {
		return
	}
	seenSessions[sid] = true
	_ = incStat("total_sessions", 1)
}

// ===== Users =====
func CreateUser(username, passHash string) error {
	if err := appInit(); err != nil {
		return err
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("empty username")
	}
	_, err := ragDB.Exec(
		`INSERT INTO app_users(username, pass_hash, created_at) VALUES(?,?,?)`,
		username, passHash, time.Now().Unix(),
	)
	return err
}

func GetUserByName(username string) (id int64, passHash string, err error) {
	if err = appInit(); err != nil {
		return 0, "", err
	}
	row := ragDB.QueryRow(`SELECT id, pass_hash FROM app_users WHERE username=?`, strings.TrimSpace(username))
	err = row.Scan(&id, &passHash)
	return
}

// ===== Messages =====
func AddMessage(userID int64, username, content string) error {
	if err := appInit(); err != nil {
		return err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("empty content")
	}
	_, err := ragDB.Exec(
		`INSERT INTO app_messages(user_id, username, content, created_at) VALUES(?,?,?,?)`,
		userID, username, content, time.Now().Unix(),
	)
	return err
}

func ListMessages(limit int) ([]map[string]interface{}, error) {
	if err := appInit(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := ragDB.Query(
		`SELECT id, username, content, created_at FROM app_messages ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]interface{}, 0, limit)
	for rows.Next() {
		var id int64
		var username, content string
		var created int64
		if err := rows.Scan(&id, &username, &content, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":         id,
			"username":   username,
			"content":    content,
			"created_at": created,
		})
	}
	return out, nil
}
