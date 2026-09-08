package Services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ConversationMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ConversationItem struct {
	ClientID  string                `json:"id"`
	Title     string                `json:"title"`
	Mode      string                `json:"mode"`
	Archived  bool                  `json:"archived"`
	Pinned    bool                  `json:"pinned"`
	CreatedAt int64                 `json:"createdAt"`
	UpdatedAt int64                 `json:"updatedAt"`
	Messages  []ConversationMessage `json:"messages"`
}

func normalizeConversationItem(item ConversationItem) (ConversationItem, error) {
	item.ClientID = strings.TrimSpace(item.ClientID)
	item.Title = strings.TrimSpace(item.Title)
	item.Mode = strings.TrimSpace(item.Mode)
	if item.ClientID == "" {
		return item, fmt.Errorf("missing conversation id")
	}
	if item.Title == "" {
		item.Title = "未命名对话"
	}
	if item.Mode == "" {
		item.Mode = "聊天"
	}
	if item.CreatedAt <= 0 {
		item.CreatedAt = time.Now().UnixMilli()
	}
	if item.UpdatedAt <= 0 {
		item.UpdatedAt = item.CreatedAt
	}
	if item.Messages == nil {
		item.Messages = []ConversationMessage{}
	}
	return item, nil
}

func ListUserConversations(userID int64) ([]ConversationItem, error) {
	if err := appInit(); err != nil {
		return nil, err
	}
	rows, err := ragDB.Query(`
SELECT client_id, title, mode, archived, pinned, created_at, updated_at, messages_json
FROM app_conversations
WHERE user_id=?
ORDER BY pinned DESC, updated_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ConversationItem, 0, 32)
	for rows.Next() {
		var item ConversationItem
		var archived, pinned int
		var messagesJSON string
		if err := rows.Scan(&item.ClientID, &item.Title, &item.Mode, &archived, &pinned, &item.CreatedAt, &item.UpdatedAt, &messagesJSON); err != nil {
			return nil, err
		}
		item.Archived = archived != 0
		item.Pinned = pinned != 0
		if strings.TrimSpace(messagesJSON) == "" {
			messagesJSON = "[]"
		}
		if err := json.Unmarshal([]byte(messagesJSON), &item.Messages); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ReplaceUserConversations(userID int64, items []ConversationItem) error {
	if err := appInit(); err != nil {
		return err
	}

	tx, err := ragDB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM app_conversations WHERE user_id=?`, userID); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
INSERT INTO app_conversations(user_id, client_id, title, mode, archived, pinned, created_at, updated_at, messages_json)
VALUES(?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, raw := range items {
		item, normErr := normalizeConversationItem(raw)
		if normErr != nil {
			err = normErr
			return err
		}
		buf, marshalErr := json.Marshal(item.Messages)
		if marshalErr != nil {
			err = marshalErr
			return err
		}
		if _, err = stmt.Exec(
			userID,
			item.ClientID,
			item.Title,
			item.Mode,
			boolToInt(item.Archived),
			boolToInt(item.Pinned),
			item.CreatedAt,
			item.UpdatedAt,
			string(buf),
		); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func GetUserConversationCount(userID int64) (int, error) {
	if err := appInit(); err != nil {
		return 0, err
	}
	var n int
	err := ragDB.QueryRow(`SELECT COUNT(*) FROM app_conversations WHERE user_id=?`, userID).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}
