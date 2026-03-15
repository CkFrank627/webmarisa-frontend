//classic_store.go
package Services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"server/Middlewares/segment"
)

type ClassicQA struct {
	ID      int64
	Q       string
	A       string
	QNorm   string
	QTokens string
	Created int64
	Updated int64
}

var classicDB *sql.DB

func classicInit() error {
	if classicDB != nil {
		return nil
	}

	dbPath := os.Getenv("CLASSIC_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/marisa_classic.db"
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
CREATE TABLE IF NOT EXISTS classic_qa(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  q TEXT NOT NULL,
  a TEXT NOT NULL,
  q_norm TEXT NOT NULL,
  q_tokens TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_classic_qnorm ON classic_qa(q_norm);
CREATE INDEX IF NOT EXISTS idx_classic_updated ON classic_qa(updated_at DESC, id DESC);
`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// 兼容旧表
	if err := classicEnsureColumn(db, `ALTER TABLE classic_qa ADD COLUMN q_tokens TEXT NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	if err := classicEnsureColumn(db, `ALTER TABLE classic_qa ADD COLUMN updated_at INTEGER NOT NULL DEFAULT 0`); err != nil {
		return err
	}

	if _, err := db.Exec(`UPDATE classic_qa SET updated_at = created_at WHERE updated_at = 0`); err != nil {
		return err
	}
	if _, err := db.Exec(`UPDATE classic_qa SET q_tokens = q_norm WHERE q_tokens = ''`); err != nil {
		return err
	}

	classicDB = db
	return nil
}

func classicEnsureColumn(db *sql.DB, ddl string) error {
	_, err := db.Exec(ddl)
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate column") || strings.Contains(msg, "already exists") {
		return nil
	}
	return err
}

func classicNormalize(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	repl := strings.NewReplacer(
		"？", " ", "?", " ",
		"！", " ", "!", " ",
		"。", " ", "，", " ", ",", " ",
		"、", " ", "：", " ", ":", " ",
		"；", " ", ";", " ",
		"（", " ", "）", " ", "(", " ", ")", " ",
		"【", " ", "】", " ", "[", " ", "]", " ",
		"“", " ", "”", " ", `"`, " ",
		"‘", " ", "’", " ", "'", " ",
		"\r", " ", "\n", " ", "\t", " ",
	)
	s = repl.Replace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func classicTokenize(s string) []string {
	qn := classicNormalize(s)
	if qn == "" {
		return nil
	}

	raw := segment.Init().Cut(qn)
	if len(raw) == 0 {
		raw = strings.Fields(qn)
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(raw)+1)

	for _, t := range raw {
		t = classicNormalize(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}

	if len(out) == 0 {
		return []string{qn}
	}
	return out
}

func classicJoinTokens(tokens []string) string {
	return strings.Join(tokens, ",")
}

func classicSplitTokens(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	raw := strings.Split(csv, ",")
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, t := range raw {
		t = classicNormalize(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// 近似沿用旧版风格：关键词命中数 / 已存问题关键词数
func classicScore(queryNorm string, queryTokens []string, item *ClassicQA) float64 {
	if item == nil || queryNorm == "" {
		return 0
	}
	if item.QNorm == queryNorm {
		return 1
	}

	itemTokens := classicSplitTokens(item.QTokens)
	if len(itemTokens) == 0 {
		itemTokens = classicTokenize(item.Q)
	}
	if len(itemTokens) == 0 || len(queryTokens) == 0 {
		return 0
	}

	set := make(map[string]struct{}, len(queryTokens))
	for _, t := range queryTokens {
		set[t] = struct{}{}
	}

	hits := 0
	for _, t := range itemTokens {
		if _, ok := set[t]; ok {
			hits++
		}
	}

	score := float64(hits) / float64(len(itemTokens))

	// 子串关系给一点保底提升，避免“你是谁”和“你到底是谁呀”完全对不上
	if strings.Contains(queryNorm, item.QNorm) || strings.Contains(item.QNorm, queryNorm) {
		if score < 0.6 {
			score = 0.6
		}
	}
	return score
}

func ClassicCount() (int, error) {
	if err := classicInit(); err != nil {
		return 0, err
	}
	var n int
	if err := classicDB.QueryRow(`SELECT COUNT(*) FROM classic_qa`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func ClassicAddQA(q, a string) (int64, error) {
	if err := classicInit(); err != nil {
		return 0, err
	}

	q = strings.TrimSpace(q)
	a = strings.TrimSpace(a)
	if q == "" || a == "" {
		return 0, fmt.Errorf("empty q/a")
	}

	qn := classicNormalize(q)
	if qn == "" {
		return 0, fmt.Errorf("empty normalized q")
	}

	qt := classicJoinTokens(classicTokenize(q))
	now := time.Now().Unix()

	// 同一个问题再次 teach：直接覆盖答案，而不是重复插一堆
	hit, err := ClassicFindExact(q)
	if err != nil {
		return 0, err
	}
	if hit != nil {
		_, err := classicDB.Exec(
			`UPDATE classic_qa SET q=?, a=?, q_norm=?, q_tokens=?, updated_at=? WHERE id=?`,
			q, a, qn, qt, now, hit.ID,
		)
		if err != nil {
			return 0, err
		}
		return hit.ID, nil
	}

	res, err := classicDB.Exec(
		`INSERT INTO classic_qa(q, a, q_norm, q_tokens, created_at, updated_at) VALUES(?,?,?,?,?,?)`,
		q, a, qn, qt, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ClassicDeleteLast() (bool, error) {
	if err := classicInit(); err != nil {
		return false, err
	}
	res, err := classicDB.Exec(`
DELETE FROM classic_qa
WHERE id = (SELECT id FROM classic_qa ORDER BY updated_at DESC, id DESC LIMIT 1)
`)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func ClassicDeleteByID(id int64) (bool, error) {
	if err := classicInit(); err != nil {
		return false, err
	}
	res, err := classicDB.Exec(`DELETE FROM classic_qa WHERE id=?`, id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func ClassicDeleteByQ(q string) (bool, error) {
	if err := classicInit(); err != nil {
		return false, err
	}
	qn := classicNormalize(q)
	if qn == "" {
		return false, nil
	}
	res, err := classicDB.Exec(`DELETE FROM classic_qa WHERE q_norm=?`, qn)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func ClassicFindExact(q string) (*ClassicQA, error) {
	if err := classicInit(); err != nil {
		return nil, err
	}
	qn := classicNormalize(q)
	if qn == "" {
		return nil, nil
	}

	row := classicDB.QueryRow(
		`SELECT id, q, a, q_norm, q_tokens, created_at, updated_at
         FROM classic_qa
         WHERE q_norm=?
         ORDER BY updated_at DESC, id DESC
         LIMIT 1`,
		qn,
	)

	var item ClassicQA
	if err := row.Scan(&item.ID, &item.Q, &item.A, &item.QNorm, &item.QTokens, &item.Created, &item.Updated); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func ClassicFindBest(q string) (*ClassicQA, error) {
	if err := classicInit(); err != nil {
		return nil, err
	}

	qn := classicNormalize(q)
	if qn == "" {
		return nil, nil
	}
	qTokens := classicTokenize(q)

	exact, err := ClassicFindExact(q)
	if err != nil || exact != nil {
		return exact, err
	}

	rows, err := classicDB.Query(`
SELECT id, q, a, q_norm, q_tokens, created_at, updated_at
FROM classic_qa
ORDER BY updated_at DESC, id DESC
LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var best *ClassicQA
	bestScore := 0.0

	for rows.Next() {
		var item ClassicQA
		if err := rows.Scan(&item.ID, &item.Q, &item.A, &item.QNorm, &item.QTokens, &item.Created, &item.Updated); err != nil {
			return nil, err
		}

		score := classicScore(qn, qTokens, &item)
		if score > bestScore {
			cp := item
			best = &cp
			bestScore = score
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 旧版 reply 阈值差不多就是 0.4 的味道
	if best != nil && bestScore >= 0.4 {
		return best, nil
	}
	return nil, nil
}

func ClassicReplyText(q string) (string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "……你刚刚说什么？再说一遍啦。", nil
	}

	hit, err := ClassicFindBest(q)
	if err != nil {
		return "", err
	}
	if hit != nil {
		// 关键：经典版命中后直接回 teach 的原答案，不要再套随机前后缀
		return hit.A, nil
	}

	return "唔嗯……这个我还不会。输入 teach，然后下一句当“问”，再下一句当“答”来教我吧。", nil
}

func ClassicList(limit int) ([]map[string]interface{}, error) {
	if err := classicInit(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := classicDB.Query(`
SELECT id, q, a, created_at
FROM classic_qa
ORDER BY updated_at DESC, id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]interface{}, 0, limit)
	for rows.Next() {
		var id int64
		var q, a string
		var created int64
		if err := rows.Scan(&id, &q, &a, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":         id,
			"q":          q,
			"a":          a,
			"created_at": created,
		})
	}
	return out, rows.Err()
}