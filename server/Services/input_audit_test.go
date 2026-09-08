package Services

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditPreservesInputAndSurvivesConversationDeletion(t *testing.T) {
	t.Setenv("RAG_DB_PATH", filepath.Join(t.TempDir(), "audit.db"))
	previous := ragDB
	ragDB = nil
	t.Cleanup(func() {
		if ragDB != nil {
			ragDB.Close()
		}
		ragDB = previous
	})
	content := "  \n" + strings.Repeat("中文🙂'\"<script>alert(1)</script>\n", 1000) + "\n  "
	if err := RecordUserInput(7, "review", "POST", "/Reply", "text/plain", content, content); err != nil {
		t.Fatal(err)
	}
	messages := make([]ConversationMessage, 450)
	for i := range messages {
		messages[i] = ConversationMessage{Name: "You", Content: content}
	}
	if err := ReplaceUserConversations(7, []ConversationItem{{ClientID: "test", Messages: messages}}); err != nil {
		t.Fatal(err)
	}
	items, err := ListUserConversations(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(items[0].Messages) != 450 || items[0].Messages[0].Content != content {
		t.Fatal("conversation was truncated")
	}
	if err := ReplaceUserConversations(7, nil); err != nil {
		t.Fatal(err)
	}
	var raw, readable string
	if err := ragDB.QueryRow("SELECT raw_body, content FROM app_input_audit WHERE user_id=?", 7).Scan(&raw, &readable); err != nil {
		t.Fatal(err)
	}
	if raw != content || readable != content {
		t.Fatal("audit content was changed")
	}
}
