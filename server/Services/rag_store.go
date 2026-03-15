package Services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type ragDoc struct {
	ID      int64
	Content string
	Created int64
}

var ragDB *sql.DB

func ragInit() error {
	if ragDB != nil {
		return nil
	}
	dbPath := os.Getenv("RAG_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/marisa_rag.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	// sqlite 简化并发
	db.SetMaxOpenConns(1)

	schema := `
CREATE TABLE IF NOT EXISTS rag_docs(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_rag_docs_created ON rag_docs(created_at DESC);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	ragDB = db
	return nil
}

func ragAdd(content string) (int64, error) {
	if err := ragInit(); err != nil {
		return 0, err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return 0, fmt.Errorf("empty content")
	}
	res, err := ragDB.Exec(`INSERT INTO rag_docs(content, created_at) VALUES (?, ?)`, content, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Forget 既支持 “id” 也支持 “按全文精确删除”
func ragForget(answer string) (bool, error) {
	if err := ragInit(); err != nil {
		return false, err
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return false, nil
	}
	// 如果是数字，当作 id
	var id int64
	if _, err := fmt.Sscanf(answer, "%d", &id); err == nil && id > 0 {
		res, err := ragDB.Exec(`DELETE FROM rag_docs WHERE id=?`, id)
		if err != nil {
			return false, err
		}
		n, _ := res.RowsAffected()
		return n > 0, nil
	}

	// 否则按内容删
	res, err := ragDB.Exec(`DELETE FROM rag_docs WHERE content=?`, answer)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func ragCount() (int, error) {
	if err := ragInit(); err != nil {
		return 0, err
	}
	var n int
	if err := ragDB.QueryRow(`SELECT COUNT(*) FROM rag_docs`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func ragSearch(query string, topK int) ([]ragDoc, error) {
	if err := ragInit(); err != nil {
		return nil, err
	}
	orig := strings.TrimSpace(query)
	if orig == "" {
		return []ragDoc{}, nil
	}
	if topK <= 0 {
		topK = 5
	}

	// 关键：同时用“原始查询”和“去标点归一化查询”去匹配，提高命中率
q := normalizeQuery(orig)
if q == "" { q = orig }
canon := extractKeyPhrase(orig) // 可能是 "你是谁" / "我是谁" / ""

rows, err := ragDB.Query(`
SELECT id, content, created_at
FROM rag_docs
WHERE content LIKE '%'||?||'%' 
   OR content LIKE '%'||?||'%'
   OR content LIKE '%'||?||'%'
ORDER BY created_at DESC
LIMIT ?`, q, orig, canon, topK)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ragDoc, 0, topK)
	for rows.Next() {
		var d ragDoc
		if err := rows.Scan(&d.ID, &d.Content, &d.Created); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func normalizeQuery(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "？", "")
	s = strings.ReplaceAll(s, "?", "")
	s = strings.ReplaceAll(s, "！", "")
	s = strings.ReplaceAll(s, "!", "")
	s = strings.ReplaceAll(s, "。", "")
	s = strings.ReplaceAll(s, "，", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "、", "")
	s = strings.ReplaceAll(s, "：", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "；", "")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func extractKeyPhrase(orig string) string {
    s := normalizeQuery(orig)

    // 特判：身份类问法——只要包含关键短语，就用关键短语去检索
    if strings.Contains(s, "你是谁") {
        return "你是谁"
    }
    if strings.Contains(s, "我是谁") {
        return "我是谁"
    }
    return ""
}
