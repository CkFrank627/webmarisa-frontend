package Services

import "time"

// Append-only audit records are independent of editable/deletable conversations.
// Keep original request bytes and readable content; never truncate or trim inputs.
func RecordUserInput(userID int64, username, method, path, contentType, rawBody, content string) error {
	if err := appInit(); err != nil {
		return err
	}
	_, err := ragDB.Exec(`INSERT INTO app_input_audit
		(user_id, username, method, path, content_type, raw_body, content, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, username, method, path, contentType, rawBody, content, time.Now().UnixMilli())
	return err
}
