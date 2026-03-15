package Services

import (
	"fmt"
	"strings"
	"time"
)

// 保存一条对话日志（user/assistant 都走这个）
func AddPersonaChatLog(userID, personaID int64, role, content string) error {
	if err := appInit(); err != nil {
		return err
	}
	role = strings.TrimSpace(role)
	content = strings.TrimSpace(content)
	if userID <= 0 || personaID <= 0 || role == "" || content == "" {
		return fmt.Errorf("bad chatlog params")
	}
	if role != "user" && role != "assistant" {
		return fmt.Errorf("bad role")
	}

	_, err := ragDB.Exec(
		`INSERT INTO app_persona_chatlog(user_id, persona_id, role, content, created_at) VALUES(?,?,?,?,?)`,
		userID, personaID, role, content, time.Now().Unix(),
	)
	if err != nil {
		return err
	}

	// ✅ 自动裁剪：每个 persona 最多保留最近 400 条（约 200 轮对话）
	// 你想更大就改 400
	_ = TrimPersonaChatLog(userID, personaID, 400)
	return nil
}

// 只保留最近 keep 条（按 id 倒序裁剪）
func TrimPersonaChatLog(userID, personaID int64, keep int) error {
	if err := appInit(); err != nil {
		return err
	}
	if keep <= 0 {
		keep = 200
	}

	// 删除不在“最新 keep 条”集合里的记录
	_, err := ragDB.Exec(`
DELETE FROM app_persona_chatlog
WHERE user_id=? AND persona_id=? AND id NOT IN (
  SELECT id FROM app_persona_chatlog
  WHERE user_id=? AND persona_id=?
  ORDER BY id DESC
  LIMIT ?
)`, userID, personaID, userID, personaID, keep)
	return err
}

// 列出某个人格的日志
func ListPersonaChatLog(userID, personaID int64, limit int) ([]map[string]interface{}, error) {
	if err := appInit(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	// 权限：必须确认 persona 属于该用户
	var owner int64
	if err := ragDB.QueryRow(`SELECT user_id FROM app_personas WHERE id=?`, personaID).Scan(&owner); err != nil {
		return nil, err
	}
	if owner != userID {
		return nil, fmt.Errorf("forbidden")
	}

	rows, err := ragDB.Query(`
SELECT id, persona_id, role, content, created_at
FROM app_persona_chatlog
WHERE user_id=? AND persona_id=?
ORDER BY id DESC
LIMIT ?`, userID, personaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]interface{}, 0, limit)
	for rows.Next() {
		var id, pid, created int64
		var role, content string
		if err := rows.Scan(&id, &pid, &role, &content, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":         id,
			"persona_id": pid,
			"role":       role,
			"content":    content,
			"created_at": created,
		})
	}
	return out, nil
}

// 列出该用户全部 persona 的日志（混合）
func ListUserChatLog(userID int64, limit int) ([]map[string]interface{}, error) {
	if err := appInit(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	rows, err := ragDB.Query(`
SELECT id, persona_id, role, content, created_at
FROM app_persona_chatlog
WHERE user_id=?
ORDER BY id DESC
LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]interface{}, 0, limit)
	for rows.Next() {
		var id, pid, created int64
		var role, content string
		if err := rows.Scan(&id, &pid, &role, &content, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":         id,
			"persona_id": pid,
			"role":       role,
			"content":    content,
			"created_at": created,
		})
	}
	return out, nil
}

// 清空某 persona 的日志
func ClearPersonaChatLog(userID, personaID int64) error {
	if err := appInit(); err != nil {
		return err
	}

	var owner int64
	if err := ragDB.QueryRow(`SELECT user_id FROM app_personas WHERE id=?`, personaID).Scan(&owner); err != nil {
		return err
	}
	if owner != userID {
		return fmt.Errorf("forbidden")
	}

	_, err := ragDB.Exec(`DELETE FROM app_persona_chatlog WHERE user_id=? AND persona_id=?`, userID, personaID)
	return err
}

// 清空该用户全部日志（可选）
func ClearUserChatLog(userID int64) error {
	if err := appInit(); err != nil {
		return err
	}
	_, err := ragDB.Exec(`DELETE FROM app_persona_chatlog WHERE user_id=?`, userID)
	return err
}
